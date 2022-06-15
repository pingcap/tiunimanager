/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/pingcap/tiunimanager/message"
	"testing"
	"time"

	resourceManagement "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/management"
	mock_allocator_recycler "github.com/pingcap/tiunimanager/test/mockresource"

	"github.com/pingcap/tiunimanager/micro-cluster/cluster/parameter"

	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/constants"
	em_errors "github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/backuprestore"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool"
	"github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiunimanager/models/common"
	wfModel "github.com/pingcap/tiunimanager/models/workflow"
	mock_br_service "github.com/pingcap/tiunimanager/test/mockbr"
	mock_deployment "github.com/pingcap/tiunimanager/test/mockdeployment"
	mock_product "github.com/pingcap/tiunimanager/test/mockmodels"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockclustermanagement"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockresource"
	mock_workflow_service "github.com/pingcap/tiunimanager/test/mockworkflow"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
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
			"fail":  {"end", "", "", workflow.SyncFuncNode, emptyNode},
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		meta := &meta.ClusterMeta{
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", fmt.Errorf("create workflow error")).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		meta := &meta.ClusterMeta{
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(fmt.Errorf("start workflow error")).AnyTimes()
		meta := &meta.ClusterMeta{
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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		mockTiup := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiup
		mockTiup.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"true\"}", nil)
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
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
				Type:   "PD",
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
		mockTiup := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiup
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{4000},
			},
			{},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		mockTiup.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"false\"}", nil).AnyTimes()
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
		mockTiup := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiup
		mockTiup.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"true\"}", nil)
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
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

	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	mockQueryTiDBFromDBAnyTimes(productRW.EXPECT())

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Type: "TiDB", Version: "v5.2.2", CpuArchitecture: "x86_64"}, []*management.ClusterInstance{
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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Version: "v5.2.2"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, []*management.DBUser{
			{
				ClusterID: "111",
				Name:      constants.DBUserName[constants.Root],
				Password:  common.PasswordInExpired{Val: "123455678"},
				RoleType:  string(constants.Root),
			},
		}, nil).AnyTimes()
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		clusterRW.EXPECT().CreateRelation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v5.2.2",
			},
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

	t.Run("not found cluster", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Version: "v5.2.2"}, []*management.ClusterInstance{
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
				Version:    "v5.2.2",
			},
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

	t.Run("clone meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Version: "v5.2.2"},
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

	t.Run("sync fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{Version: "v5.2.2"}, []*management.ClusterInstance{
			{Entity: common.Entity{ID: "instance01"}, Type: "TiDB"},
			{Entity: common.Entity{ID: "instance02"}, Type: "TiFlash"},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		clusterRW.EXPECT().CreateRelation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		_, err := manager.Clone(context.TODO(), cluster.CloneClusterReq{
			SourceClusterID: "111",
			CreateClusterParameter: structs.CreateClusterParameter{
				Name:       "testCluster",
				DBUser:     "user01",
				DBPassword: "password01",
				Type:       "TiDB",
				Version:    "v5.2.2",
			},
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

func TestManager_CreateCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, getEmptyFlow(constants.FlowCreateCluster))
	manager := Manager{}

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
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
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
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
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{})
		assert.Error(t, err)
	})

	t.Run("async fail", func(t *testing.T) {
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
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

	t.Run("validate", func(t *testing.T) {
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return em_errors.Error(em_errors.TIUNIMANAGER_PARAMETER_INVALID)
		}
		defer func() {
			validator = validateCreating
		}()

		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq{})
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
				Password:  common.PasswordInExpired{Val: "123455678"},
				RoleType:  string(constants.Root),
			},
		}, nil)
		clusterRW.EXPECT().GetMasters(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
			{
				SubjectClusterID: "01",
				ObjectClusterID:  "02",
			},
		}, nil)
		clusterRW.EXPECT().GetSlaves(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
			{
				SubjectClusterID: "01",
				ObjectClusterID:  "02",
			},
		}, nil)
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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
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
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		_, err := manager.RestoreNewCluster(context.TODO(), cluster.RestoreNewClusterReq{})
		assert.Error(t, err)
	})

	t.Run("async fail", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
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

	t.Run("duplicated name", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{}, structs.Page{
			Page:     1,
			Total:    1,
			PageSize: 1,
		}, nil).Times(1)
		manager := &Manager{}
		_, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{})
		assert.Error(t, err)
	})

	t.Run("validate", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{}, structs.Page{
			Page:     1,
			Total:    0,
			PageSize: 1,
		}, nil).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{
			CreateClusterParameter: structs.CreateClusterParameter{
				Region:          "111",
				CpuArchitecture: "111",
			},
			ResourceParameter: structs.ClusterResourceInfo{
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
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{}, structs.Page{
			Page:     1,
			Total:    0,
			PageSize: 1,
		}, nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, nil).Times(1)

		manager := &Manager{}
		resp, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{
			CreateClusterParameter: structs.CreateClusterParameter{
				Region:          "111",
				CpuArchitecture: "111",
			},
			ResourceParameter: structs.ClusterResourceInfo{
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

	t.Run("stock error", func(t *testing.T) {
		clusterRW.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.Result{}, structs.Page{
			Page:     1,
			Total:    0,
			PageSize: 1,
		}, nil).Times(1)

		validator = func(ctx context.Context, req *cluster.CreateClusterReq) error {
			return nil
		}
		defer func() {
			validator = validateCreating
		}()
		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, errors.New("")).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewCluster(context.TODO(), cluster.CreateClusterReq{})
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
			{Zone: "Zone1", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, nil).Times(1)

		manager := &Manager{}
		resp, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq{
			ClusterID: "111",
			ClusterResourceInfo: structs.ClusterResourceInfo{
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
		_, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq{})
		assert.Error(t, err)
	})
	t.Run("stock error", func(t *testing.T) {
		clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster01"}}, []*management.ClusterInstance{}, make([]*management.DBUser, 0), nil).Times(1)

		resourceRW.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Stocks{
			{Zone: "Zone1", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
			{Zone: "Zone2", FreeHostCount: 8, FreeCpuCores: 8, FreeMemory: 8, FreeDiskCount: 8, FreeDiskCapacity: 8},
		}, errors.New("")).Times(1)

		manager := &Manager{}
		_, err := manager.PreviewScaleOutCluster(context.TODO(), cluster.ScaleOutClusterReq{})
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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, nil
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		clusterRW.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq{
			TiUPIp:           "127.0.0.1",
			TiUPPath:         ".tiup/",
			TiUPUserName:     "root",
			TiUPUserPassword: "aaa",
			TiUPPort:         22,
			ClusterName:      "takeoverCluster",
			DBPassword:       "password",
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
		clusterRW.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq{
			TiUPIp:           "127.0.0.1",
			TiUPPath:         ".tiup/",
			TiUPUserName:     "root",
			TiUPUserPassword: "aaa",
			TiUPPort:         22,
			ClusterName:      "takeoverCluster",
			DBPassword:       "password",
		})

		assert.Error(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		clusterRW.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq{
			TiUPIp:           "127.0.0.1",
			TiUPPath:         ".tiup/",
			TiUPUserName:     "root",
			TiUPUserPassword: "aaa",
			TiUPPort:         22,
			DBPassword:       "password",
		})
		assert.Error(t, err)
	})
	t.Run("open fail", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, errors.New("")
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		clusterRW.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		models.SetClusterReaderWriter(clusterRW)
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq{
			TiUPIp:           "127.0.0.1",
			TiUPPath:         ".tiup/",
			TiUPUserName:     "root",
			TiUPUserPassword: "aaa",
			TiUPPort:         22,
			ClusterName:      "takeoverCluster",
			DBPassword:       "password",
		})

		assert.Error(t, err)
	})
	t.Run("async fail", func(t *testing.T) {
		openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
			return nil, nil, nil
		}
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster"}}, nil).AnyTimes()
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("fail")).AnyTimes()
		clusterRW.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := manager.Takeover(context.TODO(), cluster.TakeoverClusterReq{
			TiUPIp:           "127.0.0.1",
			TiUPPath:         ".tiup/",
			TiUPUserName:     "root",
			TiUPUserPassword: "aaa",
			TiUPPort:         22,
			ClusterName:      "takeoverCluster",
			DBPassword:       "password",
		})

		assert.Error(t, err)
	})
}

