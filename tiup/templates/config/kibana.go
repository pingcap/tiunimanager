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

package config

import (
	"bytes"
	"os"
	"path"
	"text/template"

	"github.com/pingcap-inc/tiem/tiup/embed"
	"github.com/pingcap/errors"
)

// KibanaConfig represent the data to generate Kibana config
type KibanaConfig struct {
	IP                     string
	Port                   int
	Name                   string
	DeployDir              string
	DataDir                string
	LogDir                 string
	ElasticSearchAddresses []string
}

// NewKibanaConfig returns a KibanaConfig
func NewKibanaConfig(host, deployDir, dataDir, logDir string) *KibanaConfig {
	return &KibanaConfig{
		IP:        host,
		Port:      4127,
		DeployDir: deployDir,
		DataDir:   dataDir,
		LogDir:    logDir,
	}
}

// WithPort sets Port for KibanaConfig
func (c *KibanaConfig) WithPort(port int) *KibanaConfig {
	c.Port = port
	return c
}

// WithName sets cluster name
func (c *KibanaConfig) WithName(name string) *KibanaConfig {
	c.Name = name
	return c
}

// WithElasticSearch sets es endpoints
func (c *KibanaConfig) WithElasticsearch(addrs []string) *KibanaConfig {
	c.ElasticSearchAddresses = addrs
	return c
}

// Config generate the config file data.
func (c *KibanaConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "kibana.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Kibana config content by tpl
func (c *KibanaConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Kibana").Parse(tpl)
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
func (c *KibanaConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
