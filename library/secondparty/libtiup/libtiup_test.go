package libtiup

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto/mocks"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/library/common"

	"github.com/pingcap-inc/tiem/library/framework"
)

var tiUPMicro TiUPMicro

func init() {
	tiUPMicro = TiUPMicro{}
	// make sure log would not cause nil pointer problem
	logger = framework.LogForkFile(common.LogFileTiupMgr)
	TiupMgrInit()
}

const (
	cmdDeployNonJsonStr = "nonJsonStr"
)

func TestBrMicro_MicroInit(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "tiup", "")
	glMicroTaskStatusMapLen := len(glMicroTaskStatusMap)
	glMicroCmdChanCap := cap(glMicroCmdChan)
	if glMicroCmdChanCap != 1024 {
		t.Errorf("glMicroCmdChan cap was incorrect, got: %d, want: %d.", glMicroCmdChanCap, 1024)
	}
	if glMicroTaskStatusMapLen != 0 {
		t.Errorf("glMicroTaskStatusMap len was incorrect, got: %d, want: %d.", glMicroTaskStatusMapLen, 0)
	}
}

func TestTiUPMicro_MicroSrvTiupDeploy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupDeploy("test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupDeploy_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupDeploy("test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupList_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_List
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupList(0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupList_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_List
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupList(0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupStart_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupStart("test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupStart_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupStart("test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupDestroy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupDestroy("test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupDestroy_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupDestroy("test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupDumpling_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupDumpling(0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupDumpling_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupDumpling(0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupLightning_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := tiUPMicro.MicroSrvTiupLightning(0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupLightning_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := tiUPMicro.MicroSrvTiupLightning(0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupClusterDisplay(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	resp := tiUPMicro.MicroSrvTiupClusterDisplay("test-tidb", 0, []string{})
	if resp.DisplayRespString == "" && resp.ErrorStr == "" {
		t.Errorf("case: cluster display. either DisplayRespString(%s) or ErrorStr(%v) should have zero value", resp.DisplayRespString, resp.ErrorStr)
	}
}

func TestTiUPMicro_MicroSrvTiupGetTaskStatus_Fail(t *testing.T) {
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

	_, errStr, err := tiUPMicro.MicroSrvTiupGetTaskStatus(1)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by id intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupGetTaskStatus_Success(t *testing.T) {
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

	stat, errStr, err := tiUPMicro.MicroSrvTiupGetTaskStatus(1)
	if stat != dbPb.TiupTaskStatus_Init || errStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", errStr, nil, err)
	}
}

func TestTiUPMicro_MicroSrvTiupGetTaskStatusByBizID_Fail(t *testing.T) {
	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = 0

	var resp dbPb.GetTiupTaskStatusByBizIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find tiup task by BizId"

	expectedErr := errors.New("Fail Find tiup task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupTaskStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := tiUPMicro.MicroSrvTiupGetTaskStatusByBizID(0)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by BizId intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestTiUPMicro_MicroSrvTiupGetTaskStatusByBizID_Success(t *testing.T) {
	tiUPMicro.MicroInit("../../../bin/tiupcmd", "", "")

	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = 0

	var resp dbPb.GetTiupTaskStatusByBizIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.Stat = dbPb.TiupTaskStatus_Init
	resp.StatErrStr = ""

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupTaskStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, statErrStr, err := tiUPMicro.MicroSrvTiupGetTaskStatusByBizID(0)
	if stat != dbPb.TiupTaskStatus_Init || statErrStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), statErrStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", statErrStr, nil, err)
	}
}

func loopUntilDone() {
	for {
		time.Sleep(time.Second)
		glMgrStatusMapSync()
		taskStatusMapValue := glMgrTaskStatusMap[0]
		fmt.Println("taskId 0 status: ", taskStatusMapValue.stat.Status)
		if taskStatusMapValue.stat.Status == TaskStatusError || taskStatusMapValue.stat.Status == TaskStatusFinished {
			break
		}
	}
}

func TestMgrHandleCmdDeployReqWithNonJsonStr(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		mgrHandleCmdDeployReq(cmdDeployNonJsonStr)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMgrHandleCmdDeployReqWithNonJsonStr")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdDeployReq(t *testing.T) {
	cmdDeployReq := CmdDeployReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		Version: "v5.1.1",
		ConfigStrYaml: "global:\n  user: \"root\"\n  ssh_port: 22\n  deploy_dir: \"/root/tidb-deploy\"\n  data_dir: \"/root/tidb-data\"\n  arch: \"amd64\"\n\nmonitored:\n  node_exporter_port: 9100\n  blackbox_exporter_port: 9115\n\npd_servers:\n  - host: 172.16.5.128\n\ntidb_servers:\n  - host: 172.16.5.128\n    port: 4000\n    status_port: 10080\n    deploy_dir: \"/root/tidb-deploy/tidb-4000\"\n    log_dir: \"/root/tidb-deploy/tidb-4000/log\"\n\ntikv_servers:\n  - host: 172.16.5.128\n    port: 20160\n    status_port: 20180\n    deploy_dir: \"/root/data1/tidb-deploy/tikv-20160\"\n    data_dir: \"/root/tidb-data/tikv-20160\"\n    log_dir: \"/root/tidb-deploy/tikv-20160/log\"\n\nmonitoring_servers:\n  - host: 172.16.5.128\n\ngrafana_servers:\n  - host: 172.16.5.128\n\nalertmanager_servers:\n  - host: 172.16.5.128",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdDeployReq(string(jsonMustMarshal(cmdDeployReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdStartReq(t *testing.T) {
	cmdStartReq := CmdStartReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdStartReq(string(jsonMustMarshal(cmdStartReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdDestroyReq(t *testing.T) {
	cmdDestroyReq := CmdDestroyReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdDestroyReq(string(jsonMustMarshal(cmdDestroyReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleClusterDisplayReq(t *testing.T) {
	cmdClusterDisplayReq := CmdClusterDisplayReq{
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	resp := mgrHandleClusterDisplayReq(string(jsonMustMarshal(cmdClusterDisplayReq)))
	fmt.Println("tiup cluster display resp: ", resp)
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/
