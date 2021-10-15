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
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

const (
	// Timeout in second when quering node status
	//statusQueryTimeout = 10 * time.Second

	// the prometheus metric name of start time of the process since unix epoch in seconds.
	promMetricStartTimeSeconds = "process_start_time_seconds"
)

// Topology represents specification of the cluster.
type Topology interface {
	Type() string
	BaseTopo() *BaseTopo
	// Validate validates the topology specification and produce error if the
	// specification invalid (e.g: port conflicts or directory conflicts)
	Validate() error

	// Instances() []Instance
	ComponentsByStartOrder() []Component
	ComponentsByStopOrder() []Component
	ComponentsByUpdateOrder() []Component
	IterInstance(fn func(instance Instance))
	GetMonitoredOptions() *spec.MonitoredOptions
	// count how many time a path is used by instances in cluster
	CountDir(host string, dir string) int
	Merge(that Topology) Topology
	FillHostArch(hostArchmap map[string]string) error

	ScaleOutTopology
}

// ScaleOutTopology represents a scale out metadata.
type ScaleOutTopology interface {
	// Inherit existing global configuration. We must assign the inherited values before unmarshalling
	// because some default value rely on the global options and monitored options.
	// TODO: we should separate the  unmarshal and setting default value.
	NewPart() Topology
	MergeTopo(topo Topology) Topology
}

// Metadata of a cluster.
type Metadata interface {
	GetTopology() Topology
	SetTopology(topo Topology)
	GetBaseMeta() *BaseMeta

	UpgradableMetadata
}

// BaseMeta is the base info of metadata.
type BaseMeta struct {
	User    string
	Group   string
	Version string
	OpsVer  *string `yaml:"last_ops_ver,omitempty"` // the version of ourself that updated the meta last time
}

// UpgradableMetadata represents a upgradable Metadata.
type UpgradableMetadata interface {
	SetVersion(s string)
	SetUser(u string)
}
