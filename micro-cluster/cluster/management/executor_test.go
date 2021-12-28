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
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareResource(t *testing.T) {

}

func TestBuildConfig(t *testing.T) {

}

func TestScaleOutCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ClusterScaleOut(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
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
	err := scaleOutCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)
}

func TestScaleInCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ClusterScaleIn(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("task01", nil).AnyTimes()
	secondparty.Manager = mockTiupManager

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
					Type: "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						ID: "instance03",
					},
					Type:   "TiKV",
					HostIP: []string{"127.0.0.2"},
					Ports:  []int32{8001},
				},
			},
		},
	})
	flowContext.SetData(ContextInstanceID, "instance01")
	err := scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)

	flowContext.SetData(ContextInstanceID, "instance03")
	err = scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)

	flowContext.SetData(ContextInstanceID, "instance04")
	err = scaleInCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Error(t, err)
}

func TestFreeInstanceResource(t *testing.T) {

}

func TestClearBackupData(t *testing.T) {

}

func TestBackupBeforeDelete(t *testing.T) {

}

func TestBackupSubProcess(t *testing.T) {

}

func TestBackupSourceCluster(t *testing.T) {

}

func TestRestoreNewCluster(t *testing.T) {

}

func TestWaitWorkFlow(t *testing.T) {

}

func TestSetClusterFailure(t *testing.T) {

}

func TestSetClusterOnline(t *testing.T) {

}

func TestSetClusterOffline(t *testing.T) {

}

func TestRevertResourceAfterFailure(t *testing.T) {

}

func TestEndMaintenance(t *testing.T) {

}

func TestPersistCluster(t *testing.T) {

}

func TestDeployCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ClusterDeploy(gomock.Any(), gomock.Any(), gomock.Any(),
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
	flowContext.SetData(ContextTopology, "test topology")
	err := deployCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NoError(t, err)
}

func TestStartCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
}

func TestSyncBackupStrategy(t *testing.T) {

}

func TestSyncParameters(t *testing.T) {

}

func TestRestoreCluster(t *testing.T) {

}

func TestStopCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
}

func TestDestroyCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

func TestFreedClusterResource(t *testing.T) {

}

func TestInitDatabaseAccount(t *testing.T) {

}
