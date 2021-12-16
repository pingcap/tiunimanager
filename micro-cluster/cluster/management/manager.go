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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	ContextClusterMeta   = "ClusterMeta"
	ContextTopology      = "Topology"
	ContextAllocResource = "AllocResource"
	ContextInstanceID    = "InstanceID"
	ContextSourceClusterMeta = "SourceClusterMeta"
	ContextCloneStrategy     = "CloneStrategy"
	ContextBackupID          = "BackupID"
)

type Manager struct{}

func NewClusterManager() *Manager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
	// todo revert after test
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, &createClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, &deleteClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, &restartClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, &stopClusterFlow)

	return &Manager{}
}

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

// ScaleOut
// @Description scale out a cluster
// @Parameter	request
// @Return		cluster.ScaleOutClusterResp
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
			"check cluster[%s] scale out error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Add instance into cluster topology
	if err = clusterMeta.AddInstances(ctx, request.InstanceResource); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster[%s] topology error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Update cluster maintenance status and async start workflow
	data := map[string]interface{}{
		ContextClusterMeta: clusterMeta,
	}
	flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceScaleOut, scaleOutDefine.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = clusterMeta.Cluster.ID
	resp.WorkFlowID = flowID
	return

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

// ScaleIn
// @Description scale in a cluster
// @Parameter	request
// @Return		cluster.ScaleInClusterResp
// @Return		error
func (p *Manager) ScaleIn(ctx context.Context, request cluster.ScaleInClusterReq) (resp cluster.ScaleInClusterResp, err error) {
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
			"cluster[%s] has no instance[%s]", clusterMeta.Cluster.ID, request.InstanceID)
		return
	}

	// When scale in TiFlash, ensure the number of remaining TiFlash instances is
	// greater than or equal to the maximum number of copies of all data tables
	err = handler.ScaleInPreCheck(ctx, clusterMeta, instance)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check cluster[%s] scale in error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Update cluster maintenance status and async start workflow
	data := map[string]interface{}{
		ContextClusterMeta: clusterMeta,
		ContextInstanceID:  request.InstanceID,
	}
	flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceScaleIn, scaleInDefine.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = clusterMeta.Cluster.ID
	resp.WorkFlowID = flowID
	return

}

var cloneDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowCloneCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":                   {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":            {"backupSourceCluster", "backupDone", "fail", workflow.SyncFuncNode, backupSourceCluster},
		"backupDone":              {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":              {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":              {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":               {"syncBackupStrategy", "syncBackupStrategyDone", "fail", workflow.SyncFuncNode, syncBackupStrategy},
		"syncBackupStrategyDone":  {"syncParameters", "syncParametersDone", "fail", workflow.SyncFuncNode, syncParameters},
		"syncParametersDone":      {"syncSystemVariables", "syncSystemVariablesDone", "fail", workflow.SyncFuncNode, syncSystemVariables},
		"syncSystemVariablesDone": {"restoreCluster", "restoreClusterDone", "fail", workflow.SyncFuncNode, restoreCluster},
		"restoreClusterDone":      {"syncIncrData", "syncIncrDataDone", "fail", workflow.SyncFuncNode, syncIncrData},
		"syncIncrDataDone":        {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":              {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":                    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
	},
}

// Clone
// @Description clone a cluster
// @Parameter	request
// @Return		cluster.CloneClusterResp
// @Return		error
func (p *Manager) Clone(ctx context.Context, request cluster.CloneClusterReq) (resp cluster.CloneClusterResp, err error) {
	// Get source cluster info and topology from db based by SourceClusterID
	sourceClusterMeta, err := handler.Get(ctx, request.SourceClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load source cluser[%s] meta from db error: %s", request.SourceClusterID, err.Error())
		return
	}

	// Clone source cluster meta to get cluster topology
	clusterMeta, err := sourceClusterMeta.CloneMeta(ctx, request.CreateClusterParameter)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"clone cluster[%s] meta error: %s", sourceClusterMeta.Cluster.ID, err.Error())
		return
	}

	// Update cluster maintenance status and async start workflow
	data := map[string]interface{}{
		ContextClusterMeta:       clusterMeta,
		ContextSourceClusterMeta: sourceClusterMeta,
		ContextCloneStrategy:     request.CloneStrategy,
	}
	flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceCloning, cloneDefine.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Handle response
	resp.ClusterID = clusterMeta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var createClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowCreateCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":       {"startupCluster", "startupDone", "fail", workflow.PollingNode, startCluster},
		"startupDone":      {"initAccount", "initDone", "fail", workflow.SyncFuncNode, initDatabaseAccount},
		"initDone":         {"syncTopology", "syncTopologyDone", "fail", workflow.SyncFuncNode, syncTopology},
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

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceCreating, createClusterFlow.FlowName, data)

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
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster failed, clusterId = %s", req.ClusterID)
		return
	}

	if meta.Cluster.Status != string(constants.ClusterRunning) {
		errMsg := fmt.Sprintf("cannot stop cluster [%s] under status [%s]", meta.Cluster.Name, meta.Cluster.Status)
		framework.LogWithContext(ctx).Error(errMsg)
		err = framework.NewTiEMError(common.TIEM_TASK_CONFLICT, errMsg)
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceStopping, createClusterFlow.FlowName, data)

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var deleteClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowDeleteCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":              {"destroyCluster", "destroyClusterDone", "fail", workflow.PollingNode, destroyCluster},
		"destroyClusterDone": {"deleteCluster", "deleteClusterDone", "fail", workflow.SyncFuncNode, deleteCluster},
		"deleteClusterDone":  {"freedClusterResource", "freedResourceDone", "fail", workflow.SyncFuncNode, freedClusterResource},
		"freedResourceDone":  {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":               {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func (p *Manager) DeleteCluster(ctx context.Context, req cluster.DeleteClusterReq) (resp cluster.DeleteClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster failed, clusterId = %s", req.ClusterID)
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceDeleting, deleteClusterFlow.FlowName, data)

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
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster failed, clusterId = %s", req.ClusterID)
		return
	}

	if meta.Cluster.Status != string(constants.ClusterStopped) && meta.Cluster.Status != string(constants.ClusterRunning) {
		errMsg := fmt.Sprintf("cannot restart cluster [%s] under status [%s]", meta.Cluster.Name, meta.Cluster.Status)
		framework.LogWithContext(ctx).Error(errMsg)
		err = framework.NewTiEMError(common.TIEM_TASK_CONFLICT, errMsg)
		return
	}
	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceRestarting, restartClusterFlow.FlowName, data)

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
func asyncMaintenance(ctx context.Context, meta *handler.ClusterMeta, status constants.ClusterMaintenanceStatus, flowName string, data map[string]interface{}) (flowID string, err error) {
	if err = meta.StartMaintenance(ctx, status); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed, clusterID = %s, status = %s,error = %s", meta.Cluster.ID, status, err.Error())
		return
	}

	if flow, flowError := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, flowName); flowError != nil {
		meta.EndMaintenance(ctx, status)
		framework.LogWithContext(ctx).Errorf("create flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
		err = flowError
		return
	} else {
		flowID = flow.Flow.ID
		for key, value := range data {
			flow.Context.SetData(key, value)
		}
		if err = workflow.GetWorkFlowService().AsyncStart(ctx, flow); err != nil {
			meta.EndMaintenance(ctx, status)
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
