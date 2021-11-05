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
	"github.com/pingcap/tiup/pkg/set"
)

// LogPathInfo is a list of TiEM log files
type LogPathInfo struct {
	GeneralLogs set.StringSet
	AuditLogs   set.StringSet
}

// FilebeatConfig represent the data to generate AlertManager config
type FilebeatConfig struct {
	Host              string
	DeployDir         string
	DataDir           string
	LogDir            string
	ElasticSearchHost string
	GeneralLogs       []string
	AuditLogs         []string
	HasAllLogs        bool
	HasGeneralLogs    bool
	HasAuditLogs      bool
}

// NewFilebeatConfig returns a FilebeatConfig
func NewFilebeatConfig(host, deployDir, dataDir, logDir string) *FilebeatConfig {
	return &FilebeatConfig{
		Host:      host,
		DeployDir: deployDir,
		DataDir:   dataDir,
		LogDir:    logDir,
	}
}

// WithElasticSearch sets es host
func (c *FilebeatConfig) WithElasticSearch(esAddress []string) *FilebeatConfig {
	c.ElasticSearchHost = strings.Join(esAddress, ",")
	return c
}

// WithTiEMLogs set log paths
func (c *FilebeatConfig) WithTiEMLogs(paths map[string]*LogPathInfo) *FilebeatConfig {
	if p, ok := paths[c.Host]; ok {
		c.GeneralLogs = append(c.GeneralLogs, p.GeneralLogs.Slice()...)
		c.AuditLogs = append(c.AuditLogs, p.AuditLogs.Slice()...)
	}
	if len(c.GeneralLogs) <= 0 {
		c.HasGeneralLogs = false
	} else {
		c.HasGeneralLogs = true
	}
	if len(c.AuditLogs) <= 0 {
		c.HasAuditLogs = false
	} else {
		c.HasAuditLogs = true
	}
	if len(c.GeneralLogs) <= 0 && len(c.AuditLogs) <= 0 {
		c.HasAllLogs = false
	} else {
		c.HasAllLogs = true
	}
	return c
}

// Config generate the config file data.
func (c *FilebeatConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "configs", "filebeat.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigToFile write config content to specific path
func (c *FilebeatConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ConfigWithTemplate generate the AlertManager config content by tpl
func (c *FilebeatConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Filebeat").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
