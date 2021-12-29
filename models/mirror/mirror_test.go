/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 * @File: mirror_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/29
*******************************************************************************/

package mirror

import (
	"context"
	"testing"
)

func TestGormMirrorReadWrite_Create_Fail(t *testing.T) {

}

func TestGormMirrorReadWrite(t *testing.T) {
	// create fail
	_, err := testRW.Create(context.TODO(), "", "")
	if err == nil {
		t.Error("Deliberately create error got nil")
	}

	// create success
	mirror1, err := testRW.Create(context.TODO(), TestComponentType, TestMirrorAddr)
	if err != nil || mirror1 == nil {
		t.Errorf("Create error: %v: mirror: %v", err, mirror1)
	}

	// get fail
	mirror2, err := testRW.Get(context.TODO(), "")
	if err == nil {
		t.Error("Deliberately Get error got nil")
	}

	// get success
	mirror2, err = testRW.Get(context.TODO(), mirror1.ID)
	if err != nil || mirror2 == nil || mirror1.ID != mirror2.ID {
		t.Errorf("Get error: %v, mirror1: %v, mirror2: %v", err, mirror1, mirror2)
	}

	// query fail
	mirror2, err = testRW.QueryByComponentType(context.TODO(), "")
	if err == nil {
		t.Error("Deliberately Query error got nil")
	}

	// query success
	mirror2, err = testRW.QueryByComponentType(context.TODO(), mirror1.ComponentType)
	if err != nil || mirror2 == nil || mirror1.ID != mirror2.ID {
		t.Errorf("Query error: %v, mirror1: %v, mirror2: %v", err, mirror1, mirror2)
	}

	// update fail
	mirror2.ID = ""
	mirror2.MirrorAddr = TestMirrorAddr2
	err = testRW.Update(context.TODO(), mirror2)
	if err == nil {
		t.Error("Deliberately Update error got nil")
	}

	// update success
	mirror2.ID = mirror1.ID
	mirror2.MirrorAddr = TestMirrorAddr2
	err = testRW.Update(context.TODO(), mirror2)
	if err != nil {
		t.Errorf("Update error: %v", err)
	}
}
