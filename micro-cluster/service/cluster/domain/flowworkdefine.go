package domain

import (
	copywriting2 "github.com/pingcap/tiem/library/copywriting"
)

var FlowWorkDefineMap = map[string]*FlowWorkDefine{
	FlowCreateCluster: {
		FlowName:    FlowCreateCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowCreateCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":        {"prepareResource", "resourceDone", "fail", SyncFuncTask, prepareResource},
			//"resourceDone": {"buildConfig", "configDone", "fail", SyncFuncTask, buildConfig},
			//"configDone":   {"deployCluster", "deployDone", "fail", PollingTasK, deployCluster},
			//"deployDone":   {"startupCluster", "startupDone", "fail", PollingTasK, startupCluster},
			"resourceDone":  {"end", "", "", SyncFuncTask, DefaultEnd},
			"fail":         {"fail", "", "", SyncFuncTask, DefaultFail},
		},
		ContextParser: func(s string) *FlowContext {
			// todo parse context
			c := make(FlowContext)
			return &c
		},
	},
	FlowDeleteCluster: {
		FlowName:    FlowDeleteCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowDeleteCluster),
		TaskNodes: map[string]*TaskDefine {
			"start":              {"destroyTasks", "destroyTasksDone", "fail", PollingTasK, destroyTasks},
			"destroyTasksDone":   {"destroyCluster", "destroyClusterDone", "fail", PollingTasK, destroyCluster},
			"destroyClusterDone": {"deleteCluster", "deleteClusterDone", "fail", SyncFuncTask, deleteCluster},
			"deleteClusterDone":  {"freedResource", "freedResourceDone", "fail", SyncFuncTask, freedResource},
			"freedResourceDone":  {"end", "", "", SyncFuncTask, DefaultEnd},
			"fail":               {"fail", "", "", SyncFuncTask, DefaultFail},
		},
		ContextParser: func(s string) *FlowContext {
			// todo parse context
			c := make(FlowContext)
			return &c
		},
	},
	FlowBackupCluster: {
		FlowName:    FlowBackupCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowBackupCluster),
		TaskNodes: map[string]*TaskDefine {
			"start":              {"backup", "backupDone", "fail", PollingTasK, backupCluster},
			"backupDone":  {"end", "", "", SyncFuncTask, DefaultEnd},
			"fail":               {"fail", "", "", SyncFuncTask, DefaultFail},
		},
		ContextParser: func(s string) *FlowContext {
			// todo parse context
			c := make(FlowContext)
			return &c
		},
	},
	FlowRecoverCluster: {
		FlowName:    FlowRecoverCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowRecoverCluster),
		TaskNodes: map[string]*TaskDefine {
			"start":              {"recover", "recoverDone", "fail", PollingTasK, recoverCluster},
			"recoverDone":  {"end", "", "", SyncFuncTask, DefaultEnd},
			"fail":               {"fail", "", "", SyncFuncTask, DefaultFail},
		},
		ContextParser: func(s string) *FlowContext {
			// todo parse context
			c := make(FlowContext)
			return &c
		},
	},
	FlowModifyParameters: {
		FlowName:    FlowModifyParameters,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowModifyParameters),
		TaskNodes: map[string]*TaskDefine {
			"start":              {"modifyParameter", "modifyDone", "fail", PollingTasK, modifyParameters},
			"modifyDone":  {"end", "", "", SyncFuncTask, DefaultEnd},
			"fail":               {"fail", "", "", SyncFuncTask, DefaultFail},
		},
		ContextParser: func(s string) *FlowContext {
			// todo parse context
			c := make(FlowContext)
			return &c
		},
	},
}

type FlowWorkDefine struct {
	FlowName      string
	StatusAlias   string
	TaskNodes     map[string] *TaskDefine
	ContextParser func(string) *FlowContext
}

func (define *FlowWorkDefine) getInstance(bizId string, context map[string]interface{}) *FlowWorkAggregation{

	if context == nil {
		context = make(map[string]interface{})
	}

	return &FlowWorkAggregation{
		FlowWork: &FlowWorkEntity{
			FlowName: define.FlowName,
			StatusAlias: define.StatusAlias,
			BizId: bizId,
			Status: TaskStatusInit,
		},
		Tasks: make([]*TaskEntity, 0 ,4),
		Context: context,
		Define: define,
	}
}

type TaskDefine struct {
	Name 			string
	SuccessEvent 	string
	FailEvent 		string
	ReturnType 		TaskReturnType
	Executor 		func(task *TaskEntity, context *FlowContext) bool
}

func DefaultEnd(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusFinished
	return true
}

func DefaultFail(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusError
	return true
}

