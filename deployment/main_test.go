/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: main_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/18
*******************************************************************************/

package deployment

import (
	"fmt"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/util/uuidutil"
)

const (
	TestWorkFlowID = "testworkflowid"
	TestClusterID  = "testclusterid"
	TestVersion    = "v4.0.12"
	TestTiUPHome   = "/root/.tiup"
	TestTiDBTopo   = "\nglobal:\n  user: tidb\n  group: tidb\n  ssh_port: 22\n  enable_tls: false\n  deploy_dir: -X4DGAFVRDi5KK-LmN05TA/tidb-deploy\n  data_dir: -X4DGAFVRDi5KK-LmN05TA/tidb-data\n  log_dir: -X4DGAFVRDi5KK-LmN05TA/tidb-log\n  os: linux\n\n  arch: amd64\n\nmonitored:\n  node_exporter_port: 11000\n  blackbox_exporter_port: 11001\n\npd_servers:\n  - host: 172.16.6.252\n    client_port: 10040\n    peer_port: 10041\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/pd-deploy\n    data_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/pd-data\n\nmonitoring_servers:\n  - host: 172.16.6.252\n    port: 10042\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/prometheus-deploy\n    data_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/prometheus-data\ngrafana_servers:\n  - host: 172.16.6.252\n    port: 10043\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/grafana-deploy\n    anonymous_enable: true\n    default_theme: light\n    org_name: Main Org.\n    org_role: Viewer\nalertmanager_servers:\n  - host: 172.16.6.252\n    web_port: 10044\n    cluster_port: 10045\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/alertmanagers-deploy\n    data_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/alertmanagers-data\n\ntidb_servers:\n  - host: 172.16.6.252\n    port: 10000\n    status_port: 10001\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/tidb-deploy\n\ntikv_servers:\n  - host: 172.16.6.252\n    port: 10020\n    status_port: 10021\n    deploy_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/tikv-deploy\n    data_dir: /mnt/path1/-X4DGAFVRDi5KK-LmN05TA/tikv-data\n"

	TestResult   = "testresult"
	TestErrorStr = "testerrorstr"
)

var testTiUPHome string

func TestMain(m *testing.M) {
	testTiUPHome = "testdata/" + uuidutil.ShortId()
	os.MkdirAll(fmt.Sprintf("%s/storage", testTiUPHome), 0755)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
