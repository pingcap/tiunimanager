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
 *                                                                            *
 ******************************************************************************/

package domain

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/test/mocksecondparty"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/test/mockdb"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
)

func TestExportDataPreCheck_case1(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: common.NfsStorageType,
	}
	err := ExportDataPreCheck(req)
	if checkFilePathExists(common.DefaultExportDir) {
		assert.NoError(t, err)
	} else {
		assert.NotNil(t, err)
	}
}

func TestExportDataPreCheck_case2(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: "local",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case3(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: common.S3StorageType,
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case4(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: common.S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case5(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: common.S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
		BucketUrl:   "s3://test",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case6(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: common.S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
		BucketUrl:   "s3://test",
		AccessKey:   "admin",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case7(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:       "test-abc",
		UserName:        "root",
		Password:        "",
		FileType:        FileTypeCSV,
		StorageType:     common.S3StorageType,
		EndpointUrl:     "https://minio.pingcap.net:9000",
		BucketUrl:       "s3://test",
		AccessKey:       "admin",
		SecretAccessKey: "admin",
	}
	err := ExportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestImportDataPreCheck_case1(t *testing.T) {
	req := &clusterpb.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		StorageType: common.NfsStorageType,
	}
	err := ImportDataPreCheck(context.TODO(), req)
	if checkFilePathExists(common.DefaultImportDir) {
		assert.NoError(t, err)
	} else {
		assert.NotNil(t, err)
	}
}

func TestImportDataPreCheck_case2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindTrasnportRecordByID(gomock.Any(), gomock.Any()).Return(&dbpb.DBFindTransportRecordByIDResponse{}, nil)
	client.DBClient = mockClient

	req := &clusterpb.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		StorageType: common.NfsStorageType,
		RecordId:    123,
	}
	err := ImportDataPreCheck(context.TODO(), req)
	assert.NotNil(t, err)
}

func TestImportDataPreCheck_case3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindTrasnportRecordByID(gomock.Any(), gomock.Any()).Return(&dbpb.DBFindTransportRecordByIDResponse{
		Record: &dbpb.TransportRecordDTO{
			ReImportSupport: true,
			StorageType:     common.NfsStorageType,
		},
	}, nil)
	client.DBClient = mockClient

	req := &clusterpb.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		StorageType: common.NfsStorageType,
		RecordId:    123,
	}
	err := ImportDataPreCheck(context.TODO(), req)
	assert.NotNil(t, err)
}

func TestImportDataPreCheck_case4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &clusterpb.DataImportRequest{
		ClusterId:       "test-abc",
		UserName:        "root",
		Password:        "",
		StorageType:     common.S3StorageType,
		AccessKey:       "test-ak",
		SecretAccessKey: "test-sk",
		EndpointUrl:     "test-ep",
		BucketUrl:       "test-bck",
	}
	err := ImportDataPreCheck(context.TODO(), req)
	assert.NoError(t, err)
}

func Test_checkExportParamSupportReimport_case1(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		FileType: FileTypeCSV,
		Sql:      "select * from test",
	}
	ret := checkExportParamSupportReimport(req)
	assert.Equal(t, false, ret)
}

func Test_checkExportParamSupportReimport_case2(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		FileType: FileTypeCSV,
		Filter:   "test.*",
	}
	ret := checkExportParamSupportReimport(req)
	assert.Equal(t, true, ret)
}

func TestExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataExportRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, err := ExportData(context.Background(), request)

	assert.NoError(t, err)
}

func TestImportData_case1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataImportRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, err := ImportData(context.Background(), request)

	assert.NoError(t, err)
}

func TestImportData_case2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBCreateTransportRecordResponse{}, nil)
	mockClient.EXPECT().FindTrasnportRecordByID(gomock.Any(), gomock.Any()).Return(&dbpb.DBFindTransportRecordByIDResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataImportRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
		RecordId: 123,
	}
	_, err := ImportData(context.Background(), request)

	assert.NoError(t, err)
}

func TestDescribeDataTransportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().ListTrasnportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBListTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataTransportQueryRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, _, err := DescribeDataTransportRecord(context.Background(), request.GetOperator(), 123, "123", 1, 10)
	assert.NoError(t, err)
}

func TestDeleteDataTransportRecord_case1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindTrasnportRecordByID(gomock.Any(), gomock.Any()).Return(&dbpb.DBFindTransportRecordByIDResponse{}, nil)
	mockClient.EXPECT().DeleteTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBDeleteTransportResponse{}, nil)
	client.DBClient = mockClient

	err := DeleteDataTransportRecord(context.Background(), &clusterpb.OperatorDTO{
		Id:       "123",
		Name:     "123",
		TenantId: "123",
	}, "test-abc", 123)

	assert.NoError(t, err)
}

