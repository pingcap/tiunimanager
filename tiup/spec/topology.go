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
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/pingcap-inc/tiem/tiup/templates/config"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/cluster/executor"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/meta"
	"github.com/pingcap/tiup/pkg/set"
)

const (
	DBName             = "em.db"
	statusQueryTimeout = 2 * time.Second
)

// Component names and roles
const (
	ComponentTiEMClusterServer   = "cluster-server" // tiem-cluster
	ComponentTiEMAPIServer       = "openapi-server" // tiem-api
	ComponentTiEMFileServer      = "file-server"    // tiem-file
	ComponentTiEMTracerServer    = "jaeger"         // tiem-tracer
	ComponentTiEMWebServer       = "nginx"          // tiem-web
	ComponentElasticSearchServer = "elasticsearch"  // es
	ComponentKibana              = "kibana"         // kibana
	ComponentFilebeat            = "filebeat"       // filebeat
	ComponentNodeExporter        = "node-exporter"  // different with in tiup-cluster (node_exporter)
	ComponentGrafana             = "grafana"
	ComponentAlertmanager        = "alertmanager"
	ComponentPrometheus          = "prometheus"

	RoleTiEMMetaDB  = "metadb"
	RoleTiEMCluster = "cluster"
	RoleTiEMAPI     = "api"
	RoleTiEMFile    = "file"
	RoleTiEMTracer  = "tracer"
	RoleTiEMWeb     = "web"
	RoleLogServer   = "log"
	RoleMonitor     = "monitor"

	TopoTypeTiEM = "em"
)

// MonitoredOptions represents the monitored node configuration
type MonitoredOptions struct {
	NodeExporterPort int                  `yaml:"node_exporter_port,omitempty" default:"9100"`
	DeployDir        string               `yaml:"deploy_dir,omitempty"`
	DataDir          string               `yaml:"data_dir,omitempty"`
	LogDir           string               `yaml:"log_dir,omitempty"`
	NumaNode         string               `yaml:"numa_node,omitempty" validate:"numa_node:editable"`
	ResourceControl  meta.ResourceControl `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
}

// GlobalOptions represents the global options for all groups in topology
// specification in topology.yaml
type GlobalOptions struct {
	User                     string               `yaml:"user,omitempty" default:"tidb"`
	Group                    string               `yaml:"group,omitempty"`
	SSHPort                  int                  `yaml:"ssh_port,omitempty" default:"22" validate:"ssh_port:editable"`
	SSHType                  executor.SSHType     `yaml:"ssh_type,omitempty" default:"builtin"`
	DeployDir                string               `yaml:"deploy_dir,omitempty" default:"deploy"`
	DataDir                  string               `yaml:"data_dir,omitempty" default:"data"`
	LogDir                   string               `yaml:"log_dir,omitempty"`
	LogLevel                 string               `yaml:"log_level,omitempty"`
	ExternalElasticsearchUrl string               `yaml:"external_elasticsearch_url" validate:"external_elasticsearch_url:editable"`
	ResourceControl          meta.ResourceControl `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
	OS                       string               `yaml:"os,omitempty" default:"linux"`
	Arch                     string               `yaml:"arch,omitempty"`
	LoginHostUser            string               `yaml:"login_host_user"`
	LoginPrivateKeyPath      string               `yaml:"login_private_key_path"`
	LoginPublicKeyPath       string               `yaml:"login_public_key_path"`
}

// BaseTopo is the base info to topology.
type BaseTopo struct {
	GlobalOptions    *GlobalOptions
	MonitoredOptions *MonitoredOptions
	MasterList       []string

	Monitors       []*PrometheusSpec
	Grafanas       []*GrafanaSpec
	Alertmanagers  []*AlertmanagerSpec
	ElasticSearchs []*ElasticSearchSpec

	WebServers []*WebServerSpec
}

