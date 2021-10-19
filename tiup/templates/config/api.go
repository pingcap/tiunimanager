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
	"strings"
	"text/template"

	"github.com/pingcap-inc/tiem/tiup/embed"
	"github.com/pingcap/errors"
)

// APIServerConfig represent the data to generate api server config
type APIServerConfig struct {
	PrometheusAddress    string
	GrafanaAddress       string
	AlertManagerAddress  string
	KibanaAddress        string
	JaegerAddress        string
	ElasticSearchAddress string
}

// NewAPIServerConfig returns a APIServerConfig
func NewAPIServerConfig() *APIServerConfig {
	return &APIServerConfig{}
}

// WithPrometheusAddress sets PrometheusAddress for APIServerConfig
func (c *APIServerConfig) WithPrometheusAddress(prometheusAddress []string) *APIServerConfig {
	c.PrometheusAddress = strings.Join(prometheusAddress, ",")
	return c
}

// WithGrafanaAddress sets GrafanaAddress name
func (c *APIServerConfig) WithGrafanaAddress(grafanaAddress []string) *APIServerConfig {
	c.GrafanaAddress = strings.Join(grafanaAddress, ",")
	return c
}

// WithAlertManagerAddress sets AlertManagerAddress name
func (c *APIServerConfig) WithAlertManagerAddress(alertManagerAddress []string) *APIServerConfig {
	c.AlertManagerAddress = strings.Join(alertManagerAddress, ",")
	return c
}

// WithKibanaAddress sets KibanaAddress name
func (c *APIServerConfig) WithKibanaAddress(kibanaAddress []string) *APIServerConfig {
	c.KibanaAddress = strings.Join(kibanaAddress, ",")
	return c
}

// WithJaegerAddress sets JaegerAddress name
func (c *APIServerConfig) WithJaegerAddress(jaegerAddress []string) *APIServerConfig {
	c.JaegerAddress = strings.Join(jaegerAddress, ",")
	return c
}

// WithElasticsearchAddress sets ElasticSearchAddresses endpoints
func (c *APIServerConfig) WithElasticsearchAddress(addrs []string) *APIServerConfig {
	c.ElasticSearchAddress = strings.Join(addrs, ",")
	return c
}

// Config generate the config file data.
func (c *APIServerConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "env.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the APIServer config content by tpl
func (c *APIServerConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("openapi").Parse(tpl)
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
func (c *APIServerConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
