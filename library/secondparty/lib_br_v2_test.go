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

	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocksecondparty"
)

var secondPartyManager2 *SecondPartyManager

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
	models.MockDB()

	initForTestLibbr()
}

func TestSecondPartyManager_BackUp_Fail(t *testing.T) {

	expectedErr := errors.New("fail Create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeBackup, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
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

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeBackup, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_BackUp_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName:          "testDb",
	}

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeBackup, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_BackUp_Success3_DontCareAsyncResult(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeBackup, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.BackUp(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ShowBackUpInfo_Fail(t *testing.T) {
	resp := secondPartyManager2.ShowBackUpInfo(context.TODO(), clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show backup info. either Destination(%s) or ErrorStr(%s) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func TestSecondPartyManager_Restore_Fail(t *testing.T) {
	expectedErr := errors.New("fail Create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeRestore, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
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

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeRestore, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_Restore_Success2_DontCareAsyncResult(t *testing.T) {
	defer resetVariable()
	clusterFacade = ClusterFacade{
		DbConnParameter: dbConnParam,
		DbName:          "testDb",
	}

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeRestore, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_Restore_Success3_DontCareAsyncResult(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationTypeRestore, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	operationID, err := secondPartyManager2.Restore(context.TODO(), clusterFacade, storage, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ShowRestoreInfo_Fail(t *testing.T) {
	resp := secondPartyManager2.ShowRestoreInfo(context.TODO(), clusterFacade)
	if resp.Destination == "" && resp.ErrorStr == "" {
		t.Errorf("case: show restore info. either Destination(%s) or ErrorStr(%v) should have zero value", resp.Destination, resp.ErrorStr)
	}
}

func initForTestLibbr() {
	secondPartyManager2.syncedOperationStatusMap = make(map[string]OperationStatusMapValue)
	secondPartyManager2.operationStatusCh = make(chan OperationStatusMember, 1024)
	secondPartyManager2.operationStatusMap = make(map[string]OperationStatusMapValue)
}
