package libtiup

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto/mocks"
	"testing"
	"time"
)

var tiUPMicro2 TiUPMicro

func init() {
	tiUPMicro2 = TiUPMicro{}
	// make sure log would not cause nil pointer problem
	logger = framework.LogForkFile(common.LogFileTiupMgr)
	glMicroCmdChan = make(chan CmdChanMember, 1024)
}

func TestTiUPMicro_MicroInit(t *testing.T) {
	tiUPMicro2.MicroInit("../../../bin/tiupcmd", "tiup", "")
	glMicroTaskStatusMapLen := len(glMicroTaskStatusMap)
	glMicroCmdChanCap := cap(glMicroCmdChan)
	if glMicroCmdChanCap != 1024 {
		t.Errorf("glMicroCmdChan cap was incorrect, got: %d, want: %d.", glMicroCmdChanCap, 1024)
	}
	if glMicroTaskStatusMapLen != 0 {
		t.Errorf("glMicroTaskStatusMap len was incorrect, got: %d, want: %d.", glMicroTaskStatusMapLen, 0)
	}
}

func Test_glMicroTaskStatusMapSyncer_NoNeedUpdate(t *testing.T) {
	time.Sleep(1500*time.Microsecond)
}

func Test_glMicroTaskStatusMapSyncer_NeedUpdate(t *testing.T) {
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

	taskID, err := tiUPMicro2.MicroSrvTiupDeploy("test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
	time.Sleep(1500*time.Millisecond)
}

//func TestBrMicro_MicroInit(t *testing.T) {
//	tiUPMicro.MicroInit("../../../bin/tiupcmd", "tiup", "")
//	glMicroTaskStatusMapLen := len(glMicroTaskStatusMap)
//	glMicroCmdChanCap := cap(glMicroCmdChan)
//	if glMicroCmdChanCap != 1024 {
//		t.Errorf("glMicroCmdChan cap was incorrect, got: %d, want: %d.", glMicroCmdChanCap, 1024)
//	}
//	if glMicroTaskStatusMapLen != 0 {
//		t.Errorf("glMicroTaskStatusMap len was incorrect, got: %d, want: %d.", glMicroTaskStatusMapLen, 0)
//	}
//}