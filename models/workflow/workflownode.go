/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/models/common"
	"time"
)

// WorkFlowNode work flow node information
type WorkFlowNode struct {
	common.Entity
	BizID       string `gorm:"default:null;<-:create"`
	ParentID    string `gorm:"default:null;index;comment:'ID of the workflow parent node'"`
	Name        string `gorm:"default:null;comment:'name of the workflow node'"`
	OperationID string `gorm:"default:null;comment:'ID of the operation'"`
	ReturnType  string `gorm:"default:null"`
	Parameters  string `gorm:"default:null"`
	Result      string `gorm:"default:null"`
	StartTime   time.Time
	EndTime     time.Time
}

func (node *WorkFlowNode) Processing() {
	node.Status = constants.WorkFlowStatusProcessing
}

var defaultSuccessInfo = "Completed."

func (node *WorkFlowNode) Record(result ...interface{}) {
	if result == nil {
		result = []interface{}{defaultSuccessInfo}
	}

	for _, r := range result {
		if r == nil {
			r = defaultSuccessInfo
		}
		node.Result = fmt.Sprintf("%s%s", node.Result, fmt.Sprintln(r))
	}
}

func (node *WorkFlowNode) RecordAndPersist(result ...interface{}) {
	// todo: persist record
	node.Record(result)
}

func (node *WorkFlowNode) Success(result ...interface{}) {
	node.Record(result...)

	node.Status = constants.WorkFlowStatusFinished
	node.EndTime = time.Now()
}

func (node *WorkFlowNode) Fail(e error) {
	node.Status = constants.WorkFlowStatusError
	node.EndTime = time.Now()
	node.Record(e.Error())
}