// Specification represents the specification of topology.yaml
type Specification struct {
	GlobalOptions        GlobalOptions        `yaml:"global,omitempty" validate:"global:editable"`
	MonitoredOptions     *MonitoredOptions    `yaml:"monitored,omitempty" validate:"monitored:editable"`
	ClusterServers       []*ClusterServerSpec `yaml:"em_cluster_servers"`
	APIServers           []*APIServerSpec     `yaml:"em_api_servers"`
	FileServers          []*FileServerSpec    `yaml:"em_file_servers"`
	WebServers           []*WebServerSpec     `yaml:"em_web_servers"`
	TracerServers        []*TracerServerSpec  `yaml:"tracer_servers"`
	ElasticSearchServers []*ElasticSearchSpec `yaml:"elasticsearch_servers"`
	KibanaServers        []*KibanaSpec        `yaml:"kibana_servers,omitempty"`
	FilebeatServers      []*FilebeatSpec      `yaml:"filebeat_servers,omitempty"`
	Monitors             []*PrometheusSpec    `yaml:"monitoring_servers,omitempty"`
	Grafanas             []*GrafanaSpec       `yaml:"grafana_servers,omitempty"`
	Alertmanagers        []*AlertmanagerSpec  `yaml:"alertmanager_servers,omitempty"`
}

// UnmarshalYAML sets default values when unmarshaling the topology file
func (s *Specification) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type topology Specification
	if err := unmarshal((*topology)(s)); err != nil {
		return err
	}

	if err := defaults.Set(s); err != nil {
		return errors.Trace(err)
	}

	if s.MonitoredOptions != nil {
		// Set monitored options
		if s.MonitoredOptions.DeployDir == "" {
			s.MonitoredOptions.DeployDir = filepath.Join(s.GlobalOptions.DeployDir,
				fmt.Sprintf("%s-%d", spec.RoleMonitor, s.MonitoredOptions.NodeExporterPort))
		}
		if s.MonitoredOptions.DataDir == "" {
			s.MonitoredOptions.DataDir = filepath.Join(s.GlobalOptions.DataDir,
				fmt.Sprintf("%s-%d", spec.RoleMonitor, s.MonitoredOptions.NodeExporterPort))
		}
		if s.MonitoredOptions.LogDir == "" {
			s.MonitoredOptions.LogDir = "log"
		}
		if !strings.HasPrefix(s.MonitoredOptions.LogDir, "/") &&
			!strings.HasPrefix(s.MonitoredOptions.LogDir, s.MonitoredOptions.DeployDir) {
			s.MonitoredOptions.LogDir = filepath.Join(s.MonitoredOptions.DeployDir, s.MonitoredOptions.LogDir)
		}
	}

	if err := fillTiEMCustomDefaults(&s.GlobalOptions, s); err != nil {
		return err
	}

	return s.Validate()
}

// platformConflictsDetect checks for conflicts in topology for different OS / Arch
// for set to the same host / IP
func (s *Specification) platformConflictsDetect() error {
	type (
		conflict struct {
			os   string
			arch string
			cfg  string
		}
	)

	platformStats := map[string]conflict{}
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// skip nodes imported from TiDB-Ansible
			if compSpec.Addr().Interface().(InstanceSpec).IsImported() {
				continue
			}
			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := topoType.Field(i).Tag.Get("yaml")
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}

			// platform conflicts
			stat := conflict{
				cfg: cfg,
			}
			if j, found := findField(compSpec, "OS"); found {
				stat.os = compSpec.Field(j).String()
			}
			if j, found := findField(compSpec, "Arch"); found {
				stat.arch = compSpec.Field(j).String()
			}

			prev, exist := platformStats[host]
			if exist {
				if prev.os != stat.os || prev.arch != stat.arch {
					return &meta.ValidateErr{
						Type:   meta.TypeMismatch,
						Target: "platform",
						LHS:    fmt.Sprintf("%s:%s/%s", prev.cfg, prev.os, prev.arch),
						RHS:    fmt.Sprintf("%s:%s/%s", stat.cfg, stat.os, stat.arch),
						Value:  host,
					}
				}
			}
			platformStats[host] = stat
		}
	}
	return nil
}

