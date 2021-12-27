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

package importexport

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/models/platform/config"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockimportexport"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	models.MockDB()
}

func TestGetImportExportService(t *testing.T) {
	service := GetImportExportService()
	assert.NotNil(t, service)
}

func TestImportExportManager_ExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: string(constants.StorageTypeNFS)}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	transportService := mockimportexport.NewMockReaderWriter(ctrl)
	transportService.EXPECT().CreateDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	models.SetImportExportReaderWriter(transportService)

	service := GetImportExportService()
	resp, err := service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "test-cls",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "export.zip",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestImportExportManager_ImportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: string(constants.StorageTypeNFS)}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	transportService := mockimportexport.NewMockReaderWriter(ctrl)
	transportService.EXPECT().CreateDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	models.SetImportExportReaderWriter(transportService)

	service := GetImportExportService()
	resp, err := service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "test-cls",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "export.zip",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestImportExportManager_DeleteDataTransportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockImportExportRW := mockimportexport.NewMockReaderWriter(ctrl)
	mockImportExportRW.EXPECT().GetDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{StorageType: "nfs"}, nil).AnyTimes()
	mockImportExportRW.EXPECT().DeleteDataTransportRecord(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetImportExportReaderWriter(mockImportExportRW)

	service := GetImportExportService()
	_, err := service.DeleteDataTransportRecord(context.TODO(), message.DeleteImportExportRecordReq{RecordID: "record-xxx"})
	assert.Nil(t, err)
}

func TestImportExportManager_QueryDataTransportRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	records := make([]*importexport.DataTransportRecord, 1)
	records[0] = &importexport.DataTransportRecord{
		Entity: common.Entity{
			ID: "record-xxx",
		},
		FilePath: "./testdata",
	}
	mockImportExportRW := mockimportexport.NewMockReaderWriter(ctrl)
	mockImportExportRW.EXPECT().QueryDataTransportRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(records, int64(1), nil).AnyTimes()
	models.SetImportExportReaderWriter(mockImportExportRW)

	service := GetImportExportService()
	_, _, err := service.QueryDataTransportRecords(context.TODO(), message.QueryDataImportExportRecordsReq{RecordID: "record-xxx"})
	assert.Nil(t, err)
}
