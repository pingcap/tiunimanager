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

	"github.com/pingcap/tiup/pkg/cluster/spec"
)

var (
	specManager *SpecManager
)

// TiEMMeta is the specification of generic cluster metadata
type TiEMMeta struct {
	User    string `yaml:"user"`                   // the user to run and manage cluster on remote
	Group   string `yaml:"group"`                  // the group to run and manage cluster on remote
	Version string `yaml:"tiem_version"`           // the version of TiEM
	OpsVer  string `yaml:"last_ops_ver,omitempty"` // the version of ourself that updated the meta last time

	Topology *Specification `yaml:"topology"`
}

var _ UpgradableMetadata = &TiEMMeta{}

// SetVersion implement UpgradableTiEMMeta interface.
func (m *TiEMMeta) SetVersion(s string) {
	m.Version = s
}

// SetUser implement UpgradableTiEMMeta interface.
func (m *TiEMMeta) SetUser(s string) {
	m.User = s
}

// SetGroup implement UpgradableTiEMMeta interface.
func (m *TiEMMeta) SetGroup(s string) {
	m.Group = s
}

// GetTopology implements TiEMMeta interface.
func (m *TiEMMeta) GetTopology() Topology {
	return m.Topology
}

// SetTopology implements TiEMMeta interface.
func (m *TiEMMeta) SetTopology(topo Topology) {
	tiemTopo, ok := topo.(*Specification)
	if !ok {
		panic(fmt.Sprintln("wrong type: ", reflect.TypeOf(topo)))
	}

	m.Topology = tiemTopo
}

// GetBaseMeta implements TiEMMeta interface.
func (m *TiEMMeta) GetBaseMeta() *BaseMeta {
	return &BaseMeta{
		Version: m.Version,
		User:    m.User,
		Group:   m.Group,
	}
}

// GetSpecManager return the spec manager of dm cluster.
func GetSpecManager() *SpecManager {
	if specManager == nil {
		specManager = NewSpec(
			filepath.Join(spec.ProfileDir(), spec.TiUPClusterDir),
			func() Metadata {
				return &TiEMMeta{
					Topology: new(Specification),
				}
			},
		)
	}
	return specManager
}
