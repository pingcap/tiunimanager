package libbr

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

var brMicro2 BrMicro

func init() {
	brMicro2 = BrMicro{}
	// make sure log would not cause nil pointer problem
	logger = framework.LogForkFile(common.LogFileBrMgr)
	glMicroCmdChan = make(chan CmdChanMember, 1024)
}

func TestBrMicro_MicroInit(t *testing.T) {
	brMicro2.MicroInit("../../../bin/brcmd", "")
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

	//sync part mock
	syncReq := dbPb.UpdateTiupTaskRequest{
		Id:     1,
		Status: dbPb.TiupTaskStatus_Error,
		ErrStr: "dial tcp 127.0.0.1:4000: connect: connection refused\n" ,
	}
	syncResp := dbPb.UpdateTiupTaskResponse{
		ErrCode: 0,
		ErrStr: "",
	}
	mockDBClient.EXPECT().UpdateTiupTask(context.Background(), &syncReq).Return(&syncResp, nil)

	time.Sleep(1500*time.Millisecond)
}