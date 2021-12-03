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
	common2 "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
)

type FlowWorkDefine struct {
	FlowName      string
	TaskNodes     map[string]*TaskDefine
	ContextParser func(string) *FlowContext
}

type TaskExecutor func(task *workflow.WorkFlowNode, context *FlowContext) bool

type TaskDefine struct {
	Name         string
	SuccessEvent string
	FailEvent    string
	ReturnType   workflow.TaskReturnType
	Executor     TaskExecutor
}

func DefaultContextParser(s string) *FlowContext {
	// todo parse context
	return NewFlowContext(context.TODO())
}

func (define *FlowWorkDefine) getInstance(ctx context.Context, bizId string, data map[string]interface{}) *FlowWorkAggregation {
	return &FlowWorkAggregation{
		FlowWork: &workflow.WorkFlow{
			Name:  define.FlowName,
			BizID: bizId,
			Entities: common2.Entities{
				Status: string(workflow.TaskStatusInit),
			},
		},
		Tasks:   make([]*workflow.WorkFlowNode, 0, 4),
		Context: FlowContext{ctx, data},
		Define:  define,
	}
}

func CompositeExecutor(executors ...TaskExecutor) TaskExecutor {
	return func(task *workflow.WorkFlowNode, context *FlowContext) bool {
		for _, executor := range executors {
			if executor(task, context) {
				continue
			} else {
				return false
			}
		}
		return true
	}
}

func defaultEnd(task *workflow.WorkFlowNode, context *FlowContext) bool {
	task.Status = string(workflow.TaskStatusFinished)
	return true
}

func defaultFail(task *workflow.WorkFlowNode, context *FlowContext) bool {
	task.Status = string(workflow.TaskStatusError)
	return true
}