func (s *Specification) portConflictsDetect() error {
	type (
		usedPort struct {
			host string
			port int
		}
		conflict struct {
			tp  string
			cfg string
		}
	)

	portTypes := []string{
		"Port",
		"StatusPort",
		"PeerPort",
		"ClientPort",
		"WebPort",
		"TCPPort",
		"HTTPPort",
		"ClusterPort",
		"MetricsPort",
	}

	portStats := map[usedPort]conflict{}
	uniqueHosts := set.NewStringSet()
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// skip nodes imported from TiDB-Ansible
			if compSpec.Addr().Interface().(InstanceSpec).IsImported() {
				continue
			}
			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := topoType.Field(i).Tag.Get("yaml")
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}
			uniqueHosts.Insert(host)

			// Ports conflicts
			for _, portType := range portTypes {
				if j, found := findField(compSpec, portType); found {
					item := usedPort{
						host: host,
						port: int(compSpec.Field(j).Int()),
					}
					tp := compSpec.Type().Field(j).Tag.Get("yaml")
					prev, exist := portStats[item]
					if exist {
						return &meta.ValidateErr{
							Type:   meta.TypeConflict,
							Target: "port",
							LHS:    fmt.Sprintf("%s:%s.%s", prev.cfg, item.host, prev.tp),
							RHS:    fmt.Sprintf("%s:%s.%s", cfg, item.host, tp),
							Value:  item.port,
						}
					}
					portStats[item] = conflict{
						tp:  tp,
						cfg: cfg,
					}
				}
			}
		}
	}

	// Port conflicts in monitored components
	monitoredPortTypes := []string{
		"NodeExporterPort",
	}
	monitoredOpt := topoSpec.FieldByName(monitorOptionTypeName)
	if monitoredOpt.IsZero() {
		return nil
	}
	monitoredOpt = monitoredOpt.Elem()
	for host := range uniqueHosts {
		cfg := "monitored"
		for _, portType := range monitoredPortTypes {
			f := monitoredOpt.FieldByName(portType)
			item := usedPort{
				host: host,
				port: int(f.Int()),
			}
			ft, found := monitoredOpt.Type().FieldByName(portType)
			if !found {
				return errors.Errorf("incompatible change `%s.%s`", monitorOptionTypeName, portType)
			}
			// `yaml:"node_exporter_port,omitempty"`
			tp := strings.Split(ft.Tag.Get("yaml"), ",")[0]
			prev, exist := portStats[item]
			if exist {
				return &meta.ValidateErr{
					Type:   meta.TypeConflict,
					Target: "port",
					LHS:    fmt.Sprintf("%s:%s.%s", prev.cfg, item.host, prev.tp),
					RHS:    fmt.Sprintf("%s:%s.%s", cfg, item.host, tp),
					Value:  item.port,
				}
			}
			portStats[item] = conflict{
				tp:  tp,
				cfg: cfg,
			}
		}
	}

	return nil
}

func (s *Specification) filebeatInstanceDetect() error {
	hosts := make(map[string]int, 0)
	for _, filebeat := range s.FilebeatServers {
		count := hosts[filebeat.Host]
		hosts[filebeat.Host] = count + 1
		if hosts[filebeat.Host] > 1 {
			return &meta.ValidateErr{
				Type:   meta.TypeConflict,
				Target: "instance",
				LHS:    fmt.Sprintf("%s:%d.%d", filebeat.Config, count, 1),
				RHS:    fmt.Sprintf("%s:%d.%d", filebeat.Config, count, 1),
				Value:  filebeat.Host,
			}
		}
	}
	return nil
}

