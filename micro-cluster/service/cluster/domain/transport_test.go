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

func TestExportDataPreCheck(t *testing.T) {
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

func TestBuildDataImportConfig(t *testing.T) {
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
		EndpointUrl: "http://minio.pingcap.net:9000",
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

func TestExportData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateTransportRecord(gomock.Any(), gomock.Any()).Return(&db.DBCreateTransportRecordResponse{}, nil)
	client.DBClient = mockClient

	mockClusterRepo := mock.NewMockClusterRepository(ctrl)
	mockClusterRepo.EXPECT().Persist(gomock.Any()).Return(nil)
	mockClusterRepo.EXPECT().Load(gomock.Any()).Return(&ClusterAggregation{}, nil)
	ClusterRepo = mockClusterRepo



	request := &proto.DataExportRequest{
		Operator: &proto.OperatorDTO{
			Id: "ope",
			Name: "test123",
			TenantId: "test123",
		},
	}
	_, err := ExportData(request)
	assert.Equal(t, nil, err)
}