func TestManager_QueryProductUpdatePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := Manager{}
	t.Run("not found meta", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.QueryProductUpdatePath(context.TODO(), "111")
		assert.Error(t, err)
	})
}

func Test_generatePaths(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		resp := message.QueryProductsInfoResp{
			Products: []structs.ProductConfigInfo{
				{
					ProductID:   "TiDB",
					ProductName: "TiDB",
					Versions: []structs.SpecificVersionProduct{
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.0.0",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.1.0",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.2.0",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.2.2",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.2.3",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.3.0",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.3.1",
						},
						{
							ProductID: "TiDB",
							Arch:      "X86_64",
							Version:   "v5.4.0",
						},
					},
				},
			},
		}
		paths := generatePaths(context.TODO(), resp, "v5.2.2", "X86_64")
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, string(constants.UpgradeTypeInPlace), paths[0].UpgradeType)
		assert.Equal(t, 4, len(paths[0].Versions))
		assert.True(t, meta.Contain(paths[0].Versions, "v5.2.3"))
		assert.True(t, meta.Contain(paths[0].Versions, "v5.3.0"))
		assert.True(t, meta.Contain(paths[0].Versions, "v5.3.1"))
		assert.True(t, meta.Contain(paths[0].Versions, "v5.4.0"))
		assert.Equal(t, 2, len(paths[0].UpgradeWays))
	})
	t.Run("wrong product id", func(t *testing.T) {
		resp := message.QueryProductsInfoResp{
			Products: []structs.ProductConfigInfo{
				{
					ProductID:   "DM",
					ProductName: "DM",
					Versions: []structs.SpecificVersionProduct{
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.0.0",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.1.0",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.2.0",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.2.2",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.2.3",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.3.0",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.3.1",
						},
						{
							ProductID: "DM",
							Arch:      "X86_64",
							Version:   "v5.4.0",
						},
					},
				},
			},
		}
		paths := generatePaths(context.TODO(), resp, "v5.2.2", "X86_64")
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, string(constants.UpgradeTypeInPlace), paths[0].UpgradeType)
		assert.Equal(t, 0, len(paths[0].Versions))
		assert.Equal(t, 2, len(paths[0].UpgradeWays))
	})
}