func (s *Specification) dirConflictsDetect() error {
	type (
		usedDir struct {
			host string
			dir  string
		}
		conflict struct {
			tp  string
			cfg string
		}
	)

	dirTypes := []string{
		"DataDir",
		"DeployDir",
	}

	// usedInfo => type
	var (
		dirStats    = map[usedDir]conflict{}
		uniqueHosts = set.NewStringSet()
	)

	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// skip nodes imported from TiDB-Ansible
			if compSpec.Addr().Interface().(InstanceSpec).IsImported() {
				continue
			}
			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := topoType.Field(i).Tag.Get("yaml")
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}
			uniqueHosts.Insert(host)

			// Directory conflicts
			for _, dirType := range dirTypes {
				if j, found := findField(compSpec, dirType); found {
					item := usedDir{
						host: host,
						dir:  compSpec.Field(j).String(),
					}
					// data_dir is relative to deploy_dir by default, so they can be with
					// same (sub) paths as long as the deploy_dirs are different
					if item.dir != "" && !strings.HasPrefix(item.dir, "/") {
						continue
					}
					// `yaml:"data_dir,omitempty"`
					tp := strings.Split(compSpec.Type().Field(j).Tag.Get("yaml"), ",")[0]
					prev, exist := dirStats[item]
					if exist {
						return &meta.ValidateErr{
							Type:   meta.TypeConflict,
							Target: "directory",
							LHS:    fmt.Sprintf("%s:%s.%s", prev.cfg, item.host, prev.tp),
							RHS:    fmt.Sprintf("%s:%s.%s", cfg, item.host, tp),
							Value:  item.dir,
						}
					}
					dirStats[item] = conflict{
						tp:  tp,
						cfg: cfg,
					}
				}
			}
		}
	}

	return nil
}

// CountDir counts for dir paths used by any instance in the cluster with the same
// prefix, useful to find potential path conflicts
func (s *Specification) CountDir(targetHost, dirPrefix string) int {
	dirTypes := []string{
		"DataDir",
		"DeployDir",
		"LogDir",
	}

	// host-path -> count
	dirStats := make(map[string]int)
	count := 0
	topoSpec := reflect.ValueOf(s).Elem()
	dirPrefix = spec.Abs(s.GlobalOptions.User, dirPrefix)

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// Directory conflicts
			for _, dirType := range dirTypes {
				if j, found := findField(compSpec, dirType); found {
					dir := compSpec.Field(j).String()
					host := compSpec.FieldByName("Host").String()

					switch dirType { // the same as in logic.go for (*instance)
					case "DataDir":
						deployDir := compSpec.FieldByName("DeployDir").String()
						// the default data_dir is relative to deploy_dir
						if dir != "" && !strings.HasPrefix(dir, "/") {
							dir = filepath.Join(deployDir, dir)
						}
					case "LogDir":
						deployDir := compSpec.FieldByName("DeployDir").String()
						field := compSpec.FieldByName("LogDir")
						if field.IsValid() {
							dir = field.Interface().(string)
						}

						if dir == "" {
							dir = "log"
						}
						if !strings.HasPrefix(dir, "/") {
							dir = filepath.Join(deployDir, dir)
						}
					}
					dir = spec.Abs(s.GlobalOptions.User, dir)
					dirStats[host+dir]++
				}
			}
		}
	}

	for k, v := range dirStats {
		if k == targetHost+dirPrefix || strings.HasPrefix(k, targetHost+dirPrefix+"/") {
			count += v
		}
	}

	return count
}

// RegistryEndpoints return the list of registry endpoints in the specification
func (s *Specification) RegistryEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.ClusterServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.ClientPort))
	}
	return result
}

// APIServerEndpoints return the list of registry endpoints in the specification
func (s *Specification) APIServerEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.APIServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
	}
	return result
}

// FileServerEndpoints return the list of registry endpoints in the specification
func (s *Specification) FileServerEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.FileServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
	}
	return result
}

