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

package spec

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type metaSuiteDM struct {
}

var _ = Suite(&metaSuiteDM{})

func TestDefaultDataDir(t *testing.T) {
	// Test with without global DataDir.
	topo := new(Specification)
	topo.MetaDBServers = append(topo.MetaDBServers, &MetaDBServerSpec{Host: "1.1.1.1", Port: 1111})
	topo.ClusterServers = append(topo.ClusterServers, &ClusterServerSpec{Host: "1.1.2.1", Port: 2221})
	data, err := yaml.Marshal(topo)
	assert.Nil(t, err)

	// Check default value.
	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)
	assert.Nil(t, err)
	assert.Equal(t, "data", topo.GlobalOptions.DataDir)
	assert.Equal(t, "data", topo.MetaDBServers[0].DataDir)
	assert.Equal(t, "data", topo.ClusterServers[0].DataDir)

	// Can keep the default value.
	data, err = yaml.Marshal(topo)
	assert.Nil(t, err)
	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)
	assert.Nil(t, err)
	assert.Equal(t, "data", topo.GlobalOptions.DataDir)
	assert.Equal(t, "data", topo.MetaDBServers[0].DataDir)
	assert.Equal(t, "data", topo.ClusterServers[0].DataDir)

	// Test with global DataDir.
	topo = new(Specification)
	topo.GlobalOptions.DataDir = "/gloable_data"
	topo.MetaDBServers = append(topo.MetaDBServers, &MetaDBServerSpec{Host: "1.1.1.1", Port: 1111})
	topo.MetaDBServers = append(topo.MetaDBServers, &MetaDBServerSpec{Host: "1.1.1.2", Port: 1112, DataDir: "/my_data"})
	topo.ClusterServers = append(topo.ClusterServers, &ClusterServerSpec{Host: "1.1.2.1", Port: 2221})
	topo.ClusterServers = append(topo.ClusterServers, &ClusterServerSpec{Host: "1.1.2.2", Port: 2222, DataDir: "/my_data"})
	data, err = yaml.Marshal(topo)
	assert.Nil(t, err)

	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)

	assert.Nil(t, err)
	assert.Equal(t, "/gloable_data", topo.GlobalOptions.DataDir)
	assert.Equal(t, "/gloable_data/metadb-server-1111", topo.MetaDBServers[0].DataDir)
	assert.Equal(t, "/my_data", topo.MetaDBServers[1].DataDir)
	assert.Equal(t, "/gloable_data/cluster-server-2221", topo.ClusterServers[0].DataDir)
	assert.Equal(t, "/my_data", topo.ClusterServers[1].DataDir)
}

func TestGlobalOptions(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
tiem_metadb_servers:
  - host: 172.16.5.138
    deploy_dir: "metadb-server-deploy"
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.53
    data_dir: "metadb-data"
    metrics_port: 8112
`), &topo)
	assert.Nil(t, err)
	assert.Equal(t, "test1", topo.GlobalOptions.User)

	assert.Equal(t, 220, topo.GlobalOptions.SSHPort)
	assert.Equal(t, 220, topo.MetaDBServers[0].SSHPort)
	assert.Equal(t, "metadb-server-deploy", topo.MetaDBServers[0].DeployDir)

	assert.Equal(t, 220, topo.ClusterServers[0].SSHPort)
	assert.Equal(t, "test-deploy/cluster-server-4110", topo.ClusterServers[0].DeployDir)
	assert.Equal(t, "metadb-data", topo.ClusterServers[0].DataDir)
}

func TestDirectoryConflicts(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
tiem_metadb_servers:
  - host: 172.16.5.138
    deploy_dir: "/test-1"
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.138
    data_dir: "/test-1"
    metrics_port: 8112
`), &topo)
	assert.NotNil(t, err)
	assert.Equal(t, "directory conflict for '/test-1' between 'tiem_metadb_servers:172.16.5.138.deploy_dir' and 'tiem_cluster_servers:172.16.5.138.data_dir'", err.Error())

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "/test-data"
tiem_metadb_servers:
  - host: 172.16.5.138
    data_dir: "test-1"
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.138
    data_dir: "test-1"
    metrics_port: 8112
