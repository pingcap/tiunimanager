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

package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_br_service "github.com/pingcap-inc/tiem/test/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockresource"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"testing"
	"time"
)

var emptyNode = func(task *wfModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func getEmptyFlow(name string) *workflow.WorkFlowDefine {
	return &workflow.WorkFlowDefine{
		FlowName: name,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start": {"start", "done", "fail", workflow.SyncFuncNode, emptyNode},
			"done":  {"end", "", "", workflow.SyncFuncNode, emptyNode},
			"fail":  {"fail", "", "", workflow.SyncFuncNode, emptyNode},
		},
	}
}

func TestAsyncMaintenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), "testFlow", getEmptyFlow("testFlow"))
	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		meta := &handler.ClusterMeta{
			Cluster: &management.Cluster{Entity: common.Entity{ID: "cluster01"}},
		}
		data := make(map[string]interface{})
		data["key"] = "value"
		got, err := asyncMaintenance(context.TODO(), meta, constants.ClusterMaintenanceNone, "testFlow", data)
		assert.NoError(t, err)
		assert.Equal(t, got, "flow01")
	})

	t.Run("create workflow fail", func(t *testing.T) {
		clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, fmt.Errorf("create workflow error")).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		meta := &handler.ClusterMeta{
			Cluster: &management.Cluster{Entity: common.Entity{ID: "cluster01"}},
		}
		data := make(map[string]interface{})
		data["key"] = "value"
		_, err := asyncMaintenance(context.TODO(), meta, constants.ClusterMaintenanceNone, "testFlow", data)
		assert.Error(t, err)
	})

	t.Run("start workflow fail", func(t *testing.T) {
		clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(fmt.Errorf("start workflow error")).AnyTimes()
		meta := &handler.ClusterMeta{
			Cluster: &management.Cluster{Entity: common.Entity{ID: "cluster01"}},
		}
		data := make(map[string]interface{})
		data["key"] = "value"
		_, err := asyncMaintenance(context.TODO(), meta, constants.ClusterMaintenanceNone, "testFlow", data)
		assert.Error(t, err)
	})
}

func TestManager_ScaleOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, getEmptyFlow(constants.FlowScaleOutCluster))
	manager := &Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		mockTiup := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		secondparty.Manager = mockTiup
		mockTiup.EXPECT().ClusterComponentCtl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"true\"}", nil)
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type: "PD",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{4000},
			},
			{},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.ScaleOut(context.TODO(), cluster.ScaleOutClusterReq{
			ClusterID: "111",
			ClusterResourceInfo: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 2, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 2},
					}},
					{Type: "TiFlash", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.NoError(t, err)
	})

	t.Run("not found cluster", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "112").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type: "PD",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{4000},
			},
			{},
		}, make([]*management.DBUser, 0), fmt.Errorf("not found cluster"))
		_, err := manager.ScaleOut(context.TODO(), cluster.ScaleOutClusterReq{
			ClusterID:           "112",
			ClusterResourceInfo: structs.ClusterResourceInfo{},
		})
		assert.Error(t, err)
	})

	t.Run("check fail", func(t *testing.T) {
		mockTiup := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		secondparty.Manager = mockTiup
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type: "PD",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{4000},
			},
			{},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		mockTiup.EXPECT().ClusterComponentCtl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"false\"}", nil).AnyTimes()
		_, err := manager.ScaleOut(context.TODO(), cluster.ScaleOutClusterReq{
			ClusterID: "111",
			ClusterResourceInfo: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiFlash", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("aync fail", func(t *testing.T) {
		mockTiup := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		secondparty.Manager = mockTiup
		mockTiup.EXPECT().ClusterComponentCtl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"true\"}", nil)
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type: "PD",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{4000},
			},
			{},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.ScaleOut(context.TODO(), cluster.ScaleOutClusterReq{
			ClusterID: "111",
			ClusterResourceInfo: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 2, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 2},
					}},
					{Type: "TiFlash", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_ScaleIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, getEmptyFlow(constants.FlowScaleInCluster))
	manager := &Manager{}
	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)

	clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
		{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
	}, make([]*management.DBUser, 0), nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Type: "TiDB", Version: "v4.0.12"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Version: "v4.0.12", Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Version: "v4.0.12", Type: "TiDB"},
			{Entity: common.Entity{ID: "instance03"}, Version: "v4.0.12", Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "111",
			InstanceID: "instance01",
		})
		assert.NoError(t, err)
	})

	t.Run("not found cluster", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "112").Return(&management.Cluster{Type: "TiDB", Version: "v4.0.12"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "PD"},
		}, make([]*management.DBUser, 0), fmt.Errorf("not found cluster")).AnyTimes()
		_, err := manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "112",
			InstanceID: "instance01",
		})
		assert.Error(t, err)
	})

	t.Run("not found instance", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Type: "TiDB", Version: "v4.0.12"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		_, err := manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "111",
			InstanceID: "instance03",
		})
		assert.Error(t, err)
	})

	t.Run("check fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Type: "TiDB", Version: "v4.0.12"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		_, err := manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "111",
			InstanceID: "instance01",
		})
		assert.Error(t, err)

		_, err = manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "111",
			InstanceID: "instance02",
		})
		assert.Error(t, err)
	})

	t.Run("aync fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Type: "TiDB", Version: "v4.0.12"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance03"}, Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.ScaleIn(context.TODO(), cluster.ScaleInClusterReq{
			ClusterID:  "111",
			InstanceID: "instance01",
		})
		assert.Error(t, err)
	})
}

