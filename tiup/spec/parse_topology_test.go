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
	"fmt"
	"os"
	"testing"

	"github.com/pingcap/check"
)

func TestUtils(t *testing.T) {
	check.TestingT(t)
}

type topoSuite struct{}

var _ = check.Suite(&topoSuite{})

func withTempFile(content string, fn func(string)) {
	file, err := os.CreateTemp("/tmp", "topology-test")
	if err != nil {
		panic(fmt.Sprintf("create temp file: %s", err))
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString(content)
	if err != nil {
		panic(fmt.Sprintf("write temp file: %s", err))
	}
	file.Close()

	fn(file.Name())
}

func with2TempFile(content1, content2 string, fn func(string, string)) {
	withTempFile(content1, func(file1 string) {
		withTempFile(content2, func(file2 string) {
			fn(file1, file2)
		})
	})
}

func (s *topoSuite) TestRelativePath(c *check.C) {
	// test relative path
	withTempFile(`
tiem_cluster_servers:
  - host: 172.16.5.140
    deploy_dir: my-deploy
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/my-deploy")
	})

	// test data dir & log dir
	withTempFile(`
tiem_cluster_servers:
  - host: 172.16.5.140
    deploy_dir: my-deploy
    data_dir: my-data
    log_dir: my-log
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/my-deploy")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tiem/my-deploy/my-data")
		c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "/home/tiem/my-deploy/my-log")
	})

	// test global options, case 1
	withTempFile(`
global:
  deploy_dir: my-deploy

tiem_cluster_servers:
  - host: 172.16.5.140
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "my-deploy")
		c.Assert(topo.GlobalOptions.DataDir, check.Equals, "data")

		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/my-deploy/cluster-server-4110")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tiem/my-deploy/cluster-server-4110/data")
	})

	// test global options, case 2
	withTempFile(`
global:
  deploy_dir: my-deploy

tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 20180
  - host: 172.16.5.140
    port: 20161
    metrics_port: 20181
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "my-deploy")
		c.Assert(topo.GlobalOptions.DataDir, check.Equals, "data")

		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20160")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20160/data")

		c.Assert(topo.ClusterServers[1].DeployDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20161")
		c.Assert(topo.ClusterServers[1].DataDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20161/data")
	})

	// test global options, case 3
	withTempFile(`
global:
  deploy_dir: my-deploy

tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 20180
    data_dir: my-data
    log_dir: my-log
  - host: 172.16.5.140
    port: 20161
    metrics_port: 20181
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)

		ExpandRelativeDir(&topo)
		c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "my-deploy")
		c.Assert(topo.GlobalOptions.DataDir, check.Equals, "data")

		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20160")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20160/my-data")
		c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20160/my-log")

		c.Assert(topo.ClusterServers[1].DeployDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20161")
		c.Assert(topo.ClusterServers[1].DataDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20161/data")
		//c.Assert(topo.ClusterServers[1].LogDir, check.Equals, "/home/tiem/my-deploy/cluster-server-20161/log")
	})

	// test global options, case 4
	withTempFile(`
global:
  data_dir: my-global-data
  log_dir: my-global-log

tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 20180
    data_dir: my-local-data
    log_dir: my-local-log
  - host: 172.16.5.140
    port: 20161
    metrics_port: 20181
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "deploy")
		c.Assert(topo.GlobalOptions.DataDir, check.Equals, "my-global-data")
		c.Assert(topo.GlobalOptions.LogDir, check.Equals, "my-global-log")

		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tiem/deploy/cluster-server-20160")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tiem/deploy/cluster-server-20160/my-local-data")
		c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "/home/tiem/deploy/cluster-server-20160/my-local-log")

		c.Assert(topo.ClusterServers[1].DeployDir, check.Equals, "/home/tiem/deploy/cluster-server-20161")
		c.Assert(topo.ClusterServers[1].DataDir, check.Equals, "/home/tiem/deploy/cluster-server-20161/my-global-data")
		c.Assert(topo.ClusterServers[1].LogDir, check.Equals, "/home/tiem/deploy/cluster-server-20161/my-global-log")
	})

	// test global options, case 6
	withTempFile(`
global:
  user: test
  data_dir: my-global-data
  log_dir: my-global-log

tiem_cluster_servers:
  - host: 172.16.5.140
    port: 20160
    metrics_port: 20180
    deploy_dir: my-local-deploy
    data_dir: my-local-data
    log_dir: my-local-log
  - host: 172.16.5.140
    port: 20161
    metrics_port: 20181
`, func(file string) {
		topo := Specification{}
		err := ParseTopologyYaml(file, &topo)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(&topo)
		c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "deploy")
		c.Assert(topo.GlobalOptions.DataDir, check.Equals, "my-global-data")
		c.Assert(topo.GlobalOptions.LogDir, check.Equals, "my-global-log")

		c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/test/my-local-deploy")
		c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/test/my-local-deploy/my-local-data")
		c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "/home/test/my-local-deploy/my-local-log")

		c.Assert(topo.ClusterServers[1].DeployDir, check.Equals, "/home/test/deploy/cluster-server-20161")
		c.Assert(topo.ClusterServers[1].DataDir, check.Equals, "/home/test/deploy/cluster-server-20161/my-global-data")
		c.Assert(topo.ClusterServers[1].LogDir, check.Equals, "/home/test/deploy/cluster-server-20161/my-global-log")
	})
}

