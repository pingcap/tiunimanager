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

package main

import (
	"context"
	"crypto/tls"

	operator "github.com/pingcap-inc/tiem/tiup/operation"
	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap-inc/tiem/tiup/task"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
	"github.com/spf13/cobra"
)

func newScaleInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale-in <cluster-name>",
		Short: "Scale in a EM cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}

			clusterName := args[0]

			scale := func(b *task.Builder, imetadata spec.Metadata, tlsCfg *tls.Config) {
				metadata := imetadata.(*spec.EMMeta)
				nodes := gOpt.Nodes

				b.ClusterOperate(metadata.Topology, operator.ScaleInOperation, gOpt, tlsCfg).
					UpdateMeta(clusterName, metadata, nodes)
			}

			return cm.ScaleIn(clusterName, skipConfirm, gOpt, scale)
		},
	}

	cmd.Flags().StringSliceVarP(&gOpt.Nodes, "node", "N", nil, "Specify the nodes (required)")
	cmd.Flags().BoolVar(&gOpt.Force, "force", false, "Force just try stop and destroy instance before removing the instance from topo")

	_ = cmd.MarkFlagRequired("node")

	return cmd
}

// ScaleInDMCluster scale in dm cluster.
func ScaleInDMCluster(
	ctx context.Context,
	topo *spec.Specification,
	options operator.Options,
) error {
	// instances by uuid
	instances := map[string]spec.Instance{}
	instCount := map[string]int{}

	// make sure all nodeIds exists in topology
	for _, component := range topo.ComponentsByStartOrder() {
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

	if options.Force {
		for _, component := range topo.ComponentsByStartOrder() {
			for _, instance := range component.Instances() {
				if !deletedNodes.Exist(instance.ID()) {
					continue
				}
				instCount[instance.GetHost()]--
				if err := operator.StopAndDestroyInstance(ctx, topo, instance, options, instCount[instance.GetHost()] == 0); err != nil {
					log.Warnf("failed to stop/destroy %s: %v", component.Name(), err)
				}
			}
		}
		return nil
	}

	noAgentHosts := set.NewStringSet()
	topo.IterInstance(func(inst spec.Instance) {
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		}
	})

	// Delete member from cluster
	for _, component := range topo.ComponentsByStartOrder() {
		for _, instance := range component.Instances() {
			if !deletedNodes.Exist(instance.ID()) {
				continue
			}

			if err := operator.StopComponent(ctx, []spec.Instance{instance}, noAgentHosts, options.OptTimeout); err != nil {
				return errors.Annotatef(err, "failed to stop %s", component.Name())
			}

			if err := operator.DestroyComponent(ctx, []spec.Instance{instance}, topo, options); err != nil {
				return errors.Annotatef(err, "failed to destroy %s", component.Name())
			}

			instCount[instance.GetHost()]--
			if instCount[instance.GetHost()] == 0 {
				if err := operator.DeletePublicKey(ctx, instance.GetHost()); err != nil {
					return errors.Annotatef(err, "failed to delete public key")
				}
			}
		}
	}

	return nil
}