func Test_getFullVersion(t *testing.T) {
	t.Run("len 1", func(t *testing.T) {
		version := getFullVersion("v5")
		assert.Equal(t, "v5.0.0", version)
	})
	t.Run("len 2", func(t *testing.T) {
		version := getFullVersion("v5.1")
		assert.Equal(t, "v5.1.0", version)
	})
}

func TestManager_QueryUpgradeVersionDiffInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := Manager{}
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.QueryUpgradeVersionDiffInfo(context.TODO(), "111", "v5.4.0")
		assert.Error(t, err)
	})
}

func Test_getMinorVersion(t *testing.T) {
	t.Run("len 1", func(t *testing.T) {
		version := getMinorVersion("v5")
		assert.Equal(t, "v5", version)
	})
	t.Run("len 3", func(t *testing.T) {
		version := getMinorVersion("v5.2.2")
		assert.Equal(t, "v5.2", version)
	})
}

func Test_compareConfigDifference(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cParamInfos := []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic1",
				Name:         "param1",
				InstanceType: "tidb",
				RealValue: structs.ParameterRealValue{
					ClusterValue: "v1_real",
				},
				Type:        1,
				Unit:        "MB",
				UnitOptions: []string{"KB", "MB", "GB"},
				Range:       []string{"1, 100"},
				RangeType:   1,
				HasApply:    int(parameter.DirectApply),
				Description: "param1 desc",
			},
			{
				ParamId:      "2",
				Category:     "basic2",
				Name:         "param2",
				InstanceType: "tikv",
				RealValue: structs.ParameterRealValue{
					ClusterValue: "v2",
				},
				Type:        2,
				Unit:        "GB",
				UnitOptions: []string{"KB", "MB", "GB"},
				Range:       []string{"1, 1000"},
				RangeType:   1,
				HasApply:    int(parameter.DirectApply),
				Description: "param2 desc",
			},
			{
				ParamId:      "3",
				Category:     "basic3",
				Name:         "param3",
				InstanceType: "pd",
				RealValue: structs.ParameterRealValue{
					ClusterValue: "",
				},
				Type:        3,
				Unit:        "KB",
				UnitOptions: []string{"KB", "MB", "GB"},
				Range:       []string{"1, 10000"},
				RangeType:   1,
				HasApply:    int(parameter.DirectApply),
				Description: "param3 desc",
			},
			{
				ParamId:      "4",
				Category:     "basic4",
				Name:         "param4",
				InstanceType: "tiflash",
				RealValue: structs.ParameterRealValue{
					ClusterValue: "v4_real",
				},
				Type:        4,
				Unit:        "KB",
				UnitOptions: []string{"KB", "MB", "GB"},
				Range:       []string{"1, 100000"},
				RangeType:   1,
				HasApply:    int(parameter.DirectApply),
				Description: "param4 desc",
			},
			{
				ParamId:      "5",
				Category:     "basic5",
				Name:         "param5",
				InstanceType: "tidb",
				RealValue: structs.ParameterRealValue{
					ClusterValue: "v5_real",
				},
				Type:        5,
				Unit:        "KB",
				UnitOptions: []string{"KB", "MB", "GB"},
				Range:       []string{"1, 100000"},
				RangeType:   1,
				HasApply:    int(parameter.ModifyApply),
				Description: "param5 desc",
			},
		}

		pgParamInfos := []structs.ParameterGroupParameterInfo{
			{
				ID:           "1",
				Category:     "basic1",
				Name:         "param1",
				InstanceType: "tidb",
				DefaultValue: "v1new",
				Type:         1,
				Unit:         "MB",
				UnitOptions:  []string{"KB", "MB", "GB"},
				Range:        []string{"1, 100"},
				RangeType:    1,
				HasApply:     int(parameter.DirectApply),
				Description:  "param1 desc",
			},
			{
				ID:           "2",
				Category:     "basic2",
				Name:         "param2",
				InstanceType: "tikv",
				DefaultValue: "v2",
				Type:         2,
				Unit:         "GB",
				UnitOptions:  []string{"KB", "MB", "GB"},
				Range:        []string{"1, 1000"},
				RangeType:    1,
				HasApply:     int(parameter.DirectApply),
				Description:  "param2 desc",
			},
			{
				ID:           "3",
				Category:     "basic3",
				Name:         "param3",
				InstanceType: "pd",
				DefaultValue: " ",
				Type:         3,
				Unit:         "KB",
				UnitOptions:  []string{"KB", "MB", "GB"},
				Range:        []string{"1, 10000"},
				RangeType:    1,
				HasApply:     int(parameter.DirectApply),
				Description:  "param3 desc",
			},
			{
				ID:           "4",
				Category:     "basic4",
				Name:         "param4",
				InstanceType: "tiflash",
				DefaultValue: "v4new",
				Type:         4,
				Unit:         "KB",
				UnitOptions:  []string{"KB", "MB", "GB"},
				Range:        []string{"1, 100000"},
				RangeType:    1,
				HasApply:     int(parameter.DirectApply),
				Description:  "param4 desc",
			},
			{
				ID:           "5",
				Category:     "basic5",
				Name:         "param5",
				InstanceType: "tidb",
				DefaultValue: "v5new",
				Type:         5,
				Unit:         "KB",
				UnitOptions:  []string{"KB", "MB", "GB"},
				Range:        []string{"1, 100000"},
				RangeType:    1,
				HasApply:     int(parameter.ModifyApply),
				Description:  "param5 desc",
			},
		}

		resp := compareConfigDifference(context.TODO(), cParamInfos, pgParamInfos, []string{"tidb", "tikv", "pd"})
		assert.Equal(t, 1, len(resp))
		item := structs.ProductUpgradeVersionConfigDiffItem{
			ParamId:      "1",
			Category:     "basic1",
			Name:         "param1",
			InstanceType: "tidb",
			CurrentValue: "v1_real",
			SuggestValue: "v1new",
			Type:         1,
			Unit:         "MB",
			UnitOptions:  []string{"KB", "MB", "GB"},
			Range:        []string{"1, 100"},
			RangeType:    1,
			Description:  "param1 desc",
		}
		assert.Equal(t, 1, len(resp))
		assert.Equal(t, item, *resp[0])
	})
	//t.Run("sort", func(t *testing.T) {
	//	cParamInfos := []structs.ClusterParameterInfo{
	//		{
	//			ParamId:      "1",
	//			Category:     "basic1",
	//			Name:         "param1",
	//			InstanceType: "tidb",
	//			RealValue: structs.ParameterRealValue{
	//				ClusterValue: "v1_real",
	//			},
	//			Type:        1,
	//			Unit:        "MB",
	//			UnitOptions: []string{"KB", "MB", "GB"},
	//			Range:       []string{"1, 100"},
	//			RangeType:   1,
	//			HasApply:    int(parameter.DirectApply),
	//			Description: "param1 desc",
	//		},
	//		{
	//			ParamId:      "2",
	//			Category:     "basic2",
	//			Name:         "param2",
	//			InstanceType: "tikv",
	//			RealValue: structs.ParameterRealValue{
	//				ClusterValue: "v2_real",
	//			},
	//			Type:        2,
	//			Unit:        "GB",
	//			UnitOptions: []string{"KB", "MB", "GB"},
	//			Range:       []string{"1, 1000"},
	//			RangeType:   1,
	//			HasApply:    int(parameter.DirectApply),
	//			Description: "param2 desc",
	//		},
	//		{
	//			ParamId:      "3",
	//			Category:     "basicB",
	//			Name:         "param3",
	//			InstanceType: "pd",
	//			RealValue: structs.ParameterRealValue{
	//				ClusterValue: "v3_real",
	//			},
	//			Type:        3,
	//			Unit:        "KB",
	//			UnitOptions: []string{"KB", "MB", "GB"},
	//			Range:       []string{"1, 10000"},
	//			RangeType:   1,
	//			HasApply:    int(parameter.DirectApply),
	//			Description: "param3 desc",
	//		},
	//		{
	//			ParamId:      "4",
	//			Category:     "basicA",
	//			Name:         "param4",
	//			InstanceType: "pd",
	//			RealValue: structs.ParameterRealValue{
	//				ClusterValue: "v4_real",
	//			},
	//			Type:        4,
	//			Unit:        "KB",
	//			UnitOptions: []string{"KB", "MB", "GB"},
	//			Range:       []string{"1, 100000"},
	//			RangeType:   1,
	//			HasApply:    int(parameter.DirectApply),
	//			Description: "param4 desc",
	//		},
	//	}
	//
	//	pgParamInfos := []structs.ParameterGroupParameterInfo{
	//		{
	//			ID:           "1",
	//			Category:     "basic1",
	//			Name:         "param1",
	//			InstanceType: "tidb",
	//			DefaultValue: "v1new",
	//			Type:         1,
	//			Unit:         "MB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 100"},
	//			RangeType:    1,
	//			HasApply:     int(parameter.DirectApply),
	//			Description:  "param1 desc",
	//		},
	//		{
	//			ID:           "2",
	//			Category:     "basic2",
	//			Name:         "param2",
	//			InstanceType: "tikv",
	//			DefaultValue: "v2new",
	//			Type:         2,
	//			Unit:         "GB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 1000"},
	//			RangeType:    1,
	//			HasApply:     int(parameter.DirectApply),
	//			Description:  "param2 desc",
	//		},
	//		{
	//			ID:           "3",
	//			Category:     "basicB",
	//			Name:         "param3",
	//			InstanceType: "pd",
	//			DefaultValue: "v3new",
	//			Type:         3,
	//			Unit:         "KB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 10000"},
	//			RangeType:    1,
	//			HasApply:     int(parameter.DirectApply),
	//			Description:  "param3 desc",
	//		},
	//		{
	//			ID:           "4",
	//			Category:     "basicA",
	//			Name:         "param4",
	//			InstanceType: "pd",
	//			DefaultValue: "v4new",
	//			Type:         4,
	//			Unit:         "KB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 100000"},
	//			RangeType:    1,
	//			HasApply:     int(parameter.DirectApply),
	//			Description:  "param4 desc",
	//		},
	//	}
	//
	//	resp := compareConfigDifference(context.TODO(), cParamInfos, pgParamInfos, []string{"tidb", "tikv", "pd"})
	//	items := []structs.ProductUpgradeVersionConfigDiffItem{
	//		{
	//			ParamId:      "4",
	//			Category:     "basicA",
	//			Name:         "param4",
	//			InstanceType: "pd",
	//			CurrentValue: "v4_real",
	//			SuggestValue: "v4new",
	//			Type:         4,
	//			Unit:         "KB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 100000"},
	//			RangeType:    1,
	//			Description:  "param4 desc",
	//		},
	//		{
	//			ParamId:      "3",
	//			Category:     "basicB",
	//			Name:         "param3",
	//			InstanceType: "pd",
	//			CurrentValue: "v3_real",
	//			SuggestValue: "v3new",
	//			Type:         3,
	//			Unit:         "KB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 10000"},
	//			RangeType:    1,
	//			Description:  "param3 desc",
	//		},
	//		{
	//			ParamId:      "1",
	//			Category:     "basic1",
	//			Name:         "param1",
	//			InstanceType: "tidb",
	//			CurrentValue: "v1_real",
	//			SuggestValue: "v1new",
	//			Type:         1,
	//			Unit:         "MB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 100"},
	//			RangeType:    1,
	//			Description:  "param1 desc",
	//		},
	//		{
	//			ParamId:      "2",
	//			Category:     "basic2",
	//			Name:         "param2",
	//			InstanceType: "tikv",
	//			CurrentValue: "v2_real",
	//			SuggestValue: "v2new",
	//			Type:         2,
	//			Unit:         "GB",
	//			UnitOptions:  []string{"KB", "MB", "GB"},
	//			Range:        []string{"1, 1000"},
	//			RangeType:    1,
	//			Description:  "param2 desc",
	//		},
	//	}
	//	fmt.Printf("real: \n%s %s\n%s %s\n%s %s\n%s %s\n", resp[0].InstanceType, resp[0].Category, resp[1].InstanceType, resp[1].Category, resp[2].InstanceType, resp[2].Category, resp[3].InstanceType, resp[3].Category)
	//	assert.Equal(t, 4, len(resp))
	//	assert.Equal(t, items[0].InstanceType, resp[0].InstanceType)
	//	assert.Equal(t, items[1].InstanceType, resp[1].InstanceType)
	//	assert.Equal(t, items[2].InstanceType, resp[2].InstanceType)
	//	assert.Equal(t, items[3].InstanceType, resp[3].InstanceType)
	//})
}

