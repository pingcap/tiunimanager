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
 ******************************************************************************/

package workflow

import (
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorkFlow_Finished(t *testing.T) {
	f := &WorkFlow{
		Entity: common.Entity{
			Status: constants.WorkFlowStatusFinished,
		},
	}
	assert.True(t, f.Finished())
}

func TestResult(t *testing.T) {
	node := &WorkFlowNode{}
	node.Record("aaa", 2, true)
	assert.Equal(t, "aaa\n2\ntrue\n", node.Result)
	node.Record("ddd")
	assert.Equal(t, "aaa\n2\ntrue\nddd\n", node.Result)
}