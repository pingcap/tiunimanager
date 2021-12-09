/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: task_status_manager_v2_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/test/mockdb"
)

var secondPartyManager *SecondPartyManager

func init() {
	secondPartyManager = &SecondPartyManager{
		TiupBinPath: "mock_tiup",
	}
	secondPartyManager.Init()
}

func TestSecondPartyManager_Init(t *testing.T) {
	syncedTaskStatusMapLen := len(secondPartyManager.syncedTaskStatusMap)
	taskStatusChCap := cap(secondPartyManager.taskStatusCh)
	taskStatusMapLen := len(secondPartyManager.taskStatusMap)
	if syncedTaskStatusMapLen != 0 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 0)
	}
	if taskStatusChCap != 1024 {
		t.Errorf("taskStatusChCap cap is incorrect, got: %d, want: %d.", taskStatusChCap, 1024)
	}
	if taskStatusMapLen != 0 {
		t.Errorf("taskStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_NothingUpdate(t *testing.T) {
	time.Sleep(1500 * time.Microsecond)
	syncedTaskStatusMapLen := len(secondPartyManager.syncedTaskStatusMap)
	taskStatusMapLen := len(secondPartyManager.taskStatusMap)
	if syncedTaskStatusMapLen != 0 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 0)
	}
	if taskStatusMapLen != 0 {
		t.Errorf("taskStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_updateButFail(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient

	syncReq := dbPb.UpdateTiupOperatorRecordRequest{
		Id:     1,
		Status: dbPb.TiupTaskStatus_Processing,
		ErrStr: "",
	}
	syncResp := dbPb.UpdateTiupOperatorRecordResponse{
		ErrCode: 0,
		ErrStr:  "",
	}
	mockDBClient.EXPECT().UpdateTiupOperatorRecord(context.Background(), &syncReq).Return(&syncResp, errors.New("fail update"))

	secondPartyManager.taskStatusCh <- TaskStatusMember{
		TaskID:   1,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}

	time.Sleep(1500 * time.Millisecond)

	syncedTaskStatusMapLen := len(secondPartyManager.syncedTaskStatusMap)
	if syncedTaskStatusMapLen != 1 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 1)
	}
	v := secondPartyManager.syncedTaskStatusMap[1]
	if !v.validFlag || v.stat.Status != TaskStatusProcessing {
		t.Errorf("TaskStatus for 1 is incorrect, got: %v %v, want: %v %v", v.validFlag, v.stat.Status, true, TaskStatusProcessing)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_updateAndSucceed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient

	syncReq := dbPb.UpdateTiupOperatorRecordRequest{
		Id:     1,
		Status: dbPb.TiupTaskStatus_Finished,
		ErrStr: "",
	}
	syncResp := dbPb.UpdateTiupOperatorRecordResponse{
		ErrCode: 0,
		ErrStr:  "",
	}
	mockDBClient.EXPECT().UpdateTiupOperatorRecord(context.Background(), &syncReq).Return(&syncResp, nil)

	secondPartyManager.taskStatusCh <- TaskStatusMember{
		TaskID:   1,
		Status:   TaskStatusFinished,
		ErrorStr: "",
	}

	time.Sleep(1500 * time.Millisecond)
	syncedTaskStatusMapLen := len(secondPartyManager.syncedTaskStatusMap)
	if syncedTaskStatusMapLen != 1 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 1)
	}
	v := secondPartyManager.syncedTaskStatusMap[1]
	if !v.validFlag || v.stat.Status != TaskStatusFinished {
		t.Errorf("TaskStatus for 1 is incorrect, got: %v %v, want: %v %v", v.validFlag, v.stat.Status, true, TaskStatusFinished)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_DeleteInvalidTaskStatus(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
	taskStatusMapLen := len(secondPartyManager.taskStatusMap)
	if taskStatusMapLen != 0 {
		t.Errorf("taskStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}
