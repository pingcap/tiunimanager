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
	"os"
	"path"
	"text/template"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiunimanager/tiup/embed"
)

// ElasticSearchConfig represent the data to generate ElasticSearch config
type ElasticSearchConfig struct {
	IP        string
	Port      int
	Name      string
	DeployDir string
	DataDir   string
	LogDir    string
}

// NewElasticSearchConfig returns a ElasticSearchConfig
func NewElasticSearchConfig(host, deployDir, dataDir, logDir string) *ElasticSearchConfig {
	return &ElasticSearchConfig{
		IP:        host,
		Port:      4108,
		DeployDir: deployDir,
		DataDir:   dataDir,
		LogDir:    logDir,
	}
}

// WithPort sets Port for ElasticSearchConfig
func (c *ElasticSearchConfig) WithPort(port int) *ElasticSearchConfig {
	c.Port = port
	return c
}

// WithName sets cluster name
func (c *ElasticSearchConfig) WithName(name string) *ElasticSearchConfig {
	c.Name = name
	return c
}

// Config generate the config file data.
func (c *ElasticSearchConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "elasticsearch.yml")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the ElasticSearch config content by tpl
func (c *ElasticSearchConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("ElasticSearch").Parse(tpl)
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
func (c *ElasticSearchConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