func TestManager_InPlaceUpgradeCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowOnlineInPlaceUpgradeCluster, getEmptyFlow(constants.FlowOnlineInPlaceUpgradeCluster))
	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowOfflineInPlaceUpgradeCluster, getEmptyFlow(constants.FlowOfflineInPlaceUpgradeCluster))

	manager := Manager{}
	t.Run("normal offline", func(t *testing.T) {
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := manager.InPlaceUpgradeCluster(context.TODO(), cluster.UpgradeClusterReq{
			ClusterID:  "111",
			UpgradeWay: string(constants.UpgradeWayOffline),
		})
		assert.NoError(t, err)
	})
	t.Run("normal online", func(t *testing.T) {
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
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := manager.InPlaceUpgradeCluster(context.TODO(), cluster.UpgradeClusterReq{
			ClusterID:  "111",
			UpgradeWay: string(constants.UpgradeWayOnline),
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, nil, errors.New(""))

		_, err := manager.InPlaceUpgradeCluster(context.TODO(), cluster.UpgradeClusterReq{
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

		_, err := manager.InPlaceUpgradeCluster(context.TODO(), cluster.UpgradeClusterReq{
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_DeleteMetadataPhysically(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().ClearClusterPhysically(gomock.Any(), "", gomock.Any()).Return(em_errors.Error(em_errors.TIUNIMANAGER_PARAMETER_INVALID)).AnyTimes()
	clusterRW.EXPECT().ClearClusterPhysically(gomock.Any(), "111", gomock.Any()).Return(nil).AnyTimes()

	resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
	resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)

	manager := &Manager{}
	t.Run("normal", func(t *testing.T) {
		_, err := manager.DeleteMetadataPhysically(context.TODO(), cluster.DeleteMetadataPhysicallyReq{
			ClusterID: "111",
			Reason:    "no why",
		})
		assert.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		_, err := manager.DeleteMetadataPhysically(context.TODO(), cluster.DeleteMetadataPhysicallyReq{})
		assert.Error(t, err)
	})
}
