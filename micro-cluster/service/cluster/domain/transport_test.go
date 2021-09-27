package domain

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain/mock"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExportDataPreCheck_case1(t *testing.T) {
	req := &proto.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		FileType:    FileTypeCSV,
		StorageType: NfsStorageType,
	}
	err := ExportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestExportDataPreCheck_case2(t *testing.T) {
	req := &proto.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		FileType:    FileTypeCSV,
		StorageType: "local",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case3(t *testing.T) {
	req := &proto.DataExportRequest{
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
	req := &proto.DataExportRequest{
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
	req := &proto.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
		BucketUrl: "s3://test",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case6(t *testing.T) {
	req := &proto.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
		BucketUrl: "s3://test",
		AccessKey: "admin",
	}
	err := ExportDataPreCheck(req)
	assert.NotNil(t, err)
}

func TestExportDataPreCheck_case7(t *testing.T) {
	req := &proto.DataExportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FileType:    FileTypeCSV,
		StorageType: S3StorageType,
		EndpointUrl: "https://minio.pingcap.net:9000",
		BucketUrl: "s3://test",
		AccessKey: "admin",
		SecretAccessKey: "admin",
	}
	err := ExportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestImportDataPreCheck(t *testing.T) {
	req := &proto.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		StorageType: NfsStorageType,
	}
	err := ImportDataPreCheck(req)
	assert.NoError(t, err)
}

func TestExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &proto.DataExportRequest{
		Operator: &proto.OperatorDTO{
			Id: "123",
			Name: "123",
			TenantId: "123",
		},
	}
	_, err := ExportData(request)

	assert.NoError(t, err)
}

func TestImportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &proto.DataImportRequest{
		Operator: &proto.OperatorDTO{
			Id: "123",
			Name: "123",
			TenantId: "123",
		},
	}
	_, err := ImportData(request)

	assert.NoError(t, err)
}

func TestDescribeDataTransportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().ListTrasnportRecord(gomock.Any(), gomock.Any()).Return(&db.DBListTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	request := &proto.DataTransportQueryRequest{
		Operator: &proto.OperatorDTO{
			Id: "123",
			Name: "123",
			TenantId: "123",
		},
	}
	_, _, err := DescribeDataTransportRecord(request.GetOperator(), "123", "123", 1, 10)
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
		RecordId:    "123",
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
	ret := buildDataImportConfig(task, context)
	assert.Equal(t, true, ret)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	err := os.RemoveAll(info.ConfigPath)
	t.Log(err)
}

func Test_updateDataImportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    "123",
		StorageType: S3StorageType,
		ConfigPath:  "configPath",
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := updateDataImportRecord(task, context)

	assert.Equal(t, true, ret)
}

func Test_updateDataExportRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    "123",
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := updateDataExportRecord(task, context)

	assert.Equal(t, true, ret)
}

func Test_exportDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ExportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    "123",
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := exportDataFailed(task, context)

	assert.Equal(t, true, ret)
}

func Test_importDataFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBUpdateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextDataTransportKey, &ImportInfo{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		RecordId:    "123",
		StorageType: S3StorageType,
	})
	context.put(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})

	ret := importDataFailed(task, context)

	assert.Equal(t, true, ret)
}

func Test_getDataImportConfigDir(t *testing.T) {
	dir := getDataImportConfigDir("test-abc", TransportTypeExport)
	assert.Equal(t, dir, fmt.Sprintf("%s/%s/%s", defaultTransportDirPrefix, "test-abc", TransportTypeExport))
}

func Test_getDataExportFilePath_case1(t *testing.T) {
	request := &proto.DataExportRequest{
		StorageType: S3StorageType,
		BucketUrl: "s3://test",
		AccessKey: "admin",
		SecretAccessKey: "admin",
		EndpointUrl: "https://minio.pingcap.net:9000",
	}
	path := getDataExportFilePath(request)
	assert.Equal(t, path, fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.GetBucketUrl(), request.GetAccessKey(), request.GetSecretAccessKey(), request.GetEndpointUrl()))
}

func Test_getDataExportFilePath_case2(t *testing.T) {
	request := &proto.DataExportRequest{
		StorageType: NfsStorageType,
		FilePath: "/tmp/test",
	}
	path := getDataExportFilePath(request)
	assert.Equal(t, path, "/tmp/test")
}