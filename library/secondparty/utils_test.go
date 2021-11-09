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
	"testing"
)

func Test_assert_false(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	assert(false)
}

func Test_assert_true(t *testing.T) {
	assert(true)
}

func Test_myPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	myPanic("panic info")
}

func Test_newTmpFileWithContent(t *testing.T) {
	content := []byte("temp info")
	_, err := newTmpFileWithContent(content)
	if err != nil {
		t.Errorf(err.Error())
	}
}
