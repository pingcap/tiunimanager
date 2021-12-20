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

	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/tiup/templates/config"
	"github.com/pingcap-inc/tiem/tiup/templates/scripts"
	system "github.com/pingcap-inc/tiem/tiup/templates/systemd"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/checkpoint"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/meta"
	"go.uber.org/zap"
)

// KibanaSpec represents the Master topology specification in topology.yaml
type KibanaSpec struct {
	Host    string `yaml:"host"`
	SSHPort int    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	// Use Name to get the name with a default value if it's empty.
	Name            string                 `yaml:"name,omitempty" default:"tiem-cluster"`
	Port            int                    `yaml:"port,omitempty" default:"4109"`
	DeployDir       string                 `yaml:"deploy_dir,omitempty"`
	DataDir         string                 `yaml:"data_dir,omitempty"`
	LogDir          string                 `yaml:"log_dir,omitempty"`
	JavaHome        string                 `yaml:"java_home,omitempty" validate:"java_home:editable"`
	Config          map[string]interface{} `yaml:"config,omitempty" validate:"config:ignore"`
	Arch            string                 `yaml:"arch,omitempty"`
	OS              string                 `yaml:"os,omitempty"`
	ResourceControl meta.ResourceControl   `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
}

// Status queries current status of the instance
func (s *KibanaSpec) Status(tlsCfg *tls.Config, _ ...string) string {
	return statusByHost(s.Host, s.Port, "/status", nil)
}

// Role returns the component role of the instance
func (s *KibanaSpec) Role() string {
	return ComponentKibana
}

// SSH returns the host and SSH port of the instance
func (s *KibanaSpec) SSH() (string, int) {
	return s.Host, s.SSHPort
}

// GetMainPort returns the main port of the instance
func (s *KibanaSpec) GetMainPort() int {
	return s.Port
}

// IsImported implements the instance interface, not needed for tiem
func (s *KibanaSpec) IsImported() bool {
	return false
}

// IgnoreMonitorAgent returns if the node does not have monitor agents available
func (s *KibanaSpec) IgnoreMonitorAgent() bool {
	return false
}

// KibanaComponent represents TiEM component.
type KibanaComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *KibanaComponent) Name() string {
	return ComponentKibana
}

// Role implements Component interface.
func (c *KibanaComponent) Role() string {
	return RoleLogServer
}

// Instances implements Component interface.
func (c *KibanaComponent) Instances() []Instance {
	ins := make([]Instance, 0)
	for _, s := range c.Topology.KibanaServers {
		s := s
		ins = append(ins, &KibanaInstance{
			Name: s.Name,
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				Port:         s.Port,
				SSHP:         s.SSHPort,

				Ports: []int{
					s.Port,
				},
				Dirs: []string{
					s.DeployDir,
					s.DataDir,
					s.LogDir,
				},
				StatusFn: s.Status,
				UptimeFn: func(tlsCfg *tls.Config) time.Duration {
					return spec.UptimeByHost(s.Host, s.Port, tlsCfg)
				},
			},
			topo:     c.Topology,
			JavaHome: s.JavaHome,
		})
	}
	return ins
}

// KibanaInstance represent the TiEM instance
type KibanaInstance struct {
	Name string
	BaseInstance
	topo     *Specification
	JavaHome string
}

// InitConfig implement Instance interface
func (i *KibanaInstance) InitConfig(
	ctx context.Context,
	e ctxt.Executor,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	comp := i.ComponentName()
	host := i.GetHost()
	port := i.GetPort()
	sysCfg := filepath.Join(paths.Cache, fmt.Sprintf("%s-%s-%d.service", comp, host, port))

	var err error
	// insert checkpoint
	point := checkpoint.Acquire(ctx, CopyConfigFile, map[string]interface{}{"config-file": sysCfg})
	defer func() {
		point.Release(err, zap.String("config-file", sysCfg))
	}()

	if point.Hit() != nil {
		return nil
	}

	systemCfg := system.NewJavaAppConfig(comp, deployUser, paths.Deploy, i.JavaHome)

	if err := systemCfg.ConfigToFile(sysCfg); err != nil {
		return errors.Trace(err)
	}
	tgt := filepath.Join("/tmp", comp+"_"+uuid.New().String()+".service")
	if err := e.Transfer(ctx, sysCfg, tgt, false, 0); err != nil {
		return errors.Annotatef(err, "transfer from %s to %s failed", sysCfg, tgt)
	}
	cmd := fmt.Sprintf("mv %s /etc/systemd/system/%s-%d.service", tgt, comp, port)
	if _, _, err := e.Execute(ctx, cmd, true); err != nil {
		return errors.Annotatef(err, "execute: %s", cmd)
	}

	spec := i.InstanceSpec.(*KibanaSpec)
	cfg := config.NewKibanaConfig(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
	).
		WithPort(spec.Port).
		WithName(spec.Name).
		WithElasticsearch(i.topo.ElasticSearchEndpoints())
	fp := filepath.Join(paths.Cache, fmt.Sprintf("kibana_%s_%d.yml", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx,
		fmt.Sprintf("cp -r %s/bin/config/* %s/conf/", paths.Deploy, paths.Deploy),
		false); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "conf", "kibana.yml")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	scpt := scripts.NewKibanaScript(
		paths.Deploy,
	)

	fp = filepath.Join(paths.Cache, fmt.Sprintf("run_kibana_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "scripts", "run_kibana.sh")
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
func (i *KibanaInstance) ScaleConfig(
	ctx context.Context,
	e ctxt.Executor,
	topo Topology,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	if err := i.InitConfig(ctx, e, clusterName, clusterVersion, deployUser, paths); err != nil {
		return err
	}

	scpt := scripts.NewKibanaScript(
		paths.Deploy,
	)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_kibana_%s_%d.sh", i.GetHost(), i.GetPort()))
	log.Infof("script path: %s", fp)
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}

	dst := filepath.Join(paths.Deploy, "scripts", "run_kibana.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	return nil
}