// AlertManagerEndpoints return the list of alert manager endpoints in the specification
func (s *Specification) AlertManagerEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.Alertmanagers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.WebPort))
	}
	return result
}

// MonitorEndpoints return the list of monitor prometheus endpoints in the specification
func (s *Specification) MonitorEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.Monitors {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
	}
	return result
}

// GrafanaEndpoints return the list of grafana endpoints in the specification
func (s *Specification) GrafanaEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.Grafanas {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
	}
	return result
}

// KibanaEndpoints return the list of kibana endpoints in the specification
func (s *Specification) KibanaEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.KibanaServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
	}
	return result
}

// TracerEndpoints return the list of tracer endpoints in the specification
func (s *Specification) TracerEndpoints() []string {
	result := make([]string, 0)
	for _, inst := range s.TracerServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.CompactThriftPort))
	}
	return result
}

// TracerWebAddress return the list of tracer web address in the specification
func (s *Specification) TracerWebAddress() []string {
	result := make([]string, 0)
	for _, inst := range s.TracerServers {
		result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.QueryHTTPServePort))
	}
	return result
}

// ElasticSearchEndpoints return the list of es endpoints in the specification
func (s *Specification) ElasticSearchEndpoints() []string {
	result := make([]string, 0)
	if s.GlobalOptions.ExternalElasticsearchUrl != "" {
		result = strings.Split(s.GlobalOptions.ExternalElasticsearchUrl, ",")
	}
	if len(s.ElasticSearchServers) > 0 {
		for _, inst := range s.ElasticSearchServers {
			result = append(result, fmt.Sprintf("%s:%d", inst.Host, inst.Port))
		}
	}
	return result
}

// HasEnableHttps return has enable https in the specification
func (s *Specification) HasEnableHttps() bool {
	for _, inst := range s.APIServers {
		return inst.EnableHttps == "true"
	}
	return false
}

// TiEMLogPaths return a list of paths to logs of TiEM components
func (s *Specification) TiEMLogPaths() map[string]*config.LogPathInfo {
	result := make(map[string]*config.LogPathInfo)

	for _, host := range s.APIServers {
		if _, ok := result[host.Host]; !ok {
			result[host.Host] = &config.LogPathInfo{
				GeneralLogs: set.NewStringSet(),
				AuditLogs:   set.NewStringSet(),
			}
		}
		result[host.Host].GeneralLogs.Insert(fmt.Sprintf("%s/logs/*-server.log", host.DataDir))
		result[host.Host].AuditLogs.Insert(fmt.Sprintf("%s/logs/audit.log", host.DataDir))
	}

	for _, host := range s.ClusterServers {
		if _, ok := result[host.Host]; !ok {
			result[host.Host] = &config.LogPathInfo{
				GeneralLogs: set.NewStringSet(),
				AuditLogs:   set.NewStringSet(),
			}
		}
		result[host.Host].GeneralLogs.Insert(fmt.Sprintf("%s/logs/*-server.log", host.DataDir))
		result[host.Host].AuditLogs.Insert(fmt.Sprintf("%s/logs/audit.log", host.DataDir))
	}

	for _, host := range s.FileServers {
		if _, ok := result[host.Host]; !ok {
			result[host.Host] = &config.LogPathInfo{
				GeneralLogs: set.NewStringSet(),
				AuditLogs:   set.NewStringSet(),
			}
		}
		result[host.Host].GeneralLogs.Insert(fmt.Sprintf("%s/logs/*-server.log", host.DataDir))
		result[host.Host].AuditLogs.Insert(fmt.Sprintf("%s/logs/audit.log", host.DataDir))
	}

	return result
}

// Validate validates the topology specification and produce error if the
// specification invalid (e.g: port conflicts or directory conflicts)
func (s *Specification) Validate() error {
	if err := s.platformConflictsDetect(); err != nil {
		return err
	}

	if err := s.portConflictsDetect(); err != nil {
		return err
	}

	if err := s.dirConflictsDetect(); err != nil {
		return err
	}

	if err := s.filebeatInstanceDetect(); err != nil {
		return err
	}

	return spec.RelativePathDetect(s, isSkipField)
}