func Test_importDataToCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	mockTiup.EXPECT().MicroSrvLightning(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)
	mockTiup.EXPECT().MicroSrvGetTaskStatus(gomock.Any(), gomock.Any()).Return(dbpb.TiupTaskStatus_Finished, "success", nil)
	secondparty.SecondParty = mockTiup

	task := &TaskEntity{
		Id: 123,
	}
	flowCtx := NewFlowContext(context.TODO())
	flowCtx.SetData(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
		ConfigPath:  "/tmp/test-ut",
	})
	flowCtx.SetData(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id:          123,
			StorageType: StorageTypeS3,
		},
		Cluster: &Cluster{
			Id:          "test-tidb123",
			ClusterName: "test-tidb",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
						Port: 4000,
					},
				},
			},
		},
	})
	flowCtx.SetData("backupTaskId", uint64(123))
	ret := importDataToCluster(task, flowCtx)
	assert.Equal(t, true, ret)
}

func Test_exportDataFromCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	mockTiup.EXPECT().MicroSrvDumpling(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)
	mockTiup.EXPECT().MicroSrvGetTaskStatus(gomock.Any(), gomock.Any()).Return(dbpb.TiupTaskStatus_Finished, "success", nil)
	secondparty.SecondParty = mockTiup

	task := &TaskEntity{
		Id: 123,
	}
	flowCtx := NewFlowContext(context.TODO())
	flowCtx.SetData(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
	})
	flowCtx.SetData(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id:          123,
			StorageType: StorageTypeS3,
		},
		Cluster: &Cluster{
			Id:          "test-tidb123",
			ClusterName: "test-tidb",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
						Port: 4000,
					},
				},
			},
		},
	})
	flowCtx.SetData("backupTaskId", uint64(123))
	ret := exportDataFromCluster(task, flowCtx)
	assert.Equal(t, true, ret)
}

func Test_buildDataImportConfig(t *testing.T) {
	task := &TaskEntity{}
	ctx := NewFlowContext(context.TODO())

	ctx.SetData(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
		ConfigPath:  "/tmp/test-ut",
	})
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
					},
				},
				PDServers: []*spec.PDSpec{
					{
						Host: "127.0.0.1",
					},
				},
			},
		},
	})
	info := ctx.GetData(contextDataTransportKey).(*ImportInfo)
	_ = os.MkdirAll(info.ConfigPath, os.ModePerm)
	ret := buildDataImportConfig(task, ctx)
	assert.Equal(t, true, ret)
	_ = os.RemoveAll(info.ConfigPath)
}

func Test_updateDataImportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
		ConfigPath:  "configPath",
	})
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := updateDataImportRecord(task, ctx)
	assert.Equal(t, true, ret)
}

func Test_updateDataExportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
	})
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := updateDataExportRecord(task, ctx)
	assert.Equal(t, true, ret)
}

func Test_cleanDataTransportDir(t *testing.T) {
	err := cleanDataTransportDir(context.TODO(), "/tmp/tiem")
	assert.NoError(t, err)
}

func Test_exportDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
	})
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := exportDataFailed(task, ctx)
	assert.Equal(t, true, ret)
}

func Test_importDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: common.S3StorageType,
	})
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := importDataFailed(task, ctx)
	assert.Equal(t, true, ret)
}

func Test_getDataExportFilePath_case1(t *testing.T) {
	request := &clusterpb.DataExportRequest{
		StorageType:     common.S3StorageType,
		BucketUrl:       "s3://test",
		AccessKey:       "admin",
		SecretAccessKey: "admin",
		EndpointUrl:     "https://minio.pingcap.net:9000",
	}
	path := getDataExportFilePath(request, "/tmp/test")
	assert.Equal(t, path, fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.GetBucketUrl(), request.GetAccessKey(), request.GetSecretAccessKey(), request.GetEndpointUrl()))
}

func Test_getDataExportFilePath_case2(t *testing.T) {
	request := &clusterpb.DataExportRequest{
		StorageType: common.NfsStorageType,
	}
	path := getDataExportFilePath(request, "/tmp/test")
	assert.Equal(t, path, "/tmp/test/data")
}

func Test_getDataImportFilePath_case1(t *testing.T) {
	request := &clusterpb.DataImportRequest{
		StorageType:     common.S3StorageType,
		BucketUrl:       "s3://test",
		AccessKey:       "admin",
		SecretAccessKey: "admin",
		EndpointUrl:     "https://minio.pingcap.net:9000",
	}
	path := getDataImportFilePath(request, "/tmp/test")
	assert.Equal(t, path, fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.GetBucketUrl(), request.GetAccessKey(), request.GetSecretAccessKey(), request.GetEndpointUrl()))
}

func Test_getDataImportFilePath_case2(t *testing.T) {
	request := &clusterpb.DataImportRequest{
		StorageType: common.NfsStorageType,
	}
	path := getDataImportFilePath(request, "/tmp/test")
	assert.Equal(t, path, "/tmp/test/data")
}
