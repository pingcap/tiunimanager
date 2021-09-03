package libbr

import (
	"fmt"
	"testing"
	"time"
)

var dbConnParam DbConnParam
var storage BrStorage
var cmdBackUpReq CmdBackUpReq
var cmdShowBackUpInfoReq CmdShowBackUpInfoReq
var cmdRestoreReq CmdRestoreReq
var cmdShowRestoreInfoReq CmdShowRestoreInfoReq

func init() {
	fmt.Println("init before test")
	dbConnParam = DbConnParam {
		Username: "root",
		Ip: "127.0.0.1",
		Port: "57395",
	}
	storage = BrStorage {
		StorageType: StorageTypeLocal,
		Root: "/tmp/backup",
	}

	cmdBackUpReq = CmdBackUpReq{
		TaskID: 0,
		DbConnParameter: dbConnParam,
		StorageAddress: fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowBackUpInfoReq = CmdShowBackUpInfoReq {
		TaskID: 0,
		DbConnParameter: dbConnParam,
	}
	cmdRestoreReq = CmdRestoreReq{
		TaskID: 0,
		DbConnParameter: dbConnParam,
		StorageAddress: fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowRestoreInfoReq = CmdShowRestoreInfoReq{
		TaskID: 0,
		DbConnParameter: dbConnParam,
	}
	//log = logger.GetRootLogger("config_key_test_log")
	BrMgrInit()
	// MicroInit("../../../bin/micro-cluster/brmgr/brmgr", "/tmp/log/br/")
}

func TestBrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
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