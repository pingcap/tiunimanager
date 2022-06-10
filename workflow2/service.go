/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package workflow2

import (
	"context"
	"github.com/pingcap-inc/tiunimanager/common/structs"
	"github.com/pingcap-inc/tiunimanager/message"
)

// WorkFlowService workflow interface
type WorkFlowService interface {
	// RegisterWorkFlow
	// @Description: register workflow define
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Parameter flowDefine
	RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *WorkFlowDefine)

	// GetWorkFlowDefine
	// @Description: get workflow define by flowName
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Return *WorkFlowDefine
	// @Return error
	GetWorkFlowDefine(ctx context.Context, flowName string) (*WorkFlowDefine, error)

	// CreateWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter bizType
	// @Parameter flowName
	// @Return flowId
	// @Return error
	CreateWorkFlow(ctx context.Context, bizId string, bizType string, flowName string) (string, error)

	// ListWorkFlows
	// @Description: list workflows by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowsResp
	// @Return total
	// @Return error
	ListWorkFlows(ctx context.Context, request message.QueryWorkFlowsReq) (message.QueryWorkFlowsResp, structs.Page, error)

	// DetailWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowDetailResp
	// @Return error
	DetailWorkFlow(ctx context.Context, request message.QueryWorkFlowDetailReq) (message.QueryWorkFlowDetailResp, error)

	// InitContext
	// @Description: init flow context for workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Parameter key
	// @Parameter value
	// @Return error
	InitContext(ctx context.Context, flowId string, key string, value interface{}) error

	// Start
	// @Description: async start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return error
	Start(ctx context.Context, flowId string) error

	// Stop
	// @Description: stop workflow in current node
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return error
	Stop(ctx context.Context, flowId string) error

	// Cancel
	// @Description: cancel workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Parameter reason
	// @Return error
	Cancel(ctx context.Context, flowId string, reason string) error
}
