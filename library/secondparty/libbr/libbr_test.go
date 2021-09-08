package libbr

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto/mocks"
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
	fmt.Println("init before test")
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
	fmt.Println("initialized")
}

func TestBrCmd_BrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	mgrTaskStatusMapLen := len(glMgrTaskStatusMap)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
	if mgrTaskStatusMapLen != 0 {
		t.Errorf("mgrTaskStatusMap len war incorrect, got: %d, want: %d.", mgrTaskStatusMapLen, 0)
	}
}

func TestBrMicro_MicroInit(t *testing.T) {
	brMicro.MicroInit("../../../bin/brcmd", "")
	glMicroTaskStatusMapLen := len(glMicroTaskStatusMap)
	glMicroCmdChanCap := cap(glMicroCmdChan)
	if glMicroCmdChanCap != 1024 {
		t.Errorf("glMicroCmdChan cap was incorrect, got: %d, want: %d.", glMicroCmdChanCap, 1024)
	}
	if glMicroTaskStatusMapLen != 0 {
		t.Errorf("glMicroTaskStatusMap len war incorrect, got: %d, want: %d.", glMicroTaskStatusMapLen, 0)
	}
}

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
	brMicro.MicroInit("../../../bin/brcmd", "")

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
	brMicro.MicroInit("../../../bin/brcmd", "")

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
	brMicro.MicroInit("../../../bin/brcmd", "")

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
	brMicro.MicroInit("../../../bin/brcmd", "")

	resp := brMicro.ShowRestoreInfo(clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show restore info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
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
