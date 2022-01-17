/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: operation_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package operation

import (
	"context"
	"testing"
)

const (
	TestWorkFlowID = "testworkflowid"
)

func TestGormOperationReadWrite_Create_Fail(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), "deploy", "")

	if err == nil || operation1 != nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}
}

func TestGormOperationReadWrite_Create_Success(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), "deploy", TestWorkFlowID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation2, err := testRW.Get(context.TODO(), operation1.ID)
	if err != nil || operation2 == nil || operation1.ID != operation2.ID {
		t.Errorf("Create Get() error: %v, operation1: %v, operation2: %v", err, operation1, operation2)
	}
}

func TestGormOperationReadWrite_Update_Fail(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), "deploy", TestWorkFlowID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation1.ID = ""
	operation1.Status = Processing
	err = testRW.Update(context.TODO(), operation1)

	if err == nil {
		t.Errorf("Update() error: %v", err)
	}
}

func TestGormOperationReadWrite_Update_Success(t *testing.T) {
	operation1, err := testRW.Create(context.TODO(), "deploy", TestWorkFlowID)

	if err != nil || operation1 == nil {
		t.Errorf("Create() error: %v: operation1: %v", err, operation1)
	}

	operation1.Status = Processing
	err = testRW.Update(context.TODO(), operation1)

	if err != nil {
		t.Errorf("Update() error: %v", err)
	}

	operation2, err := testRW.Get(context.TODO(), operation1.ID)
	if err != nil || operation2 == nil || operation1.ID != operation2.ID || operation2.Status != Processing {
		t.Errorf("Update Get() error: %v, operation1: %v, operation2: %v", err, operation1, operation2)
	}
}

func TestGormOperationReadWrite_Get_Fail(t *testing.T) {
	operation1, err := testRW.Get(context.TODO(), "")
	if err == nil || operation1 != nil {
		t.Errorf("Get() error: %v: operation1: %v", err, operation1)
	}
}
