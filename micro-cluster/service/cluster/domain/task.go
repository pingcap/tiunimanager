package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// FlowWorkEntity
type FlowWorkEntity struct {
	Id          uint
	FlowName    string
	StatusAlias string
	BizId       string
	Status      TaskStatus
	ContextContent string
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
		r,err := json.Marshal(result)
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
	FlowWork 	*FlowWorkEntity
	Define 		*FlowWorkDefine
	CurrentTask *TaskEntity
	Tasks    	[]*TaskEntity
	Context     FlowContext
}

func CreateFlowWork(bizId string, defineName string) (*FlowWorkAggregation, error){
	define := FlowWorkDefineMap[defineName]
	if define == nil {
		return nil, errors.New("workflow undefined")
	}
	context := make(map[string]interface{})

	flow := define.getInstance(bizId, context)
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

func (flow *FlowWorkAggregation) handle(taskDefine *TaskDefine) {
	if taskDefine == nil {
		flow.FlowWork.Status = TaskStatusFinished
		return
	}
	task := &TaskEntity{
		Status:   TaskStatusInit,
		TaskName: taskDefine.Name,
		TaskReturnType: taskDefine.ReturnType,
	}

	TaskRepo.AddFlowTask(task, flow.FlowWork.Id)
	flow.Tasks = append(flow.Tasks, task)
	task.Processing()

	handleSuccess := taskDefine.Executor(task, &flow.Context)

	if !handleSuccess {
		if "" == taskDefine.FailEvent {
			return
		}
		if e,ok := flow.Define.TaskNodes[taskDefine.FailEvent]; ok {
			flow.handle(e)
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
		ticker := time.NewTicker(5 * time.Second)
		bizId := task.Id
		for _ = range ticker.C {
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
	ID uint
	Name string
	BizId string
	CronTaskType CronTaskType
	Cron string
	Parameter string
	NextTime time.Time
	Status CronStatus
}

type CronTaskAggregation struct {
	CronTask CronTaskEntity
	Tasks    []TaskEntity
}

func GetDefaultMaintainTask() *CronTaskEntity {
	return &CronTaskEntity{
		Name: "maintain",
		Cron: "0 0 21 ? ? ? ",
		Status: CronStatusValid,
	}
}


