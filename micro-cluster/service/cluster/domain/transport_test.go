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
	ctx "context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	mock "github.com/pingcap-inc/tiem/test/mock"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
)

func TestExportDataPreCheck_case1(t *testing.T) {
	req := &clusterpb.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: NfsStorageType,
	}
	err := ExportDataPreCheck(req)
	assert.NoError(t, err)
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
		StorageType: S3StorageType,
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
		StorageType: S3StorageType,
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
		StorageType: S3StorageType,
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
		StorageType: S3StorageType,
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
		StorageType:     S3StorageType,
		EndpointUrl:     "https://minio.pingcap.net:9000",
		BucketUrl:       "s3://test",
		AccessKey:       "admin",
		SecretAccessKey: "admin",
	}
	err := ExportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestImportDataPreCheck(t *testing.T) {
	req := &clusterpb.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		StorageType: NfsStorageType,
	}
	err := ImportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataExportRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, err := ExportData(ctx.Background(), request)

	assert.NoError(t, err)
}

func TestImportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataImportRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, err := ImportData(ctx.Background(), request)

	assert.NoError(t, err)
}

func TestDescribeDataTransportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().ListTrasnportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBListTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &clusterpb.DataTransportQueryRequest{
		Operator: &clusterpb.OperatorDTO{
			Id:       "123",
			Name:     "123",
			TenantId: "123",
		},
	}
	_, _, err := DescribeDataTransportRecord(ctx.Background(), request.GetOperator(), 123, "123", 1, 10)
	assert.NoError(t, err)
}

func Test_buildDataImportConfig(t *testing.T) {
	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: S3StorageType,
		ConfigPath:  "configPath",
	})
	context.put(contextClusterKey, &ClusterAggregation{
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
	context.put(contextCtxKey, ctx.Background())
	ret := buildDataImportConfig(task, context)
	assert.Equal(t, true, ret)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	_ = os.RemoveAll(info.ConfigPath)
}

func Test_updateDataImportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: S3StorageType,
		ConfigPath:  "configPath",
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})
	context.put(contextCtxKey, ctx.Background())

	ret := updateDataImportRecord(task, context)

	assert.Equal(t, true, ret)
}

func Test_updateDataExportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})
	context.put(contextCtxKey, ctx.Background())

	ret := updateDataExportRecord(task, context)

	assert.Equal(t, true, ret)
}

func Test_exportDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})
	context.put(contextCtxKey, ctx.Background())

	ret := exportDataFailed(task, context)

	assert.Equal(t, true, ret)
}

func Test_importDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&dbpb.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    123,
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})
	context.put(contextCtxKey, ctx.Background())

	ret := importDataFailed(task, context)

	assert.Equal(t, true, ret)
}

func Test_getDataExportFilePath_case1(t *testing.T) {
	request := &clusterpb.DataExportRequest{
		StorageType:     S3StorageType,
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
		StorageType: NfsStorageType,
	}
	path := getDataExportFilePath(request, "/tmp/test")
	assert.Equal(t, path, "/tmp/test/data")
}

func Test_getDataImportFilePath_case1(t *testing.T) {
	request := &clusterpb.DataImportRequest{
		StorageType:     S3StorageType,
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
		StorageType: NfsStorageType,
	}
	path := getDataImportFilePath(request, "/tmp/test")
	assert.Equal(t, path, "/tmp/test/data")
}
