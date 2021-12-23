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

	"github.com/pingcap-inc/tiem/test/mockmodels/mocksecondparty"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"github.com/golang/mock/gomock"
)

var secondPartyManager *SecondPartyManager

func init() {
	secondPartyManager = &SecondPartyManager{
		TiUPBinPath: "mock_tiup",
	}
	models.MockDB()
	secondPartyManager.Init()
}

func TestSecondPartyManager_Init(t *testing.T) {
	syncedTaskStatusMapLen := len(secondPartyManager.syncedOperationStatusMap)
	taskStatusChCap := cap(secondPartyManager.operationStatusCh)
	taskStatusMapLen := len(secondPartyManager.operationStatusMap)
	if syncedTaskStatusMapLen != 0 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 0)
	}
	if taskStatusChCap != 1024 {
		t.Errorf("taskStatusChCap cap is incorrect, got: %d, want: %d.", taskStatusChCap, 1024)
	}
	if taskStatusMapLen != 0 {
		t.Errorf("operationStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_NothingUpdate(t *testing.T) {
	time.Sleep(1500 * time.Microsecond)
	syncedTaskStatusMapLen := len(secondPartyManager.syncedOperationStatusMap)
	taskStatusMapLen := len(secondPartyManager.operationStatusMap)
	if syncedTaskStatusMapLen != 0 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 0)
	}
	if taskStatusMapLen != 0 {
		t.Errorf("operationStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_updateButFail(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID:       TestOperationID,
		Status:   secondparty.OperationStatus_Processing,
		Result:   "",
		ErrorStr: "",
	}
	expectedErr := errors.New("fail update second party operation")

	mockReaderWriter.EXPECT().Update(context.Background(), &secondPartyOperation).Return(expectedErr)

	secondPartyManager.operationStatusCh <- OperationStatusMember{
		OperationID: TestOperationID,
		Status:      secondparty.OperationStatus_Processing,
		Result:      "",
		ErrorStr:    "",
	}

	time.Sleep(1500 * time.Millisecond)

	syncedTaskStatusMapLen := len(secondPartyManager.syncedOperationStatusMap)
	if syncedTaskStatusMapLen != 1 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 1)
	}
	v := secondPartyManager.syncedOperationStatusMap[TestOperationID]
	if !v.validFlag || v.stat.Status != secondparty.OperationStatus_Processing {
		t.Errorf("TaskStatus for 1 is incorrect, got: %v %v, want: %v %v", v.validFlag, v.stat.Status, true, secondparty.OperationStatus_Processing)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_updateAndSucceed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID:       TestOperationID,
		Status:   secondparty.OperationStatus_Finished,
		Result:   "",
		ErrorStr: "",
	}
	mockReaderWriter.EXPECT().Update(context.Background(), &secondPartyOperation).Return(nil)

	secondPartyManager.operationStatusCh <- OperationStatusMember{
		OperationID: TestOperationID,
		Status:      secondparty.OperationStatus_Finished,
		Result:      "",
		ErrorStr:    "",
	}

	time.Sleep(1500 * time.Millisecond)
	syncedTaskStatusMapLen := len(secondPartyManager.syncedOperationStatusMap)
	if syncedTaskStatusMapLen != 1 {
		t.Errorf("syncedTaskStatusMapLen len is incorrect, got: %d, want: %d.", syncedTaskStatusMapLen, 1)
	}
	v := secondPartyManager.syncedOperationStatusMap[TestOperationID]
	if !v.validFlag || v.stat.Status != secondparty.OperationStatus_Finished {
		t.Errorf("TaskStatus for 1 is incorrect, got: %v %v, want: %v %v", v.validFlag, v.stat.Status, true, secondparty.OperationStatus_Finished)
	}
}

func TestSecondPartyManager_taskStatusMapSyncer_DeleteInvalidTaskStatus(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
	taskStatusMapLen := len(secondPartyManager.operationStatusMap)
	if taskStatusMapLen != 0 {
		t.Errorf("operationStatusMap len is incorrect, got: %d, want: %d.", taskStatusMapLen, 0)
	}
}
