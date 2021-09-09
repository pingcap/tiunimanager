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

package spec

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

var (
	globalOptionTypeName  = reflect.TypeOf(spec.GlobalOptions{}).Name()
	monitorOptionTypeName = reflect.TypeOf(spec.MonitoredOptions{}).Name()
)

func setDefaultDir(parent, role, port string, field reflect.Value) {
	if field.String() != "" {
		return
	}
	if defaults.CanUpdate(field.Interface()) {
		dir := fmt.Sprintf("%s-%s", role, port)
		field.Set(reflect.ValueOf(filepath.Join(parent, dir)))
	}
}

func findField(v reflect.Value, fieldName string) (int, bool) {
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Name == fieldName {
			return i, true
		}
	}
	return -1, false
}

// Skip global/monitored/job options
func isSkipField(field reflect.Value) bool {
	if field.Kind() == reflect.Ptr {
		if field.IsZero() {
			return true
		}
		field = field.Elem()
	}
	tp := field.Type().Name()
	return tp == globalOptionTypeName || tp == monitorOptionTypeName
}

type (
	// InstanceSpec represent a instance specification
	InstanceSpec interface {
		Role() string
		SSH() (string, int)
		GetMainPort() int
		IsImported() bool
		IgnoreMonitorAgent() bool
	}
)

// Component represents a component of the cluster.
type Component = spec.Component

// Instance represents an instance
type Instance = spec.Instance

// GetGlobalOptions returns cluster topology
func (topo *Specification) GetGlobalOptions() spec.GlobalOptions {
	return topo.GlobalOptions
}

// GetMonitoredOptions returns MonitoredOptions
func (topo *Specification) GetMonitoredOptions() *spec.MonitoredOptions {
	return topo.MonitoredOptions
}

// ComponentsByStopOrder return component in the order need to stop.
func (topo *Specification) ComponentsByStopOrder() (comps []Component) {
	comps = topo.ComponentsByStartOrder()
	// revert order
	i := 0
	j := len(comps) - 1
	for i < j {
		comps[i], comps[j] = comps[j], comps[i]
		i++
		j--
	}
	return
}

// ComponentsByStartOrder return component in the order need to start.
func (topo *Specification) ComponentsByStartOrder() (comps []Component) {
	// "elasticsearch", "monitor", "metadb", "api-server", "web", "cluster-server"
	comps = append(comps, &ElasticSearchComponent{topo})
	comps = append(comps, &spec.MonitorComponent{Topology: topo})
	comps = append(comps, &spec.GrafanaComponent{Topology: topo})
	comps = append(comps, &spec.AlertManagerComponent{Topology: topo})
	// TODO: add tracer
	comps = append(comps, &MetaDBComponent{topo})
	comps = append(comps, &APIServerComponent{topo})
	comps = append(comps, &WebServerComponent{topo})
	comps = append(comps, &ClusterServerComponent{topo})
	return
}

// ComponentsByUpdateOrder return component in the order need to be updated.
func (topo *Specification) ComponentsByUpdateOrder() (comps []Component) {
	// "metadb", "api-server", "web", "cluster-server", "monitor", "elasticsearch"
	comps = append(comps, &MetaDBComponent{topo})
	comps = append(comps, &APIServerComponent{topo})
	comps = append(comps, &WebServerComponent{topo})
	comps = append(comps, &ClusterServerComponent{topo})
	// TODO: add tracer
	comps = append(comps, &spec.MonitorComponent{Topology: topo})
	comps = append(comps, &spec.GrafanaComponent{Topology: topo})
	comps = append(comps, &spec.AlertManagerComponent{Topology: topo})
	comps = append(comps, &ElasticSearchComponent{topo})
	return
}

// IterComponent iterates all components in component starting order
func (topo *Specification) IterComponent(fn func(comp Component)) {
	for _, comp := range topo.ComponentsByStartOrder() {
		fn(comp)
	}
}

// IterInstance iterates all instances in component starting order
func (topo *Specification) IterInstance(fn func(instance Instance)) {
	for _, comp := range topo.ComponentsByStartOrder() {
		for _, inst := range comp.Instances() {
			fn(inst)
		}
	}
}

// IterHost iterates one instance for each host
func (topo *Specification) IterHost(fn func(instance Instance)) {
	hostMap := make(map[string]bool)
	for _, comp := range topo.ComponentsByStartOrder() {
		for _, inst := range comp.Instances() {
			host := inst.GetHost()
			_, ok := hostMap[host]
			if !ok {
				hostMap[host] = true
				fn(inst)
			}
		}
	}
}

// AllTiEMComponentNames contains the names of all tiem components.
// should include all components in ComponentsByStartOrder
func AllTiEMComponentNames() (roles []string) {
	tp := &Specification{}
	tp.IterComponent(func(c Component) {
		roles = append(roles, c.Name())
	})

	return
}
