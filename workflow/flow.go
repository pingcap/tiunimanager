/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"time"
)

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
	FlowError   error
}

func createFlowWork(ctx context.Context, bizId string, define *FlowWorkDefine) (*FlowWorkAggregation, error) {
	framework.LogWithContext(ctx).Infof("create flowwork %v for bizId %s", define, bizId)
	if define == nil {
		return nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(framework.ForkMicroCtx(ctx), bizId, flowData)
	TaskRepo.AddFlowWork(ctx, flow.FlowWork)
	return flow, nil
}

func (flow *FlowWorkAggregation) start() {
	flow.FlowWork.Status = TaskStatusProcessing
	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.complete(result)
	TaskRepo.Persist(flow.Context, flow)
}

func (flow *FlowWorkAggregation) asyncStart() {
	go flow.start()
}

func (flow *FlowWorkAggregation) destroy(reason string) {
	flow.FlowWork.Status = TaskStatusCanceled

	if flow.CurrentTask != nil {
		flow.CurrentTask.Fail(framework.NewTiEMError(common.TIEM_TASK_CANCELED, reason))
	}

	TaskRepo.Persist(flow.Context, flow)
}

func (flow FlowWorkAggregation) complete(success bool) {
	if success {
		flow.FlowWork.Status = TaskStatusFinished
	} else {
		flow.FlowWork.Status = TaskStatusError
	}
}

func (flow *FlowWorkAggregation) addContext(key string, value interface{}) {
	flow.Context.setData(key, value)
}

func (flow *FlowWorkAggregation) executeTask(task *TaskEntity, taskDefine *TaskDefine) bool {
	flow.CurrentTask = task
	flow.Tasks = append(flow.Tasks, task)
	task.Processing()
	TaskRepo.Persist(flow.Context, flow)

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
		StartTime:      time.Now().Unix(),
	}

	TaskRepo.AddFlowTask(flow.Context, task, flow.FlowWork.Id)
	handleSuccess := flow.executeTask(task, taskDefine)

	if !handleSuccess {
		flow.handleTaskError(task, taskDefine)
		return false
	}

	switch taskDefine.ReturnType {
	case SyncFuncTask:
		return flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
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
				task.Success(statString)
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