`), &topo)
	assert.Nil(t, err)
}

func TestPortConflicts(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
tiem_metadb_servers:
  - host: 172.16.5.138
    registry_peer_port: 1234
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.138
    port: 1234
    metrics_port: 8112
`), &topo)
	assert.NotNil(t, err)
	assert.Equal(t, "port conflict for '1234' between 'tiem_metadb_servers:172.16.5.138.registry_peer_port,omitempty' and 'tiem_cluster_servers:172.16.5.138.port,omitempty'", err.Error())
}

func TestPlatformConflicts(t *testing.T) {
	// aarch64 and arm64 are equal
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  os: "linux"
  arch: "aarch64"
tiem_metadb_servers:
  - host: 172.16.5.138
    arch: "arm64"
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.138
    metrics_port: 8112
`), &topo)
	assert.Nil(t, err)

	// different arch defined for the same host
	topo = Specification{}
	err = yaml.Unmarshal([]byte(`
global:
  os: "linux"
tiem_metadb_servers:
  - host: 172.16.5.138
    arch: "aarch64"
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.138
    arch: "amd64"
    metrics_port: 8112
`), &topo)
	assert.NotNil(t, err)
	assert.Equal(t, "platform mismatch for '172.16.5.138' between 'tiem_metadb_servers:linux/arm64' and 'tiem_cluster_servers:linux/amd64'", err.Error())

	// different os defined for the same host
	topo = Specification{}
	err = yaml.Unmarshal([]byte(`
global:
  os: "linux"
  arch: "aarch64"
tiem_metadb_servers:
  - host: 172.16.5.138
    metrics_port: 8111
    os: "darwin"
tiem_cluster_servers:
  - host: 172.16.5.138
    metrics_port: 8112
`), &topo)
	assert.NotNil(t, err)
	assert.Equal(t, "platform mismatch for '172.16.5.138' between 'tiem_metadb_servers:darwin/arm64' and 'tiem_cluster_servers:linux/arm64'", err.Error())
}

func TestCountDir(t *testing.T) {
	topo := Specification{}

	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
tiem_metadb_servers:
  - host: 172.16.5.138
    metrics_port: 8111
    deploy_dir: "metadb-server-deploy"
    data_dir: "/test-data/data-1"
tiem_cluster_servers:
  - host: 172.16.5.53
    metrics_port: 8112
    data_dir: "test-1"
`), &topo)
	assert.Nil(t, err)
	cnt := topo.CountDir("172.16.5.53", "test-deploy/cluster-server-4110")
	assert.Equal(t, 3, cnt)
	cnt = topo.CountDir("172.16.5.138", "/test-data/data")
	assert.Equal(t, 0, cnt) // should not match partial path

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "/test-deploy"
tiem_metadb_servers:
  - host: 172.16.5.138
    metrics_port: 8111
    deploy_dir: "metadb-server-deploy"
    data_dir: "/test-data/data-1"
tiem_cluster_servers:
  - host: 172.16.5.138
    metrics_port: 8112
    data_dir: "/test-data/data-2"
