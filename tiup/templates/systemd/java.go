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

package system

import (
	"bytes"
	"os"
	"path"
	"text/template"

	"github.com/pingcap/tiunimanager/tiup/embed"
)

// JavaAppConfig is template for Java apps
type JavaAppConfig struct {
	ServiceName string
	User        string
	DeployDir   string
	JavaHome    string
	// Takes one of no, on-success, on-failure, on-abnormal, on-watchdog, on-abort, or always.
	// The Template set as always if this is not setted.
	Restart string
	Type    string // service type
}

// NewJavaAppConfig returns a Config with given arguments
func NewJavaAppConfig(service, user, deployDir, javaHome string) *JavaAppConfig {
	return &JavaAppConfig{
		ServiceName: service,
		User:        user,
		DeployDir:   deployDir,
		JavaHome:    javaHome,
	}
}

// WithType set the systemd service type
func (c *JavaAppConfig) WithType(typ string) *JavaAppConfig {
	c.Type = typ
	return c
}

// ConfigToFile write config content to specific path
func (c *JavaAppConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// Config generate the config file data.
func (c *JavaAppConfig) Config() ([]byte, error) {
	fp := path.Join("templates", "systemd", "java_app.service.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the system config content by tpl
func (c *JavaAppConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("system").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
