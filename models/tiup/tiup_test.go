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
 * @File: tiup_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/30
*******************************************************************************/

package tiup

import (
	"context"
	"testing"
)

func TestGormMirrorReadWrite(t *testing.T) {
	// create fail
	_, err := testRW.Create(context.TODO(), "", "")
	if err == nil {
		t.Error("Deliberately create error got nil")
	}

	// create success
	config1, err := testRW.Create(context.TODO(), TestComponentType, TestTiUPHome)
	if err != nil || config1 == nil {
		t.Errorf("Create error: %v: config: %v", err, config1)
	}

	// get fail
	config2, err := testRW.Get(context.TODO(), "")
	if err == nil {
		t.Error("Deliberately Get error got nil")
	}

	// get success
	config2, err = testRW.Get(context.TODO(), config1.ID)
	if err != nil || config2 == nil || config1.ID != config2.ID {
		t.Errorf("Get error: %v, config1: %v, config2: %v", err, config1, config2)
	}

	// query fail
	config2, err = testRW.QueryByComponentType(context.TODO(), "")
	if err == nil {
		t.Error("Deliberately Query error got nil")
	}

	// query success
	config2, err = testRW.QueryByComponentType(context.TODO(), config1.ComponentType)
	if err != nil || config2 == nil || config1.ID != config2.ID {
		t.Errorf("Query error: %v, config1: %v, config2: %v", err, config1, config2)
	}

	// update fail
	config2.ID = ""
	config2.TiupHome = TestTiUPHome2
	err = testRW.Update(context.TODO(), config2)
	if err == nil {
		t.Error("Deliberately Update error got nil")
	}

	// update success
	config2.ID = config1.ID
	config2.TiupHome = TestTiUPHome2
	err = testRW.Update(context.TODO(), config2)
	if err != nil {
		t.Errorf("Update error: %v", err)
	}
}