`), &topo)
	assert.Nil(t, err)
	cnt = topo.CountDir("172.16.5.138", "/test-deploy/cluster-server-4110")
	assert.Equal(t, 2, cnt)
	cnt = topo.CountDir("172.16.5.138", "")
	assert.Equal(t, 2, cnt)
	cnt = topo.CountDir("172.16.5.138", "test-data")
	assert.Equal(t, 0, cnt)
	cnt = topo.CountDir("172.16.5.138", "/test-data")
	assert.Equal(t, 2, cnt)

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "/test-deploy"
  data_dir: "/test-data"
tiem_metadb_servers:
  - host: 172.16.5.138
    metrics_port: 8111
    data_dir: "data-1"
tiem_cluster_servers:
  - host: 172.16.5.138
    metrics_port: 8112
    data_dir: "data-2"
  - host: 172.16.5.53
`), &topo)
	assert.Nil(t, err)
	// if per-instance data_dir is set, the global data_dir is ignored, and if it
	// is a relative path, it will be under the instance's deploy_dir
	cnt = topo.CountDir("172.16.5.138", "/test-deploy/cluster-server-4110")
	assert.Equal(t, 3, cnt)
	cnt = topo.CountDir("172.16.5.138", "")
	assert.Equal(t, 0, cnt)
	cnt = topo.CountDir("172.16.5.53", "/test-data")
	assert.Equal(t, 1, cnt)
}

func TestRelativePath(t *testing.T) {
	// base test
	withTempFile(`
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    metrics_port: 8112
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)
		assert.Equal(t, "/home/tiem/deploy/metadb-server-4100", topo.MetaDBServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-4110", topo.ClusterServers[0].DeployDir)
	})

	// test data dir & log dir
	withTempFile(`
tiem_metadb_servers:
  - host: 172.16.5.140
    deploy_dir: my-deploy
    data_dir: my-data
    log_dir: my-log
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)

		assert.Equal(t, "/home/tiem/my-deploy", topo.MetaDBServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/my-data", topo.MetaDBServers[0].DataDir)
		assert.Equal(t, "/home/tiem/my-deploy/my-log", topo.MetaDBServers[0].LogDir)
	})

	// test global options, case 1
	withTempFile(`
global:
  deploy_dir: my-deploy
tiem_metadb_servers:
  - host: 172.16.5.140
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)

		assert.Equal(t, "/home/tiem/my-deploy/metadb-server-4100", topo.MetaDBServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/metadb-server-4100/data", topo.MetaDBServers[0].DataDir)
		assert.Equal(t, "", topo.MetaDBServers[0].LogDir)
	})

	// test global options, case 2
	withTempFile(`
global:
  deploy_dir: my-deploy
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 8112
  - host: 172.16.5.140
    port: 20161
    metrics_port: 8113
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)

		assert.Equal(t, "my-deploy", topo.GlobalOptions.DeployDir)
		assert.Equal(t, "data", topo.GlobalOptions.DataDir)

		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20160", topo.ClusterServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20160/data", topo.ClusterServers[0].DataDir)

		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20161", topo.ClusterServers[1].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20161/data", topo.ClusterServers[1].DataDir)
	})

	// test global options, case 3
	withTempFile(`
global:
  deploy_dir: my-deploy
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 8112
    data_dir: my-data
    log_dir: my-log
  - host: 172.16.5.140
    port: 20161
    metrics_port: 8113
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)

		assert.Equal(t, "my-deploy", topo.GlobalOptions.DeployDir)
		assert.Equal(t, "data", topo.GlobalOptions.DataDir)

		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20160", topo.ClusterServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20160/my-data", topo.ClusterServers[0].DataDir)
		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20160/my-log", topo.ClusterServers[0].LogDir)

		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20161", topo.ClusterServers[1].DeployDir)
		assert.Equal(t, "/home/tiem/my-deploy/cluster-server-20161/data", topo.ClusterServers[1].DataDir)
		assert.Equal(t, "", topo.ClusterServers[1].LogDir)
	})

	// test global options, case 4
	withTempFile(`
global:
  data_dir: my-global-data
  log_dir: my-global-log
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 8112
    data_dir: my-local-data
    log_dir: my-local-log
  - host: 172.16.5.140
    port: 20161
    metrics_port: 8113
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		ExpandRelativeDir(&topo)

		assert.Equal(t, "deploy", topo.GlobalOptions.DeployDir)
		assert.Equal(t, "my-global-data", topo.GlobalOptions.DataDir)
		assert.Equal(t, "my-global-log", topo.GlobalOptions.LogDir)

		assert.Equal(t, "/home/tiem/deploy/cluster-server-20160", topo.ClusterServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-20160/my-local-data", topo.ClusterServers[0].DataDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-20160/my-local-log", topo.ClusterServers[0].LogDir)

		assert.Equal(t, "/home/tiem/deploy/cluster-server-20161", topo.ClusterServers[1].DeployDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-20161/my-global-data", topo.ClusterServers[1].DataDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-20161/my-global-log", topo.ClusterServers[1].LogDir)
	})
}

