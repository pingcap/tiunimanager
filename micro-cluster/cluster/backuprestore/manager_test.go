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
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/platform/config"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBRService(t *testing.T) {
	service := GetBRService()
	assert.NotNil(t, service)
}

func TestBRManager_BackupCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), make([]*management.DBUser, 0), nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: string(constants.StorageTypeNFS)}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	brService := mockbr.NewMockReaderWriter(ctrl)
	brService.EXPECT().CreateBackupRecord(gomock.Any(), gomock.Any()).Return(&backuprestore.BackupRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	models.SetBRReaderWriter(brService)

	service := GetBRService()
	resp, err := service.BackupCluster(context.TODO(), cluster.BackupClusterDataReq{
		ClusterID:  "test-cls",
		BackupMode: string(constants.BackupModeManual),
	}, true)

	assert.Nil(t, err)
	assert.NotNil(t, resp.BackupID)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestBRManager_RestoreExistCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), make([]*management.DBUser, 0), nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: string(constants.StorageTypeNFS)}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	brService := mockbr.NewMockReaderWriter(ctrl)
	brService.EXPECT().GetBackupRecord(gomock.Any(), gomock.Any()).Return(&backuprestore.BackupRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	models.SetBRReaderWriter(brService)

	service := GetBRService()
	resp, err := service.RestoreExistCluster(context.TODO(), cluster.RestoreExistClusterReq{
		ClusterID: "test-cls",
		BackupID:  "xxx",
	}, true)

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestBRManager_DeleteBackupRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	records := make([]*backuprestore.BackupRecord, 1)
	records[0] = &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxx",
		},
		FilePath: "./testdata",
	}
	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(records, int64(1), nil)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make([]*backuprestore.BackupRecord, 0), int64(0), nil)
	brRW.EXPECT().DeleteBackupRecord(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.DeleteBackupRecords(context.TODO(), cluster.DeleteBackupDataReq{
		ClusterID: "testCluster",
		BackupID:  "testBackup",
	})
	assert.Nil(t, err)
}

func TestBRManager_GetBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().GetBackupStrategy(gomock.Any(), gomock.Any()).Return(&backuprestore.BackupStrategy{
		ClusterID:  "cls-xxxx",
		BackupDate: "Monday,Friday",
		StartHour:  0,
		EndHour:    1,
	}, nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	resp, err := service.GetBackupStrategy(context.TODO(), cluster.GetBackupStrategyReq{})
	assert.Nil(t, err)
	assert.Equal(t, "0:00-1:00", resp.Strategy.Period)
}

func TestBRManager_DeleteBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().DeleteBackupStrategy(gomock.Any(), gomock.Any()).Return(nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.DeleteBackupStrategy(context.TODO(), cluster.DeleteBackupStrategyReq{})
	assert.Nil(t, err)
}

func TestBRManager_SaveBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().SaveBackupStrategy(gomock.Any(), gomock.Any()).Return(nil, nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.SaveBackupStrategy(context.TODO(), cluster.SaveBackupStrategyReq{
		ClusterID: "cls-xxxx",
		Strategy: structs.BackupStrategy{
			ClusterID:  "cls-xxxx",
			BackupDate: "Monday",
			Period:     "0:00-1:00",
		},
	})
	assert.Nil(t, err)
}

func TestBRManager_QueryClusterBackupRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	records := make([]*backuprestore.BackupRecord, 1)
	records[0] = &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxx",
		},
		FilePath: "./testdata",
	}
	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(records, int64(1), nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	resp, _, err := service.QueryClusterBackupRecords(context.TODO(), cluster.QueryBackupRecordsReq{})
	assert.Nil(t, err)
	assert.Equal(t, records[0].FilePath, resp.BackupRecords[0].FilePath)
}
