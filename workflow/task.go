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
	"fmt"
	"time"
)

//
// FlowWorkEntity
// @Description: flowwork entity
type FlowWorkEntity struct {
	Id             string
	FlowName       string
	BizId          string
	Status         TaskStatus
	ContextContent string
	CreateTime     time.Time
	UpdateTime     time.Time
	//todo to model
}

type FlowContext struct {
	context.Context
	FlowData map[string]interface{}
}

func NewFlowContext(ctx context.Context) *FlowContext {
	return &FlowContext{
		ctx,
		map[string]interface{}{},
	}
}

func (c FlowContext) getData(key string) interface{} {
	return c.FlowData[key]
}

func (c FlowContext) setData(key string, value interface{}) {
	c.FlowData[key] = value
}

func (c FlowWorkEntity) finished() bool {
	return c.Status.Finished()
}

//
// TaskEntity
// @Description: task entity
//
type TaskEntity struct {
	Id             uint
	Status         TaskStatus
	TaskName       string
	TaskReturnType TaskReturnType
	BizId          string
	Parameters     string
	Result         string
	StartTime      int64
	EndTime        int64
	TaskError      error
}

func (t *TaskEntity) Processing() {
	t.Status = TaskStatusProcessing
}

var defaultSuccessInfo = "success"

func (t *TaskEntity) Success(result ...interface{}) {
	t.Status = TaskStatusFinished
	t.EndTime = time.Now().Unix()

	if result == nil {
		result = []interface{}{defaultSuccessInfo}
	}

	for _, r := range result {
		if r == nil {
			r = defaultSuccessInfo
		}
		t.Result = fmt.Sprintln(t.Result, r)
	}
}

func (t *TaskEntity) Fail(e error) {
	t.Status = TaskStatusError
	t.TaskError = e
	t.EndTime = time.Now().Unix()
	t.Result = e.Error()
}