func TestManager_Clone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, getEmptyFlow(constants.FlowScaleInCluster))
	manager := &Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, []*management.DBUser{
			{
				ClusterID: "111",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType: string(constants.Root),
			},
		}, nil).AnyTimes()
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().CreateRelation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v5.0.0",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("not found cluster", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "PD"},
		}, make([]*management.DBUser, 0), fmt.Errorf("not found cluster")).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v5.0.0",
			},
		})
		assert.Error(t, err)
	})

	t.Run("clone meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Version: "v5.0.0"},
			[]*management.ClusterInstance{
				{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
				{Entity: common.Entity{ID: "instance02"}, Type: "PD"},
			}, make([]*management.DBUser, 0), nil).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v3.0.0",
			},
		})
		assert.Error(t, err)
	})

	t.Run("sync fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().CreateRelation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v5.0.0",
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_CreateCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, getEmptyFlow(constants.FlowCreateCluster))
	manager := Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.NoError(t, err)
	})

	t.Run("build cluster fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("fail")).AnyTimes()
		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("no computes", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{})
		assert.Error(t, err)
	})

	t.Run("async fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_StopCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, getEmptyFlow(constants.FlowStopCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq{
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_RestartCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, getEmptyFlow(constants.FlowRestartCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := manager.RestartCluster(context.TODO(), cluster.RestartClusterReq{
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.RestartCluster(context.TODO(), cluster.RestartClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.RestartCluster(context.TODO(), cluster.RestartClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_DeleteCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, getEmptyFlow(constants.FlowDeleteCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil).AnyTimes()
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := manager.DeleteCluster(context.TODO(), cluster.DeleteClusterReq{
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.DeleteCluster(context.TODO(), cluster.DeleteClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.DeleteCluster(context.TODO(), cluster.DeleteClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_DetailCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Type:              "TiDB",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "TiDB",
				Version:  "v5.0.0",
				Ports:    []int32{10001, 10002, 10003, 10004},
				HostIP:   []string{"127.0.0.1"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 3,
				Memory:   7,
				Type:     "TiDB",
				Version:  "v5.0.0",
				Ports:    []int32{10001, 10002, 10003, 10004},
				HostIP:   []string{"127.0.0.1"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "TiKV",
				Version:  "v5.0.0",
				Ports:    []int32{20001, 20002, 20003, 20004},
				HostIP:   []string{"127.0.0.2"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 3,
				Memory:   7,
				Type:     "TiKV",
				Version:  "v5.0.0",
				Ports:    []int32{20001, 20002, 20003, 20004},
				HostIP:   []string{"127.0.0.2"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "PD",
				Version:  "v5.0.0",
				Ports:    []int32{30001, 30002, 30003, 30004},
				HostIP:   []string{"127.0.0.3"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 3,
				Memory:   7,
				Type:     "PD",
				Version:  "v5.0.0",
				Ports:    []int32{30001, 30002, 30003, 30004},
				HostIP:   []string{"127.0.0.3"},
			},
		}, []*management.DBUser{
			{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType: string(constants.Root),
			},
		},nil)
		got, err := manager.DetailCluster(context.TODO(), cluster.QueryClusterDetailReq{
			ClusterID: "111",
		})
		assert.NoError(t, err)
		assert.Equal(t, got.Info.Region, "Region1")
		assert.Equal(t, len(got.ClusterTopologyInfo.Topology), 6)
	})

	t.Run("not found cluster", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(nil, nil, nil, fmt.Errorf("fail"))
		_, err := manager.DetailCluster(context.TODO(), cluster.QueryClusterDetailReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_RestoreNewCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	brService := mock_br_service.NewMockBRService(ctrl)
	brService.EXPECT().QueryClusterBackupRecords(gomock.Any(), gomock.Any()).Return(
		cluster.QueryBackupRecordsResp{
			BackupRecords: []*structs.BackupRecord{
				{
					Status: string(constants.ClusterBackupFinished),
				},
			},
		}, structs.Page{}, nil).AnyTimes()
	backuprestore.MockBRService(brService)
	defer backuprestore.MockBRService(backuprestore.NewBRManager())

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.RestoreNewCluster(context.TODO(), cluster.RestoreNewClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
			BackupID: "backup123",
		})
		assert.NoError(t, err)
	})

	t.Run("build cluster fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("fail")).AnyTimes()
		_, err := manager.RestoreNewCluster(context.TODO(), cluster.RestoreNewClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
			BackupID: "backup123",
		})
		assert.Error(t, err)
	})

	t.Run("no computes", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		_, err := manager.RestoreNewCluster(context.TODO(), cluster.RestoreNewClusterReq{})
		assert.Error(t, err)
	})

	t.Run("async fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.RestoreNewCluster(context.TODO(), cluster.RestoreNewClusterReq{
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
			BackupID: "backup123",
		})
		assert.Error(t, err)
	})
}

func TestManager_GetMonitorInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{
				ID: "2145635758",
			},
		}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Grafana",
				Version:  "v5.0.0",
				Ports:    []int32{50001, 50002},
				HostIP:   []string{"127.0.0.5"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "AlertManger",
				Version:  "v5.0.0",
				Ports:    []int32{60001, 60002},
				HostIP:   []string{"127.0.0.6"},
			},
		}, make([]*management.DBUser, 0), nil)
		manager := Manager{}
		got, err := manager.GetMonitorInfo(context.TODO(), cluster.QueryMonitorInfoReq{
			ClusterID: "2145635758",
		})
		assert.NoError(t, err)
		assert.Equal(t, got.ClusterID, "2145635758")
		assert.Equal(t, got.AlertUrl, "http://127.0.0.6:60001")
		assert.Equal(t, got.GrafanaUrl, "http://127.0.0.5:50001")
	})

	t.Run("err ip", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{
				ID: "2145635758",
			},
		}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Grafana",
				Version:  "v5.0.0",
				Ports:    []int32{50001, 50002},
				HostIP:   []string{"127.0.0.5"},
			},
		}, make([]*management.DBUser, 0), nil)
		manager := Manager{}
		_, err := manager.GetMonitorInfo(context.TODO(), cluster.QueryMonitorInfoReq{
			ClusterID: "2145635758",
		})
		assert.Error(t, err)
	})

	t.Run("grafana err port", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{
				ID: "2145635758",
			},
		}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Grafana",
				Version:  "v5.0.0",
				Ports:    []int32{-50001, -50002},
				HostIP:   []string{"127.0.0.5"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "AlertManger",
				Version:  "v5.0.0",
				Ports:    []int32{60001, 60002},
				HostIP:   []string{"127.0.0.6"},
			},
		}, make([]*management.DBUser, 0), nil)
		manager := Manager{}
		_, err := manager.GetMonitorInfo(context.TODO(), cluster.QueryMonitorInfoReq{
			ClusterID: "2145635758",
		})
		assert.Error(t, err)
	})

	t.Run("alert err port", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{
				ID: "2145635758",
			},
		}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Grafana",
				Version:  "v5.0.0",
				Ports:    []int32{50001, 50002},
				HostIP:   []string{"127.0.0.5"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "AlertManger",
				Version:  "v5.0.0",
				Ports:    []int32{-60001, -60002},
				HostIP:   []string{"127.0.0.6"},
			},
		}, make([]*management.DBUser, 0), nil)
		manager := Manager{}
		_, err := manager.GetMonitorInfo(context.TODO(), cluster.QueryMonitorInfoReq{
			ClusterID: "2145635758",
		})
		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, make([]*management.DBUser, 0), fmt.Errorf("not found"))
		manager := Manager{}
		_, err := manager.GetMonitorInfo(context.TODO(), cluster.QueryMonitorInfoReq{
			ClusterID: "2145635758",
		})
		assert.Error(t, err)
	})
}

