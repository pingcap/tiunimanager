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
)

type FlowWorkDefine struct {
	FlowName      string
	TaskNodes     map[string]*TaskDefine
	ContextParser func(string) *FlowContext
}

type TaskExecutor func(task *TaskEntity, context *FlowContext) bool

type TaskDefine struct {
	Name         string
	SuccessEvent string
	FailEvent    string
	ReturnType   TaskReturnType
	Executor     TaskExecutor
}

func DefaultContextParser(s string) *FlowContext {
	// todo parse context
	return NewFlowContext(context.TODO())
}

func (define *FlowWorkDefine) getInstance(ctx context.Context, bizId string, data map[string]interface{}) *FlowWorkAggregation {
	if data == nil {
		data = make(map[string]interface{})
	}

	return &FlowWorkAggregation{
		FlowWork: &FlowWorkEntity{
			FlowName: define.FlowName,
			BizId:    bizId,
			Status:   TaskStatusInit,
		},
		Tasks:   make([]*TaskEntity, 0, 4),
		Context: FlowContext{ctx, data},
		Define:  define,
	}
}

func CompositeExecutor(executors ...TaskExecutor) TaskExecutor {
	return func(task *TaskEntity, context *FlowContext) bool {
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

func defaultEnd(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusFinished
	return true
}

func defaultFail(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusError
	return true
}