func merge4test(base, scale string) (*Specification, error) {
	baseTopo := Specification{}
	if err := ParseTopologyYaml(base, &baseTopo); err != nil {
		return nil, err
	}

	scaleTopo := baseTopo.NewPart()
	if err := ParseTopologyYaml(scale, scaleTopo); err != nil {
		return nil, err
	}

	mergedTopo := baseTopo.MergeTopo(scaleTopo)
	if err := mergedTopo.Validate(); err != nil {
		return nil, err
	}

	return mergedTopo.(*Specification), nil
}

func (s *topoSuite) TestTopologyMerge(c *check.C) {
	// base test
	with2TempFile(`
tiem_metadb_servers:
  - host: 172.16.5.140
`, `
tiem_metadb_servers:
  - host: 172.16.5.139
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		c.Assert(err, check.IsNil)
		ExpandRelativeDir(topo)
		ExpandRelativeDir(topo) // should be idempotent

		c.Assert(topo.MetaDBServers[0].DeployDir, check.Equals, "/home/tiem/deploy/metadb-server-4100")
		c.Assert(topo.MetaDBServers[0].DataDir, check.Equals, "/home/tiem/deploy/metadb-server-4100/data")

		c.Assert(topo.MetaDBServers[1].DeployDir, check.Equals, "/home/tiem/deploy/metadb-server-4100")
		c.Assert(topo.MetaDBServers[1].DataDir, check.Equals, "/home/tiem/deploy/metadb-server-4100/data")
	})

	// test global option overwrite
	with2TempFile(`
global:
  user: test
  deploy_dir: /my-global-deploy

tiem_metadb_servers:
  - host: 172.16.5.140
    log_dir: my-local-log-tiflash
    data_dir: my-local-data-tiflash
  - host: 172.16.5.175
    deploy_dir: flash-deploy
  - host: 172.16.5.141
`, `
tiem_metadb_servers:
  - host: 172.16.5.139
    deploy_dir: flash-deploy
  - host: 172.16.5.134
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		c.Assert(err, check.IsNil)

		ExpandRelativeDir(topo)

		c.Assert(topo.MetaDBServers[0].DeployDir, check.Equals, "/my-global-deploy/metadb-server-4100")
		c.Assert(topo.MetaDBServers[0].DataDir, check.Equals, "/my-global-deploy/metadb-server-4100/my-local-data-tiflash")
		c.Assert(topo.MetaDBServers[0].LogDir, check.Equals, "/my-global-deploy/metadb-server-4100/my-local-log-tiflash")

		c.Assert(topo.MetaDBServers[1].DeployDir, check.Equals, "/home/test/flash-deploy")
		c.Assert(topo.MetaDBServers[1].DataDir, check.Equals, "/home/test/flash-deploy/data")
		c.Assert(topo.MetaDBServers[3].DeployDir, check.Equals, "/home/test/flash-deploy")
		c.Assert(topo.MetaDBServers[3].DataDir, check.Equals, "/home/test/flash-deploy/data")

		c.Assert(topo.MetaDBServers[2].DeployDir, check.Equals, "/my-global-deploy/metadb-server-4100")
		c.Assert(topo.MetaDBServers[2].DataDir, check.Equals, "/my-global-deploy/metadb-server-4100/data")
		c.Assert(topo.MetaDBServers[4].DeployDir, check.Equals, "/my-global-deploy/metadb-server-4100")
		c.Assert(topo.MetaDBServers[4].DataDir, check.Equals, "/my-global-deploy/metadb-server-4100/data")
	})
}

func (s *topoSuite) TestFixRelativePath(c *check.C) {
	// base test
	topo := Specification{
		ClusterServers: []*ClusterServerSpec{
			{
				DeployDir: "my-deploy",
			},
		},
	}
	expandRelativePath("tidb", &topo)
	c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tidb/my-deploy")

	// test data dir & log dir
	topo = Specification{
		ClusterServers: []*ClusterServerSpec{
			{
				DeployDir: "my-deploy",
				DataDir:   "my-data",
				LogDir:    "my-log",
			},
		},
	}
	expandRelativePath("tidb", &topo)
	c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "/home/tidb/my-deploy")
	c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "/home/tidb/my-deploy/my-data")
	c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "/home/tidb/my-deploy/my-log")

	// test global options
	topo = Specification{
		GlobalOptions: GlobalOptions{
			DeployDir: "my-deploy",
			DataDir:   "my-data",
			LogDir:    "my-log",
		},
		ClusterServers: []*ClusterServerSpec{
			{},
		},
	}
	expandRelativePath("tidb", &topo)
	c.Assert(topo.GlobalOptions.DeployDir, check.Equals, "my-deploy")
	c.Assert(topo.GlobalOptions.DataDir, check.Equals, "my-data")
	c.Assert(topo.GlobalOptions.LogDir, check.Equals, "my-log")
	c.Assert(topo.ClusterServers[0].DeployDir, check.Equals, "")
	c.Assert(topo.ClusterServers[0].DataDir, check.Equals, "")
	c.Assert(topo.ClusterServers[0].LogDir, check.Equals, "")
}
