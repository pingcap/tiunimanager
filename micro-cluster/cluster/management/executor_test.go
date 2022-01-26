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
	"fmt"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/changefeed"
	"github.com/pingcap-inc/tiem/test/mockchangefeed"
	"strconv"

	"github.com/pingcap-inc/tiem/models/parametergroup"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockparametergroup"
	"reflect"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"os"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockresource"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	structs2 "github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/tiup"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_br_service "github.com/pingcap-inc/tiem/test/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclusterparameter"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocktiupconfig"
	mock_allocator_recycler "github.com/pingcap-inc/tiem/test/mockresource"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestPrepareResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				Status:    string(constants.ClusterInitializing),
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
			Exclusive:         false,
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiDB",
					Version:      "v5.0.0",
					Ports:        []int32{10001, 10002, 10003, 10004},
					HostIP:       []string{"127.0.0.1", "127.0.0.6"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiKV",
					Version:      "v5.0.0",
					Ports:        []int32{20001, 20002, 20003, 20004},
					HostIP:       []string{"127.0.0.2"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"PD": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "PD",
					Version:      "v5.0.0",
					Ports:        []int32{30001, 30002, 30003, 30004},
					HostIP:       []string{"127.0.0.3"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
		},
	})
	t.Run("normal", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().AllocResources(gomock.Any(), gomock.Any()).Return(&structs.BatchAllocResponse{
			BatchResults: []*structs.AllocRsp{
				{
					Applicant: structs.Applicant{
						RequestId: "123",
					},
				},
			},
		}, nil)
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := prepareResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("alloc failed", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().AllocResources(gomock.Any(),
			gomock.Any()).Return(nil, fmt.Errorf("alloc failed"))
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := prepareResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestBuildConfig(t *testing.T) {
	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:     "111",
				Status: string(constants.ClusterRunning),
			},
			Version: "v4.1.1",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{
						1, 2, 3, 4, 5, 6,
					},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{
						1, 2, 3, 4, 5, 6,
					},
				},
			},
		},
	})
	err := buildConfig(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:     "111",
				Status: string(constants.ClusterRunning),
			},
			Version: "v4.1.1",
		},
		Instances: map[string][]*management.ClusterInstance{},
	})
	err = buildConfig(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestScaleOutCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ClusterScaleOut(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
	secondparty.Manager = mockTiupManager

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
			Version: "v5.0.0",
		},
	})

	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterScaleOut(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext.SetData(ContextTopology, "test topology")
		err := scaleOutCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("topology not found", func(t *testing.T) {
		flowContext.SetData(ContextTopology, nil)
		err := scaleOutCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("scale out fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterScaleOut(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext.SetData(ContextTopology, "test topology")
		err := scaleOutCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestScaleInCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
			Version: "v5.0.0",
			Type:    "TiDB",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						ID: "instance01",
					},
					Type:   "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8000},
				},
				{
					Entity: common.Entity{
						ID: "instance02",
					},
					Type:   "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{
						ID: "instance03",
					},
					Type:   "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						ID: "instance04",
					},
					Type:   "TiKV",
					HostIP: []string{"127.0.0.2"},
					Ports:  []int32{8001},
				},
			},
		},
	})
	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterScaleIn(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager
		flowContext.SetData(ContextInstanceID, "instance01")
		err := scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		flowContext.SetData(ContextInstanceID, "instance05")
		err := scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("scale in fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterScaleIn(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager
		flowContext.SetData(ContextInstanceID, "instance02")
		err := scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestFreeInstanceResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().DeleteInstance(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "123",
				},
			},
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						Entity: common.Entity{
							ID: "111",
						},
						ClusterID: "123",
						HostID:    "123",
						CpuCores:  8,
						Memory:    12,
						DiskID:    "id12",
						Ports:     []int32{12, 34},
					},
				},
			},
		})
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(nil)
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		flowContext.SetData(ContextInstanceID, "111")
		err := freeInstanceResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("delete not found", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "123",
				},
			},
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						Entity: common.Entity{
							ID: "111",
						},
						ClusterID: "123",
						HostID:    "123",
						CpuCores:  8,
						Memory:    12,
						DiskID:    "id12",
						Ports:     []int32{12, 34},
					},
				},
			},
		})
		flowContext.SetData(ContextInstanceID, "112")
		err := freeInstanceResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("recycle fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "123",
				},
			},
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						Entity: common.Entity{
							ID: "111",
						},
						ClusterID: "123",
						HostID:    "123",
						CpuCores:  8,
						Memory:    12,
						DiskID:    "id12",
						Ports:     []int32{12, 34},
					},
				},
			},
		})
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(fmt.Errorf("recycle fail"))
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		flowContext.SetData(ContextInstanceID, "111")
		err := freeInstanceResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestClearBackupData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
	})
	flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{KeepHistoryBackupRecords: false})

	t.Run("normal", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().DeleteBackupStrategy(gomock.Any(), gomock.Any()).Return(cluster.DeleteBackupStrategyResp{}, nil).AnyTimes()
		brService.EXPECT().DeleteBackupRecords(gomock.Any(), gomock.Any()).Return(cluster.DeleteBackupDataResp{}, nil).AnyTimes()
		err := clearBackupData(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("delete backup strategy fail", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().DeleteBackupStrategy(gomock.Any(),
			gomock.Any()).Return(cluster.DeleteBackupStrategyResp{}, fmt.Errorf("delete backup strategy fail")).AnyTimes()
		err := clearBackupData(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("delete backup data fail", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().DeleteBackupStrategy(gomock.Any(),
			gomock.Any()).Return(cluster.DeleteBackupStrategyResp{}, nil).AnyTimes()
		brService.EXPECT().DeleteBackupRecords(gomock.Any(),
			gomock.Any()).Return(cluster.DeleteBackupDataResp{}, fmt.Errorf("delete backup data fail")).AnyTimes()
		err := clearBackupData(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestBackupBeforeDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetCurrentClusterTopologySnapshot(gomock.Any(), "testCluster").Return(management.ClusterTopologySnapshot{}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{AutoBackup: true})
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusFinished}}, nil).AnyTimes()
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().BackupCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.BackupClusterDataResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				}}, nil)
		err := backupBeforeDelete(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("backup fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{AutoBackup: true})
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusFinished}}, nil).AnyTimes()
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().BackupCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.BackupClusterDataResp{}, fmt.Errorf("backup fail"))
		err := backupBeforeDelete(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("wait workflow fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{AutoBackup: true})
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusError}}, nil).AnyTimes()
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().BackupCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.BackupClusterDataResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				}}, nil)
		err := backupBeforeDelete(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("no backup", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{AutoBackup: false})
		err := backupBeforeDelete(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("skip", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "skip",
				},
			},
		})
		flowContext.SetData(ContextDeleteRequest, cluster.DeleteClusterReq{AutoBackup: false})

		clusterRW.EXPECT().GetCurrentClusterTopologySnapshot(gomock.Any(), "skip").Return(management.ClusterTopologySnapshot{}, errors.NewError(errors.TIEM_PANIC, "")).Times(1)
		node := &workflowModel.WorkFlowNode{}
		err := backupBeforeDelete(node, flowContext)
		assert.NoError(t, err)
		assert.NotEmpty(t, node.Result)
	})

}

func TestApplyParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				ParameterGroupID: "211",
			},
		})
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusFinished}}, nil).AnyTimes()

	})

}

func TestBackupSourceCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextSourceClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
	})
	flowContext.SetData(ContextCloneStrategy, string(constants.SnapShotClone))
	t.Run("normal", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().BackupCluster(gomock.Any(),
			gomock.Any(), true).Return(
			cluster.BackupClusterDataResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				},
				BackupID: "123",
			}, nil)
		err := backupSourceCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("backup fail", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().BackupCluster(gomock.Any(),
			gomock.Any(), true).Return(
			cluster.BackupClusterDataResp{}, fmt.Errorf("backup fail"))
		err := backupSourceCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestRestoreNewCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextBackupID, "backup123")
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().RestoreExistCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.RestoreExistClusterResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				},
			}, nil)
		err := restoreNewCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("restore fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextBackupID, "backup123")
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().RestoreExistCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.RestoreExistClusterResp{}, fmt.Errorf("restore fail"))
		err := restoreNewCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("no backup id", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		err := restoreNewCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})
}

func TestWaitWorkFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no found workflow", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		err := waitWorkFlow(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextWorkflowID, "111")
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusFinished}}, nil).AnyTimes()
		err := waitWorkFlow(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("workflow fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextWorkflowID, "111")
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs2.WorkFlowInfo{
					Status: constants.WorkFlowStatusError}}, nil).AnyTimes()
		err := waitWorkFlow(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestSetClusterFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
	})
	err := setClusterFailure(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(),
		gomock.Any()).Return(fmt.Errorf("update status fail"))
	err = setClusterFailure(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestSetClusterOnline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
				},
			},
		},
	})

	err := setClusterOnline(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(),
		gomock.Any()).Return(fmt.Errorf("update status fail"))
	err = setClusterOnline(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestSetClusterOffline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
			},
		},
	})

	err := setClusterOffline(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	clusterRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(),
		gomock.Any()).Return(fmt.Errorf("update status fail"))
	err = setClusterOffline(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestRevertResourceAfterFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextAllocResource, &structs.BatchAllocResponse{
			BatchResults: []*structs.AllocRsp{
				{
					Applicant: structs.Applicant{
						RequestId: "123",
					},
				},
			},
		})
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(nil)
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := revertResourceAfterFailure(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("recycle fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextAllocResource, &structs.BatchAllocResponse{
			BatchResults: []*structs.AllocRsp{
				{
					Applicant: structs.Applicant{
						RequestId: "123",
					},
				},
			},
		})
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(fmt.Errorf("recycle fail"))
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := revertResourceAfterFailure(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("no alloc response", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		err := revertResourceAfterFailure(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})
}

func TestEndMaintenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
			MaintenanceStatus: constants.ClusterMaintenanceTakeover,
		},
	})

	err := endMaintenance(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(),
		gomock.Any()).Return(fmt.Errorf("clear maintenance status fail"))
	err = endMaintenance(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestPersistCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().UpdateMeta(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
			},
		},
	})
	err := persistCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	clusterRW.EXPECT().UpdateMeta(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update meta fail"))
	err = persistCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestDeployCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterDeploy(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		flowContext.SetData(ContextTopology, "test topology")
		err := deployCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("no topology", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := deployCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("deploy fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterDeploy(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return("task01", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		flowContext.SetData(ContextTopology, "test topology")
		err := deployCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestStartCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterStart(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := startCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("start fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterStart(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := startCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestSyncBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextSourceClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "sourceCluster",
			},
		},
	})
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "targetCluster",
			},
		},
	})

	t.Run("normal", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().GetBackupStrategy(gomock.Any(),
			gomock.Any()).Return(
			cluster.GetBackupStrategyResp{
				Strategy: structs2.BackupStrategy{
					BackupDate: "2021-12-23",
				},
			}, nil)
		brService.EXPECT().SaveBackupStrategy(gomock.Any(),
			gomock.Any()).Return(cluster.SaveBackupStrategyResp{}, nil)
		err := syncBackupStrategy(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("get strategy fail", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().GetBackupStrategy(gomock.Any(),
			gomock.Any()).Return(
			cluster.GetBackupStrategyResp{}, fmt.Errorf("get backup strategy fail"))
		err := syncBackupStrategy(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("no backup strategy", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().GetBackupStrategy(gomock.Any(),
			gomock.Any()).Return(
			cluster.GetBackupStrategyResp{
				Strategy: structs2.BackupStrategy{
					BackupDate: "",
				},
			}, nil)
		err := syncBackupStrategy(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("save strategy fail", func(t *testing.T) {
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().GetBackupStrategy(gomock.Any(),
			gomock.Any()).Return(
			cluster.GetBackupStrategyResp{
				Strategy: structs2.BackupStrategy{
					BackupDate: "2021-12-23",
				},
			}, nil)
		brService.EXPECT().SaveBackupStrategy(gomock.Any(),
			gomock.Any()).Return(cluster.SaveBackupStrategyResp{}, fmt.Errorf("save backup strategy fail"))
		err := syncBackupStrategy(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestSyncParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
			return "1", []*parameter.ClusterParamDetail{}, 1, fmt.Errorf("query cluster fail")
		})

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextSourceClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "sourceCluster",
			},
		},
	})
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "targetCluster",
			},
		},
	})

	err := syncParameters(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestRestoreCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextBackupID, "backup123")
		flowContext.SetData(ContextCloneStrategy, string(constants.SnapShotClone))
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().RestoreExistCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.RestoreExistClusterResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				},
			}, nil)
		err := restoreCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("restore fail", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextBackupID, "backup123")
		flowContext.SetData(ContextCloneStrategy, string(constants.SnapShotClone))
		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().RestoreExistCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.RestoreExistClusterResp{}, fmt.Errorf("restore fail"))
		err := restoreCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("no backup id", func(t *testing.T) {
		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
			},
		})
		flowContext.SetData(ContextCloneStrategy, string(constants.SnapShotClone))

		err := restoreCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})
}

func TestStopCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterStop(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := stopCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("stop fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterStop(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := stopCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestDestroyCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetCurrentClusterTopologySnapshot(gomock.Any(), "testCluster").Return(management.ClusterTopologySnapshot{}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterDestroy(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := destroyCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("destroy fail", func(t *testing.T) {
		mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		mockTiupManager.EXPECT().ClusterDestroy(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", fmt.Errorf("fail")).AnyTimes()
		secondparty.Manager = mockTiupManager

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "testCluster",
				},
				Version: "v5.0.0",
			},
		})
		err := destroyCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("skip", func(t *testing.T) {
		clusterRW.EXPECT().GetCurrentClusterTopologySnapshot(gomock.Any(), "skip").Return(management.ClusterTopologySnapshot{}, errors.NewError(errors.TIEM_PANIC, "")).AnyTimes()

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "skip",
				},
				Version: "v5.0.0",
			},
		})
		node := &workflowModel.WorkFlowNode{}
		err := destroyCluster(node, flowContext)
		assert.NoError(t, err)
		assert.NotEmpty(t, node.Result)
	})
}

func TestDeleteCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().Delete(gomock.Any(), "111").Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v5.0.0",
		},
	})
	err := deleteCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

}

func TestDeleteClusterPhysically(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().ClearClusterPhysically(gomock.Any(), "111").Return(nil)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v5.0.0",
		},
	})
	err := clearClusterPhysically(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

}
func TestFreedClusterResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
		},
	})

	t.Run("normal", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(nil)
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := freedClusterResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("recycle fail", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(fmt.Errorf("recycle fail"))
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)
		err := freedClusterResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestInitDatabaseAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
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
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
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
			},
		},
		DBUsers: map[string]*management.DBUser{
			string(constants.Root): &management.DBUser{
				ClusterID: "2145635758",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
		},
	})

	t.Run("normal", func(t *testing.T) {
		//mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		//mockTiupManager.EXPECT().(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		//secondparty.Manager = mockTiupManager
		err := initDatabaseAccount(&workflowModel.WorkFlowNode{}, flowContext)
		//assert.NoError(t, err)
		fmt.Println(err)
	})

	t.Run("init fail", func(t *testing.T) {
		//	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
		//	mockTiupManager.EXPECT().SetClusterDbPassword(gomock.Any(),
		//		gomock.Any(), gomock.Any()).Return(fmt.Errorf("init fail")).AnyTimes()
		//	secondparty.Manager = mockTiupManager
		err := initDatabaseAccount(&workflowModel.WorkFlowNode{}, flowContext)
		fmt.Println(err)
		assert.Error(t, err)
	})
}

func Test_testConnectivity(t *testing.T) {
	/* run this case with real tidb ip/port/user/password
	t.Run("normal", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				DBUser: "root",
				DBPassword: "fdasfsdafa",
			},
			Instances: map[string][]*management.ClusterInstance{
				string(constants.ComponentIDTiDB) : {
					{
						Entity: common.Entity{
							Status: string(constants.ClusterRunning),
						},
						HostIP: []string{"172.16.5.148"},
						Ports: []int32{10000},
					},
				},
			},
		})
		err := testConnectivity(&workflowModel.WorkFlowNode{}, ctx)
		assert.NoError(t, err)
	})
	*/
	t.Run("error", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				//DBUser:     "root",
				//DBPassword: "wrong",
			},
			Instances: map[string][]*management.ClusterInstance{
				string(constants.ComponentIDTiDB): {
					{
						Entity: common.Entity{
							Status: string(constants.ClusterRunning),
						},
						HostIP: []string{"172.16.6.176"},
						Ports:  []int32{10000},
					},
				},
			},
			DBUsers: map[string]*management.DBUser{
				string(constants.Root): &management.DBUser{
					ClusterID: "2145635758",
					Name:      constants.DBUserName[constants.Root],
					Password:  "wrong",
					RoleType:  string(constants.Root),
				},
			},
		})
		err := testConnectivity(&workflowModel.WorkFlowNode{}, ctx)
		fmt.Println(err)
		assert.Error(t, err)
	})
}

func Test_initDatabaseData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				//DBUser:     "root",
				//DBPassword: "ssssssss",
			},
			Instances: map[string][]*management.ClusterInstance{
				string(constants.ComponentIDTiDB): {
					{
						Entity: common.Entity{
							Status: string(constants.ClusterRunning),
						},
						HostIP: []string{"172.16.6.176"},
						Ports:  []int32{10000},
					},
				},
			},
			DBUsers: map[string]*management.DBUser{
				string(constants.Root): {
					ClusterID: "testID",
					Name:      "root",
					Password:  "ssssssss",
					RoleType:  string(constants.Root),
				},
			},
		})

		ctx.SetData(ContextBackupID, "iddddd")

		brService := mock_br_service.NewMockBRService(ctrl)
		backuprestore.MockBRService(brService)
		brService.EXPECT().RestoreExistCluster(gomock.Any(),
			gomock.Any(), false).Return(
			cluster.RestoreExistClusterResp{
				AsyncTaskWorkFlowInfo: structs2.AsyncTaskWorkFlowInfo{
					WorkFlowID: "111",
				},
			}, nil)

		node := &workflowModel.WorkFlowNode{}
		err := initDatabaseData(node, ctx)
		assert.NoError(t, err)
		assert.Contains(t, node.Result, "recover data from backup record")
	})
	t.Run("skip", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				//DBUser:     "root",
				//DBPassword: "ssssssss",
			},
			Instances: map[string][]*management.ClusterInstance{
				string(constants.ComponentIDTiDB): {
					{
						Entity: common.Entity{
							Status: string(constants.ClusterRunning),
						},
						HostIP: []string{"172.16.6.176"},
						Ports:  []int32{10000},
					},
				},
			},
			DBUsers: map[string]*management.DBUser{
				string(constants.Root): {
					ClusterID: "testID",
					Name:      "root",
					Password:  "ssssssss",
					RoleType:  string(constants.Root),
				},
			},
		})

		node := &workflowModel.WorkFlowNode{}
		err := initDatabaseData(node, ctx)
		assert.NoError(t, err)
		assert.Contains(t, node.Result, "skip")
	})
}

