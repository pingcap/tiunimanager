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

package scripts

import (
	"bytes"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pingcap/tiunimanager/tiup/embed"
)

// TiUniManagerFileServerScript represent the data to generate TiUniManagerFileServer config
type TiUniManagerFileServerScript struct {
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

// NewTiUniManagerFileServerScript returns a TiUniManagerFileServerScript with given arguments
func NewTiUniManagerFileServerScript(ip, deployDir, dataDir, logDir, logLevel string) *TiUniManagerFileServerScript {
	return &TiUniManagerFileServerScript{
		Host:        ip,
		Port:        4102,
		MetricsPort: 4105,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiUniManagerFileServerScript
func (c *TiUniManagerFileServerScript) WithPort(port int) *TiUniManagerFileServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiUniManagerFileServerScript
func (c *TiUniManagerFileServerScript) WithMetricsPort(port int) *TiUniManagerFileServerScript {
	c.MetricsPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiUniManagerFileServerScript) WithRegistry(addr []string) *TiUniManagerFileServerScript {
	c.RegistryEndpoints = strings.Join(addr, ",")
	return c
}

// WithTracer set TracerAddress
func (c *TiUniManagerFileServerScript) WithTracer(addr []string) *TiUniManagerFileServerScript {
	c.TracerAddress = strings.Join(addr, ",")
	return c
}

// WithElasticsearch set ElasticsearchAddress
func (c *TiUniManagerFileServerScript) WithElasticsearch(addr []string) *TiUniManagerFileServerScript {
	c.ElasticsearchAddress = strings.Join(addr, ",")
	return c
}

// WithEnableHttps set EnableHttps
func (c *TiUniManagerFileServerScript) WithEnableHttps(enableHttps string) *TiUniManagerFileServerScript {
	c.EnableHttps = enableHttps
	return c
}

// Script generate the config file data.
func (c *TiUniManagerFileServerScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_em_file.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiUniManagerFileServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiUniManagerFileServer config content by tpl
func (c *TiUniManagerFileServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiUniManagerFileServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