// Type implements Topology interface.
func (s *Specification) Type() string {
	return TopoTypeTiEM
}

// BaseTopo implements Topology interface.
func (s *Specification) BaseTopo() *BaseTopo {
	return &BaseTopo{
		GlobalOptions:    &s.GlobalOptions,
		MonitoredOptions: s.GetMonitoredOptions(),
		Monitors:         s.Monitors,
		Grafanas:         s.Grafanas,
		Alertmanagers:    s.Alertmanagers,
		ElasticSearchs:   s.ElasticSearchServers,
		WebServers:       s.WebServers,
	}
}

// NewPart implements ScaleOutTopology interface.
func (s *Specification) NewPart() Topology {
	return &Specification{
		GlobalOptions:    s.GlobalOptions,
		MonitoredOptions: s.MonitoredOptions,
	}
}

// MergeTopo implements ScaleOutTopology interface.
func (s *Specification) MergeTopo(rhs Topology) Topology {
	other, ok := rhs.(*Specification)
	if !ok {
		panic("topo should be EM Topology")
	}

	return s.Merge(other)
}

// FillHostArch fills the topology with the given host->arch
func (s *Specification) FillHostArch(hostArch map[string]string) error {
	return spec.FillHostArch(s, hostArch)
}

// Merge returns a new Topology which sum old ones
func (s *Specification) Merge(that Topology) Topology {
	spec := that.(*Specification)
	return &Specification{
		GlobalOptions:        s.GlobalOptions,
		MonitoredOptions:     s.MonitoredOptions,
		ClusterServers:       append(s.ClusterServers, spec.ClusterServers...),
		APIServers:           append(s.APIServers, spec.APIServers...),
		FileServers:          append(s.FileServers, spec.FileServers...),
		WebServers:           append(s.WebServers, spec.WebServers...),
		TracerServers:        append(s.TracerServers, spec.TracerServers...),
		ElasticSearchServers: append(s.ElasticSearchServers, spec.ElasticSearchServers...),
		KibanaServers:        append(s.KibanaServers, spec.KibanaServers...),
		FilebeatServers:      append(s.FilebeatServers, spec.FilebeatServers...),
		Monitors:             append(s.Monitors, spec.Monitors...),
		Grafanas:             append(s.Grafanas, spec.Grafanas...),
		Alertmanagers:        append(s.Alertmanagers, spec.Alertmanagers...),
	}
}

