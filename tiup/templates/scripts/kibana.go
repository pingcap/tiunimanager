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

	"github.com/pingcap/tiunimanager/tiup/embed"
)

// KibanaScript represent the data to generate Kibana config
type KibanaScript struct {
	DeployDir string
}

// NewKibanaScript returns a KibanaScript with given arguments
func NewKibanaScript(deployDir string) *KibanaScript {
	return &KibanaScript{
		DeployDir: deployDir,
	}
}

// Script generate the config file data.
func (c *KibanaScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_kibana.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *KibanaScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the Kibana config content by tpl
func (c *KibanaScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("Kibana").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
