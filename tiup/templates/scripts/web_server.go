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

// TiEMWebServerScript represent the data to generate TiEMWebServer config
type TiEMWebServerScript struct {
	IP          string
	Port        int
	MetricsPort int
	DeployDir   string
	DataDir     string
	LogDir      string
}

// NewTiEMWebServerScript returns a TiEMWebServerScript with given arguments
func NewTiEMWebServerScript(ip, deployDir, dataDir, logDir string) *TiEMWebServerScript {
	return &TiEMWebServerScript{
		IP:          ip,
		Port:        4120,
		MetricsPort: 4121,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
	}
}

// WithPort set Port field of TiEMWebServerScript
func (c *TiEMWebServerScript) WithPort(port int) *TiEMWebServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiEMWebServerScript
func (c *TiEMWebServerScript) WithMetricsPort(port int) *TiEMWebServerScript {
	c.MetricsPort = port
	return c
}

// Script generate the config file data.
func (c *TiEMWebServerScript) Script() ([]byte, error) {
	fp := path.Join("scripts", "run_tiem_metadb.sh.tpl")
	tpl, err := embed.ReadScriptTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiEMWebServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiEMWebServer config content by tpl
func (c *TiEMWebServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiEMWebServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