func TestNewClusterManager(t *testing.T) {
	got := NewClusterManager()
	assert.NotNil(t, got)

}

func TestPreviewCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceRW := mockresource.NewMockReaderWriter(ctrl)
	models.SetResourceReaderWriter(resourceRW)

	provider := resourcepool.GetResourcePool().GetHostProvider().(*hostprovider.FileHostProvider)
	provider.SetResourceReaderWriter(resourceRW)

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)

	t.Run("normal", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{
		}, structs.Page{
			Page: 1,
			Total: 0,
			PageSize: 1,
		}, nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, nil).Times(1)

		manager := &Manager{}
		resp, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Region: "111",
				CpuArchitecture: "111",
			},
			ResourceParameter: structs.ClusterResourceInfo {
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Zone1", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
						{Zone: "Zone3", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
					}},
					{Type: "TiKV", Count: 4, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Zone1", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
						{Zone: "Zone2", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
					}},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resp.StockCheckResult))
		assert.True(t, resp.StockCheckResult[0].Enough)
		assert.False(t, resp.StockCheckResult[1].Enough)
		assert.False(t, resp.StockCheckResult[2].Enough)
		assert.True(t, resp.StockCheckResult[3].Enough)
	})
	t.Run("duplicated name", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{}, structs.Page{
			Page: 1,
			Total: 1,
			PageSize: 1,
		}, nil).Times(1)
		manager := &Manager{}
		_, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{
		})
		assert.Error(t, err)
	})
	t.Run("stock error", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{
		}, structs.Page{
			Page: 1,
			Total: 0,
			PageSize: 1,
		}, nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, errors.New("")).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{
		})
		assert.Error(t, err)
	})
}

func TestPreviewScaleOutCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceRW := mockresource.NewMockReaderWriter(ctrl)
	models.SetResourceReaderWriter(resourceRW)

	provider := resourcepool.GetResourcePool().GetHostProvider().(*hostprovider.FileHostProvider)
	provider.SetResourceReaderWriter(resourceRW)

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)

	t.Run("normal", func(t *testing.T) {
		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster01"}}, []*management.ClusterInstance{}, make([]*management.DBUser, 0), nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, nil).Times(1)

		manager := &Manager{}
		resp, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq {
			ClusterID: "111",
			ClusterResourceInfo: structs.ClusterResourceInfo {
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Zone1", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
						{Zone: "Zone3", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
					}},
					{Type: "TiKV", Count: 4, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Zone1", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
						{Zone: "Zone2", Count: 1, Spec: "4C8G", DiskCapacity: 1, DiskType: "SATA"},
					}},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resp.StockCheckResult))
		assert.True(t, resp.StockCheckResult[0].Enough)
		assert.False(t, resp.StockCheckResult[1].Enough)
		assert.False(t, resp.StockCheckResult[2].Enough)
		assert.True(t, resp.StockCheckResult[3].Enough)
	})
	t.Run("cluster is not existed", func(t *testing.T) {
		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster01"}}, []*management.ClusterInstance{}, make([]*management.DBUser, 0), errors.New("")).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq{
		})
		assert.Error(t, err)
	})
	t.Run("stock error", func(t *testing.T) {
		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster01"}}, []*management.ClusterInstance{}, make([]*management.DBUser, 0), nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount:8, FreeCpuCores:8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, errors.New("")).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq{
		})
		assert.Error(t, err)
	})
}

func TestManager_openSftpClient(t *testing.T) {
	t.Run("Dial err", func(t *testing.T) {
		_, _, err := openSftpClient(context.TODO(), cluster.TakeoverClusterReq{})
		assert.Error(t, err)
	})
}

func TestManager_TakeoverCluster(t *testing.T) {
	original := openSftpClient
	defer func() {
		openSftpClient = original
	}()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowTakeoverCluster, getEmptyFlow(constants.FlowTakeoverCluster))
	manager := Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, nil
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq {
			TiUPIp: "127.0.0.1",
			TiUPPath: ".tiup/",
			TiUPUserName: "root",
			TiUPUserPassword: "aaa",
			TiUPPort: 22,
			ClusterName: "takeoverCluster",
			DBUser: "dbUserName",
			DBPassword: "password",
		})
		assert.NoError(t, err)
	})

	t.Run("build cluster fail", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, nil
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("fail")).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq {
			TiUPIp: "127.0.0.1",
			TiUPPath: ".tiup/",
			TiUPUserName: "root",
			TiUPUserPassword: "aaa",
			TiUPPort: 22,
			ClusterName: "takeoverCluster",
			DBUser: "dbUserName",
			DBPassword: "password",
		})

		assert.Error(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq {
			TiUPIp: "127.0.0.1",
			TiUPPath: ".tiup/",
			TiUPUserName: "root",
			TiUPUserPassword: "aaa",
			TiUPPort: 22,
			DBUser: "dbUserName",
			DBPassword: "password",
		})
		assert.Error(t, err)
	})
	t.Run("open fail", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, errors.New("")
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq {
			TiUPIp: "127.0.0.1",
			TiUPPath: ".tiup/",
			TiUPUserName: "root",
			TiUPUserPassword: "aaa",
			TiUPPort: 22,
			ClusterName: "takeoverCluster",
			DBUser: "dbUserName",
			DBPassword: "password",
		})

		assert.Error(t, err)
	})
	t.Run("async fail", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, nil
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq {
			TiUPIp: "127.0.0.1",
			TiUPPath: ".tiup/",
			TiUPUserName: "root",
			TiUPUserPassword: "aaa",
			TiUPPort: 22,
			ClusterName: "takeoverCluster",
			DBUser: "dbUserName",
			DBPassword: "password",
		})

		assert.Error(t, err)
	})
}
