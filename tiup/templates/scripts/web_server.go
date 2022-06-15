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
	"text/template"

	"github.com/pingcap/tiunimanager/tiup/embed"
)

// TiUniManagerWebServerScript represent the data to generate TiUniManagerWebServer config
type TiUniManagerWebServerScript struct {
	Host      string
	DeployDir string
	LogDir    string
}

// NewTiUniManagerWebServerScript returns a TiUniManagerWebServerScript with given arguments
func NewTiUniManagerWebServerScript(ip, deployDir, logDir string) *TiUniManagerWebServerScript {
	return &TiUniManagerWebServerScript{
		Host:      ip,
		DeployDir: deployDir,
		LogDir:    logDir,
	}
}

// Script generate the config file data.
func (c *TiUniManagerWebServerScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_em_web.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiUniManagerWebServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiUniManagerWebServer config content by tpl
func (c *TiUniManagerWebServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiUniManagerWebServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
