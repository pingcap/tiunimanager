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
 * @File: second_party_operation_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/13
*******************************************************************************/

package secondparty

import (
	"context"
	"testing"
)

var operationIDs []string

func TestGormSecondPartyOperationReadWrite_Create_Fail(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), OperationType_ClusterDeploy, "")

	if err == nil || operation1 != nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}
}

func TestGormSecondPartyOperationReadWrite_Create_Success(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), OperationType_ClusterDeploy, TestWorkFlowNodeID)
	operationIDs = append(operationIDs, operation1.ID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation2, err := testRW.Get(context.TODO(), operation1.ID)
	if err != nil || operation2 == nil || operation1.ID != operation2.ID {
		t.Errorf("Create Get() error: %v, operation1: %v, operation2: %v", err, operation1, operation2)
	}
}

func TestGormSecondPartyOperationReadWrite_Update_Fail(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), OperationType_ClusterDeploy, TestWorkFlowNodeID)
	operationIDs = append(operationIDs, operation1.ID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation1.ID = ""
	operation1.Status = OperationStatus_Processing
	err = testRW.Update(context.TODO(), operation1)

	if err == nil {
		t.Errorf("Update() error: %v", err)
	}
}

func TestGormSecondPartyOperationReadWrite_Update_Success(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), OperationType_ClusterDeploy, TestWorkFlowNodeID)
	operationIDs = append(operationIDs, operation1.ID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation1.Status = OperationStatus_Processing
	err = testRW.Update(context.TODO(), operation1)

	if err != nil {
		t.Errorf("Update() error: %v", err)
	}

	operation2, err := testRW.Get(context.TODO(), operation1.ID)
	if err != nil || operation2 == nil || operation1.ID != operation2.ID || operation2.Status != OperationStatus_Processing {
		t.Errorf("Update Get() error: %v, operation1: %v, operation2: %v", err, operation1, operation2)
	}
}

func TestGormSecondPartyOperationReadWrite_Get_Fail(t *testing.T) {
	operation1, err := testRW.Get(context.TODO(), "")
	if err == nil || operation1 != nil {
		t.Errorf("Get() error: %v: operation1: %v", err, operation1)
	}
}

func TestGormSecondPartyOperationReadWrite_QueryByWorkFlowNodeID_Fail1(t *testing.T) {
	operation1, err := testRW.QueryByWorkFlowNodeID(context.TODO(), "")
	if err == nil || operation1 != nil {
		t.Errorf("QueryByWorkFlowNodeID() error: %v: operation1: %v", err, operation1)
	}
}

func TestGormSecondPartyOperationReadWrite_QueryByWorkFlowNodeID_Fail2(t *testing.T) {
	operation1, err := testRW.QueryByWorkFlowNodeID(context.TODO(), TestNoSuchWorkFlowNodeID)

	if err == nil || operation1 != nil {
		t.Errorf("QueryByWorkFlowNodeID() error: %v: operation1: %v", err, operation1)
	}
}

func TestGormSecondPartyOperationReadWrite_QueryByWorkFlowNodeID_Success1(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), OperationType_ClusterDeploy, TestWorkFlowNodeID)
	operationIDs = append(operationIDs, operation1.ID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	for _, id := range operationIDs {
		operation, err := testRW.Get(context.TODO(), id)
		if err != nil {
			t.Error(err)
		}
		operation.Status = OperationStatus_Error
		err = testRW.Update(context.TODO(), operation)
		if err != nil {
			t.Error(err)
		}
	}
	operation2, err := testRW.QueryByWorkFlowNodeID(context.TODO(), TestWorkFlowNodeID)
	if err != nil || operation2.Status != OperationStatus_Error {
		t.Errorf("err: %v, operation: %v", err, operation2)
	}

	for _, id := range operationIDs {
		operation, err := testRW.Get(context.TODO(), id)
		if err != nil {
			t.Error(err)
		}
		operation.Status = OperationStatus_Init
		err = testRW.Update(context.TODO(), operation)
		if err != nil {
			t.Error(err)
		}
	}
	operation2, err = testRW.QueryByWorkFlowNodeID(context.TODO(), TestWorkFlowNodeID)
	if err != nil || operation2.Status != OperationStatus_Init {
		t.Errorf("err: %v, operation: %v", err, operation2)
	}

	for _, id := range operationIDs {
		operation, err := testRW.Get(context.TODO(), id)
		if err != nil {
			t.Error(err)
		}
		operation.Status = OperationStatus_Processing
		err = testRW.Update(context.TODO(), operation)
		if err != nil {
			t.Error(err)
		}
	}
	operation2, err = testRW.QueryByWorkFlowNodeID(context.TODO(), TestWorkFlowNodeID)
	if err != nil || operation2.Status != OperationStatus_Processing {
		t.Errorf("err: %v, operation: %v", err, operation2)
	}
}
