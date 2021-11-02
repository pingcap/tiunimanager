package secondparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	db "github.com/pingcap-inc/tiem/library/client/metadb/dbpb/mocks"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var secondMicro2 *SecondMicro

var dbConnParam DbConnParam
var storage BrStorage
var clusterFacade ClusterFacade

func init() {
	secondMicro2 = &SecondMicro{}
	dbConnParam = DbConnParam{
		Username: "root",
		Ip:       "127.0.0.1",
		Port:     "4000",
	}
	storage = BrStorage{
		StorageType: StorageTypeLocal,
		Root:        "/tmp/backup",
	}
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
	}

	microInitForTestLibbr("")
}

func TestSecondMicro_BackUp_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro2.MicroSrvBackUp(clusterFacade, storage, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_BackUp_Success1_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		TableName: "testTbl",
		DbName: "testDb",
		RateLimitM: "1",
		Concurrency: "1",
		CheckSum: "1",
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvBackUp(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_BackUp_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName: "testDb",
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvBackUp(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_BackUp_Success3_DontCareAsyncResult(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvBackUp(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_ShowBackUpInfo_Fail(t *testing.T) {
	resp := secondMicro2.MicroSrvShowBackUpInfo(clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show backup info. either Destination(%s) or ErrorStr(%s) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func Test_execShowBackUpInfoThruSQL_Fail(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW BACKUPS").
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	resp := execShowBackUpInfoThruSQL(db, "SHOW BACKUPS")
	if resp.Destination == "" && resp.ErrorStr != "some error" {
		t.Errorf("case: show backup info. Destination(%s) should have zero value, and ErrorStr(%v) should be 'some error'", resp.Destination, resp.ErrorStr)
	}
}

func Test_execShowBackUpInfoThruSQL_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW BACKUPS").
		WillReturnError(fmt.Errorf("sql: no rows in result set"))
	mock.ExpectRollback()

	resp := execShowBackUpInfoThruSQL(db, "SHOW BACKUPS")
	if resp.Progress != 100 && resp.ErrorStr != "" {
		t.Errorf("case: show backup info. Progress(%f) should be 100, and ErrorStr(%v) should have zero value", resp.Progress, resp.ErrorStr)
	}
}

func TestSecondMicro_Restore_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro2.MicroSrvRestore(clusterFacade, storage, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_Restore_Success1_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		TableName: "testTbl",
		DbName: "testDb",
		RateLimitM: "1",
		Concurrency: "1",
		CheckSum: "1",
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvRestore(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_Restore_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName: "testDb",
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvRestore(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_Restore_Success3_DontCareAsyncResult(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro2.MicroSrvRestore(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_ShowRestoreInfo_Fail(t *testing.T) {
	resp := secondMicro2.MicroSrvShowRestoreInfo(clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show restore info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func Test_execShowRestoreInfoThruSQL_Fail(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW RESTORES").
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	resp := execShowRestoreInfoThruSQL(db, "SHOW RESTORES")
	if resp.Destination == "" && resp.ErrorStr != "some error" {
		t.Errorf("case: show restore info. Destination(%s) should have zero value, and ErrorStr(%v) should be 'some error'", resp.Destination, resp.ErrorStr)
	}
}

func Test_execShowRestoreInfoThruSQL_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW RESTORES").
		WillReturnError(fmt.Errorf("sql: no rows in result set"))
	mock.ExpectRollback()

	resp := execShowRestoreInfoThruSQL(db, "SHOW RESTORES")
	if resp.Progress != 100 && resp.ErrorStr != "" {
		t.Errorf("case: show restore info. Progress(%f) should be 100, and ErrorStr(%v) should have zero value", resp.Progress, resp.ErrorStr)
	}
}

func microInitForTestLibbr(mgrLogFilePath string) {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileLibTiup)

	secondMicro2.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondMicro2.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondMicro2.taskStatusMap = make(map[uint64]TaskStatusMapValue)
}

func resetVariable() {
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
	}
}