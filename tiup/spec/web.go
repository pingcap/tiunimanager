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
	"github.com/pingcap/tiup/pkg/utils"
	"go.uber.org/zap"
)

// WebServerSpec represents the Master topology specification in topology.yaml
type WebServerSpec struct {
	Host            string               `yaml:"host"`
	SSHPort         int                  `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	ServerName      string               `yaml:"server_name,omitempty" default:"localhost" validate:"server_name:editable"`
	Port            int                  `yaml:"port,omitempty" default:"80"`
	TlsPort         int                  `yaml:"tls_port,omitempty" default:"443"`
	DeployDir       string               `yaml:"deploy_dir,omitempty"`
	DataDir         string               `yaml:"data_dir,omitempty"`
	LogDir          string               `yaml:"log_dir,omitempty"`
	Arch            string               `yaml:"arch,omitempty"`
	OS              string               `yaml:"os,omitempty"`
	ResourceControl meta.ResourceControl `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
}

// Status queries current status of the instance
func (s *WebServerSpec) Status(tlsCfg *tls.Config, _ ...string) string {
	client := utils.NewHTTPClient(statusQueryTimeout, tlsCfg)

	path := "/"
	url := fmt.Sprintf("http://%s:%d%s", s.Host, s.Port, path)

	// body doesn't have any status section needed
	_, err := client.Get(context.TODO(), url)
	if err != nil {
		return "Down"
	}
	return "Up"
}

// Role returns the component role of the instance
func (s *WebServerSpec) Role() string {
	return ComponentTiEMWebServer
}

// SSH returns the host and SSH port of the instance
func (s *WebServerSpec) SSH() (string, int) {
	return s.Host, s.SSHPort
}

// GetMainPort returns the main port of the instance
func (s *WebServerSpec) GetMainPort() int {
	return s.Port
}

// IsImported implements the instance interface, not needed for tiem
func (s *WebServerSpec) IsImported() bool {
	return false
}

// IgnoreMonitorAgent returns if the node does not have monitor agents available
func (s *WebServerSpec) IgnoreMonitorAgent() bool {
	return false
}

// WebServerComponent represents TiEM component.
type WebServerComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *WebServerComponent) Name() string {
	return ComponentTiEMWebServer
}

// Role implements Component interface.
func (c *WebServerComponent) Role() string {
	return RoleTiEMWeb
}

// Instances implements Component interface.
func (c *WebServerComponent) Instances() []Instance {
	ins := make([]Instance, 0)
	for _, s := range c.Topology.WebServers {
		s := s
		ports := make([]int, 0)
		if c.Topology.HasEnableHttps() {
			ports = append(ports, s.TlsPort)
		}
		ports = append(ports, s.Port)
		ins = append(ins, &WebServerInstance{
			ServerName: s.ServerName,
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				Port:         s.Port,
				SSHP:         s.SSHPort,

				Ports: ports,
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
			topo: c.Topology,
		})
	}
	return ins
}

// WebServerInstance represent the TiEM instance
type WebServerInstance struct {
	ServerName string
	BaseInstance
	topo *Specification
}

// InitConfig implement Instance interface
func (i *WebServerInstance) InitConfig(
	ctx context.Context,
	e ctxt.Executor,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	// build systemd service file
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

	resource := MergeResourceControl(i.topo.GlobalOptions.ResourceControl, i.resourceControl())
	systemCfg := system.NewConfig(comp, deployUser, paths.Deploy).
		WithMemoryLimit(resource.MemoryLimit).
		WithCPUQuota(resource.CPUQuota).
		WithLimitCORE(resource.LimitCORE).
		WithIOReadBandwidthMax(resource.IOReadBandwidthMax).
		WithIOWriteBandwidthMax(resource.IOWriteBandwidthMax).
		WithType("forking")

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

	spec := i.InstanceSpec.(*WebServerSpec)
	// render startup script
	scpt := scripts.NewTiEMWebServerScript(
		i.GetHost(),
		paths.Deploy,
		paths.Log,
	)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_nginx_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_nginx.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	// copy default configs
	if _, _, err := e.Execute(ctx,
		fmt.Sprintf("cp -r %s/bin/conf/* %s/conf/", paths.Deploy, paths.Deploy),
		false); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx,
		fmt.Sprintf("cp -r %s/bin/cert %s/cert", paths.Deploy, paths.Deploy),
		false); err != nil {
		return err
	}

	// render config file
	cfg := config.NewNginxConfig(
		i.GetHost(),
		paths.Deploy,
		paths.Log,
	).
		WithPort(i.Port).
		WithServerName(i.ServerName).
		WithRegistryEndpoints(i.topo.RegistryEndpoints()).
		WithEnableHttps(i.topo.HasEnableHttps()).
		WithTlsPort(spec.TlsPort).
		WithGrafanaAddress(i.topo.GrafanaEndpoints()).
		WithKibanaAddress(i.topo.KibanaEndpoints()).
		WithAlertManagerAddress(i.topo.AlertManagerEndpoints()).
		WithTracerAddress(i.topo.TracerWebAddress())

	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "nginx.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	openApiSrvList := config.NewNginxServerList(i.topo.APIServerEndpoints())
	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_server_config_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := openApiSrvList.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "server_list.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	fileSrvList := config.NewNginxServerList(i.topo.FileServerEndpoints())
	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_server_config_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := fileSrvList.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "file_server_list.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	return e.Transfer(ctx, fp, dst, false, 0)
}

// ScaleConfig deploy temporary config on scaling
func (i *WebServerInstance) ScaleConfig(
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

	spec := i.InstanceSpec.(*WebServerSpec)
	scpt := scripts.NewTiEMWebServerScript(
		i.GetHost(),
		paths.Deploy,
		paths.Log,
	)

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_nginx_%s_%d.sh", i.GetHost(), i.GetPort()))
	log.Infof("script path: %s", fp)
	if err := scpt.ScriptToFile(fp); err != nil {
		return err
	}

	dst := filepath.Join(paths.Deploy, "scripts", "run_nginx.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	// render config file
	cfg := config.NewNginxConfig(
		i.GetHost(),
		paths.Deploy,
		paths.Log,
	).
		WithPort(i.Port).
		WithServerName(i.ServerName).
		WithRegistryEndpoints(i.topo.RegistryEndpoints()).
		WithEnableHttps(i.topo.HasEnableHttps()).
		WithTlsPort(spec.TlsPort).
		WithGrafanaAddress(i.topo.GrafanaEndpoints()).
		WithKibanaAddress(i.topo.KibanaEndpoints()).
		WithAlertManagerAddress(i.topo.AlertManagerEndpoints()).
		WithTracerAddress(i.topo.TracerWebAddress())

	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "nginx.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	openApiSrvList := config.NewNginxServerList(i.topo.APIServerEndpoints())
	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_server_config_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := openApiSrvList.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "server_list.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	fileSrvList := config.NewNginxServerList(i.topo.FileServerEndpoints())
	fp = filepath.Join(paths.Cache, fmt.Sprintf("nginx_server_config_%s_%d.conf", i.GetHost(), i.GetPort()))
	if err := fileSrvList.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "file_server_list.conf")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}
	return e.Transfer(ctx, fp, dst, false, 0)
}
