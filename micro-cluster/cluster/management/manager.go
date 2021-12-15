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
 ******************************************************************************/

package management

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	ContextClusterMeta       = "ClusterMeta"
	ContextTopology          = "Topology"
	ContextAllocId           = "AllocResource"
	ContextInstanceID        = "InstanceID"
	ContextSourceClusterMeta = "SourceClusterMeta"
	ContextCloneStrategy     = "CloneStrategy"
	ContextBackupID          = "BackupID"
)

type Manager struct{}

var scaleOutDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleOutCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":        {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone": {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":   {"scaleOutCluster", "scaleOutDone", "fail", workflow.PollingNode, scaleOutCluster},
		"scaleOutDone": {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":   {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":         {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

var scaleInDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleInCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":       {"scaleInCluster", "scaleInDone", "fail", workflow.PollingNode, scaleInCluster},
		"scaleInDone": {"freeInstanceResource", "freeDone", "fail", workflow.SyncFuncNode, freeInstanceResource},
		"freeDone":    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":        {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

var cloneDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowCloneCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":                   {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":            {"backupSourceCluster", "backupDone", "fail", workflow.SyncFuncNode, backupSourceCluster},
		"backupDone":              {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":              {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":              {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":               {"saveClusterMeta", "saveClusterMetaDone", "fail", workflow.SyncFuncNode, saveClusterMeta},
		"saveClusterMetaDone":     {"syncBackupStrategy", "syncBackupStrategyDone", "fail", workflow.SyncFuncNode, syncBackupStrategy},
		"syncBackupStrategyDone":  {"syncParameters", "syncParametersDone", "fail", workflow.SyncFuncNode, syncParameters},
		"syncParametersDone":      {"syncSystemVariables", "syncSystemVariablesDone", "fail", workflow.SyncFuncNode, syncSystemVariables},
		"syncSystemVariablesDone": {"restoreCluster", "restoreClusterDone", "fail", workflow.SyncFuncNode, restoreCluster},
		"restoreClusterDone":      {"syncIncrData", "syncIncrDataDone", "fail", workflow.SyncFuncNode, syncIncrData},
		"syncIncrDataDone":        {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":              {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":                    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

func NewClusterManager() *Manager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
	// todo revert after test
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, &testClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, &deleteClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, &restartClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, &stopClusterFlow)

	return &Manager{}
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	operator
// @Parameter	request
// @Return		*cluster.ScaleOutClusterResp
// @Return		error
func (p *Manager) ScaleOut(ctx context.Context, request cluster.ScaleOutClusterReq) (resp cluster.ScaleOutClusterResp, err error) {
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, request.ClusterID)

	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser[%s] meta from db error: %s", request.ClusterID, err.Error())
		return
	}

	// When scale out TiFlash, Judge whether enable-placement-rules is true
	err = handler.ScaleOutPreCheck(ctx, clusterMeta, request.InstanceResource)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check cluster[%s] scale out error: %s", clusterMeta.Cluster.Name, err.Error())
		return
	}

	// Add instance into cluster topology
	if err = clusterMeta.AddInstances(ctx, request.InstanceResource); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster[%s] topology error: %s", clusterMeta.Cluster.Name, err.Error())
		return
	}

	flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceScaleOut, scaleOutDefine.FlowName)

	resp.ClusterID = clusterMeta.Cluster.ID
	resp.WorkFlowID = flowID
	return

}

func(p *Manager) ScaleIn(ctx context.Context, request cluster.ScaleInClusterReq) (resp cluster.ScaleInClusterResp, err error) {
		// Get cluster info and topology from db based by clusterID
		clusterMeta, err := handler.Get(ctx, request.ClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"load cluser[%s] meta from db error: %s", request.ClusterID, err.Error())
			return
		}

		// Judge whether the instance exists
		instance, err := clusterMeta.GetInstance(ctx, request.InstanceID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"cluster[%s] has no instance[%s]", clusterMeta.Cluster.Name, request.InstanceID)
			return
		}

		// When scale in TiFlash, ensure the number of remaining TiFlash instances is
		// greater than or equal to the maximum number of copies of all data tables
		err = handler.ScaleInPreCheck(ctx, clusterMeta, instance)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"check cluster[%s] scale in error: %s", clusterMeta.Cluster.Name, err.Error())
			return
		}

		flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceScaleIn, scaleInDefine.FlowName)

		resp.ClusterID = clusterMeta.Cluster.ID
		resp.WorkFlowID = flowID
		return

}

	func(p *Manager) Clone(ctx context.Context, request cluster.CloneClusterReq) (cluster.CloneClusterResp, error) {
		response := cluster.CloneClusterResp{}

		// Get source cluster info and topology from db based by SourceClusterID
		sourceClusterMeta, err := handler.Get(ctx, request.SourceClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"load source cluser[%s] meta from db error: %s", request.SourceClusterID, err.Error())
			return response, err
		}

		// Clone source cluster meta to get cluster topology
		clusterMeta, err := sourceClusterMeta.CloneMeta(ctx, request.CreateClusterParameter)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"clone cluster[%s] meta error: %s", sourceClusterMeta.Cluster.Name, err.Error())
			return response, err
		}
		// add cluster meta into db
		// TODO: add cluster meta into db

		// Start the workflow to clone a cluster
		workflowManager := workflow.GetWorkFlowService()
		flow, err := workflowManager.CreateWorkFlow(ctx, clusterMeta.Cluster.ID, constants.FlowCloneCluster)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create workflow error: %s", err.Error())
			return response, err
		}
		workflowManager.AddContext(flow, ContextClusterMeta, clusterMeta)
		workflowManager.AddContext(flow, ContextSourceClusterMeta, sourceClusterMeta)
		workflowManager.AddContext(flow, ContextCloneStrategy, request.CloneStrategy)

		if err = workflowManager.AsyncStart(ctx, flow); err != nil {
			framework.LogWithContext(ctx).Errorf("async start workflow[%s] error: %s", flow.Flow.ID, err.Error())
			return response, err
		}

		// Handle response
		response.ClusterID = clusterMeta.Cluster.ID
		response.WorkFlowID = flow.Flow.ID
		return response, nil
	}

