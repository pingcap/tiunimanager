package domain

import (
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/secondparty"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	mock "github.com/pingcap-inc/tiem/micro-metadb/proto/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSaveBackupStrategyPreCheck_case1(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:00-16:00",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NoError(t, err)
}

func TestSaveBackupStrategyPreCheck_case3(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:0016:00",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}

func TestSaveBackupStrategyPreCheck_case4(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "18-16",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}

func TestSaveBackupStrategyPreCheck_case5(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:00-16:00",
		BackupDate: "Monday,weekday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}

func TestBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().SaveBackupRecord(gomock.Any(), gomock.Any()).Return(&db.DBSaveBackupRecordResponse{}, nil)
	client.DBClient = mockClient

	_, err := Backup(&proto.OperatorDTO{
		Id: "123",
		Name: "123",
		TenantId: "123",
	}, "test-abc", "", "", BackupModeManual, "")

	assert.NoError(t, err)
}

func TestRecoverPreCheck(t *testing.T) {
	request := &proto.RecoverRequest{
		Operator: &proto.OperatorDTO{
			Id: "123",
			Name: "123",
		},
		Cluster: &proto.ClusterBaseInfoDTO{
			RecoverInfo: &proto.RecoverInfoDTO{
				BackupRecordId: 123,
				SourceClusterId: "test-tidb",
			},
		},
	}

	err := RecoverPreCheck(request)

	assert.NoError(t, err)
}

func TestRecover(t *testing.T) {
	_, err := Recover(&proto.OperatorDTO{
		Id: "123",
		Name: "123",
		TenantId: "123",
	}, &proto.ClusterBaseInfoDTO{
			ClusterName: "test-tidb",
			ClusterVersion: &proto.ClusterVersionDTO{Code: "v4.0.12", Name: "v4.0.12"},
			ClusterType: &proto.ClusterTypeDTO{Code: "TiDB", Name: "TiDB"},
	}, nil)

	assert.NoError(t, err)
}

func TestDeleteBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any()).Return(&db.DBQueryBackupRecordResponse{}, nil)
	mockClient.EXPECT().DeleteBackupRecord(gomock.Any(), gomock.Any()).Return(&db.DBDeleteBackupRecordResponse{}, nil)
	client.DBClient = mockClient

	err := DeleteBackup(&proto.OperatorDTO{
		Id: "123",
		Name: "123",
		TenantId: "123",
	}, "test-abc", 123)

	assert.NoError(t, err)
}

func TestSaveBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().SaveBackupStrategy(gomock.Any(), gomock.Any()).Return(&db.DBSaveBackupStrategyResponse{}, nil)
	client.DBClient = mockClient

	err := SaveBackupStrategy(&proto.OperatorDTO{
		Id: "123",
		Name: "123",
		TenantId: "123",
	}, &proto.BackupStrategy{
		ClusterId: "test-abc",
		BackupDate: "Monday,Sunday",
		Period: "12:00-13:00",
	})
	assert.NoError(t, err)
}

func TestQueryBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().QueryBackupStrategy(gomock.Any(), gomock.Any()).Return(&db.DBQueryBackupStrategyResponse{
		Strategy: &db.DBBackupStrategyDTO{
			TenantId: "123",
			OperatorId: "123",
			ClusterId: "test-abc",
			BackupDate: "Monday,Sunday",
			StartHour: 12,
			EndHour: 13,
		},
	}, nil)
	client.DBClient = mockClient

	_, err := QueryBackupStrategy(&proto.OperatorDTO{
		Id: "123",
		Name: "123",
		TenantId: "123",
	}, "test-abc")

	assert.NoError(t, err)
}

func Test_calculateNextBackupTime_case1(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Monday,Tuesday,Thursday"
	startHour := 10
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 20, 10, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case2(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Monday,Tuesday,Thursday,Sunday"
	startHour := 10
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 19, 10, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case3(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Friday"
	startHour := 11
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 21, 11, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case4(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Friday"
	startHour := 12
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 17, 12, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case5(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Monday,Tuesday,Wednesday,Thursday,Friday,Saturday,Sunday"
	startHour := 12
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	t.Log(nextBackupTime)
	assert.Equal(t, time.Date(2021, 9, 17, 12, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case6(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Weekday"
	startHour := 12
	_, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NotNil(t, err)
}

func Test_convertBrStorageType(t *testing.T) {
	result, _ := convertBrStorageType(string(StorageTypeS3))
	assert.Equal(t, secondparty.StorageTypeS3, result)
	result, _ = convertBrStorageType(string(StorageTypeLocal))
	assert.Equal(t, secondparty.StorageTypeLocal, result)
	_, err := convertBrStorageType("data")
	assert.NotNil(t, err)
}

func Test_updateBackupRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateBackupRecord(gomock.Any(), gomock.Any()).Return(&db.DBUpdateBackupRecordResponse{}, nil)
	client.DBClient = mockClient

	task := &TaskEntity{}
	context := &FlowContext{}
	context.put(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id: 123,
			Size: 1000,
		},
	})
	ret := updateBackupRecord(task, context)

	assert.Equal(t, true, ret)
}