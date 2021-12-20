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

// TiEMFileServerScript represent the data to generate TiEMFileServer config
type TiEMFileServerScript struct {
	Host                 string
	Port                 int
	ClientPort           int
	MetricsPort          int
	DeployDir            string
	DataDir              string
	LogDir               string
	LogLevel             string
	RegistryEndpoints    string
	TracerAddress        string
	ElasticsearchAddress string
	EnableHttps          string
}

// NewTiEMFileServerScript returns a TiEMFileServerScript with given arguments
func NewTiEMFileServerScript(ip, deployDir, dataDir, logDir, logLevel string) *TiEMFileServerScript {
	return &TiEMFileServerScript{
		Host:        ip,
		Port:        4102,
		MetricsPort: 4105,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiEMFileServerScript
func (c *TiEMFileServerScript) WithPort(port int) *TiEMFileServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiEMFileServerScript
func (c *TiEMFileServerScript) WithMetricsPort(port int) *TiEMFileServerScript {
	c.MetricsPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiEMFileServerScript) WithRegistry(addr []string) *TiEMFileServerScript {
	c.RegistryEndpoints = strings.Join(addr, ",")
	return c
}

// WithTracer set TracerAddress
func (c *TiEMFileServerScript) WithTracer(addr []string) *TiEMFileServerScript {
	c.TracerAddress = strings.Join(addr, ",")
	return c
}

// WithElasticsearch set ElasticsearchAddress
func (c *TiEMFileServerScript) WithElasticsearch(addr []string) *TiEMFileServerScript {
	c.ElasticsearchAddress = strings.Join(addr, ",")
	return c
}

// WithEnableHttps set EnableHttps
func (c *TiEMFileServerScript) WithEnableHttps(enableHttps string) *TiEMFileServerScript {
	c.EnableHttps = enableHttps
	return c
}

// Script generate the config file data.
func (c *TiEMFileServerScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_tiem_file.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiEMFileServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiEMFileServer config content by tpl
func (c *TiEMFileServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiEMFileServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
