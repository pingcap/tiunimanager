package clustermanage

import (
	"time"
)

// FlowWorkEntity 流程化的工作，用FlowWork记录其状态
type FlowWorkEntity struct {
	id          string
	flowName    string
	alias       string
	bizId       string
	status      int
	context     FlowContext
	currentTask TaskEntity
}

type FlowWorkAggregation struct {
	FlowWork 	FlowWorkEntity
	Tasks 		[]TaskEntity
}

type FlowContext map[string]interface{}

func (c FlowContext) Value(key string) interface{} {
	return c[key]
}

// Put 向流程添加上下文变量
// TODO 变量需要定义作用域，决定flowWork要不要存储它
func (c FlowContext) Put(key string, value interface{}) {
	c[key] = value
}

func ClusterInitFlowWork() FlowWorkEntity {
	return FlowWorkEntity{
		id: "flow1",
		flowName: "创建集群",
		bizId: "1212213",
		status: 0,
		context: make(map[string]interface{}),
	}
}

// 应该按生成不同的task来驱动
func (f *FlowWorkEntity) moveOn(event string) {
	if event == "start" {
		// 修改flow的task信息
		cluster := f.context.Value("cluster").(*Cluster)
		cluster.PrepareResource(f)
	}
	if event == "allocDone" {
		// 修改flow的task信息
		f.context.Value("cluster").(*Cluster).BuildConfig(f)
	}
	if event == "configDone" {
		// 修改flow的task信息
		f.context.Value("cluster").(*Cluster).ExecuteTiUP(f)
	}

	if event == "tiUPStart" {
		// 修改flow的task信息
		f.context.Value("cluster").(*Cluster).CheckTiUPResult(f)
	}

	if event == "tiUPDone" {
		// 修改flow的task信息
	}
}

// CronTaskEntity 定时任务，比如每周备份一次
type CronTaskEntity struct {
	name string
	cron string
	handlerName string
	parameter string

	nextTime time.Time
	// 0 启用，1 暂停，2 未启用，3 删除
	status int
	config string
}
type CronTaskAggregation struct {
	CronTask 	CronTaskEntity
	Tasks 		[]TaskEntity
}

// TaskEntity 异步的任务或工作。需要用户手动完成或系统回调的
type TaskEntity struct {
	id         		uint
	// doing success fail
	status 			int
	taskType    	TaskType
	taskCode		string
	taskExecutor 	string
	parameter 		string
	result 			string
}

var taskIdSeq uint = 100000000

func CreateTask() (t TaskEntity) {
	t.id = taskIdSeq
	taskIdSeq ++
	return
}

type TaskType int8

const (
	UserTaskType 		TaskType = 1 // 用户任务，如审批
	SyncCallFuncTask 	TaskType = 2 // 立刻同步执行的方法，如申请主机
	DelayedFuncTask 	TaskType = 3 // 延迟执行方法，如5分钟后检查审批状态
	WaitNotifyTask		TaskType = 4 // 等待通知的任务，外部根据bizId来通知成功或失败
)

