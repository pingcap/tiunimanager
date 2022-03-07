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

	"github.com/pingcap-inc/tiem/tiup/templates/config"

	"github.com/pingcap-inc/tiem/tiup/templates/scripts"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/meta"
)

// FilebeatSpec represents the Filebeat topology specification in topology.yaml
type FilebeatSpec struct {
	Host            string                 `yaml:"host"`
	SSHPort         int                    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	Port            int                    `yaml:"port,omitempty" default:"0"`
	Imported        bool                   `yaml:"imported,omitempty"`
	IgnoreExporter  bool                   `yaml:"ignore_exporter,omitempty"`
	DeployDir       string                 `yaml:"deploy_dir,omitempty"`
	DataDir         string                 `yaml:"data_dir,omitempty"`
	LogDir          string                 `yaml:"log_dir,omitempty"`
	Config          map[string]interface{} `yaml:"config,omitempty" validate:"config:ignore"`
	ResourceControl meta.ResourceControl   `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
	Arch            string                 `yaml:"arch,omitempty"`
	OS              string                 `yaml:"os,omitempty"`
}

// Role returns the component role of the instance
func (s *FilebeatSpec) Role() string {
	return ComponentFilebeat
}

// SSH returns the host and SSH port of the instance
func (s *FilebeatSpec) SSH() (string, int) {
	return s.Host, s.SSHPort
}

// GetMainPort returns the main port of the instance
func (s *FilebeatSpec) GetMainPort() int {
	return s.Port
}

// IsImported returns if the node is imported from TiDB-Ansible
func (s *FilebeatSpec) IsImported() bool {
	return s.Imported
}

// IgnoreMonitorAgent returns if the node does not have monitor agents available
func (s *FilebeatSpec) IgnoreMonitorAgent() bool {
	return s.IgnoreExporter
}

// FilebeatComponent represents Filebeat component.
type FilebeatComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *FilebeatComponent) Name() string {
	return ComponentFilebeat
}

// Role implements Component interface.
func (c *FilebeatComponent) Role() string {
	return RoleLogServer
}

// Instances implements Component interface.
func (c *FilebeatComponent) Instances() []Instance {
	ins := make([]Instance, 0)
	for _, s := range c.Topology.FilebeatServers {
		s := s
		ins = append(ins, &FilebeatInstance{
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				Port:         0,
				SSHP:         s.SSHPort,
				Ports:        []int{},
				Dirs: []string{
					s.DeployDir,
					s.DataDir,
					s.LogDir,
				},
				StatusFn: func(_ *tls.Config, _ ...string) string {
					return "-"
				},
				UptimeFn: func(tlsCfg *tls.Config) time.Duration {
					return 0
				},
			},
			topo: c.Topology,
		})
	}
	return ins
}

// FilebeatInstance represent the filebeat instance
type FilebeatInstance struct {
	BaseInstance
	topo *Specification
}

// InitConfig implement Instance interface
func (i *FilebeatInstance) InitConfig(
	ctx context.Context,
	e ctxt.Executor,
	clusterName,
	clusterVersion,
	deployUser string,
	deployGroup string,
	paths meta.DirPaths,
) error {
	gOpts := *i.topo.BaseTopo().GlobalOptions
	if err := i.BaseInstance.InitConfig(ctx, e, gOpts, deployUser, deployGroup, paths); err != nil {
		return err
	}
	cfg := config.NewFilebeatConfig(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
	).
		WithElasticSearch(i.topo.ElasticSearchEndpoints()).
		WithTiEMLogs(i.topo.TiEMLogPaths())
	fp := filepath.Join(paths.Cache, fmt.Sprintf("filebeat_%s_%d.yml", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "conf", "filebeat.yml")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	// Transfer start script
	scpt := scripts.NewFilebeatScript(paths.Deploy)
	fp = filepath.Join(paths.Cache, fmt.Sprintf("run_filebeat_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := scpt.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "scripts", "run_filebeat.sh")
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
func (i *FilebeatInstance) ScaleConfig(
	ctx context.Context,
	e ctxt.Executor,
	topo Topology,
	clusterName,
	clusterVersion,
	deployUser string,
	deployGroup string,
	paths meta.DirPaths,
) error {
	if err := i.InitConfig(ctx, e, clusterName, clusterVersion, deployUser, deployGroup, paths); err != nil {
		return err
	}

	cfg := config.NewFilebeatConfig(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
	).
		WithElasticSearch(i.topo.ElasticSearchEndpoints()).
		WithTiEMLogs(i.topo.TiEMLogPaths())
	fp := filepath.Join(paths.Cache, fmt.Sprintf("filebeat_%s_%d.yml", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "conf", "filebeat.yml")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	scpt := scripts.NewFilebeatScript(paths.Deploy)
	fp = filepath.Join(paths.Cache, fmt.Sprintf("run_filebeat_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := scpt.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "scripts", "run_filebeat.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}
	return nil
}
