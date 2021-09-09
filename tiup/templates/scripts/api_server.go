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

// TiEMAPIServerScript represent the data to generate TiEMAPIServer config
type TiEMAPIServerScript struct {
	IP                   string
	Port                 int
	MetricsPort          int
	DeployDir            string
	DataDir              string
	LogDir               string
	LogLevel             string
	RegistryEndpoints    string
	TracerAddress        string
	ElasticSearchAddress string
}

// NewTiEMAPIServerScript returns a TiEMAPIServerScript with given arguments
func NewTiEMAPIServerScript(ip, deployDir, dataDir, logDir, logLevel string) *TiEMAPIServerScript {
	return &TiEMAPIServerScript{
		IP:          ip,
		Port:        4116,
		MetricsPort: 4121,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiEMAPIServerScript
func (c *TiEMAPIServerScript) WithPort(port int) *TiEMAPIServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiEMAPIServerScript
func (c *TiEMAPIServerScript) WithMetricsPort(port int) *TiEMAPIServerScript {
	c.MetricsPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiEMAPIServerScript) WithRegistry(addr string) *TiEMAPIServerScript {
	c.RegistryEndpoints = addr
	return c
}

// WithTracer set TracerAddress
func (c *TiEMAPIServerScript) WithTracer(addr string) *TiEMAPIServerScript {
	c.TracerAddress = addr
	return c
}

// WithElasticSearch set ElasticSearchAddress
func (c *TiEMAPIServerScript) WithElasticSearch(addr string) *TiEMAPIServerScript {
	c.ElasticSearchAddress = addr
	return c
}

// Script generate the config file data.
func (c *TiEMAPIServerScript) Script() ([]byte, error) {
	fp := path.Join("scripts", "run_tiem_metadb.sh.tpl")
	tpl, err := embed.ReadScriptTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiEMAPIServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiEMAPIServer config content by tpl
func (c *TiEMAPIServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiEMAPIServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
