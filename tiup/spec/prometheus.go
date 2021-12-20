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
	"reflect"
	"time"

	"github.com/pingcap-inc/tiem/tiup/templates/config"
	"github.com/pingcap-inc/tiem/tiup/templates/scripts"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/meta"
	"github.com/pingcap/tiup/pkg/set"
)

// PrometheusSpec represents the Prometheus Server topology specification in topology.yaml
type PrometheusSpec struct {
	Host                  string                 `yaml:"host"`
	SSHPort               int                    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	Imported              bool                   `yaml:"imported,omitempty"`
	Patched               bool                   `yaml:"patched,omitempty"`
	IgnoreExporter        bool                   `yaml:"ignore_exporter,omitempty"`
	Port                  int                    `yaml:"port" default:"4110"`
	DeployDir             string                 `yaml:"deploy_dir,omitempty"`
	DataDir               string                 `yaml:"data_dir,omitempty"`
	LogDir                string                 `yaml:"log_dir,omitempty"`
	NumaNode              string                 `yaml:"numa_node,omitempty" validate:"numa_node:editable"`
	RemoteConfig          Remote                 `yaml:"remote_config,omitempty" validate:"remote_config:ignore"`
	ExternalAlertmanagers []ExternalAlertmanager `yaml:"external_alertmanagers" validate:"external_alertmanagers:ignore"`
	Retention             string                 `yaml:"storage_retention,omitempty" default:"30d" validate:"storage_retention:editable"`
	ResourceControl       meta.ResourceControl   `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
	Arch                  string                 `yaml:"arch,omitempty"`
	OS                    string                 `yaml:"os,omitempty"`
	RuleDir               string                 `yaml:"rule_dir,omitempty" validate:"rule_dir:editable"`
}

// Remote prometheus remote config
type Remote struct {
	RemoteWrite []map[string]interface{} `yaml:"remote_write,omitempty" validate:"remote_write:ignore"`
	RemoteRead  []map[string]interface{} `yaml:"remote_read,omitempty" validate:"remote_read:ignore"`
}

// ExternalAlertmanager configs prometheus to include alertmanagers not deployed in current cluster
type ExternalAlertmanager struct {
	Host    string `yaml:"host"`
	WebPort int    `yaml:"web_port" default:"4112"`
}

// Role returns the component role of the instance
func (s *PrometheusSpec) Role() string {
	return ComponentPrometheus
}

// SSH returns the host and SSH port of the instance
func (s *PrometheusSpec) SSH() (string, int) {
	return s.Host, s.SSHPort
}

// GetMainPort returns the main port of the instance
func (s *PrometheusSpec) GetMainPort() int {
	return s.Port
}

// IsImported returns if the node is imported from TiDB-Ansible
func (s *PrometheusSpec) IsImported() bool {
	return s.Imported
}

// IgnoreMonitorAgent returns if the node does not have monitor agents available
func (s *PrometheusSpec) IgnoreMonitorAgent() bool {
	return s.IgnoreExporter
}

// MonitorComponent represents Monitor component.
type MonitorComponent struct{ Topology }

// Name implements Component interface.
func (c *MonitorComponent) Name() string {
	return ComponentPrometheus
}

// Role implements Component interface.
func (c *MonitorComponent) Role() string {
	return RoleMonitor
}

// Instances implements Component interface.
func (c *MonitorComponent) Instances() []Instance {
	servers := c.BaseTopo().Monitors
	ins := make([]Instance, 0, len(servers))

	for _, s := range servers {
		ins = append(ins, &MonitorInstance{BaseInstance{
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
			StatusFn: func(_ *tls.Config, _ ...string) string {
				return statusByHost(s.Host, s.Port, "/-/ready", nil)
			},
			UptimeFn: func(tlsCfg *tls.Config) time.Duration {
				return UptimeByHost(s.Host, s.Port, tlsCfg)
			},
		}, c.Topology})
	}
	return ins
}

// MonitorInstance represent the monitor instance
type MonitorInstance struct {
	BaseInstance
	topo Topology
}

// InitConfig implement Instance interface
func (i *MonitorInstance) InitConfig(
	ctx context.Context,
	e ctxt.Executor,
	clusterName,
	clusterVersion,
	deployUser string,
	paths meta.DirPaths,
) error {
	gOpts := *i.topo.BaseTopo().GlobalOptions
	if err := i.BaseInstance.InitConfig(ctx, e, gOpts, deployUser, paths); err != nil {
		return err
	}

	//enableTLS := gOpts.TLSEnabled
	// transfer run script
	spec := i.InstanceSpec.(*PrometheusSpec)
	cfg := scripts.NewPrometheusScript(
		i.GetHost(),
		paths.Deploy,
		paths.Data[0],
		paths.Log,
	).
		WithPort(spec.Port).
		WithRetention(spec.Retention)
	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_prometheus_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ScriptToFile(fp); err != nil {
		return err
	}

	dst := filepath.Join(paths.Deploy, "scripts", "run_prometheus.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0); err != nil {
		return err
	}

	if _, _, err := e.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	topoHasField := func(field string) (reflect.Value, bool) {
		return findSliceField(i.topo, field)
	}
	monitoredOptions := i.topo.GetMonitoredOptions()

	// transfer config
	fp = filepath.Join(paths.Cache, fmt.Sprintf("prometheus_%s_%d.yml", i.GetHost(), i.GetPort()))
	cfig := config.NewPrometheusConfig(clusterName)
	uniqueHosts := set.NewStringSet()

	if servers, found := topoHasField("ClusterServers"); found {
		for i := 0; i < servers.Len(); i++ {
			srv := servers.Index(i).Interface().(*ClusterServerSpec)
			uniqueHosts.Insert(srv.Host)
			cfig.AddClusterServer(srv.Host, uint64(srv.MetricsPort))
		}
	}

	if servers, found := topoHasField("APIServers"); found {
		for i := 0; i < servers.Len(); i++ {
			srv := servers.Index(i).Interface().(*APIServerSpec)
			uniqueHosts.Insert(srv.Host)
			cfig.AddAPIServer(srv.Host, uint64(srv.MetricsPort))
		}
	}

	if servers, found := topoHasField("FileServers"); found {
		for i := 0; i < servers.Len(); i++ {
			srv := servers.Index(i).Interface().(*FileServerSpec)
			uniqueHosts.Insert(srv.Host)
			cfig.AddFileServer(srv.Host, uint64(srv.MetricsPort))
		}
	}

	if servers, found := topoHasField("Alertmanagers"); found {
		for i := 0; i < servers.Len(); i++ {
			alertmanager := servers.Index(i).Interface().(*AlertmanagerSpec)
			uniqueHosts.Insert(alertmanager.Host)
			cfig.AddAlertmanager(alertmanager.Host, uint64(alertmanager.WebPort))
		}
	}

	if monitoredOptions != nil {
		for host := range uniqueHosts {
			cfig.AddNodeExpoertor(host, uint64(monitoredOptions.NodeExporterPort))
			cfig.AddBlackboxExporter(host, uint64(monitoredOptions.BlackboxExporterPort))
		}
	}

	for _, alertmanager := range spec.ExternalAlertmanagers {
		cfig.AddAlertmanager(alertmanager.Host, uint64(alertmanager.WebPort))
	}

	/*
		if spec.RuleDir != "" {
			filter := func(name string) bool { return strings.HasSuffix(name, ".rules.yml") }
			err := i.IteratorLocalConfigDir(ctx, spec.RuleDir, filter, func(name string) error {
				cfig.AddLocalRule(name)
				return nil
			})
			if err != nil {
				return errors.Annotate(err, "add local rule")
			}
		}
	*/
	if err := i.installRules(ctx, e, paths.Deploy, clusterName, clusterVersion); err != nil {
		return errors.Annotate(err, "install rules")
	}

	if err := i.initRules(ctx, e, spec, paths, clusterName); err != nil {
		return err
	}

	if err := cfig.ConfigToFile(fp); err != nil {
		return err
	}
	dst = filepath.Join(paths.Deploy, "conf", "prometheus.yml")
	return e.Transfer(ctx, fp, dst, false, 0)
}

// We only really installRules for dm cluster because the rules(*.rules.yml) packed with the prometheus
// component is designed for tidb cluster (the dm cluster use the same prometheus component with tidb
// cluster), and the rules for dm cluster is packed in the dm-master component. So if deploying tidb
// cluster, the rules is correct, if deploying dm cluster, we should remove rules for tidb and install
// rules for dm.
func (i *MonitorInstance) installRules(ctx context.Context, e ctxt.Executor, deployDir, clusterName, clusterVersion string) error {
	if i.topo.Type() != TopoTypeTiEM {
		return nil
	}

	/*
		tmp := filepath.Join(deployDir, "_tiup_tmp")
		_, stderr, err := e.Execute(ctx, fmt.Sprintf("mkdir -p %s", tmp), false)
		if err != nil {
			return errors.Annotatef(err, "stderr: %s", string(stderr))
		}
			srcPath := PackagePath(ComponentTiEMMetaDBServer, clusterVersion, i.OS(), i.Arch())
			dstPath := filepath.Join(tmp, filepath.Base(srcPath))

			err = e.Transfer(ctx, srcPath, dstPath, false, 0)
			if err != nil {
				return err
			}

			cmd := fmt.Sprintf(`tar --no-same-owner -zxf %s -C %s && rm %s`, dstPath, tmp, dstPath)
			_, stderr, err = e.Execute(ctx, cmd, false)
			if err != nil {
				return errors.Annotatef(err, "stderr: %s", string(stderr))
			}
	*/
	/*
		// copy dm-master/conf/*.rules.yml
		targetDir := filepath.Join(deployDir, "bin", "prometheus")
		cmds := []string{
			"mkdir -p %[1]s",
			`find %[1]s -type f -name "*.rules.yml" -delete`,
			`find %[2]s/dm-master/conf -type f -name "*.rules.yml" -exec cp {} %[1]s \;`,
			"rm -rf %[2]s",
			`find %[1]s -maxdepth 1 -type f -name "*.rules.yml" -exec sed -i "s/ENV_LABELS_ENV/%[3]s/g" {} \;`,
		}
		_, stderr, err = e.Execute(ctx, fmt.Sprintf(strings.Join(cmds, " && "), targetDir, tmp, clusterName), false)
		if err != nil {
			return errors.Annotatef(err, "stderr: %s", string(stderr))
		}
	*/
	return nil
}

func (i *MonitorInstance) initRules(ctx context.Context, e ctxt.Executor, spec *PrometheusSpec, paths meta.DirPaths, clusterName string) error {
	/*
		// To make this step idempotent, we need cleanup old rules first
		cmds := []string{
			"mkdir -p %[1]s/conf",
			`find %[1]s/conf -type f -name "*.rules.yml" -delete`,
			`find %[1]s/bin/prometheus -maxdepth 1 -type f -name "*.rules.yml" -exec cp {} %[1]s/conf/ \;`,
			`find %[1]s/conf -maxdepth 1 -type f -name "*.rules.yml" -exec sed -i -e "s/ENV_LABELS_ENV/%[2]s/g" {} \;`,
		}
		_, stderr, err := e.Execute(ctx, fmt.Sprintf(strings.Join(cmds, " && "), paths.Deploy, clusterName), false)
		if err != nil {
			return errors.Annotatef(err, "stderr: %s", string(stderr))
		}
	*/
	return nil
}

// ScaleConfig deploy temporary config on scaling
func (i *MonitorInstance) ScaleConfig(
	ctx context.Context,
	e ctxt.Executor,
	topo Topology,
	clusterName string,
	clusterVersion string,
	deployUser string,
	paths meta.DirPaths,
) error {
	s := i.topo
	defer func() { i.topo = s }()
	i.topo = topo
	return i.InitConfig(ctx, e, clusterName, clusterVersion, deployUser, paths)
}
