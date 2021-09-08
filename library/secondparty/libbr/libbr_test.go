package libbr

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto/mocks"
	"os"
	"testing"
	"time"
)

var brMicro BrMicro
var dbConnParam DbConnParam
var storage BrStorage
var clusterFacade ClusterFacade

var cmdBackUpReq CmdBackUpReq
var cmdShowBackUpInfoReq CmdShowBackUpInfoReq
var cmdRestoreReq CmdRestoreReq
var cmdShowRestoreInfoReq CmdShowRestoreInfoReq

func init() {
	brMicro = BrMicro{}
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

	cmdBackUpReq = CmdBackUpReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowBackUpInfoReq = CmdShowBackUpInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}
	cmdRestoreReq = CmdRestoreReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowRestoreInfoReq = CmdShowRestoreInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}
	//log = logger.GetRootLogger("config_key_test_log")
	// MicroInit("../../../bin/micro-cluster/brmgr/brmgr", "/tmp/log/br/")
}

//func TestBrMicro_MicroInit(t *testing.T) {
//	brMicro.MicroInit("../../../bin/brcmd", "")
//	glMicroTaskStatusMapLen := len(glMicroTaskStatusMap)
//	glMicroCmdChanCap := cap(glMicroCmdChan)
//	if glMicroCmdChanCap != 1024 {
//		t.Errorf("glMicroCmdChan cap was incorrect, got: %d, want: %d.", glMicroCmdChanCap, 1024)
//	}
//	if glMicroTaskStatusMapLen != 0 {
//		t.Errorf("glMicroTaskStatusMap len was incorrect, got: %d, want: %d.", glMicroTaskStatusMapLen, 0)
//	}
//}

func TestBrMicro_BackUp_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := brMicro.BackUp(clusterFacade, storage, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestBrMicro_BackUp_Success(t *testing.T) {
	brMicro.MicroInitForTest("../../../bin/brcmd", "")

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

	taskID, err := brMicro.BackUp(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestBrMicro_ShowBackUpInfo(t *testing.T) {
	brMicro.MicroInitForTest("../../../bin/brcmd", "")

	resp := brMicro.ShowBackUpInfo(clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show backup info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func TestBrMicro_Restore_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := brMicro.Restore(clusterFacade, storage, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestBrMicro_Restore_Success(t *testing.T) {
	brMicro.MicroInitForTest("../../../bin/brcmd", "")

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

	taskID, err := brMicro.Restore(clusterFacade, storage, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestBrMicro_ShowRestoreInfo(t *testing.T) {
	brMicro.MicroInitForTest("../../../bin/brcmd", "")

	resp := brMicro.ShowRestoreInfo(clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show restore info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func TestBrMicro_MicroSrvTiupGetTaskStatus_Fail(t *testing.T) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupTaskByIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find tiup task by Id"

	expectedErr := errors.New("Fail Find tiup task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupTaskByID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := brMicro.MicroSrvTiupGetTaskStatus(1)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by id intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestBrMicro_MicroSrvTiupGetTaskStatus_Success(t *testing.T) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupTaskByIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.TiupTask = &dbPb.TiupTask{
		ID: 1,
		Status: dbPb.TiupTaskStatus_Init,
	}

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupTaskByID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, errStr, err := brMicro.MicroSrvTiupGetTaskStatus(1)
	if stat != dbPb.TiupTaskStatus_Init || errStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", errStr, nil, err)
	}
}

//// todo: need to start a TiDB cluster and generate some data before running this test
//func TestBackUpAndInfo(t *testing.T) {
//	mgrHandleCmdBackUpReq(string(jsonMustMarshal(cmdBackUpReq)))
//	loopUntilBackupDone()
//}
//
//// todo: need to start a TiDB cluster and generate some data before running this test
//func TestRestoreAndInfo(t *testing.T) {
//	mgrHandleCmdRestoreReq(string(jsonMustMarshal(cmdRestoreReq)))
//	loopUntilRestoreDone()
//}

func loopUntilBackupDone() {
	for {
		time.Sleep(time.Second)
		resp := mgrHandleCmdShowBackUpInfoReq(string(jsonMustMarshal(cmdShowBackUpInfoReq)))
		fmt.Println("taskId 0 status from api: ", resp)
		glMgrStatusMapSync()
		taskStatusMapValue := glMgrTaskStatusMap[0]
		fmt.Println("taskId 0 status: ", taskStatusMapValue.stat.Status)
		if taskStatusMapValue.stat.Status == TaskStatusError || taskStatusMapValue.stat.Status == TaskStatusFinished {
			break
		}
	}
}

func loopUntilRestoreDone() {
	for {
		time.Sleep(time.Second)
		resp := mgrHandleCmdShowRestoreInfoReq(string(jsonMustMarshal(cmdShowRestoreInfoReq)))
		fmt.Println("taskId 0 status from api: ", resp)
		glMgrStatusMapSync()
		taskStatusMapValue := glMgrTaskStatusMap[0]
		fmt.Println("taskId 0 status: ", taskStatusMapValue.stat.Status)
		if taskStatusMapValue.stat.Status == TaskStatusError || taskStatusMapValue.stat.Status == TaskStatusFinished {
			break
		}
	}
}

func (brMicro *BrMicro) MicroInitForTest(brMgrPath, mgrLogFilePath string) {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileLibBr)

	glBrMgrPath = brMgrPath
	glMicroTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	glMicroCmdChan = microStartBrMgr(mgrLogFilePath)
}