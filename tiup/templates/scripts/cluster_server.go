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

	"github.com/pingcap/tiunimanager/tiup/embed"
)

// TiUniManagerClusterServerScript represent the data to generate TiUniManagerClusterServer config
type TiUniManagerClusterServerScript struct {
	Host                 string
	Port                 int
	ClientPort           int
	PeerPort             int
	MetricsPort          int
	DeployDir            string
	DataDir              string
	LogDir               string
	LogLevel             string
	RegistryEndpoints    string
	ElasticsearchAddress string
	TracerAddress        string
	ClusterName          string
	ClusterVersion       string
	DeployUser           string
	DeployGroup          string
	LoginHostUser        string
	LoginPrivateKeyPath  string
	LoginPublicKeyPath   string
}

// NewTiUniManagerClusterServerScript returns a TiUniManagerClusterServerScript with given arguments
func NewTiUniManagerClusterServerScript(ip, deployDir, dataDir, logDir, logLevel string) *TiUniManagerClusterServerScript {
	return &TiUniManagerClusterServerScript{
		Host:        ip,
		Port:        4101,
		MetricsPort: 4104,
		ClientPort:  4106,
		PeerPort:    4107,
		DeployDir:   deployDir,
		DataDir:     dataDir,
		LogDir:      logDir,
		LogLevel:    logLevel,
	}
}

// WithPort set Port field of TiUniManagerClusterServerScript
func (c *TiUniManagerClusterServerScript) WithPort(port int) *TiUniManagerClusterServerScript {
	c.Port = port
	return c
}

// WithMetricsPort set PeerPort field of TiUniManagerClusterServerScript
func (c *TiUniManagerClusterServerScript) WithMetricsPort(port int) *TiUniManagerClusterServerScript {
	c.MetricsPort = port
	return c
}

// WithClientPort set ClientPort field of TiUniManagerMetaDBScript
func (c *TiUniManagerClusterServerScript) WithClientPort(port int) *TiUniManagerClusterServerScript {
	c.ClientPort = port
	return c
}

// WithPeerPort set PeerPort field of TiUniManagerMetaDBScript
func (c *TiUniManagerClusterServerScript) WithPeerPort(port int) *TiUniManagerClusterServerScript {
	c.PeerPort = port
	return c
}

// WithRegistry set RegistryEndpoints
func (c *TiUniManagerClusterServerScript) WithRegistry(addr []string) *TiUniManagerClusterServerScript {
	c.RegistryEndpoints = strings.Join(addr, ",")
	return c
}

// WithElasticsearch set ElasticsearchAddress
func (c *TiUniManagerClusterServerScript) WithElasticsearch(addr []string) *TiUniManagerClusterServerScript {
	c.ElasticsearchAddress = strings.Join(addr, ",")
	return c
}

// WithTracer set TracerAddress
func (c *TiUniManagerClusterServerScript) WithTracer(addr []string) *TiUniManagerClusterServerScript {
	c.TracerAddress = strings.Join(addr, ",")
	return c
}

// WithClusterName set ClusterName
func (c *TiUniManagerClusterServerScript) WithClusterName(clusterName string) *TiUniManagerClusterServerScript {
	c.ClusterName = clusterName
	return c
}

// WithClusterVersion set ClusterVersion
func (c *TiUniManagerClusterServerScript) WithClusterVersion(clusterVersion string) *TiUniManagerClusterServerScript {
	c.ClusterVersion = clusterVersion
	return c
}

// WithDeployUser set DeployUser
func (c *TiUniManagerClusterServerScript) WithDeployUser(deployUser string) *TiUniManagerClusterServerScript {
	c.DeployUser = deployUser
	return c
}

// WithDeployGroup set DeployGroup
func (c *TiUniManagerClusterServerScript) WithDeployGroup(deployGroup string) *TiUniManagerClusterServerScript {
	c.DeployGroup = deployGroup
	return c
}

// WithLoginHostUser set LoginHostUser
func (c *TiUniManagerClusterServerScript) WithLoginHostUser(loginHostUser string) *TiUniManagerClusterServerScript {
	c.LoginHostUser = loginHostUser
	return c
}

// WithLoginPrivateKeyPath set LoginPrivateKeyPath
func (c *TiUniManagerClusterServerScript) WithLoginPrivateKeyPath(loginPrivateKeyPath string) *TiUniManagerClusterServerScript {
	c.LoginPrivateKeyPath = loginPrivateKeyPath
	return c
}

// WithLoginPublicKeyPath set LoginPublicKeyPath
func (c *TiUniManagerClusterServerScript) WithLoginPublicKeyPath(loginPublicKeyPath string) *TiUniManagerClusterServerScript {
	c.LoginPublicKeyPath = loginPublicKeyPath
	return c
}

// Script generate the config file data.
func (c *TiUniManagerClusterServerScript) Script() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_em_cluster.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, err
	}
	return c.ScriptWithTemplate(string(tpl))
}

// ScriptToFile write config content to specific path
func (c *TiUniManagerClusterServerScript) ScriptToFile(file string) error {
	config, err := c.Script()
	if err != nil {
		return err
	}
	return os.WriteFile(file, config, 0755)
}

// ScriptWithTemplate generate the TiUniManagerClusterServer config content by tpl
func (c *TiUniManagerClusterServerScript) ScriptWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiUniManagerClusterServer").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
