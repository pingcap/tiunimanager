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

package workflow

import (
	"context"
)

type ReaderWriter interface {
	// CreateWorkFlow
	// @Description: create new workflow record
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Return *WorkFlow
	// @Return error
	CreateWorkFlow(ctx context.Context, flow *WorkFlow) (*WorkFlow, error)

	// UpdateWorkFlow
	// @Description: update workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Parameter status
	// @Parameter flowContext
	// @Return error
	UpdateWorkFlow(ctx context.Context, flowId string, status string, flowContext string) (err error)

	// GetWorkFlow
	// @Description: get workflow by flowID
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return *WorkFlow
	// @Return error
	GetWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, err error)

	// QueryWorkFlows
	// @Description: query workflows by condition with page
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter bizType
	// @Parameter fuzzyName
	// @Parameter status
	// @Parameter page
	// @Parameter pageSize
	// @Return []*WorkFlow
	// @Return error
	QueryWorkFlows(ctx context.Context, bizId, bizType, fuzzyName, status string, page int, pageSize int) (flows []*WorkFlow, total int64, err error)

	// CreateWorkFlowNode
	// @Description: create new workflow node
	// @Receiver m
	// @Parameter ctx
	// @Parameter node
	// @Return *WorkFlowNode
	// @Return error
	CreateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (*WorkFlowNode, error)

	// UpdateWorkFlowNode
	// @Description: update workflow node
	// @Receiver m
	// @Parameter ctx
	// @Parameter node
	// @Return error
	UpdateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (err error)

	// GetWorkFlowNode
	// @Description: update workflow node
	// @Receiver m
	// @Parameter ctx
	// @Parameter nodeId
	// @Return *WorkFlowNode
	// @Return error
	GetWorkFlowNode(ctx context.Context, nodeId string) (node *WorkFlowNode, err error)

	// UpdateWorkFlowDetail
	// @Description: update workflow detail nodes
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter nodes
	// @Return error
	UpdateWorkFlowDetail(ctx context.Context, flow *WorkFlow, nodes []*WorkFlowNode) (err error)

	// QueryDetailWorkFlow
	// @Description: detail workflow with nodes
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return *WorkFlow
	// @Return []*WorkFlowNode
	// @Return error
	QueryDetailWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, nodes []*WorkFlowNode, err error)
}
