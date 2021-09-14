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
	"os"
	"path"

	"github.com/pingcap-inc/tiem/tiup/templates/embed"
	"github.com/pingcap/errors"
)

// NginxConfig represent the data to generate Nginx config
type NginxConfig struct {
	IP                string
	Port              int
	ServerName        string
	DeployDir         string
	LogDir            string
	RegistryEndpoints []string
}

// NewNginxConfig returns a NginxConfig
func NewNginxConfig(host, deployDir, logDir string) *NginxConfig {
	return &NginxConfig{
		IP:        host,
		Port:      4120,
		DeployDir: deployDir,
		LogDir:    logDir,
	}
}

// WithPort sets Port for NginxConfig
func (c *NginxConfig) WithPort(port int) *NginxConfig {
	c.Port = port
	return c
}

// WithServerName sets ServerName for NginxConfig
func (c *NginxConfig) WithServerName(name string) *NginxConfig {
	c.ServerName = name
	return c
}

// WithRegistryEndpoints sets registry endpoints for NginxConfig
func (c *NginxConfig) WithRegistryEndpoints(endpoints []string) *NginxConfig {
	c.RegistryEndpoints = endpoints
	return c
}

// Config generate the config file data.
func (c *NginxConfig) Config() ([]byte, error) {
	fp := path.Join("configs", "nginx.conf.tpl")
	tpl, err := embed.ReadConfigTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Nginx config content by tpl
func (c *NginxConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	return []byte(tpl), nil
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
	fp := path.Join("configs", "nginx_server_list.conf.tpl")
	tpl, err := embed.ReadConfigTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Nginx config content by tpl
func (c *NginxServerList) ConfigWithTemplate(tpl string) ([]byte, error) {
	return []byte(tpl), nil
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
