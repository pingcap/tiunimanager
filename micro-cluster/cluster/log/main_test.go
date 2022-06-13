/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/*******************************************************************************
 * @File: main_test.go
 * @Description: main unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/16 11:32
*******************************************************************************/

package log

import (
	"os"
	"testing"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models/cluster/management"

	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/common"
)

var mockManager = NewManager()

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)
			models.MockDB()
			return models.Open(d)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func mockClusterMeta() *meta.ClusterMeta {
	return &meta.ClusterMeta{
		Cluster: mockCluster(),
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": mockClusterInstances(),
			"TiKV": mockClusterInstances(),
			"PD":   mockClusterInstances(),
		},
		NodeExporterPort:     9091,
		BlackboxExporterPort: 9092,
	}
}

func mockCluster() *management.Cluster {
	return &management.Cluster{
		Entity:            common.Entity{ID: "123", TenantId: "1", Status: "1"},
		Name:              "testCluster",
		Type:              "0",
		Version:           "5.0",
		TLS:               false,
		Tags:              nil,
		TagInfo:           "",
		OwnerId:           "1",
		ParameterGroupID:  "1",
		Copies:            0,
		Exclusive:         false,
		Region:            "1",
		CpuArchitecture:   "x64",
		MaintenanceStatus: "1",
		MaintainWindow:    "1",
	}
}

func mockClusterInstances() []*management.ClusterInstance {
	return []*management.ClusterInstance{
		{
			Entity:       common.Entity{ID: "123", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "0",
			Version:      "5.0",
			ClusterID:    "123",
			Role:         "admin",
			DiskType:     "ssd",
			DiskCapacity: 16,
			CpuCores:     32,
			Memory:       8,
			HostID:       "127.0.0.1",
			Zone:         "1",
			Rack:         "1",
			HostIP:       []string{"127.0.0.1"},
			Ports:        []int32{10000, 10001},
			DiskID:       "1",
			DiskPath:     "/tmp",
			HostInfo:     "host",
			PortInfo:     "port",
		},
	}
}

func mockDBUsers() []*management.DBUser {
	return []*management.DBUser{
		{
			ClusterID: "123",
			Name:      "backup",
			Password:  common.PasswordInExpired{Val: "123455678"},
			RoleType:  string(constants.DBUserBackupRestore),
		},
		{
			ClusterID: "123",
			Name:      "root",
			Password:  common.PasswordInExpired{Val: "123455678"},
			RoleType:  string(constants.Root),
		},
		{
			ClusterID: "123",
			Name:      "parameter",
			Password:  common.PasswordInExpired{Val: "123455678"},
			RoleType:  string(constants.DBUserParameterManagement),
		},
		{
			ClusterID: "123",
			Name:      "data_sync",
			Password:  common.PasswordInExpired{Val: "123455678"},
			RoleType:  string(constants.DBUserCDCDataSync),
		},
	}
}
