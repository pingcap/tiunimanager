/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package backuprestore

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/platform/config"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockmanagement"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	models.MockDB()
}

func TestExecutor_backupCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().BackUp(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	secondparty.Manager = mockTiupManager

	confingRW := mockconfig.NewMockReaderWriter(ctrl)
	confingRW.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "test"}, nil).AnyTimes()
	models.SetConfigReaderWriter(confingRW)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		StorageType: "s3",
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-test",
			},
			Name: "cls-test",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiDB",
					Version:      "v5.0.0",
					Ports:        []int32{10001, 10002, 10003, 10004},
					HostIP:       []string{"127.0.0.1"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
		},
		DBUsers: map[string]*management.DBUser{
			string(constants.DBUserBackupRestore): &management.DBUser{
				ClusterID: "cls-test",
				Name:      constants.DBUserName[constants.DBUserBackupRestore],
				Password:  "12345678",
				RoleType:  string(constants.DBUserBackupRestore),
			},
		},
	})
	err := backupCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_updateBackupRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().UpdateBackupRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().ShowBackUpInfoThruMetaDB(gomock.Any(), gomock.Any()).Return(secondparty.CmdBrResp{
		Size:     123,
		BackupTS: 234,
	}, nil).AnyTimes()
	secondparty.Manager = mockTiupManager

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-xxxx",
			},
		},
	})
	flowContext.SetData(contextBackupTiupTaskIDKey, "123")
	err := updateBackupRecord(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_restoreFromSrcCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiupManager := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	mockTiupManager.EXPECT().Restore(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	secondparty.Manager = mockTiupManager

	confingRW := mockconfig.NewMockReaderWriter(ctrl)
	confingRW.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "test"}, nil).AnyTimes()
	models.SetConfigReaderWriter(confingRW)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		StorageType: "s3",
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-test",
			},
			Name: "cls-test",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Type:   "TiDB",
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8000},
				},
			},
		},
		DBUsers: map[string]*management.DBUser{
			string(constants.DBUserBackupRestore): &management.DBUser{
				ClusterID: "cls-test",
				Name:      constants.DBUserName[constants.DBUserBackupRestore],
				Password:  "12345678",
				RoleType:  string(constants.DBUserBackupRestore),
			},
		},
	})
	err := restoreFromSrcCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_backupFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().UpdateBackupRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	clusterRW := mockmanagement.NewMockReaderWriter(ctrl)
	clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetClusterReaderWriter(clusterRW)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-xxxx",
			},
		},
	})
	flowContext.SetData(contextMaintenanceStatusChangeKey, true)
	err := backupFail(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_restoreFail(t *testing.T) {
	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-xxxx",
			},
		},
	})
	flowContext.SetData(contextMaintenanceStatusChangeKey, true)
	err := restoreFail(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_defaultEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockmanagement.NewMockReaderWriter(ctrl)
	clusterRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetClusterReaderWriter(clusterRW)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &handler.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-xxxx",
			},
		},
	})
	flowContext.SetData(contextMaintenanceStatusChangeKey, true)
	err := defaultEnd(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}
