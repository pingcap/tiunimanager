
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

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// FlowWorkEntity
type FlowWorkEntity struct {
	Id             uint
	FlowName       string
	StatusAlias    string
	BizId          string
	Status         TaskStatus
	ContextContent string
	Operator	*Operator
	CreateTime time.Time
	UpdateTime time.Time
}

type FlowContext map[string]interface{}

func (c FlowContext) value(key string) interface{} {
	return c[key]
}

func (c FlowContext) put(key string, value interface{}) {
	c[key] = value
}

func (c FlowWorkEntity) Finished() bool {
	return c.Status.Finished()
}

// TaskEntity
type TaskEntity struct {
	Id             uint
	Status         TaskStatus
	TaskName       string
	TaskReturnType TaskReturnType
	BizId          string
	Parameters     string
	Result         string
}

func (t *TaskEntity) Processing() {
	t.Status = TaskStatusProcessing
}

func (t *TaskEntity) Success(result interface{}) {
	t.Status = TaskStatusFinished
	if result != nil {
		r, err := json.Marshal(result)
		if err != nil {
			getLogger().Error(err)
		} else {
			t.Result = string(r)
		}
	}
}

func (t *TaskEntity) Fail(e error) {
	t.Status = TaskStatusError
	t.Result = e.Error()
}

// FlowWorkAggregation
type FlowWorkAggregation struct {
	FlowWork    *FlowWorkEntity
	Define      *FlowWorkDefine
	CurrentTask *TaskEntity
	Tasks       []*TaskEntity
	Context     FlowContext
}

func CreateFlowWork(bizId string, defineName string, operator *Operator) (*FlowWorkAggregation, error) {
	define := FlowWorkDefineMap[defineName]
	if define == nil {
		return nil, errors.New("workflow undefined")
	}
	context := make(map[string]interface{})

	flow := define.getInstance(bizId, context, operator)
	TaskRepo.AddFlowWork(flow.FlowWork)
	return flow, nil
}

func (flow *FlowWorkAggregation) Start() {
	flow.FlowWork.Status = TaskStatusProcessing
	start := flow.Define.TaskNodes["start"]
	flow.handle(start)
	TaskRepo.Persist(flow)
}

func (flow *FlowWorkAggregation) Destroy() {
	flow.FlowWork.Status = TaskStatusError
	flow.CurrentTask.Fail(errors.New("workflow destroy"))
	TaskRepo.Persist(flow)
}

func (flow *FlowWorkAggregation) AddContext(key string, value interface{}) {
	flow.Context.put(key, value)
}

func (flow *FlowWorkAggregation) executeTask(task *TaskEntity, taskDefine *TaskDefine) bool {
	flow.CurrentTask = task
	flow.Tasks = append(flow.Tasks, task)
	task.Processing()
	return taskDefine.Executor(task, &flow.Context)
}

func (flow *FlowWorkAggregation) handle(taskDefine *TaskDefine) {
	if taskDefine == nil {
		flow.FlowWork.Status = TaskStatusFinished
		return
	}
	task := &TaskEntity{
		Status:         TaskStatusInit,
		TaskName:       taskDefine.Name,
		TaskReturnType: taskDefine.ReturnType,
	}

	TaskRepo.AddFlowTask(task, flow.FlowWork.Id)
	handleSuccess := flow.executeTask(task, taskDefine)

	if !handleSuccess {
		if "" == taskDefine.FailEvent {
			return
		}
		if e, ok := flow.Define.TaskNodes[taskDefine.FailEvent]; ok {
			flow.handle(e)
			return
		}
		panic("workflow define error")
	}

	switch taskDefine.ReturnType {

	case UserTask:

	case SyncFuncTask:
		flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
	case CallbackTask:

	case PollingTasK:
		// receive the taskId and start ticker
		ticker := time.NewTicker(1 * time.Second)
		bizId := task.Id
		for range ticker.C {
			// todo check bizId
			//status, s, err := Operator.CheckProgress(uint64(f.CurrentTask.id))
			//			if err != nil {
			//				getLogger().Error(err)
			//				continue
			//			}
			//
			//			switch status {
			//			case dbPb.TiupTaskStatus_Init:
			//			getLogger().Info(s)
			fmt.Println(bizId)
			flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
			break
		}
	}
}

type CronTaskEntity struct {
	ID           uint
	Name         string
	BizId        string
	CronTaskType CronTaskType
	Cron         string
	Parameter    string
	NextTime     time.Time
	Status       CronStatus
}

type CronTaskAggregation struct {
	CronTask CronTaskEntity
	Tasks    []TaskEntity
}

func GetDefaultMaintainTask() *CronTaskEntity {
	return &CronTaskEntity{
		Name:   "maintain",
		Cron:   "0 0 21 ? ? ? ",
		Status: CronStatusValid,
	}
}