// fillDefaults tries to fill custom fields to their default values
func fillTiEMCustomDefaults(globalOptions *GlobalOptions, data interface{}) error {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var err error
	for i := 0; i < t.NumField(); i++ {
		if err = setTiEMCustomDefaults(globalOptions, v.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func setTiEMCustomDefaults(globalOptions *GlobalOptions, field reflect.Value) error {
	if !field.CanSet() || isSkipField(field) {
		return nil
	}

	switch field.Kind() {
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			if err := setTiEMCustomDefaults(globalOptions, field.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		ref := reflect.New(field.Type())
		ref.Elem().Set(field)
		if err := fillTiEMCustomDefaults(globalOptions, ref.Interface()); err != nil {
			return err
		}
		field.Set(ref.Elem())
	case reflect.Ptr:
		if err := setTiEMCustomDefaults(globalOptions, field.Elem()); err != nil {
			return err
		}
	}

	if field.Kind() != reflect.Struct {
		return nil
	}

	for j := 0; j < field.NumField(); j++ {
		switch field.Type().Field(j).Name {
		case "SSHPort":
			if field.Field(j).Int() != 0 {
				continue
			}
			field.Field(j).Set(reflect.ValueOf(globalOptions.SSHPort))
		case "Name":
			if field.Field(j).String() != "" {
				continue
			}
			host := field.FieldByName("Host").String()
			port := field.FieldByName("Port").Int()
			field.Field(j).Set(reflect.ValueOf(fmt.Sprintf("tiem-%s-%d", host, port)))
		case "DataDir":
			dataDir := field.Field(j).String()
			if dataDir != "" { // already have a value, skip filling default values
				continue
			}
			// If the data dir in global options is an obsolute path, it appends to
			// the global and has a comp-port sub directory
			if strings.HasPrefix(globalOptions.DataDir, "/") {
				field.Field(j).Set(reflect.ValueOf(filepath.Join(
					globalOptions.DataDir,
					fmt.Sprintf("%s-%s", field.Addr().Interface().(InstanceSpec).Role(), getPort(field)),
				)))
				continue
			}

			// If the data dir in global options is empty or a relative path, keep it be relative
			// Our run_*.sh start scripts are run inside deploy_path, so the final location
			// will be deploy_path/global.data_dir
			// (the default value of global.data_dir is "data")
			if globalOptions.DataDir == "" {
				field.Field(j).Set(reflect.ValueOf("data"))
			} else {
				field.Field(j).Set(reflect.ValueOf(globalOptions.DataDir))
			}
		case "DeployDir":
			setDefaultDir(globalOptions.DeployDir, field.Addr().Interface().(InstanceSpec).Role(), getPort(field), field.Field(j))
		case "LogDir":
			if field.Field(j).String() == "" && defaults.CanUpdate(field.Field(j).Interface()) {
				field.Field(j).Set(reflect.ValueOf(globalOptions.LogDir))
			}
		case "LogLevel":
			if field.Field(j).String() == "" && defaults.CanUpdate(field.Field(j).Interface()) {
				field.Field(j).Set(reflect.ValueOf(globalOptions.LogLevel))
			}
		case "ExternalElasticsearchUrl":
			if field.Field(j).String() == "" && defaults.CanUpdate(field.Field(j).Interface()) {
				field.Field(j).Set(reflect.ValueOf(globalOptions.ExternalElasticsearchUrl))
			}
		case "Arch":
			// default values of globalOptions are set before fillCustomDefaults in Unmarshal
			// so the globalOptions.Arch already has its default value set, no need to check again
			if field.Field(j).String() == "" {
				field.Field(j).Set(reflect.ValueOf(globalOptions.Arch))
			}

			switch strings.ToLower(field.Field(j).String()) {
			// replace "x86_64" with amd64, they are the same in our repo
			case "x86_64":
				field.Field(j).Set(reflect.ValueOf("amd64"))
			// replace "aarch64" with arm64
			case "aarch64":
				field.Field(j).Set(reflect.ValueOf("arm64"))
			}

			// convert to lower case
			if field.Field(j).String() != "" {
				field.Field(j).Set(reflect.ValueOf(strings.ToLower(field.Field(j).String())))
			}
		case "OS":
			// default value of globalOptions.OS is already set, same as "Arch"
			if field.Field(j).String() == "" {
				field.Field(j).Set(reflect.ValueOf(globalOptions.OS))
			}
			// convert to lower case
			if field.Field(j).String() != "" {
				field.Field(j).Set(reflect.ValueOf(strings.ToLower(field.Field(j).String())))
			}
		}
	}

	return nil
}

func getPort(v reflect.Value) string {
	for i := 0; i < v.NumField(); i++ {
		switch v.Type().Field(i).Name {
		// all field names that could be a listening port of some process
		case "Port",
			"ClientPort",
			"WebPort",
			"TCPPort",
			"NodeExporterPort",
			"ZipkinThriftPort",
			"CompactThriftPort",
			"BinaryThriftPort",
			"AgentServePort",
			"QueryServePort",
			"JaegerThriftPort",
			"JaegerHostPort",
			"CollectorPort",
			"ZipkinHostPort":
			return fmt.Sprintf("%d", v.Field(i).Int())
		}
	}
	return "0"
}