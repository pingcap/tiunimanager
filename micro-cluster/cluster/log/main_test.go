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
	"context"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models/cluster/management"

	"github.com/pingcap-inc/tiem/models/common"
	workflowModels "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

var mockManager = NewManager()

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)
			models.MockDB()
			return models.Open(d, false)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func mockClusterMeta() *handler.ClusterMeta {
	return &handler.ClusterMeta{
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
