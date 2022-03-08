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
	"os"
	"testing"
	"time"
)

func init() {
	models.MockDB()
}

func TestGetImportExportService(t *testing.T) {
	service := GetImportExportService()
	assert.NotNil(t, service)
}

func TestImportExportManager_ExportData_case1(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: common.PasswordInExpired{Val: "123455678", UpdateTime: time.Now()}, RoleType: string(constants.Root)},
	}, nil).AnyTimes()
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
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyExportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
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

func TestImportExportManager_ExportData_case2(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: common.PasswordInExpired{Val: "123455678", UpdateTime: time.Now()}, RoleType: string(constants.Root)},
	}, nil).AnyTimes()
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
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyExportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
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
		StorageType:     string(constants.StorageTypeNFS),
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

func TestImportExportManager_ExportData_case3(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyExportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	service := GetImportExportService()
	_, err := service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "",
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
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "",
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
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		FileType:        "xxx",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "export.zip",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "",
		EndpointUrl:     "",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ExportData(context.TODO(), message.DataExportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		FileType:        "csv",
		Filter:          "filter",
		StorageType:     string(constants.StorageTypeS3),
		ZipName:         "",
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "",
		Comment:         "comment",
	})
	assert.NotNil(t, err)
}

func TestImportExportManager_ImportData_case1(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: common.PasswordInExpired{Val: "123455678", UpdateTime: time.Now()}, RoleType: string(constants.Root)},
	}, nil).AnyTimes()
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
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyImportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	transportService := mockimportexport.NewMockReaderWriter(ctrl)
	transportService.EXPECT().CreateDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	models.SetImportExportReaderWriter(transportService)

	service := GetImportExportService()
	resp, err := service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "test-cls",
		UserName:        "userName",
		Password:        "password",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestImportExportManager_ImportData_case2(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

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
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyImportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	transportService := mockimportexport.NewMockReaderWriter(ctrl)
	transportService.EXPECT().CreateDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{Entity: common.Entity{
		ID: "xxx",
	}}, nil).AnyTimes()
	transportService.EXPECT().GetDataTransportRecord(gomock.Any(), gomock.Any()).Return(&importexport.DataTransportRecord{
		Entity: common.Entity{
			ID: "xxx",
		},
		FilePath:        "./testdata",
		StorageType:     string(constants.StorageTypeNFS),
		ReImportSupport: true,
	}, nil).AnyTimes()
	models.SetImportExportReaderWriter(transportService)

	service := GetImportExportService()
	resp, err := service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "test-cls",
		UserName:        "userName",
		Password:        "password",
		RecordId:        "record-xxx",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

func TestImportExportManager_ImportData_case3(t *testing.T) {
	os.MkdirAll("./testdata", 0755)
	defer os.RemoveAll("./testdata")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: common.PasswordInExpired{Val: "123455678", UpdateTime: time.Now()}, RoleType: string(constants.Root)},
	}, nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	configService := mockconfig.NewMockReaderWriter(ctrl)
	configService.EXPECT().GetConfig(gomock.Any(), constants.ConfigKeyImportShareStoragePath).Return(&config.SystemConfig{ConfigValue: "./testdata"}, nil).AnyTimes()
	models.SetConfigReaderWriter(configService)

	service := GetImportExportService()
	_, err := service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "",
		UserName:        "userName",
		Password:        "password",
		RecordId:        "record-xxx",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "",
		Password:        "password",
		RecordId:        "record-xxx",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "",
		RecordId:        "record-xxx",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "",
		AccessKey:       "ak",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "",
		SecretAccessKey: "sk",
		Comment:         "comment",
	})
	assert.NotNil(t, err)

	_, err = service.ImportData(context.TODO(), message.DataImportReq{
		ClusterID:       "clusterId",
		UserName:        "userName",
		Password:        "password",
		StorageType:     string(constants.StorageTypeS3),
		EndpointUrl:     "endpointUrl",
		BucketUrl:       "bucketUrl",
		AccessKey:       "ak",
		SecretAccessKey: "",
		Comment:         "comment",
	})
	assert.NotNil(t, err)
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
