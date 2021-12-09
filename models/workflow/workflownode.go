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

package workflow

import (
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models/common"
	"time"
)

// WorkFlowNode work flow node infomation
type WorkFlowNode struct {
	common.Entities
	BizID      string `gorm:"default:null;<-:create"`
	ParentID   string `gorm:"default:null;index;comment:'ID of the workflow parent node'"`
	Name       string `gorm:"default:null;comment:'name of the workflow node'"`
	ReturnType string `gorm:"default:null"`
	Parameters string `gorm:"default:null"`
	Result     string `gorm:"default:null"`
	StartTime  time.Time
	EndTime    time.Time
}

func (node *WorkFlowNode) Processing() {
	node.Status = constants.WorkFlowStatusProcessing
}

var defaultSuccessInfo = "success"

func (node *WorkFlowNode) Success(result ...interface{}) {
	node.Status = constants.WorkFlowStatusFinished
	node.EndTime = time.Now()

	if result == nil {
		result = []interface{}{defaultSuccessInfo}
	}

	for _, r := range result {
		if r == nil {
			r = defaultSuccessInfo
		}
		node.Result = fmt.Sprintln(node.Result, r)
	}
}

func (node *WorkFlowNode) Fail(e error) {
	node.Status = constants.WorkFlowStatusError
	node.EndTime = time.Now()
	node.Result = e.Error()
}