// todo delete after test
var testClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowCreateCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start": {"start", "succeed", "fail", workflow.SyncFuncNode, func(task *wfModel.WorkFlowNode, context *workflow.FlowContext) error {
			return nil
		}},
		"succeed": {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

var createClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowCreateCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":       {"startupCluster", "startupDone", "fail", workflow.PollingNode, startCluster},
		"startupDone":      {"syncTopology", "syncTopologyDone", "fail", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

// CreateCluster
// @Description: See createClusterFlow
// @Receiver p
// @Parameter ctx
// @Parameter req
// @return resp
// @return err
func (p *Manager) CreateCluster(ctx context.Context, req cluster.CreateClusterReq) (resp cluster.CreateClusterResp, err error) {
	meta := &handler.ClusterMeta{}
	if err = meta.BuildCluster(ctx, req.CreateClusterParameter); err != nil {
		return
	}
	if err = meta.AddInstances(ctx, req.ResourceParameter.InstanceResource); err != nil {
		return
	}

	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceCreating, createClusterFlow.FlowName)

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var stopClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowStopCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":       {"clusterStop", "stopDone", "fail", workflow.PollingNode, stopCluster},
		"stopDone":    {"setClusterOffline", "offlineDone", "fail", workflow.SyncFuncNode, setClusterOffline},
		"offlineDone": {"end", "", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":        {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func (p *Manager) StopCluster(ctx context.Context, req cluster.StopClusterReq) (resp cluster.StopClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceStopping, createClusterFlow.FlowName)

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var deleteClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowDeleteCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":              {"destroyCluster", "destroyClusterDone", "fail", workflow.PollingNode, destroyCluster},
		"destroyClusterDone": {"deleteCluster", "deleteClusterDone", "fail", workflow.SyncFuncNode, deleteCluster},
		"deleteClusterDone":  {"freedResource", "freedResourceDone", "fail", workflow.SyncFuncNode, freedResource},
		"freedResourceDone":  {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":               {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func (p *Manager) DeleteCluster(ctx context.Context, req cluster.DeleteClusterReq) (resp cluster.DeleteClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceDeleting, deleteClusterFlow.FlowName)

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var restartClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowRestartCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":      {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":  {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone": {"end", "", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":       {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func (p *Manager) RestartCluster(ctx context.Context, req cluster.RestartClusterReq) (resp cluster.RestartClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceRestarting, restartClusterFlow.FlowName)

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

// asyncMaintenance
// @Description: common asynchronous process for cluster maintenance
// @Parameter ctx
// @Parameter meta
// @Parameter status
// @Parameter flowName
// @return flowID
// @return err
func asyncMaintenance(ctx context.Context, meta *handler.ClusterMeta, status constants.ClusterMaintenanceStatus, flowName string) (flowID string, err error) {
	if err = meta.StartMaintenance(ctx, status); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed, clusterID = %s, status = %s,error = %s", meta.Cluster.ID, status, err.Error())
		return
	}

	if flow, flowError := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, flowName); flowError != nil {
		framework.LogWithContext(ctx).Errorf("create flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
		err = flowError
		return
	} else {
		flowID = flow.Flow.ID
		flow.Context.SetData(ContextClusterMeta, meta)
		if err = workflow.GetWorkFlowService().AsyncStart(ctx, flow); err != nil {
			framework.LogWithContext(ctx).Errorf("start flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
			return
		}
		framework.LogWithContext(ctx).Infof("create flow %s succeed, clusterID = %s", flow.Flow.Name, meta.Cluster.ID)
	}
	return
}

func (p *Manager) QueryCluster(ctx context.Context, req cluster.QueryClustersReq) (resp cluster.QueryClusterResp, total int, err error) {
	return
}

func (p *Manager) DetailCluster(ctx context.Context, req cluster.QueryClusterDetailReq) (resp cluster.QueryClusterDetailResp, err error) {
	return
}

func (manager *Manager) GetClusterDashboardInfo(ctx context.Context, request *cluster.GetDashboardInfoReq) (*cluster.GetDashboardInfoResp, error) {
	return GetDashboardInfo(ctx, request)
}