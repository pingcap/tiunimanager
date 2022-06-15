// Copyright 2021 PingCAP
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

package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiunimanager/tiup/embed"
)

// PrometheusConfig represent the data to generate Prometheus config
type PrometheusConfig struct {
	ClusterName                string
	NodeExporterAddrs          []string
	MonitoredServers           []string
	AlertmanagerAddrs          []string
	BlackboxExporterAddrs      []string
	TiUniManagerAPIServers     []string
	TiUniManagerMetaDBServers  []string
	TiUniManagerClusterServers []string
	TiUniManagerFileServers    []string
}

// NewPrometheusConfig returns a PrometheusConfig
func NewPrometheusConfig(clusterName string) *PrometheusConfig {
	return &PrometheusConfig{
		ClusterName: clusterName,
	}
}

// AddMetaDB add an alertmanager address
func (c *PrometheusConfig) AddMetaDB(ip string, port uint64) *PrometheusConfig {
	c.TiUniManagerMetaDBServers = append(c.TiUniManagerMetaDBServers, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddAPIServer add an alertmanager address
func (c *PrometheusConfig) AddAPIServer(ip string, port uint64) *PrometheusConfig {
	c.TiUniManagerAPIServers = append(c.TiUniManagerAPIServers, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddFileServer add an alertmanager address
func (c *PrometheusConfig) AddFileServer(ip string, port uint64) *PrometheusConfig {
	c.TiUniManagerFileServers = append(c.TiUniManagerFileServers, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddClusterServer add an alertmanager address
func (c *PrometheusConfig) AddClusterServer(ip string, port uint64) *PrometheusConfig {
	c.TiUniManagerClusterServers = append(c.TiUniManagerClusterServers, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddNodeExpoertor add a node expoter address
func (c *PrometheusConfig) AddNodeExpoertor(ip string, port uint64) *PrometheusConfig {
	c.NodeExporterAddrs = append(c.NodeExporterAddrs, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddBlackboxExporter add a BlackboxExporter address
func (c *PrometheusConfig) AddBlackboxExporter(ip string, port uint64) *PrometheusConfig {
	c.BlackboxExporterAddrs = append(c.BlackboxExporterAddrs, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// AddAlertmanager add an alertmanager address
func (c *PrometheusConfig) AddAlertmanager(ip string, port uint64) *PrometheusConfig {
	c.AlertmanagerAddrs = append(c.AlertmanagerAddrs, fmt.Sprintf("%s:%d", ip, port))
	return c
}

// Config generate the config file data.
func (c *PrometheusConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "prometheus.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Prometheus config content by tpl
func (c *PrometheusConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Prometheus").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}

// ConfigToFile write config content to specific path
func (c *PrometheusConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
