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
 *                                                                            *
 ******************************************************************************/

package workflow

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
)

type WorkFlowDefine struct {
	FlowName      string
	TaskNodes     map[string]*NodeDefine
	ContextParser func(string) *FlowContext
}

type NodeExecutor func(task *workflow.WorkFlowNode, context *FlowContext) error

type NodeDefine struct {
	Name         string
	SuccessEvent string
	FailEvent    string
	ReturnType   NodeReturnType
	Executor     NodeExecutor
}

func (define *WorkFlowDefine) getInstance(ctx context.Context, bizId string, data map[string]interface{}) *WorkFlowAggregation {
	return &WorkFlowAggregation{
		Flow: &workflow.WorkFlow{
			Name:  define.FlowName,
			BizID: bizId,
			Entity: common.Entity{
				Status:   constants.WorkFlowStatusInitializing,
				TenantId: framework.GetTenantIDFromContext(ctx),
			},
		},
		Nodes:   make([]*workflow.WorkFlowNode, 0, 4),
		Context: FlowContext{ctx, data},
		Define:  define,
	}
}

func (define *WorkFlowDefine) getNodeNameList() []string {
	var nodeNames []string
	node := define.TaskNodes["start"]

	for node != nil && node.Name != "end" && node.Name != "fail" {
		nodeNames = append(nodeNames, node.Name)
		node = define.TaskNodes[node.SuccessEvent]
	}
	return nodeNames
}

func CompositeExecutor(executors ...NodeExecutor) NodeExecutor {
	return func(node *workflow.WorkFlowNode, context *FlowContext) error {
		for _, executor := range executors {
			if err := executor(node, context); err == nil {
				continue
			} else {
				return err
			}
		}
		return nil
	}
}
