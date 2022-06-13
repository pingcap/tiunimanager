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
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/backuprestore"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiunimanager/models/common"
	"github.com/pingcap/tiunimanager/models/platform/config"
	workflowModel "github.com/pingcap/tiunimanager/models/workflow"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockbr"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockconfig"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockmanagement"
	"github.com/pingcap/tiunimanager/util/api/tidb/sql"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func init() {
	models.MockDB()
}

func TestExecutor_backupCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	confingRW := mockconfig.NewMockReaderWriter(ctrl)
	confingRW.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "test"}, nil).AnyTimes()
	models.SetConfigReaderWriter(confingRW)

	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		StorageType: "nfs",
		FilePath:    "./testdata",
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-test",
			},
			Name:    "cls-test",
			Version: "v5.2.2",
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
			string(constants.DBUserBackupRestore): {
				ClusterID: "cls-test",
				Name:      constants.DBUserName[constants.DBUserBackupRestore],
				Password:  common.PasswordInExpired{Val: "12345678", UpdateTime: time.Now()},

				RoleType: string(constants.DBUserBackupRestore),
			},
		},
	})
	err := backupCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NotNil(t, err)
	os.Remove("./testdata")
}

func TestExecutor_updateBackupRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().UpdateBackupRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-xxxx",
			},
		},
	})
	flowContext.SetData(contextBRInfoKey, &sql.BRSQLResp{
		Destination: "test",
		Size:        1234,
		BackupTS:    2534534534,
	})
	err := updateBackupRecord(&workflowModel.WorkFlowNode{}, flowContext)
	assert.Nil(t, err)
}

func TestExecutor_restoreFromSrcCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	confingRW := mockconfig.NewMockReaderWriter(ctrl)
	confingRW.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "test"}, nil).AnyTimes()
	models.SetConfigReaderWriter(confingRW)

	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		StorageType: "s3",
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "cls-test",
			},
			Name:    "cls-test",
			Version: "v5.2.2",
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
			string(constants.DBUserBackupRestore): {
				ClusterID: "cls-test",
				Name:      constants.DBUserName[constants.DBUserBackupRestore],
				Password:  common.PasswordInExpired{Val: "12345678", UpdateTime: time.Now()},
				RoleType:  string(constants.DBUserBackupRestore),
			},
		},
	})
	err := restoreFromSrcCluster(&workflowModel.WorkFlowNode{}, flowContext)
	assert.NotNil(t, err)
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

	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
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
	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
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

	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData(contextBackupRecordKey, &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxxx",
		},
	})
	flowContext.SetData(contextClusterMetaKey, &meta.ClusterMeta{
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
