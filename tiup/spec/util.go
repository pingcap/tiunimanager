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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/pingcap-inc/tiem/tiup/templates/scripts"
	"github.com/pingcap-inc/tiem/tiup/version"
	"github.com/pingcap/tiup/pkg/utils"
	"github.com/prometheus/common/expfmt"
	"go.etcd.io/etcd/client/pkg/v3/transport"
)

// AuditDir return the directory for saving audit log.
func AuditDir() string {
	return filepath.Join(profileDir, TiUPAuditDir)
}

// SaveMetadata saves the cluster meta information to profile directory
func SaveMetadata(clusterName string, cmeta *TiEMMeta) error {
	// set the cmd version
	cmeta.OpsVer = version.NewComponentVersion().String()
	return GetSpecManager().SaveMeta(clusterName, cmeta)
}

// TiEMMetadata tries to read the metadata of a cluster from file
func TiEMMetadata(clusterName string) (*Metadata, error) {
	var cm Metadata
	err := GetSpecManager().Metadata(clusterName, &cm)
	if err != nil {
		// Return the value of cm even on error, to make sure the caller can get the data
		// we read, if there's any.
		// This is necessary when, either by manual editing of meta.yaml file, by not fully
		// validated `edit-config`, or by some unexpected operations from a broken legacy
		// release, we could provide max possibility that operations like `display`, `scale`
		// and `destroy` are still (more or less) working, by ignoring certain errors.
		return &cm, err
	}

	return &cm, nil
}

// LoadClientCert read and load the client cert key pair and CA cert
func LoadClientCert(dir string) (*tls.Config, error) {
	return transport.TLSInfo{
		TrustedCAFile: filepath.Join(dir, TLSCACert),
		CertFile:      filepath.Join(dir, TLSClientCert),
		KeyFile:       filepath.Join(dir, TLSClientKey),
	}.ClientConfig()
}

// statusByHost queries current status of the instance by http status api.
func statusByHost(host string, port int, path string, tlsCfg *tls.Config) string {
	client := utils.NewHTTPClient(statusQueryTimeout, tlsCfg)

	scheme := "http"
	if tlsCfg != nil {
		scheme = "https"
	}
	if path == "" {
		path = "/"
	}
	url := fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)

	// body doesn't have any status section needed
	body, err := client.Get(context.TODO(), url)
	if err != nil || body == nil {
		return "Down"
	}
	return "Up"
}

// UptimeByHost queries current uptime of the instance by http Prometheus metric api.
func UptimeByHost(host string, port int, tlsCfg *tls.Config) time.Duration {
	scheme := "http"
	if tlsCfg != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d/metrics", scheme, host, port)

	client := utils.NewHTTPClient(statusQueryTimeout, tlsCfg)

	body, err := client.Get(context.TODO(), url)
	if err != nil || body == nil {
		return 0
	}

	var parser expfmt.TextParser
	reader := bytes.NewReader(body)
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return 0
	}

	now := time.Now()
	for k, v := range mf {
		if k == promMetricStartTimeSeconds {
			ms := v.GetMetric()
			if len(ms) >= 1 {
				startTime := ms[0].Gauge.GetValue()
				return now.Sub(time.Unix(int64(startTime), 0))
			}
			return 0
		}
	}

	return 0
}

// Abs returns the absolute path
func Abs(user, path string) string {
	// trim whitespaces before joining
	user = strings.TrimSpace(user)
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = filepath.Join("/home", user, path)
	}
	return filepath.Clean(path)
}

// MultiDirAbs returns the absolute path for multi-dir separated by comma
func MultiDirAbs(user, paths string) []string {
	var dirs []string
	for _, path := range strings.Split(paths, ",") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		dirs = append(dirs, Abs(user, path))
	}
	return dirs
}

// PackagePath return the tar bar path
func PackagePath(comp string, version string, os string, arch string) string {
	fileName := fmt.Sprintf("%s-%s-%s-%s.tar.gz", comp, version, os, arch)
	return ProfilePath(TiUPPackageCacheDir, fileName)
}

// AlertManagerEndpoints returns the AlertManager endpoints configurations
func AlertManagerEndpoints(alertmanager []*AlertmanagerSpec, user string, enableTLS bool) []*scripts.AlertManagerScript {
	var ends []*scripts.AlertManagerScript
	for _, spec := range alertmanager {
		deployDir := Abs(user, spec.DeployDir)
		// data dir would be empty for components which don't need it
		dataDir := spec.DataDir
		// the default data_dir is relative to deploy_dir
		if dataDir != "" && !strings.HasPrefix(dataDir, "/") {
			dataDir = filepath.Join(deployDir, dataDir)
		}
		// log dir will always be with values, but might not used by the component
		logDir := Abs(user, spec.LogDir)

		script := scripts.NewAlertManagerScript(
			spec.Host,
			deployDir,
			dataDir,
			logDir,
		).
			WithWebPort(spec.WebPort).
			WithClusterPort(spec.ClusterPort)
		ends = append(ends, script)
	}
	return ends
}

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

func findSliceField(v Topology, fieldName string) (reflect.Value, bool) {
	topo := reflect.ValueOf(v)
	if topo.Kind() == reflect.Ptr {
		topo = topo.Elem()
	}

	j, found := findField(topo, fieldName)
	if found {
		val := topo.Field(j)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			return val, true
		}
	}
	return reflect.Value{}, false
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
