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
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pingcap-inc/tiem/tiup/templates/scripts"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/meta"
)

// MetaDBServerSpec represents the Master topology specification in topology.yaml
type MetaDBServerSpec struct {
	Host    string `yaml:"host"`
	SSHPort int    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	// Use Name to get the name with a default value if it's empty.
	Name            string                 `yaml:"name,omitempty"`
	Port            int                    `yaml:"port,omitempty" default:"4100"`
	ClientPort      int                    `yaml:"registry_client_port,omitempty" default:"4101"`
	PeerPort        int                    `yaml:"registry_peer_port,omitempty" default:"4102"`
	MetricsPort     int                    `yaml:"metrics_port,omitempty" default:"4121"`
	DeployDir       string                 `yaml:"deploy_dir,omitempty"`
	DataDir         string                 `yaml:"data_dir,omitempty"`
	LogDir          string                 `yaml:"log_dir,omitempty"`
	Config          map[string]interface{} `yaml:"config,omitempty" validate:"config:ignore"`
	Arch            string                 `yaml:"arch,omitempty"`
	OS              string                 `yaml:"os,omitempty"`
	LogLevel        string                 `yaml:"log_level,omitempty" default:"info" validate:"log_level:editable"`
	ResourceControl *meta.ResourceControl  `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
}

// Status queries current status of the instance
func (s *MetaDBServerSpec) Status(tlsCfg *tls.Config, _ ...string) string {
	return "N/A"
}

// Role returns the component role of the instance
func (s *MetaDBServerSpec) Role() string {
	return ComponentTiEMMetaDBServer
}

// SSH returns the host and SSH port of the instance
func (s *MetaDBServerSpec) SSH() (string, int) {
	return s.Host, s.SSHPort
}

// GetMainPort returns the main port of the instance
func (s *MetaDBServerSpec) GetMainPort() int {
	return s.Port
}

// IsImported implements the instance interface, not needed for tiem
func (s *MetaDBServerSpec) IsImported() bool {
	return false
}

// IgnoreMonitorAgent returns if the node does not have monitor agents available
func (s *MetaDBServerSpec) IgnoreMonitorAgent() bool {
	return false
}

// MetaDBComponent represents TiEM component.
type MetaDBComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *MetaDBComponent) Name() string {
	return ComponentTiEMMetaDBServer
}

// Role implements Component interface.
func (c *MetaDBComponent) Role() string {
	return RoleTiEMMetaDB
}

// Instances implements Component interface.
func (c *MetaDBComponent) Instances() []Instance {
	ins := make([]Instance, 0)
	for _, s := range c.Topology.MetaDBServers {
		s := s
		ins = append(ins, &MetaDBInstance{
			Name: s.Name,
			BaseInstance: spec.BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				Port:         s.Port,
				SSHP:         s.SSHPort,

				Ports: []int{
					s.Port,
					s.PeerPort,
				},
				Dirs: []string{
					s.DeployDir,
					s.DataDir,
				},
				StatusFn: s.Status,
				UptimeFn: func(tlsCfg *tls.Config) time.Duration {
					return spec.UptimeByHost(s.Host, s.Port, tlsCfg)
				},
			},
			topo: c.Topology,
		})
	}
	return ins
}

// MetaDBInstance represent the TiEM instance
type MetaDBInstance struct {
	Name string
	spec.BaseInstance
	topo *Specification
}

// InitConfig implement Instance interface
func (i *MetaDBInstance) InitConfig(
	ctx context.Context,
	e ctxt.Executor,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	if err := i.BaseInstance.InitConfig(ctx, e, i.topo.GlobalOptions, deployUser, paths); err != nil {
		return err
	}

	spec := i.InstanceSpec.(*MetaDBServerSpec)
	scpt := scripts.NewTiEMMetaDBScript(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
		spec.LogLevel,
	).
		WithPort(spec.Port).
		WithPeerPort(spec.PeerPort).
		WithClientPort(spec.ClientPort).
		WithMetricsPort(spec.MetricsPort)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_tiem_metadb_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_tiem_metadb.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	// no config file needed
	return nil
}

// ScaleConfig deploy temporary config on scaling
func (i *MetaDBInstance) ScaleConfig(
	ctx context.Context,
	e ctxt.Executor,
	topo spec.Topology,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	if err := i.InitConfig(ctx, e, clusterName, clusterVersion, deployUser, paths); err != nil {
		return err
	}

	spec := i.InstanceSpec.(*MetaDBServerSpec)
	scpt := scripts.NewTiEMMetaDBScript(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
		spec.LogLevel,
	).
		WithPort(spec.Port).
		WithPeerPort(spec.PeerPort).
		WithClientPort(spec.ClientPort).
		WithMetricsPort(spec.MetricsPort)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_tiem_metadb_%s_%d.sh", i.GetHost(), i.GetPort()))
	log.Infof("script path: %s", fp)
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}

	dst := filepath.Join(paths.Deploy, "scripts", "run_tiem_metadb.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	return nil
}
