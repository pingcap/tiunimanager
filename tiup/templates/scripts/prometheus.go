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

	"github.com/pingcap-inc/tiem/tiup/embed"
)

// PrometheusScript represent the data to generate Prometheus config
type PrometheusScript struct {
	IP        string
	Port      int
	DeployDir string
	DataDir   string
	LogDir    string
	Retention string
}

// NewPrometheusScript returns a PrometheusScript with given arguments
func NewPrometheusScript(ip, deployDir, dataDir, logDir string) *PrometheusScript {
	return &PrometheusScript{
		IP:        ip,
		Port:      4129,
		DeployDir: deployDir,
		DataDir:   dataDir,
		LogDir:    logDir,
	}
}

// WithPort set Port field of PrometheusScript
func (c *PrometheusScript) WithPort(port int) *PrometheusScript {
	c.Port = port
	return c
}

// WithRetention set retention policy of PrometheusScript
func (c *PrometheusScript) WithRetention(r string) *PrometheusScript {
	c.Retention = r
	return c
}

// Script generate the config file data.
func (c *PrometheusScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_prometheus.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *PrometheusScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the Prometheus config content by tpl
func (c *PrometheusScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Prometheus").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
