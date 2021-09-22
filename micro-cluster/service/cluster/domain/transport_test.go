package domain

import (
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
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
		StorageType: S3StorageType,
	}
	if err := ExportDataPreCheck(req); err != nil {
		t.Errorf("TestExportDataPreCheck failed, %s", err.Error())
		return
	}
	t.Log("TestExportDataPreCheck success")
}

func TestImportDataPreCheck(t *testing.T) {
	req := &proto.DataImportRequest{
		ClusterId:   "test-abc",
		UserName:    "root",
		Password:    "",
		FilePath:    "filePath",
		StorageType: S3StorageType,
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
		CurrentTiUPConfigRecord: &TiUPConfigRecord{
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
	os.RemoveAll(info.ConfigPath)
}
