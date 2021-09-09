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

package scripts

import (
	"bytes"
	"os"
	"path"
	"text/template"

	"github.com/pingcap-inc/tiem/tiup/templates/embed"
)

// TiEMClusterServerScript represent the data to generate TiEMClusterServer config
type TiEMClusterServerScript struct {
	IP                string
	Port              int
	MetricsPort       int
	DeployDir         string
	DataDir           string
	LogDir            string
	LogLevel          string
	RegistryEndpoints string
	TracerAddress     string
}

// NewTiEMClusterServerScript returns a TiEMClusterServerScript with given arguments
func NewTiEMClusterServerScript(ip, deployDir, dataDir, logDir, logLevel string) *TiEMClusterServerScript {
	return &TiEMClusterServerScript{
		IP:          ip,
		Port:        4110,
		MetricsPort: 4121,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiEMClusterServerScript
func (c *TiEMClusterServerScript) WithPort(port int) *TiEMClusterServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiEMClusterServerScript
func (c *TiEMClusterServerScript) WithMetricsPort(port int) *TiEMClusterServerScript {
	c.MetricsPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiEMClusterServerScript) WithRegistry(addr string) *TiEMClusterServerScript {
	c.RegistryEndpoints = addr
	return c
}

// WithTracer set TracerAddress
func (c *TiEMClusterServerScript) WithTracer(addr string) *TiEMClusterServerScript {
	c.TracerAddress = addr
	return c
}

// Script generate the config file data.
func (c *TiEMClusterServerScript) Script() ([]byte, error) {
	fp := path.Join("scripts", "run_tiem_cluster.sh.tpl")
	tpl, err := embed.ReadScriptTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiEMClusterServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiEMClusterServer config content by tpl
func (c *TiEMClusterServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiEMClusterServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
