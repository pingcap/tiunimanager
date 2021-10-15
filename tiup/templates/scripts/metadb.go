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
	"strings"
	"text/template"

	"github.com/pingcap-inc/tiem/tiup/embed"
)

// TiEMMetaDBScript represent the data to generate TiEMMetaDB config
type TiEMMetaDBScript struct {
	Host              string
	Port              int
	ClientPort        int
	PeerPort          int
	MetricsPort       int
	DeployDir         string
	DataDir           string
	LogDir            string
	LogLevel          string
	RegistryEndpoints string
	TracerAddress     string
}

// NewTiEMMetaDBScript returns a TiEMMetaDBScript with given arguments
func NewTiEMMetaDBScript(ip, deployDir, dataDir, logDir, logLevel string) *TiEMMetaDBScript {
	return &TiEMMetaDBScript{
		Host:        ip,
		Port:        4100,
		ClientPort:  4101,
		PeerPort:    4102,
		MetricsPort: 4121,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiEMMetaDBScript
func (c *TiEMMetaDBScript) WithPort(port int) *TiEMMetaDBScript {
	c.Port = port
	return c
}

// WithClientPort set ClientPort field of TiEMMetaDBScript
func (c *TiEMMetaDBScript) WithClientPort(port int) *TiEMMetaDBScript {
	c.ClientPort = port
	return c
}

// WithPeerPort set PeerPort field of TiEMMetaDBScript
func (c *TiEMMetaDBScript) WithPeerPort(port int) *TiEMMetaDBScript {
	c.PeerPort = port
	return c
}

// WithMetricsPort set PeerPort field of TiEMMetaDBScript
func (c *TiEMMetaDBScript) WithMetricsPort(port int) *TiEMMetaDBScript {
	c.MetricsPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiEMMetaDBScript) WithRegistry(addr []string) *TiEMMetaDBScript {
	c.RegistryEndpoints = strings.Join(addr, ",")
	return c
}

// WithTracer set TracerAddress
func (c *TiEMMetaDBScript) WithTracer(addr []string) *TiEMMetaDBScript {
	c.TracerAddress = strings.Join(addr, ",")
	return c
}

// Script generate the config file data.
func (c *TiEMMetaDBScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_tiem_metadb.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiEMMetaDBScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiEMMetaDB config content by tpl
func (c *TiEMMetaDBScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiEMMetaDB").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
