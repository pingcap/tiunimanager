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
 ******************************************************************************/

/*******************************************************************************
 * @File: workflow.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type QueryWorkFlowDetailReq struct {
	WorkFlowID string `json:"workFlowId" swaggerignore:"true"`
}

type QueryWorkFlowDetailResp struct {
	Info      *structs.WorkFlowInfo       `json:"info"`
	NodeInfo  []*structs.WorkFlowNodeInfo `json:"nodes"`
	NodeNames []string                    `json:"nodeNames"`
}

type QueryWorkFlowsReq struct {
	structs.PageRequest
	Status   string `json:"status" form:"status"`
	FlowName string `json:"flowName" form:"flowName"`
	BizID    string `json:"bizId" form:"bizId"`
	BizType  string `json:"bizType" form:"bizType"`
}

type QueryWorkFlowsResp struct {
	WorkFlows []*structs.WorkFlowInfo `json:"workFlows" form:"workFlows"`
}
