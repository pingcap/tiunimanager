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

package secondparty

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	db "github.com/pingcap-inc/tiem/test/mockdb"
)

var secondMicro1 *SecondMicro

func init() {
	secondMicro1 = &SecondMicro{
		TiupBinPath: "mock_tiup",
	}
	microInitForTestLibtiup("")
}

func TestSecondMicro_MicroSrvTiupDeploy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupDeploy(ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDeploy_Success(t *testing.T) {

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

	taskID, err := secondMicro1.MicroSrvTiupDeploy(ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupList(t *testing.T) {
	_, err := secondMicro1.MicroSrvTiupList(ClusterComponentTypeStr, 0, []string{})
	if err == nil {
		t.Errorf("case: create tiup task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupList_WithTimeOut(t *testing.T) {
	_, err := secondMicro1.MicroSrvTiupList(ClusterComponentTypeStr, 1, []string{})
	if err == nil {
		t.Errorf("case: create tiup task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupStart_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupStart(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupStart_Success(t *testing.T) {

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

	taskID, err := secondMicro1.MicroSrvTiupStart(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupRestart_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupRestart(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupRestart_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupRestart(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupStop_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupStop(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupStop_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupStop(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupDestroy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupDestroy(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDestroy_Success(t *testing.T) {

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

	taskID, err := secondMicro1.MicroSrvTiupDestroy(ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupDumpling_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvDumpling(0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDumpling_Success(t *testing.T) {

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

	taskID, err := secondMicro1.MicroSrvDumpling(0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupLightning_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = 0

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvLightning(0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupLightning_Success(t *testing.T) {

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

	taskID, err := secondMicro1.MicroSrvLightning(0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupClusterDisplay(t *testing.T) {

	_, err := secondMicro1.MicroSrvTiupDisplay(ClusterComponentTypeStr, "test-tidb", 0, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupClusterDisplay_WithTimeout(t *testing.T) {

	_, err := secondMicro1.MicroSrvTiupDisplay(ClusterComponentTypeStr, "test-tidb", 1, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatus_Fail(t *testing.T) {
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

	_, errStr, err := secondMicro1.MicroSrvGetTaskStatus(1)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by id intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatus_Success(t *testing.T) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupTaskByIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.TiupTask = &dbPb.TiupTask{
		ID:     1,
		Status: dbPb.TiupTaskStatus_Init,
	}

	mockCtl := gomock.NewController(t)
	mockDBClient := db.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupTaskByID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, errStr, err := secondMicro1.MicroSrvGetTaskStatus(1)
	if stat != dbPb.TiupTaskStatus_Init || errStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", errStr, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatusByBizID_Fail(t *testing.T) {
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

	_, errStr, err := secondMicro1.MicroSrvGetTaskStatusByBizID(0)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by BizId intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatusByBizID_Success(t *testing.T) {
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

	stat, statErrStr, err := secondMicro1.MicroSrvGetTaskStatusByBizID(0)
	if stat != dbPb.TiupTaskStatus_Init || statErrStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), statErrStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", statErrStr, nil, err)
	}
}

func TestSecondMicro_startNewTiupTask_Wrong(t *testing.T) {
	secondMicro1.startNewTiupTask(1, "ls", []string{"-2"}, 1)
}

func TestSecondMicro_startNewTiupTask(t *testing.T) {
	secondMicro1.startNewTiupTask(1, "ls", []string{}, 1)
}

func microInitForTestLibtiup(mgrLogFilePath string) {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileLibTiup)

	secondMicro1.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondMicro1.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondMicro1.taskStatusMap = make(map[uint64]TaskStatusMapValue)
}
