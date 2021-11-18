
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
	"context"
	"encoding/json"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"time"
)

//
// FlowWorkEntity
// @Description: flowwork entity
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

func (c FlowContext) GetData(key string) interface{} {
	return c.FlowData[key]
}

func (c FlowContext) SetData(key string, value interface{}) {
	c.FlowData[key] = value
}

func (c FlowWorkEntity) Finished() bool {
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

func (t *TaskEntity) Success(result interface{}) {
	t.Status = TaskStatusFinished
	if result != nil {
		r, err := json.Marshal(result)
		if err != nil {
			getLogger().Error(err)
		} else {
			t.Result = string(r)
		}
	} else {
		t.Result = "success"
	}
	t.EndTime = time.Now().Unix()
}

func (t *TaskEntity) Fail(e error) {
	t.Status = TaskStatusError
	t.TaskError = e
	t.EndTime = time.Now().Unix()
	t.Result = e.Error()
}

//
// FlowWorkAggregation
// @Description: flowwork aggregation with flowwork definition and tasks
//
type FlowWorkAggregation struct {
	FlowWork    *FlowWorkEntity
	Define      *FlowWorkDefine
	CurrentTask *TaskEntity
	Tasks       []*TaskEntity
	Context     FlowContext
	FlowError 	error
}

func CreateFlowWork(ctx context.Context, bizId string, defineName string, operator *Operator) (*FlowWorkAggregation, error) {
	framework.LogWithContext(ctx).Infof("create flowwork %s for bizId %s", defineName, bizId)
	define := FlowWorkDefineMap[defineName]
	if define == nil {
		return nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(framework.ForkMicroCtx(ctx) , bizId, flowData, operator)
	TaskRepo.AddFlowWork(ctx, flow.FlowWork)
	return flow, nil
}

func (flow *FlowWorkAggregation) Start() {
	flow.FlowWork.Status = TaskStatusProcessing
	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.Complete(result)
	TaskRepo.Persist(flow.Context, flow)
}

func (flow *FlowWorkAggregation) AsyncStart() {
	go flow.Start()
}

func (flow *FlowWorkAggregation) Destroy(reason string) {
	flow.FlowWork.Status = TaskStatusCanceled
	flow.CurrentTask.Fail(framework.NewTiEMError(common.TIEM_TASK_CANCELED, reason))

	TaskRepo.Persist(flow.Context, flow)
}

func (flow FlowWorkAggregation) Complete(success bool) {
	if success {
		flow.FlowWork.Status = TaskStatusFinished
	} else {
		flow.FlowWork.Status = TaskStatusError
	}
}

func (flow *FlowWorkAggregation) AddContext(key string, value interface{}) {
	flow.Context.SetData(key, value)
}

func (flow *FlowWorkAggregation) executeTask(task *TaskEntity, taskDefine *TaskDefine) bool {
	flow.CurrentTask = task
	flow.Tasks = append(flow.Tasks, task)
	task.Processing()
	return taskDefine.Executor(task, &flow.Context)
}

func (flow *FlowWorkAggregation) handleTaskError(task *TaskEntity, taskDefine *TaskDefine) {
	flow.FlowError = task.TaskError
	if "" != taskDefine.FailEvent {
		flow.handle(flow.Define.TaskNodes[taskDefine.FailEvent])
	} else {
		framework.Log().Fatalf("no fail event in flow definition, flowname %s", taskDefine.Name)
	}
}

func (flow *FlowWorkAggregation) handle(taskDefine *TaskDefine) bool {
	if taskDefine == nil {
		flow.FlowWork.Status = TaskStatusFinished
		return true
	}
	task := &TaskEntity{
		Status:         TaskStatusInit,
		TaskName:       taskDefine.Name,
		TaskReturnType: taskDefine.ReturnType,
		StartTime: time.Now().Unix(),
	}

	TaskRepo.AddFlowTask(flow.Context, task, flow.FlowWork.Id)
	handleSuccess := flow.executeTask(task, taskDefine)

	if !handleSuccess {
		flow.handleTaskError(task, taskDefine)
		return false
	}

	switch taskDefine.ReturnType {

	case UserTask:
		return true
	case SyncFuncTask:
		return flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
	case CallbackTask:
		return true
	case PollingTasK:
		ticker := time.NewTicker(3 * time.Second)
		sequence := 0
		for range ticker.C {
			if sequence += 1; sequence > 200 {
				task.Fail(framework.SimpleError(common.TIEM_TASK_POLLING_TIME_OUT))
				flow.handleTaskError(task, taskDefine)
				return false
			}
			framework.LogWithContext(flow.Context).Infof("polling task waiting, sequence %d, taskId %d, taskName %s", sequence, task.Id, task.TaskName)

			stat, statString, err := secondparty.SecondParty.MicroSrvGetTaskStatusByBizID(flow.Context, uint64(task.Id))
			if err != nil {
				framework.LogWithContext(flow.Context).Error(err)
				task.Fail(framework.WrapError(common.TIEM_TASK_FAILED, common.TIEM_TASK_FAILED.Explain(), err))
				flow.handleTaskError(task, taskDefine)
				return false
			}
			if stat == dbpb.TiupTaskStatus_Error {
				task.Fail(framework.NewTiEMError(common.TIEM_TASK_FAILED, statString))
				flow.handleTaskError(task, taskDefine)
				return false
			}
			if stat == dbpb.TiupTaskStatus_Finished {
				task.Success(nil)
				return flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
			}
		}
	}
	return true
}

func (flow *FlowWorkAggregation) GetAllTaskDef() []string {
	var nodeNames []string
	node := flow.Define.TaskNodes["start"]

	for node != nil && node.Name != "end" && node.Name != "fail" {
		nodeNames = append(nodeNames, node.Name)
		node = flow.Define.TaskNodes[node.SuccessEvent]
	}
	return nodeNames
}

func (flow *FlowWorkAggregation) ExtractTaskDTO() []*clusterpb.TaskDTO {
	var tasks []*clusterpb.TaskDTO
	for _, task := range flow.Tasks {
		tasks = append(tasks, &clusterpb.TaskDTO{
			Id:         int64(task.Id),
			Status:     int32(task.Status),
			TaskName:   task.TaskName,
			Result:     task.Result,
			Parameters: task.Parameters,
			StartTime:  task.StartTime,
			EndTime:    task.EndTime,
		})
	}
	return tasks
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