func Test_testRebuildTopologyFromConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().UpdateInstance(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		clusterMeta := &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID:       "clusterId",
					TenantId: "tenantId",
				},
			},
		}
		ctx.SetData(ContextClusterMeta, clusterMeta)

		metadata := &spec.ClusterMeta{
			Version: "v5.2.2",
			Topology: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{Host: "127.0.0.1", Port: 1, StatusPort: 2},
					{Host: "127.0.0.2", Port: 3, StatusPort: 4},
				},
				TiKVServers: []*spec.TiKVSpec{
					{Host: "127.0.0.4", Port: 5, StatusPort: 6},
					{Host: "127.0.0.5", Port: 7, StatusPort: 8},
				},
				TiFlashServers: []*spec.TiFlashSpec{
					{Host: "127.0.0.6", TCPPort: 9, HTTPPort: 10, FlashServicePort: 11, FlashProxyPort: 12, FlashProxyStatusPort: 13, StatusPort: 14},
				},
				CDCServers: []*spec.CDCSpec{
					{Host: "127.0.0.7", Port: 15},
					{Host: "127.0.0.8", Port: 16},
				},
				PDServers: []*spec.PDSpec{
					{Host: "127.0.0.9", ClientPort: 17, PeerPort: 18},
					{Host: "127.0.0.10", ClientPort: 19, PeerPort: 20},
				},
				Grafanas: []*spec.GrafanaSpec{
					{Host: "127.0.0.11", Port: 21},
				},
				Alertmanagers: []*spec.AlertmanagerSpec{
					{Host: "127.0.0.12", WebPort: 22, ClusterPort: 23},
				},
				Monitors: []*spec.PrometheusSpec{
					{Host: "127.0.0.13", Port: 24},
				},
			},
		}
		bytes, err := yaml.Marshal(metadata)
		assert.NoError(t, err)
		ctx.SetData(ContextTopologyConfig, bytes)

		err = rebuildTopologyFromConfig(&workflowModel.WorkFlowNode{}, ctx)
		assert.NoError(t, err)
		assert.Equal(t, "v5.2.2", clusterMeta.Cluster.Version)
		assert.NotEmpty(t, clusterMeta.Instances)

	})
}

func TestTakeoverResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
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
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
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
			},
		},
		// todo: need user?
		DBUsers: map[string]*management.DBUser{
			string(constants.Root): &management.DBUser{
				ClusterID: "2145635758",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
		},
	})

	t.Run("normal", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().AllocResources(gomock.Any(), gomock.Any()).Return(&structs.BatchAllocResponse{
			BatchResults: []*structs.AllocRsp{
				{
					Applicant: structs.Applicant{
						RequestId: "123",
					},
					Results: []structs.Compute{
						{
							Location: structs2.Location{
								Region: "region01",
								Zone:   "zone01",
								Rack:   "rack01",
							},
							HostId: "001",
						},
					},
				},
			},
		}, nil)
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)

		err := takeoverResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("alloc fail", func(t *testing.T) {
		resourceManager := mock_allocator_recycler.NewMockAllocatorRecycler(ctrl)
		resourceManager.EXPECT().AllocResources(gomock.Any(),
			gomock.Any()).Return(&structs.BatchAllocResponse{}, fmt.Errorf("fail"))
		resourceManagement.GetManagement().SetAllocatorRecycler(resourceManager)

		err := takeoverResource(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func Test_syncTopology(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiupRW := mocktiupconfig.NewMockReaderWriter(ctrl)
	models.SetTiUPConfigReaderWriter(tiupRW)
	tiupRW.EXPECT().QueryByComponentType(gomock.Any(), gomock.Any()).Return(&tiup.TiupConfig{TiupHome: "testdata"}, nil).AnyTimes()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().UpdateTopologySnapshotConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v5.0.0",
		},
	})

	testFilePath := "testdata"
	os.MkdirAll(testFilePath, 0755)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()
	path := getClusterSpaceInTiUP(context.TODO(), "111")
	os.MkdirAll(path, 0755)

	t.Run("normal", func(t *testing.T) {
		f, _ := os.Create(path + "/meta.yaml")
		f.Write([]byte{'a', 'b'})
		defer f.Close()
		defer os.RemoveAll(path)

		err := syncTopology(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		err := syncTopology(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func Test_syncConnectionKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiupRW := mocktiupconfig.NewMockReaderWriter(ctrl)
	models.SetTiUPConfigReaderWriter(tiupRW)
	tiupRW.EXPECT().QueryByComponentType(gomock.Any(), gomock.Any()).Return(&tiup.TiupConfig{TiupHome: "testdata"}, nil).AnyTimes()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().CreateClusterTopologySnapshot(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v5.0.0",
		},
	})

	testFilePath := "testdata"
	os.MkdirAll(testFilePath, 0755)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()
	path := getClusterSpaceInTiUP(context.TODO(), "111")
	os.MkdirAll(path, 0755)

	t.Run("normal", func(t *testing.T) {
		os.MkdirAll(path+"/ssh", 0755)
		defer os.RemoveAll(path + "/ssh")

		f1, err := os.Create(path + "/ssh/id_rsa")
		assert.NoError(t, err)
		defer f1.Close()

		f1.Write([]byte{'a', 'b'})

		f2, err := os.Create(path + "/ssh/id_rsa.pub")
		assert.NoError(t, err)
		defer f2.Close()

		f2.Write([]byte{'c', 'd'})

		err = syncConnectionKey(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		err := syncConnectionKey(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
	t.Run("without public", func(t *testing.T) {
		os.MkdirAll(path+"/ssh", 0755)
		defer os.RemoveAll(path + "/ssh")

		f1, err := os.Create(path + "/ssh/id_rsa")
		assert.NoError(t, err)
		defer f1.Close()

		f1.Write([]byte{'a', 'b'})

		err = syncConnectionKey(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
	t.Run("without private", func(t *testing.T) {
		os.MkdirAll(path+"/ssh", 0755)
		defer os.RemoveAll(path + "/ssh")

		f2, err := os.Create(path + "/ssh/id_rsa.pub")
		assert.NoError(t, err)
		defer f2.Close()

		f2.Write([]byte{'c', 'd'})

		err = syncConnectionKey(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func Test_rebuildTiupSpaceForCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiupRW := mocktiupconfig.NewMockReaderWriter(ctrl)
	models.SetTiUPConfigReaderWriter(tiupRW)
	tiupRW.EXPECT().QueryByComponentType(gomock.Any(), gomock.Any()).Return(&tiup.TiupConfig{
		TiupHome: "testdata",
	}, nil).AnyTimes()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetCurrentClusterTopologySnapshot(gomock.Any(), gomock.Any()).Return(management.ClusterTopologySnapshot{
		Config:     "111",
		PublicKey:  "222",
		PrivateKey: "333",
	}, nil).AnyTimes()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v5.0.0",
		},
	})

	t.Run("normal", func(t *testing.T) {
		err := rebuildTiupSpaceForCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
		publicKey, err := os.ReadFile("testdata/storage/cluster/clusters/111/ssh/id_rsa.pub")
		assert.NoError(t, err)
		assert.Equal(t, "222", string(publicKey))
	})
}

func Test_validateHostsStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceRW := mockresource.NewMockReaderWriter(ctrl)
	models.SetResourceReaderWriter(resourceRW)

	provider := resourcepool.GetResourcePool().GetHostProvider().(*hostprovider.FileHostProvider)
	provider.SetResourceReaderWriter(resourceRW)

	t.Run("normal", func(t *testing.T) {
		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]rp.Host{
			{Status: string(constants.HostOnline)},
		}, int64(1), nil).Times(1)

		node := &workflowModel.WorkFlowNode{}

		context := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		context.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						HostIP: []string{
							"127.0.0.1",
						},
					},
				},
			},
		})
		err := validateHostsStatus(node, context)
		assert.NoError(t, err)
		assert.NotEmpty(t, node.Result)
	})
	t.Run("ip not existed", func(t *testing.T) {
		node := &workflowModel.WorkFlowNode{}

		context := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		context.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						HostIP: []string{},
					},
				},
			},
		})
		err := validateHostsStatus(node, context)
		assert.Error(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]rp.Host{
			{Status: string(constants.HostFailed)},
		}, int64(1), nil).Times(1)

		node := &workflowModel.WorkFlowNode{}

		context := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		context.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						HostIP: []string{
							"127.0.0.1",
						},
					},
				},
			},
		})
		err := validateHostsStatus(node, context)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), strconv.Itoa(int(errors.TIEM_RESOURCE_CREATE_HOST_ERROR)))
	})

	t.Run("init + succeed", func(t *testing.T) {
		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]rp.Host{
			{Status: string(constants.HostInit)},
		}, int64(1), nil).Times(1)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]rp.Host{
			{Status: string(constants.HostOnline)},
		}, int64(1), nil).Times(1)

		node := &workflowModel.WorkFlowNode{}

		context := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		context.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						HostIP: []string{
							"127.0.0.1",
						},
					},
				},
			},
		})
		err := validateHostsStatus(node, context)
		assert.NoError(t, err)
		assert.NotEmpty(t, node.Result)
	})
	t.Run("timeout", func(t *testing.T) {
		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]rp.Host{
			{Status: string(constants.HostInit)},
		}, int64(1), nil).Times(2)

		node := &workflowModel.WorkFlowNode{}
		validateHostTimeout = time.Second * 6
		context := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		context.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						HostIP: []string{
							"127.0.0.1",
						},
					},
				},
			},
		})
		err := validateHostsStatus(node, context)
		assert.Error(t, err)
		assert.NotEmpty(t, node.Result)
		assert.Contains(t, node.Result, "importing")
	})

}

