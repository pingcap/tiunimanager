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
 * @File: lib_br_v2_test
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

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/test/mockdb"
)

var secondPartyManager2 *SecondPartyManager
var TESTBIZID = "bizid"

func init() {
	secondPartyManager2 = &SecondPartyManager{}
	dbConnParam = DbConnParam{
		Username: "root",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
	storage = BrStorage{
		StorageType: StorageTypeLocal,
		Root:        "/tmp/backup",
	}
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
	}

	initForTestLibbr()
}

func TestSecondPartyManager_BackUp_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_BackUp_Success1_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		TableName:       "testTbl",
		DbName:          "testDb",
		RateLimitM:      "1",
		Concurrency:     "1",
		CheckSum:        "1",
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_BackUp_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName:          "testDb",
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_BackUp_Success3_DontCareAsyncResult(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ShowBackUpInfo_Fail(t *testing.T) {
	resp := secondPartyManager2.ShowBackUpInfo(context.TODO(), clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show backup info. either Destination(%s) or ErrorStr(%s) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func TestSecondPartyManager_Restore_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_Restore_Success1_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		TableName:       "testTbl",
		DbName:          "testDb",
		RateLimitM:      "1",
		Concurrency:     "1",
		CheckSum:        "1",
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_Restore_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName:          "testDb",
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_Restore_Success3_DontCareAsyncResult(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ShowRestoreInfo_Fail(t *testing.T) {
	resp := secondPartyManager2.ShowRestoreInfo(context.TODO(), clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show restore info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func initForTestLibbr() {
	secondPartyManager2.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondPartyManager2.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondPartyManager2.taskStatusMap = make(map[uint64]TaskStatusMapValue)
}