func TestTopologyMerge(t *testing.T) {
	// base test
	with2TempFile(`
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    metrics_port: 8112
`, `
tiem_cluster_servers:
  - host: 172.16.5.139
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		assert.Nil(t, err)
		ExpandRelativeDir(topo)

		assert.Equal(t, "/home/tiem/deploy/cluster-server-4110", topo.ClusterServers[0].DeployDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-4110/data", topo.ClusterServers[0].DataDir)
		assert.Equal(t, "", topo.ClusterServers[0].LogDir)

		assert.Equal(t, "/home/tiem/deploy/cluster-server-4110", topo.ClusterServers[1].DeployDir)
		assert.Equal(t, "/home/tiem/deploy/cluster-server-4110/data", topo.ClusterServers[1].DataDir)
		assert.Equal(t, "", topo.ClusterServers[1].LogDir)
	})

	// test global option overwrite
	with2TempFile(`
global:
  user: test
  deploy_dir: /my-global-deploy
tiem_metadb_servers:
  - host: 172.16.5.140
    metrics_port: 8111
tiem_cluster_servers:
  - host: 172.16.5.140
    metrics_port: 8112
    log_dir: my-local-log
    data_dir: my-local-data
  - host: 172.16.5.175
    deploy_dir: flash-deploy
  - host: 172.16.5.141
`, `
tiem_cluster_servers:
  - host: 172.16.5.139
    deploy_dir: flash-deploy
  - host: 172.16.5.134
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		assert.Nil(t, err)

		ExpandRelativeDir(topo)

		assert.Equal(t, "/my-global-deploy/cluster-server-4110", topo.ClusterServers[0].DeployDir)
		assert.Equal(t, "/my-global-deploy/cluster-server-4110/my-local-data", topo.ClusterServers[0].DataDir)
		assert.Equal(t, "/my-global-deploy/cluster-server-4110/my-local-log", topo.ClusterServers[0].LogDir)

		assert.Equal(t, "/home/test/flash-deploy", topo.ClusterServers[1].DeployDir)
		assert.Equal(t, "/home/test/flash-deploy/data", topo.ClusterServers[1].DataDir)
		assert.Equal(t, "/home/test/flash-deploy", topo.ClusterServers[3].DeployDir)
		assert.Equal(t, "/home/test/flash-deploy/data", topo.ClusterServers[3].DataDir)

		assert.Equal(t, "/my-global-deploy/cluster-server-4110", topo.ClusterServers[2].DeployDir)
		assert.Equal(t, "/my-global-deploy/cluster-server-4110/data", topo.ClusterServers[2].DataDir)
		assert.Equal(t, "/my-global-deploy/cluster-server-4110", topo.ClusterServers[4].DeployDir)
		assert.Equal(t, "/my-global-deploy/cluster-server-4110/data", topo.ClusterServers[4].DataDir)
	})
}

func TestMonitorLogDir(t *testing.T) {
	withTempFile(`
monitored:
  node_exporter_port: 39100
  blackbox_exporter_port: 39115
  deploy_dir: "test-deploy"
  log_dir: "test-deploy/log"
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		assert.Nil(t, err)
		assert.Equal(t, 39100, topo.MonitoredOptions.NodeExporterPort)
		assert.Equal(t, 39115, topo.MonitoredOptions.BlackboxExporterPort)
		assert.Equal(t, "test-deploy/log", topo.MonitoredOptions.LogDir)
		assert.Equal(t, "test-deploy", topo.MonitoredOptions.DeployDir)
	})
}