func Test_syncIncrData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cluster01",
			},
		},
	})
	flowContext.SetData(ContextSourceClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cluster02",
			},
		},
	})
	flowContext.SetData(ContextCloneStrategy, string(constants.CDCSyncClone))

	t.Run("normal", func(t *testing.T) {
		service := mockchangefeed.NewMockService(ctrl)
		changefeed.MockChangeFeedService(service)

		service.EXPECT().CreateBetweenClusters(gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return("task01", nil)
		flowContext.SetData(ContextGCLifeTime, "10m0s")
		service.EXPECT().Detail(gomock.Any(), gomock.Any()).Return(
			cluster.DetailChangeFeedTaskResp{
				ChangeFeedTaskInfo: cluster.ChangeFeedTaskInfo{
					UpstreamUpdateUnix: 12000,
					DownstreamSyncUnix: 11000,
				}}, nil)
		err := syncIncrData(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

}

func Test_fetchTopologyFile(t *testing.T) {
	originalOpen := openSftpClient
	openSftpClient = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
		return nil, nil, nil
	}
	defer func() {
		openSftpClient = originalOpen
	}()

	originalRead := readRemoteFile
	readRemoteFile = func(ctx context.Context, sftp *sftp.Client, clusterHome string, file string) ([]byte, error) {
		return []byte{}, nil
	}
	defer func() {
		readRemoteFile = originalRead
	}()
	node := &workflowModel.WorkFlowNode{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().CreateClusterTopologySnapshot(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	rw.EXPECT().UpdateTopologySnapshotConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	context := &workflow.FlowContext{
		Context:  context.TODO(),
		FlowData: map[string]interface{}{},
	}
	context.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					HostIP: []string{
						"127.0.0.1",
					},
				},
			},
		},
	})
	context.SetData(ContextTakeoverRequest, cluster.TakeoverClusterReq{})

	err := fetchTopologyFile(node, context)
	assert.NoError(t, err)
}

func TestCheckInstanceStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
			Version: "v5.0.0",
			Type:    "TiDB",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						ID: "instance01",
					},
					Type:   "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						ID: "instance02",
					},
					Type:   "TiKV",
					HostIP: []string{"127.0.0.2"},
					Ports:  []int32{8001},
				},
			},
			"PD": {
				{
					Entity: common.Entity{
						ID:     "instance03",
						Status: string(constants.ClusterInstanceRunning),
					},
					Type:   "PD",
					HostIP: []string{"127.0.0.3"},
					Ports:  []int32{8001},
				},
			},
		},
	})

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ClusterComponentCtl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		[]string{"-u", "127.0.0.3:8001", "store", "--state", "Tombstone,Up,Offline"}, gomock.Any()).Return(`
{
  "count": 3,
  "stores": [
    {
      "store": {
        "id": 1,
        "address": "127.0.0.2:8001",
        "state_name": "Offline"
      },
      "status": {
        "leader_count": 0,
        "region_count": 1
      }
    },
    {
      "store": {
        "id": 4,
        "address": "172.16.4.187:10020",
        "state_name": "Up"
      },
      "status": {
        "leader_count": 1,
        "region_count": 1
      }
    },
    {
      "store": {
        "id": 1001,
        "address": "172.16.4.187:10022",
        "state_name": "Up"
      },
      "status": {
        "leader_count": 0,
        "region_count": 1
      }
    }
  ]
}`, nil)
	mockTiupManager.EXPECT().ClusterComponentCtl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), []string{"-u", "127.0.0.3:8001", "store", "1"}, gomock.Any()).Return(`
{
  "store": {
    "id": 1,
    "address": "127.0.0.2:8001",
    "state_name": "Tombstone"
  },
  "status": {
    "leader_count": 0,
    "region_count": 0
  }
}`, nil)
	mockTiupManager.EXPECT().ClusterPrune(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil)
	secondparty.Manager = mockTiupManager
	flowContext.SetData(ContextInstanceID, "instance02")
	err := checkInstanceStatus(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)
}

