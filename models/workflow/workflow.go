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
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/models/common"
)

// WorkFlow work flow infomation
type WorkFlow struct {
	common.Entity
	Name    string `gorm:"default:null;comment:'name of the workflow'"`
	BizID   string `gorm:"default:null;<-:create"`
	BizType string `gorm:"default:null"`
	Context string `gorm:"default:null"`
}

func (flow *WorkFlow) Finished() bool {
	return constants.WorkFlowStatusFinished == flow.Status || constants.WorkFlowStatusError == flow.Status || constants.WorkFlowStatusCanceled == flow.Status
}

func (flow *WorkFlow) Stopped() bool {
	return constants.WorkFlowStatusStopped == flow.Status
}
