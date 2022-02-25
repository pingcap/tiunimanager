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

package parameter

import (
	"context"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	workflowModels "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
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

var apiContent = []byte(`
{
	"oom-action": "cancel",
	"security": {
		"enable-sem": true
	},
	"log": {
		"file": {
			"max-size": 102400
		}
	},
	"coprocessor": {
		"region-max-size": "144MiB"
	},
	"rocksdb": {
		"defaultcf": {
			"compression-per-level": ["no","no","lz4","lz4","lz4","zstd","zstd"]
		}
	}
}
`)

var tomlConfigContent = []byte(`
# WARNING: This file is auto-generated. Do not edit! All your modification will be overwritten!
# You can use 'tiup cluster edit-config' and 'tiup cluster reload' to update the configuration
# All configuration items you want to change can be added to:
# server_configs:
#   pd:
#     aa.b1.c3: value
#     aa.b2.c4: value
auto-compaction-mod = "periodic"
auto-compaction-retention = "1h"
lease = 3
quota-backend-bytes = 8589934592

[dashboard]
public-path-prefix = "/dashboard"

[label-property]
key = ["test"]
value = ["test"]

[log]
[log.file]
max-backups = 7
max-days = 28
max-size = 300

[metric]
interval = "15s"

[replication]
location-labels = ["region", "zone", "rack", "host"]

[security]
redact-info-log = false
`)

func mockClusterMeta() *meta.ClusterMeta {
	return &meta.ClusterMeta{
		Cluster: mockCluster(),
		Instances: map[string][]*management.ClusterInstance{
			"TiDB":    mockClusterInstances(),
			"TiKV":    mockClusterInstances(),
			"PD":      mockClusterInstances(),
			"CDC":     mockClusterInstances(),
			"TiFlash": mockClusterInstances(),
		},
		DBUsers: map[string]*management.DBUser{
			string(constants.Root): &management.DBUser{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
			string(constants.DBUserBackupRestore): &management.DBUser{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.DBUserBackupRestore],
				Password:  "12345678",
				RoleType:  string(constants.DBUserBackupRestore),
			},
			string(constants.DBUserParameterManagement): &management.DBUser{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.DBUserParameterManagement],
				Password:  "12345678",
				RoleType:  string(constants.DBUserParameterManagement),
			},
			string(constants.DBUserCDCDataSync): &management.DBUser{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.DBUserCDCDataSync],
				Password:  "12345678",
				RoleType:  string(constants.DBUserCDCDataSync),
			},
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
			Entity:       common.Entity{ID: "1", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "TiDB",
			Version:      "v5.0.0",
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

func mockClusterAllInstances() []*management.ClusterInstance {
	return []*management.ClusterInstance{
		{
			Entity:       common.Entity{ID: "1", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "TiDB",
			Version:      "v5.0.0",
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
		}, {
			Entity:       common.Entity{ID: "2", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "TiKV",
			Version:      "v5.0.0",
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
		}, {
			Entity:       common.Entity{ID: "3", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "PD",
			Version:      "v5.0.0",
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
		}, {
			Entity:       common.Entity{ID: "4", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "CDC",
			Version:      "v5.0.0",
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
		}, {
			Entity:       common.Entity{ID: "5", TenantId: "1", Status: string(constants.ClusterInstanceRunning)},
			Type:         "TiFlash",
			Version:      "v5.0.0",
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
			ClusterID: "clusterId",
			Name:      "backup",
			Password:  "123455678",
			RoleType:  string(constants.DBUserBackupRestore),
		},
		{
			ClusterID: "clusterId",
			Name:      "root",
			Password:  "123455678",
			RoleType:  string(constants.Root),
		},
		{
			ClusterID: "clusterId",
			Name:      "parameter",
			Password:  "123455678",
			RoleType:  string(constants.DBUserParameterManagement),
		},
		{
			ClusterID: "clusterId",
			Name:      "data_sync",
			Password:  "123455678",
			RoleType:  string(constants.DBUserCDCDataSync),
		},
	}
}

func mockModifyParameter() *ModifyParameter {
	return &ModifyParameter{
		Reboot: false,
		Params: []*ModifyClusterParameterInfo{
			{
				ParamId:        "1",
				Category:       "basic",
				Name:           "test_param_1",
				InstanceType:   "TiDB",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				Range:          []string{"0", "1024"},
				RangeType:      1,
				ReadOnly:       1,
				RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
			},
			{
				ParamId:        "2",
				Category:       "log",
				Name:           "test_param_2",
				InstanceType:   "TiKV",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				Range:          []string{"2"},
				RangeType:      2,
				RealValue:      structs.ParameterRealValue{ClusterValue: "2"},
			},
			{
				ParamId:        "3",
				Name:           "test_param_3",
				InstanceType:   "PD",
				UpdateSource:   3,
				HasApply:       1,
				SystemVariable: "test_param_3",
				Type:           1,
				Unit:           "kB",
				UnitOptions:    []string{"KB", "MB"},
				Range:          []string{"10KB", "1MB"},
				RangeType:      1,
				RealValue:      structs.ParameterRealValue{ClusterValue: "1024KB"},
			},
			{
				ParamId:        "4",
				Name:           "test_param_4",
				InstanceType:   "TiDB",
				UpdateSource:   3,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				RealValue:      structs.ParameterRealValue{ClusterValue: "4"},
			},
			{
				ParamId:        "5",
				Name:           "test_param_5",
				InstanceType:   "TiDB",
				UpdateSource:   1,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				RealValue:      structs.ParameterRealValue{ClusterValue: "5"},
			},
			{
				ParamId:        "6",
				Name:           "test_param_6",
				InstanceType:   "TiKV",
				UpdateSource:   3,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				RealValue:      structs.ParameterRealValue{ClusterValue: "2"},
			},
			{
				ParamId:        "7",
				Name:           "test_param_7",
				InstanceType:   "CDC",
				UpdateSource:   3,
				HasApply:       1,
				SystemVariable: "",
				Type:           1,
				RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
			},
			{
				ParamId:        "8",
				Name:           "test_param_8",
				InstanceType:   "PD",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				RealValue:      structs.ParameterRealValue{ClusterValue: "3"},
			},
			{
				ParamId:        "9",
				Name:           "test_param_9",
				InstanceType:   "tiflash-learner",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           3,
				Range:          []string{"0.1", "1.0"},
				RangeType:      2,
				RealValue:      structs.ParameterRealValue{ClusterValue: "1.0"},
			},
			{
				ParamId:        "10",
				Name:           "test_param_10",
				InstanceType:   "pump",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				RealValue:      structs.ParameterRealValue{ClusterValue: "3"},
			},
			{
				ParamId:        "11",
				Name:           "test_param_11",
				InstanceType:   "drainer",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           4,
				Range:          []string{"10"},
				RangeType:      2,
				RealValue:      structs.ParameterRealValue{ClusterValue: "[]"},
			},
			{
				ParamId:        "12",
				Name:           "test_param_12",
				InstanceType:   "CDC",
				UpdateSource:   0,
				HasApply:       1,
				SystemVariable: "",
				Type:           0,
				Range:          []string{"0", "1024"},
				RangeType:      1,
				RealValue:      structs.ParameterRealValue{ClusterValue: ""},
			},
		},
		Nodes: []string{"172.16.1.12:9000", "172.16.1.12:9001"},
	}
}

func mockWorkFlowAggregation() *workflow.WorkFlowAggregation {
	return &workflow.WorkFlowAggregation{
		Flow: &workflowModels.WorkFlow{
			Entity: common.Entity{
				ID:       "1",
				TenantId: "1",
				Status:   "1",
			},
		},
		Define: &workflow.WorkFlowDefine{
			FlowName: "test",
			TaskNodes: map[string]*workflow.NodeDefine{
				"start": {
					Name:         "testNode",
					SuccessEvent: "",
					FailEvent:    "",
					ReturnType:   "",
					Executor:     nil,
				},
			},
		},
		CurrentNode: &workflowModels.WorkFlowNode{
			Entity: common.Entity{
				ID:       "1",
				TenantId: "1",
				Status:   "1",
			},
			BizID:      "1",
			ParentID:   "1",
			Name:       "start",
			ReturnType: "1",
			Parameters: "1",
			Result:     "1",
		},
		Nodes: []*workflowModels.WorkFlowNode{
			{
				Entity: common.Entity{
					ID:       "1",
					TenantId: "1",
					Status:   "1",
				},
				BizID:      "1",
				ParentID:   "1",
				Name:       "start",
				ReturnType: "1",
				Parameters: "1",
				Result:     "1",
			},
		},
		Context: workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		},
		FlowError: nil,
	}
}