func Test_applyParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.2", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{
			{},
		}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.1", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.0", 1, 1, gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.Error(errors.TIEM_PANIC)).AnyTimes()

	t.Run("query group error", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.0.0",
				ParameterGroupID: "",
			},
		})
		err := applyParameterGroup(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})
	t.Run("query group empty", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.1.22",
				ParameterGroupID: "",
			},
		})
		err := applyParameterGroup(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})
}

func Test_chooseParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.2", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{
			{},
		}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.1", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.0", 1, 1, gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.Error(errors.TIEM_PANIC)).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		node := &workflowModel.WorkFlowNode{}
		err := chooseParameterGroup(&handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.2.3",
				Type: "TiDB",
			},
		}, node, ctx)
		assert.NoError(t, err)
		assert.Contains(t, node.Result, "parameter group id is empty")
		assert.Contains(t, node.Result, "will be applied to cluster")
	})
	t.Run("empty", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		node := &workflowModel.WorkFlowNode{}
		err := chooseParameterGroup(&handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.1.9",
				Type: "TiDB",
			},
		}, node, ctx)
		assert.Error(t, err)
		assert.Contains(t, node.Result, "no default group found")
	})
	t.Run("error", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		node := &workflowModel.WorkFlowNode{}
		err := chooseParameterGroup(&handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.0.1",
				Type: "TiDB",
			},
		}, node, ctx)
		assert.Error(t, err)
	})
}

func Test_Test_applyParameterGroupForTakeover(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.2", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{
			{},
		}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.1", 1, 1, gomock.Any(), gomock.Any()).
		Return([]*parametergroup.ParameterGroup{}, int64(0), nil).AnyTimes()

	parameterGroupRW.EXPECT().
		QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), "v5.0", 1, 1, gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.Error(errors.TIEM_PANIC)).AnyTimes()

	t.Run("query group error", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.0.0",
				ParameterGroupID: "",
			},
		})
		err := applyParameterGroupForTakeover(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})
	t.Run("query group empty", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.1.22",
				ParameterGroupID: "",
			},
		})
		err := applyParameterGroupForTakeover(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})
}

func Test_adjustParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(parameterRW)

	t.Run("query parameter error", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.1.22",
				ParameterGroupID: "",
			},
		})
		parameterRW.EXPECT().
			QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("111", nil, int64(0), errors.Error(errors.TIEM_PANIC)).
			Times(1)

		err := adjustParameters(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})
	t.Run("query parameter empty", func(t *testing.T) {
		ctx := workflow.NewFlowContext(context.TODO())
		ctx.SetData(ContextClusterMeta, &handler.ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID: "1111",
				},
				Version:          "v5.1.22",
				ParameterGroupID: "",
			},
		})
		parameterRW.EXPECT().
			QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("111", []*parameter.ClusterParamDetail{
				{Parameter: parametergroup.Parameter{
					ID:   "aaa",
					Name: "aaa",
				}},
			}, int64(0), nil).
			Times(1)

		err := adjustParameters(&workflowModel.WorkFlowNode{}, ctx)
		assert.Error(t, err)
	})

}

var dbConnParam1 secondparty.DbConnParam
var dbConnParam2 secondparty.DbConnParam

func init() {
	dbConnParam1 = secondparty.DbConnParam{
		Username: "root",
		Password: "",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
	dbConnParam2 = secondparty.DbConnParam{
		Username: "root",
		Password: "12345678",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
}

func TestGenerateDBUser(t *testing.T) {
	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(ContextClusterMeta, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "testCluster",
			},
			Version: "v5.0.0",
		},
	})
	clusterMeta := flowContext.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	type args struct {
		context *workflow.FlowContext
		roleTyp constants.DBUserRoleType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"normal", args{flowContext, constants.DBUserBackupRestore}, clusterMeta.Cluster.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateDBUser(tt.args.context, tt.args.roleTyp); !reflect.DeepEqual(got.ClusterID, tt.want) {
				t.Errorf("GenerateDBUser() = %v, want %v", got, tt.want)
			} else {
				fmt.Println(got)
				fmt.Println(got.ID)
			}
		})
	}
}
