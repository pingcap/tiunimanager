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
	copywriting2 "github.com/pingcap-inc/tiem/library/copywriting"
)

func defaultContextParser(s string) *FlowContext {
	// todo parse context
	c := make(FlowContext)
	return &c
}

var FlowWorkDefineMap = map[string]*FlowWorkDefine{
	FlowCreateCluster: {
		FlowName:    FlowCreateCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowCreateCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":        {"prepareResource", "resourceDone", "fail", SyncFuncTask, prepareResource},
			"resourceDone": {"buildConfig", "configDone", "fail", SyncFuncTask, buildConfig},
			"configDone":   {"deployCluster", "deployDone", "fail", SyncFuncTask, deployCluster},
			"deployDone":   {"startupCluster", "startupDone", "fail", SyncFuncTask, startupCluster},
			"startupDone":  {"setClusterOnline", "onlineDone", "fail", SyncFuncTask, setClusterOnline},
			"onlineDone":   {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":         {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowDeleteCluster: {
		FlowName:    FlowDeleteCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowDeleteCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":              {"destroyTasks", "destroyTasksDone", "fail", PollingTasK, destroyTasks},
			"destroyTasksDone":   {"destroyCluster", "destroyClusterDone", "fail", PollingTasK, destroyCluster},
			"destroyClusterDone": {"deleteCluster", "deleteClusterDone", "fail", SyncFuncTask, deleteCluster},
			"deleteClusterDone":  {"freedResource", "freedResourceDone", "fail", SyncFuncTask, freedResource},
			"freedResourceDone":  {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":               {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowBackupCluster: {
		FlowName:    FlowBackupCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowBackupCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":            {"backup", "backupDone", "fail", PollingTasK, backupCluster},
			"backupDone":       {"updateBackupRecord", "updateRecordDone", "fail", SyncFuncTask, updateBackupRecord},
			"updateRecordDone": {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":             {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowRecoverCluster: {
		FlowName:    FlowRecoverCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowRecoverCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":        {"prepareResource", "resourceDone", "fail", SyncFuncTask, prepareResource},
			"resourceDone": {"buildConfig", "configDone", "fail", SyncFuncTask, buildConfig},
			"configDone":   {"deployCluster", "deployDone", "fail", SyncFuncTask, deployCluster},
			"deployDone":   {"startupCluster", "startupDone", "fail", SyncFuncTask, startupCluster},
			"startupDone":  {"recoverFromSrcCluster", "recoverDone", "fail", SyncFuncTask, recoverFromSrcCluster},
			"recoverDone":  {"setClusterOnline", "onlineDone", "fail", SyncFuncTask, setClusterOnline},
			"onlineDone":   {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":         {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowModifyParameters: {
		FlowName:    FlowModifyParameters,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowModifyParameters),
		TaskNodes: map[string]*TaskDefine{
			"start":      {"modifyParameter", "modifyDone", "fail", PollingTasK, modifyParameters},
			"modifyDone": {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":       {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowExportData: {
		FlowName:    FlowExportData,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowExportData),
		TaskNodes: map[string]*TaskDefine{
			"start":            {"exportDataFromCluster", "exportDataDone", "fail", PollingTasK, exportDataFromCluster},
			"exportDataDone":   {"updateDataExportRecord", "updateRecordDone", "fail", SyncFuncTask, updateDataExportRecord},
			"updateRecordDone": {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":             {"fail", "", "", SyncFuncTask, exportDataFailed},
		},
		ContextParser: defaultContextParser,
	},
	FlowImportData: {
		FlowName:    FlowImportData,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowImportData),
		TaskNodes: map[string]*TaskDefine{
			"start":            {"buildDataImportConfig", "buildConfigDone", "fail", SyncFuncTask, buildDataImportConfig},
			"buildConfigDone":  {"importDataToCluster", "importDataDone", "fail", PollingTasK, importDataToCluster},
			"importDataDone":   {"updateDataImportRecord", "updateRecordDone", "fail", SyncFuncTask, updateDataImportRecord},
			"updateRecordDone": {"end", "", "", SyncFuncTask, ClusterEnd},
			"fail":             {"fail", "", "", SyncFuncTask, importDataFailed},
		},
		ContextParser: defaultContextParser,
	},
	FlowRestartCluster: {
		FlowName:    FlowRestartCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowRestartCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":       {"clusterRestart", "restartDone", "fail", SyncFuncTask, clusterRestart},
			"restartDone": {"end", "", "fail", SyncFuncTask, ClusterEnd},
			"fail":        {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
	FlowStopCluster: {
		FlowName:    FlowStopCluster,
		StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowStopCluster),
		TaskNodes: map[string]*TaskDefine{
			"start":    {"clusterStop", "stopDone", "fail", SyncFuncTask, clusterStop},
			"stopDone": {"end", "", "fail", SyncFuncTask, ClusterEnd},
			"fail":     {"fail", "", "", SyncFuncTask, ClusterFail},
		},
		ContextParser: defaultContextParser,
	},
}

type FlowWorkDefine struct {
	FlowName      string
	StatusAlias   string
	TaskNodes     map[string]*TaskDefine
	ContextParser func(string) *FlowContext
}

func (define *FlowWorkDefine) getInstance(bizId string, context map[string]interface{}, operator *Operator) *FlowWorkAggregation {

	if context == nil {
		context = make(map[string]interface{})
	}

	return &FlowWorkAggregation{
		FlowWork: &FlowWorkEntity{
			FlowName:    define.FlowName,
			StatusAlias: define.StatusAlias,
			BizId:       bizId,
			Status:      TaskStatusInit,
			Operator:    operator,
		},
		Tasks:   make([]*TaskEntity, 0, 4),
		Context: context,
		Define:  define,
	}
}

type TaskDefine struct {
	Name         string
	SuccessEvent string
	FailEvent    string
	ReturnType   TaskReturnType
	Executor     func(task *TaskEntity, context *FlowContext) bool
}

func ClusterEnd(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusFinished
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true
	return true
}

func ClusterFail(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusError
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true
	return true
}

func DefaultEnd(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusFinished
	return true
}

func DefaultFail(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusError
	return true
}
