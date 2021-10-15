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

package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap/tiup/pkg/set"
)

// UpdateMeta is used to maintain the cluster meta information
type UpdateMeta struct {
	cluster        string
	metadata       *spec.TiEMMeta
	deletedNodeIDs []string
}

// Execute implements the Task interface
// the metadata especially the topology is in wide use,
// the other callers point to this field by a pointer,
// so we should update the original topology directly, and don't make a copy
func (u *UpdateMeta) Execute(ctx context.Context) error {
	deleted := set.NewStringSet(u.deletedNodeIDs...)
	topo := u.metadata.Topology

	metadbServers := make([]*spec.MetaDBServerSpec, 0)
	for i, instance := range (&spec.MetaDBComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		metadbServers = append(metadbServers, topo.MetaDBServers[i])
	}
	topo.MetaDBServers = metadbServers

	clsServers := make([]*spec.ClusterServerSpec, 0)
	for i, instance := range (&spec.ClusterServerComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		clsServers = append(clsServers, topo.ClusterServers[i])
	}
	topo.ClusterServers = clsServers

	apiServers := make([]*spec.APIServerSpec, 0)
	for i, instance := range (&spec.APIServerComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		apiServers = append(apiServers, topo.APIServers[i])
	}
	topo.APIServers = apiServers

	webServers := make([]*spec.WebServerSpec, 0)
	for i, instance := range (&spec.WebServerComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		webServers = append(webServers, topo.WebServers[i])
	}
	topo.WebServers = webServers

	tracerServers := make([]*spec.TracerServerSpec, 0)
	for i, instance := range (&spec.JaegerComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		tracerServers = append(tracerServers, topo.TracerServers[i])
	}
	topo.TracerServers = tracerServers

	esServers := make([]*spec.ElasticSearchSpec, 0)
	for i, instance := range (&spec.ElasticSearchComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		esServers = append(esServers, topo.ElasticSearchServers[i])
	}
	topo.ElasticSearchServers = esServers

	monitors := make([]*spec.PrometheusSpec, 0)
	for i, instance := range (&spec.MonitorComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		monitors = append(monitors, topo.Monitors[i])
	}
	topo.Monitors = monitors

	grafanas := make([]*spec.GrafanaSpec, 0)
	for i, instance := range (&spec.GrafanaComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		grafanas = append(grafanas, topo.Grafanas[i])
	}
	topo.Grafanas = grafanas

	alertmanagers := make([]*spec.AlertmanagerSpec, 0)
	for i, instance := range (&spec.AlertManagerComponent{Topology: topo}).Instances() {
		if deleted.Exist(instance.ID()) {
			continue
		}
		alertmanagers = append(alertmanagers, topo.Alertmanagers[i])
	}
	topo.Alertmanagers = alertmanagers

	return spec.SaveMetadata(u.cluster, u.metadata)
}

// Rollback implements the Task interface
func (u *UpdateMeta) Rollback(ctx context.Context) error {
	return spec.SaveMetadata(u.cluster, u.metadata)
}

// String implements the fmt.Stringer interface
func (u *UpdateMeta) String() string {
	return fmt.Sprintf("UpdateMeta: cluster=%s, deleted=`'%s'`", u.cluster, strings.Join(u.deletedNodeIDs, "','"))
}
