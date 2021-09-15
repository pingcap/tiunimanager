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

// ElasticSearchScript represent the data to generate ElasticSearch config
type ElasticSearchScript struct {
	Host      string
	Port      int
	DeployDir string
	DataDir   string
	LogDir    string
}

// NewElasticSearchScript returns a ElasticSearchScript with given arguments
func NewElasticSearchScript(ip, deployDir, dataDir, logDir string) *ElasticSearchScript {
	return &ElasticSearchScript{
		Host:      ip,
		Port:      4122,
		DeployDir: deployDir,
		DataDir:   dataDir,
		LogDir:    logDir,
	}
}

// WithPort set Port field of ElasticSearchScript
func (c *ElasticSearchScript) WithPort(port int) *ElasticSearchScript {
	c.Port = port
	return c
}

// Script generate the config file data.
func (c *ElasticSearchScript) Script() ([]byte, error) {
	fp := path.Join("scripts", "run_elasticsearch.sh.tpl")
	tpl, err := embed.ReadScriptTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *ElasticSearchScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the ElasticSearch config content by tpl
func (c *ElasticSearchScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("ElasticSearch").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
