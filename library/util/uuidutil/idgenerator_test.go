
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

package uuidutil

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := GenerateID()
	if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}

	if len(got) != ENTITY_UUID_LENGTH {
		t.Errorf("GenerateID() want len = %d, got = %v", ENTITY_UUID_LENGTH, len(got))
	}

}

func TestGenerateIDReplace(t *testing.T) {
	time := 0
	for time < 100 {
		got := GenerateID()
		if strings.Contains(got, "/") {
			t.Errorf("GenerateID() got /")
		}
		if strings.Contains(got, "-") {
			break
		}
		time++
	}
	for time < 200 {
		got := GenerateID()
		if strings.Contains(got, "/") {
			t.Errorf("GenerateID() got /")
		}
		if strings.Contains(got, "-") {
			break
		}
		time++
	}

}
