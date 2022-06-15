/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: status_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/20
*******************************************************************************/

package deployment

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		op := Operation{
			Type:       CMDDeploy,
			Operation:  TestOperation,
			WorkFlowID: TestWorkFlowID,
			Status:     Init,
			Result:     TestResult,
			ErrorStr:   TestErrorStr,
		}
		fileName, err := Create(testTiUPHome, op)
		assert.NoError(t, err)
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%s/storage/operation-.+\\.json", testTiUPHome)), fileName)
	})
	t.Run("Update", func(t *testing.T) {
		op := Operation{
			Type:       CMDDeploy,
			Operation:  TestOperation,
			WorkFlowID: TestWorkFlowID,
			Status:     Init,
			Result:     TestResult,
			ErrorStr:   TestErrorStr,
		}
		fileName, err := Create(testTiUPHome, op)
		assert.NoError(t, err)
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%s/storage/operation-.+\\.json", testTiUPHome)), fileName)

		op.Status = Processing
		err = Update(fileName, op)
		assert.NoError(t, err)

		op2, err := Read(fileName)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(op, op2))
	})
	t.Run("Delete", func(t *testing.T) {
		op := Operation{
			Type:       CMDDeploy,
			Operation:  TestOperation,
			WorkFlowID: TestWorkFlowID,
			Status:     Init,
			Result:     TestResult,
			ErrorStr:   TestErrorStr,
		}
		fileName, err := Create(testTiUPHome, op)
		assert.NoError(t, err)
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%s/storage/operation-.+\\.json", testTiUPHome)), fileName)

		_, err = Read(fileName)
		assert.NoError(t, err)
		err = Delete(fileName)
		assert.NoError(t, err)
		_, err = Read(fileName)
		assert.Error(t, err)
	})
}
