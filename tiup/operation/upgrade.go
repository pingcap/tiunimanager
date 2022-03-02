// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package operator

import (
	"context"
	"crypto/tls"
	"reflect"
	"strconv"

	"github.com/pingcap-inc/tiem/tiup/spec"
	perrs "github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/checkpoint"
	"github.com/pingcap/tiup/pkg/cluster/api"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
	"go.uber.org/zap"
)

var (
	// register checkpoint for upgrade operation
	upgradePoint       = checkpoint.Register(checkpoint.Field("instance", reflect.DeepEqual))
	increaseLimitPoint = checkpoint.Register()
)

// Upgrade the cluster.
func Upgrade(
	ctx context.Context,
	topo spec.Topology,
	options Options,
	tlsCfg *tls.Config,
) error {
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := topo.ComponentsByUpdateOrder()
	components = FilterComponent(components, roleFilter)

	for _, component := range components {
		instances := FilterInstance(component.Instances(), nodeFilter)
		if len(instances) < 1 {
			continue
		}

		log.Infof("Upgrading component %s", component.Name())

		for _, instance := range instances {
			if err := upgradeInstance(ctx, topo, instance, options, tlsCfg); err != nil {
				return err
			}
		}
	}

	return nil
}

func upgradeInstance(ctx context.Context, topo spec.Topology, instance spec.Instance, options Options, tlsCfg *tls.Config) (err error) {
	// insert checkpoint
	point := checkpoint.Acquire(ctx, upgradePoint, map[string]interface{}{"instance": instance.ID()})
	defer func() {
		point.Release(err, zap.String("instance", instance.ID()))
	}()

	if point.Hit() != nil {
		return nil
	}

	var rollingInstance spec.RollingUpdateInstance
	var isRollingInstance bool

	if !options.Force {
		rollingInstance, isRollingInstance = instance.(spec.RollingUpdateInstance)
	}

	if isRollingInstance {
		err := rollingInstance.PreRestart(topo, int(options.APITimeout), tlsCfg)
		if err != nil && !options.Force {
			return err
		}
	}

	if err := restartInstance(ctx, instance, options.OptTimeout); err != nil && !options.Force {
		return err
	}

	if isRollingInstance {
		err := rollingInstance.PostRestart(topo, tlsCfg)
		if err != nil && !options.Force {
			return err
		}
	}

	return nil
}

// Addr returns the address of the instance.
func Addr(ins spec.Instance) string {
	if ins.GetPort() == 0 || ins.GetPort() == 80 {
		panic(ins)
	}
	return ins.GetHost() + ":" + strconv.Itoa(ins.GetPort())
}

var (
	leaderScheduleLimitOffset = 32
	regionScheduleLimitOffset = 512
	// storeLimitOffset             = 512
	leaderScheduleLimitThreshold = 64
	regionScheduleLimitThreshold = 1024
	// storeLimitThreshold          = 1024
)

// increaseScheduleLimit increases the schedule limit of leader and region for faster
// rebalancing during the rolling restart / upgrade process
func increaseScheduleLimit(ctx context.Context, pc *api.PDClient) (
	currLeaderScheduleLimit int,
	currRegionScheduleLimit int,
	err error) {
	// insert checkpoint
	point := checkpoint.Acquire(ctx, increaseLimitPoint, map[string]interface{}{})
	defer func() {
		point.Release(err,
			zap.Int("currLeaderScheduleLimit", currLeaderScheduleLimit),
			zap.Int("currRegionScheduleLimit", currRegionScheduleLimit),
		)
	}()

	if data := point.Hit(); data != nil {
		currLeaderScheduleLimit = int(data["currLeaderScheduleLimit"].(float64))
		currRegionScheduleLimit = int(data["currRegionScheduleLimit"].(float64))
		return
	}

	// query current values
	cfg, err := pc.GetConfig()
	if err != nil {
		return
	}
	val, ok := cfg["schedule.leader-schedule-limit"].(float64)
	if !ok {
		return currLeaderScheduleLimit, currRegionScheduleLimit, perrs.New("cannot get current leader-schedule-limit")
	}
	currLeaderScheduleLimit = int(val)
	val, ok = cfg["schedule.region-schedule-limit"].(float64)
	if !ok {
		return currLeaderScheduleLimit, currRegionScheduleLimit, perrs.New("cannot get current region-schedule-limit")
	}
	currRegionScheduleLimit = int(val)

	// increase values
	if currLeaderScheduleLimit < leaderScheduleLimitThreshold {
		newLimit := currLeaderScheduleLimit + leaderScheduleLimitOffset
		if newLimit > leaderScheduleLimitThreshold {
			newLimit = leaderScheduleLimitThreshold
		}
		if err := pc.SetReplicationConfig("leader-schedule-limit", newLimit); err != nil {
			return currLeaderScheduleLimit, currRegionScheduleLimit, err
		}
	}
	if currRegionScheduleLimit < regionScheduleLimitThreshold {
		newLimit := currRegionScheduleLimit + regionScheduleLimitOffset
		if newLimit > regionScheduleLimitThreshold {
			newLimit = regionScheduleLimitThreshold
		}
		if err := pc.SetReplicationConfig("region-schedule-limit", newLimit); err != nil {
			// try to revert leader scheduler limit by our best effort, does not make sense
			// to handle this error again
			_ = pc.SetReplicationConfig("leader-schedule-limit", currLeaderScheduleLimit)
			return currLeaderScheduleLimit, currRegionScheduleLimit, err
		}
	}

	return
}

// decreaseScheduleLimit tries to set the schedule limit back to it's original with
// the same offset value as increaseScheduleLimit added, with some sanity checks
func decreaseScheduleLimit(pc *api.PDClient, origLeaderScheduleLimit, origRegionScheduleLimit int) error {
	if err := pc.SetReplicationConfig("leader-schedule-limit", origLeaderScheduleLimit); err != nil {
		return err
	}
	return pc.SetReplicationConfig("region-schedule-limit", origRegionScheduleLimit)
}