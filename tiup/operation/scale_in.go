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

	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
)

// ScaleIn scales in the cluster
func ScaleIn(
	ctx context.Context,
	cluster *spec.Specification,
	options Options,
	tlsCfg *tls.Config,
) error {
	return ScaleInCluster(ctx, cluster, options, tlsCfg)
}

// ScaleInCluster scales in the cluster
func ScaleInCluster(
	ctx context.Context,
	cluster *spec.Specification,
	options Options,
	tlsCfg *tls.Config,
) error {
	// instances by uuid
	instances := map[string]spec.Instance{}
	instCount := map[string]int{}

	// make sure all nodeIds exists in topology
	for _, component := range cluster.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			instances[instance.ID()] = instance
			instCount[instance.GetHost()]++
		}
	}

	// Clean components
	deletedDiff := map[string][]spec.Instance{}
	deletedNodes := set.NewStringSet(options.Nodes...)
	for nodeID := range deletedNodes {
		inst, found := instances[nodeID]
		if !found {
			return errors.Errorf("cannot find node id '%s' in topology", nodeID)
		}
		deletedDiff[inst.ComponentName()] = append(deletedDiff[inst.ComponentName()], inst)
	}

	// Cannot delete all MetaDB servers
	if len(deletedDiff[spec.ComponentTiEMClusterServer]) == len(cluster.ClusterServers) {
		return errors.New("cannot delete all Cluster servers")
	}

	// Cannot delete all API servers
	if len(deletedDiff[spec.ComponentTiEMAPIServer]) == len(cluster.APIServers) {
		return errors.New("cannot delete all API servers")
	}

	if options.Force {
		for _, component := range cluster.ComponentsByStartOrder() {
			for _, instance := range component.Instances() {
				if !deletedNodes.Exist(instance.ID()) {
					continue
				}
				compName := component.Name()

				instCount[instance.GetHost()]--
				if err := StopAndDestroyInstance(ctx, cluster, instance, options, instCount[instance.GetHost()] == 0); err != nil {
					log.Warnf("failed to stop/destroy %s: %v", compName, err)
				}
			}
		}
		return nil
	}

	// Delete member from cluster
	for _, component := range cluster.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			if !deletedNodes.Exist(instance.ID()) {
				continue
			}

			instCount[instance.GetHost()]--
			if err := StopAndDestroyInstance(ctx, cluster, instance, options, instCount[instance.GetHost()] == 0); err != nil {
				return err
			}
		}
	}

	return nil
}
