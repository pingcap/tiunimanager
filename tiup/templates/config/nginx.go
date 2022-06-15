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

// NginxConfig represent the data to generate Nginx config
type NginxConfig struct {
	IP                  string
	Port                int
	TlsPort             int
	ServerName          string
	DeployDir           string
	LogDir              string
	RegistryEndpoints   []string
	EnableHttps         bool
	Protocol            string
	GrafanaAddress      string
	KibanaAddress       string
	AlertManagerAddress string
	TracerAddress       string
}

// NewNginxConfig returns a NginxConfig
func NewNginxConfig(host, deployDir, logDir string) *NginxConfig {
	return &NginxConfig{
		IP:        host,
		Port:      4180,
		TlsPort:   4181,
		DeployDir: deployDir,
		LogDir:    logDir,
	}
}

// WithPort sets Port for NginxConfig
func (c *NginxConfig) WithPort(port int) *NginxConfig {
	c.Port = port
	return c
}

// WithTlsPort sets TlsPort for NginxConfig
func (c *NginxConfig) WithTlsPort(tlsPort int) *NginxConfig {
	c.TlsPort = tlsPort
	if !c.EnableHttps {
		c.TlsPort = 0
	}
	return c
}

// WithServerName sets ServerName for NginxConfig
func (c *NginxConfig) WithServerName(name string) *NginxConfig {
	c.ServerName = name
	return c
}

// WithRegistryEndpoints sets RegistryEndpoints for NginxConfig
func (c *NginxConfig) WithRegistryEndpoints(endpoints []string) *NginxConfig {
	c.RegistryEndpoints = endpoints
	return c
}

// WithEnableHttps sets EnableHttps for NginxConfig
func (c *NginxConfig) WithEnableHttps(enableHttps bool) *NginxConfig {
	c.EnableHttps = enableHttps
	if enableHttps {
		c.Protocol = "https"
	} else {
		c.Protocol = "http"
	}
	return c
}

// WithGrafanaAddress sets GrafanaAddress for NginxConfig
func (c *NginxConfig) WithGrafanaAddress(grafanaAddress []string) *NginxConfig {
	c.GrafanaAddress = strings.Join(grafanaAddress, ",")
	return c
}

// WithKibanaAddress sets AlertManagerAddress for NginxConfig
func (c *NginxConfig) WithKibanaAddress(kibanaAddress []string) *NginxConfig {
	c.KibanaAddress = strings.Join(kibanaAddress, ",")
	return c
}

// WithAlertManagerAddress sets AlertManagerAddress for NginxConfig
func (c *NginxConfig) WithAlertManagerAddress(alertManagerAddress []string) *NginxConfig {
	c.AlertManagerAddress = strings.Join(alertManagerAddress, ",")
	return c
}

// WithTracerAddress sets TracerAddress for NginxConfig
func (c *NginxConfig) WithTracerAddress(tracerAddress []string) *NginxConfig {
	c.TracerAddress = strings.Join(tracerAddress, ",")
	return c
}

// Config generate the config file data.
func (c *NginxConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "nginx.conf.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Nginx config content by tpl
func (c *NginxConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Nginx").Parse(tpl)
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
func (c *NginxConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}

// NginxServerList represent the data to generate Nginx config
type NginxServerList struct {
	APIServers []string
}

// NewNginxServerList returns a NginxConfig
func NewNginxServerList(apis []string) *NginxServerList {
	return &NginxServerList{
		APIServers: apis,
	}
}

// Config generate the config file data.
func (c *NginxServerList) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "nginx_server_list.conf.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Nginx config content by tpl
func (c *NginxServerList) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("NginxServerList").Parse(tpl)
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
func (c *NginxServerList) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, config, 0755); err != nil {
		return errors.AddStack(err)
	}
	return nil
}
