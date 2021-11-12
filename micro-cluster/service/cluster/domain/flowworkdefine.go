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
	copywriting2 "github.com/pingcap-inc/tiem/library/copywriting"
	"github.com/pingcap-inc/tiem/library/framework"
	"time"
)

func defaultContextParser(s string) *FlowContext {
	// todo parse context
	return NewFlowContext(context.TODO())
}

func InitFlowMap() {
	FlowWorkDefineMap = map[string]*FlowWorkDefine{
		FlowCreateCluster: {
			FlowName:    FlowCreateCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowCreateCluster),
			TaskNodes: map[string]*TaskDefine{
				"start":        {"prepareResource", "resourceDone", "fail", SyncFuncTask, prepareResource},
				"resourceDone": {"buildConfig", "configDone", "fail", SyncFuncTask, buildConfig},
				"configDone":   {"deployCluster", "deployDone", "fail", PollingTasK, deployCluster},
				"deployDone":   {"startupCluster", "startupDone", "fail", PollingTasK, startupCluster},
				"startupDone":  {"syncTopology", "syncTopologyDone", "fail", SyncFuncTask, syncTopology},
				"syncTopologyDone":  {"setClusterOnline", "onlineDone", "fail", SyncFuncTask, setClusterOnline},
				"onlineDone":   {"end", "", "", SyncFuncTask, ClusterEnd},
				"fail":         {"fail", "", "", SyncFuncTask, ClusterFail},
			},
			ContextParser: defaultContextParser,
		},
		FlowDeleteCluster: {
			FlowName:    FlowDeleteCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowDeleteCluster),
			TaskNodes: map[string]*TaskDefine{
				"start":              {"destroyTasks", "destroyTasksDone", "fail", SyncFuncTask, destroyTasks},
				"destroyTasksDone":   {"destroyCluster", "destroyClusterDone", "fail", SyncFuncTask, destroyCluster},
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
				"configDone":   {"deployCluster", "deployDone", "fail", PollingTasK, deployCluster},
				"deployDone":   {"startupCluster", "startupDone", "fail", PollingTasK, startupCluster},
				"startupDone":  {"recoverFromSrcCluster", "recoverDone", "fail", PollingTasK, recoverFromSrcCluster},
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
				"start":       {"clusterRestart", "restartDone", "fail", PollingTasK, clusterRestart},
				"restartDone": {"setClusterOnline", "onlineDone", "fail", SyncFuncTask, setClusterOnline},
				"onlineDone":  {"end", "", "fail", SyncFuncTask, ClusterEnd},
				"fail":        {"fail", "", "", SyncFuncTask, ClusterFail},
			},
			ContextParser: defaultContextParser,
		},
		FlowStopCluster: {
			FlowName:    FlowStopCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowStopCluster),
			TaskNodes: map[string]*TaskDefine{
				"start":       {"clusterStop", "stopDone", "fail", PollingTasK, clusterStop},
				"stopDone":    {"setClusterOffline", "offlineDone", "fail", SyncFuncTask, setClusterOffline},
				"offlineDone": {"end", "", "fail", SyncFuncTask, ClusterEnd},
				"fail":        {"fail", "", "", SyncFuncTask, ClusterFail},
			},
			ContextParser: defaultContextParser,
		},
		FlowTakeoverCluster: {
			FlowName:    FlowTakeoverCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowTakeoverCluster),
			TaskNodes: map[string]*TaskDefine{
				"start":   {"fetchTopologyFile", "fetched", "fail", SyncFuncTask, fetchTopologyFile},
				"fetched": {"buildTopology", "built", "fail", SyncFuncTask, buildTopology},
				"built":   {"takeoverResource", "success", "", SyncFuncTask, takeoverResource},
				"success": {"end", "", "", SyncFuncTask, ClusterEnd},
				"fail":    {"fail", "", "", SyncFuncTask, ClusterFail},
			},
			ContextParser: defaultContextParser,
		},
		FlowBuildLogConfig: {
			FlowName:    FlowBuildLogConfig,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowBuildLogConfig),
			TaskNodes: map[string]*TaskDefine{
				"start":   {"collect", "success", "fail", SyncFuncTask, collectorTiDBLogConfig},
				"success": {"end", "", "", SyncFuncTask, ClusterEnd},
				"fail":    {"fail", "", "", SyncFuncTask, ClusterFail},
			},
			ContextParser: defaultContextParser,
		},
	}
}

var FlowWorkDefineMap map[string]*FlowWorkDefine

type FlowWorkDefine struct {
	FlowName      string
	StatusAlias   string
	TaskNodes     map[string]*TaskDefine
	ContextParser func(string) *FlowContext
}

func (define *FlowWorkDefine) getInstance(ctx context.Context, bizId string, data map[string]interface{}, operator *Operator) *FlowWorkAggregation {
	if data == nil {
		data = make(map[string]interface{})
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
		Context: FlowContext{ctx, data},
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
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true

	if clusterAggregation.ConfigModified {
		go time.AfterFunc(time.Second*3, func() {
			framework.LogWithContext(context.Context).Infof("BuildClusterLogConfig for cluster %s", clusterAggregation.Cluster.Id)
			err := BuildClusterLogConfig(context.Context, clusterAggregation.Cluster.Id)
			if err != nil {
				framework.LogWithContext(context.Context).Error(err)
			}
		})
	}
	return true
}

func ClusterFail(task *TaskEntity, context *FlowContext) bool {
	task.Status = TaskStatusError
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
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
