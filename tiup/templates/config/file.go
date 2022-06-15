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
	"strings"
	"text/template"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiunimanager/tiup/embed"
)

// FileServerConfig represent the data to generate api server config
type FileServerConfig struct {
	PrometheusAddress    string
	GrafanaAddress       string
	AlertManagerAddress  string
	KibanaAddress        string
	JaegerAddress        string
	ElasticSearchAddress string
}

// NewFileServerConfig returns a FileServerConfig
func NewFileServerConfig() *FileServerConfig {
	return &FileServerConfig{}
}

// WithPrometheusAddress sets PrometheusAddress for APIServerConfig
func (c *FileServerConfig) WithPrometheusAddress(prometheusAddress []string) *FileServerConfig {
	c.PrometheusAddress = strings.Join(prometheusAddress, ",")
	return c
}

// WithGrafanaAddress sets GrafanaAddress name
func (c *FileServerConfig) WithGrafanaAddress(grafanaAddress []string) *FileServerConfig {
	c.GrafanaAddress = strings.Join(grafanaAddress, ",")
	return c
}

// WithAlertManagerAddress sets AlertManagerAddress name
func (c *FileServerConfig) WithAlertManagerAddress(alertManagerAddress []string) *FileServerConfig {
	c.AlertManagerAddress = strings.Join(alertManagerAddress, ",")
	return c
}

// WithKibanaAddress sets KibanaAddress name
func (c *FileServerConfig) WithKibanaAddress(kibanaAddress []string) *FileServerConfig {
	c.KibanaAddress = strings.Join(kibanaAddress, ",")
	return c
}

// WithJaegerAddress sets JaegerAddress name
func (c *FileServerConfig) WithJaegerAddress(jaegerAddress []string) *FileServerConfig {
	c.JaegerAddress = strings.Join(jaegerAddress, ",")
	return c
}

// WithElasticsearchAddress sets ElasticSearchAddresses endpoints
func (c *FileServerConfig) WithElasticsearchAddress(addrs []string) *FileServerConfig {
	c.ElasticSearchAddress = strings.Join(addrs, ",")
	return c
}

// Config generate the config file data.
func (c *FileServerConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "file.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the APIServer config content by tpl
func (c *FileServerConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("File").Parse(tpl)
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
func (c *FileServerConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
