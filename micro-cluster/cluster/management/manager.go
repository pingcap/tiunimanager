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
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
)

const (
	ContextClusterMeta       = "ClusterMeta"
	ContextTopology          = "Topology"
	ContextAllocResource     = "AllocResource"
	ContextInstanceID        = "InstanceID"
	ContextSourceClusterMeta = "SourceClusterMeta"
	ContextCloneStrategy     = "CloneStrategy"
	ContextBackupID          = "BackupID"
	ContextWorkflowID        = "WorkflowID"
	ContextTopologyConfig    = "TopologyConfig"
	ContextDeleteRequest     = "DeleteRequest"
	ContextTakeoverRequest   = "TakeoverRequest"
)

type Manager struct{}

func NewClusterManager() *Manager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, &createClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestoreNewCluster, &restoreNewClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, &deleteClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, &restartClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, &stopClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCloneCluster, &cloneDefine)

	return &Manager{}
}

var scaleOutDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleOutCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"scaleOutCluster", "scaleOutDone", "fail", workflow.PollingNode, scaleOutCluster},
		"scaleOutDone":     {"syncTopology", "syncTopologyDone", "fail", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, revertResourceAfterFailure, endMaintenance)},
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
			"load cluser %s meta from db error: %s", request.ClusterID, err.Error())
		return
	}

	// When scale out TiFlash, Judge whether enable-placement-rules is true
	err = handler.ScaleOutPreCheck(ctx, clusterMeta, request.InstanceResource)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check cluster %s scale out error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Add instance into cluster topology
	if err = clusterMeta.AddInstances(ctx, request.InstanceResource); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster %s topology error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	// Update cluster maintenance status and async start workflow
	data := map[string]interface{}{
		ContextClusterMeta: clusterMeta,
	}
	flowID, err := asyncMaintenance(ctx, clusterMeta, constants.ClusterMaintenanceScaleOut, scaleOutDefine.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
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
		"freeDone":    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":        {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, revertResourceAfterFailure, endMaintenance)},
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
			"load cluser %s meta from db error: %s", request.ClusterID, err.Error())
		return
	}

	// Judge whether the instance exists
	instance, err := clusterMeta.GetInstance(ctx, request.InstanceID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s has no instance %s", clusterMeta.Cluster.ID, request.InstanceID)
		return
	}

	// When scale in TiFlash, ensure the number of remaining TiFlash instances is
	// greater than or equal to the maximum number of copies of all data tables
	err = handler.ScaleInPreCheck(ctx, clusterMeta, instance)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check cluster %s scale in error: %s", clusterMeta.Cluster.ID, err.Error())
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
			"cluster %s async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = clusterMeta.Cluster.ID
	resp.WorkFlowID = flowID
	return

}

var cloneDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowCloneCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":                  {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":           {"backupSourceCluster", "backupDone", "fail", workflow.SyncFuncNode, backupSourceCluster},
		"backupDone":             {"waitBackup", "waitBackupDone", "fail", workflow.SyncFuncNode, waitWorkFlow},
		"waitBackupDone":         {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":             {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":             {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":              {"syncBackupStrategy", "syncBackupStrategyDone", "fail", workflow.SyncFuncNode, syncBackupStrategy},
		"syncBackupStrategyDone": {"syncParameters", "syncParametersDone", "fail", workflow.SyncFuncNode, syncParameters},
		"syncParametersDone":     {"waitSyncParam", "waitSyncParamDone", "fail", workflow.SyncFuncNode, waitWorkFlow},
		"waitSyncParamDone":      {"restoreCluster", "restoreClusterDone", "fail", workflow.SyncFuncNode, restoreCluster},
		"restoreClusterDone":     {"waitRestore", "waitRestoreDone", "fail", workflow.SyncFuncNode, waitWorkFlow},
		"waitRestoreDone":        {"syncIncrData", "syncIncrDataDone", "fail", workflow.SyncFuncNode, syncIncrData},
		"syncIncrDataDone":       {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":             {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":                   {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, revertResourceAfterFailure, endMaintenance)},
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
			"load source cluser %s meta from db error: %s", request.SourceClusterID, err.Error())
		return
	}

	// Clone source cluster meta to get cluster topology
	clusterMeta, err := sourceClusterMeta.CloneMeta(ctx, request.CreateClusterParameter)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"clone cluster %s meta error: %s", sourceClusterMeta.Cluster.ID, err.Error())
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
			"cluster %s async maintenance error: %s", clusterMeta.Cluster.ID, err.Error())
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
		"deployDone":       {"startupCluster", "startupDone", "failAfterDeploy", workflow.PollingNode, startCluster},
		"startupDone":      {"initAccount", "initDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
		"initDone":         {"syncTopology", "syncTopologyDone", "failAfterDeploy", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"setClusterOnline", "onlineDone", "failAfterDeploy", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
		"failAfterDeploy":  {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
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
		framework.LogWithContext(ctx).Errorf("build cluser %s error: %s", req.Name, err.Error())
		return
	}
	if err = meta.AddInstances(ctx, req.ResourceParameter.InstanceResource); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster %s topology error: %s", meta.Cluster.ID, err.Error())
		return
	}
	if err = meta.AddDefaultInstances(ctx); err != nil {
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceCreating, createClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var restoreNewClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowRestoreNewCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":       {"startupCluster", "startupDone", "failAfterDeploy", workflow.PollingNode, startCluster},
		"startupDone":      {"initAccount", "initDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
		"initDone":         {"syncTopology", "syncTopologyDone", "failAfterDeploy", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"persistCluster", "persistDone", "failAfterDeploy", workflow.SyncFuncNode, persistCluster},
		"persistDone":      {"restoreData", "restoreDone", "failAfterDeploy", workflow.SyncFuncNode, restoreNewCluster},
		"restoreDone":      {"waitWorkFlow", "waitDone", "failAfterDeploy", workflow.SyncFuncNode, waitWorkFlow},
		"waitDone":         {"setClusterOnline", "onlineDone", "failAfterDeploy", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure, revertResourceAfterFailure)},
		"failAfterDeploy":  {"failAfterDeploy", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

// RestoreNewCluster
// @Description: restore a new cluster by backup record
// @Receiver m
// @Parameter ctx
// @Parameter request
// @Return cluster.RestoreNewClusterResp
// @Return error
func (p *Manager) RestoreNewCluster(ctx context.Context, req cluster.RestoreNewClusterReq) (resp cluster.RestoreNewClusterResp, err error) {
	meta := &handler.ClusterMeta{}

	if err = meta.BuildCluster(ctx, req.CreateClusterParameter); err != nil {
		framework.LogWithContext(ctx).Errorf("build cluser %s error: %s", req.Name, err.Error())
		return
	}
	if err = meta.AddInstances(ctx, req.ResourceParameter.InstanceResource); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster %s topology error: %s", meta.Cluster.ID, err.Error())
		return
	}
	if err = meta.AddDefaultInstances(ctx); err != nil {
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
		ContextBackupID:    req.BackupID,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceRestore, restoreNewClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

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
		framework.LogWithContext(ctx).Errorf(
			"load cluser %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceStopping, stopClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var deleteClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowDeleteCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":              {"destroyCluster", "destroyClusterDone", "fail", workflow.PollingNode, destroyCluster},
		"destroyClusterDone": {"freedClusterResource", "freedResourceDone", "fail", workflow.SyncFuncNode, freedClusterResource},
		"freedResourceDone":  {"backupBeforeDelete", "backupDone", "fail", workflow.SyncFuncNode, backupBeforeDelete},
		"backupDone":         {"clearBackupData", "clearDone", "fail", workflow.SyncFuncNode, clearBackupData},
		"clearDone":          {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(deleteCluster)},
		"fail":               {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
	},
}

func (p *Manager) DeleteCluster(ctx context.Context, req cluster.DeleteClusterReq) (resp cluster.DeleteClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta:   meta,
		ContextDeleteRequest: req,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceDeleting, deleteClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

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
		framework.LogWithContext(ctx).Errorf(
			"load cluser %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta: meta,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceRestarting, restartClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

	resp.ClusterID = meta.Cluster.ID
	resp.WorkFlowID = flowID
	return
}

var takeoverClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowTakeoverCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":   {"fetchTopologyFile", "fetched", "fail", workflow.SyncFuncNode, fetchTopologyFile},
		"fetched": {"rebuildTopologyFromConfig", "built", "fail", workflow.SyncFuncNode, rebuildTopologyFromConfig},
		"built":   {"takeoverResource", "success", "", workflow.SyncFuncNode, takeoverResource},
		"success": {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func openSftpClient(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
	conf := ssh.ClientConfig{User: req.TiUPUserName,
		Auth: []ssh.AuthMethod{ssh.Password(req.TiUPUserPassword)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(req.TiUPIp, strconv.Itoa(req.TiUPPort)), &conf)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("connect error, error: %s", err.Error())
		return nil, nil, errors.WrapError(errors.TIEM_TAKEOVER_SSH_CONNECT_ERROR, "ssh dial error", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("new sftp client failed, error: %s", err.Error())
		client.Close()
		return nil, nil, errors.WrapError(errors.TIEM_TAKEOVER_SFTP_ERROR, "new sftp client failed", err)
	}
	return client, sftpClient, nil
}

func (p *Manager) Takeover(ctx context.Context, req cluster.TakeoverClusterReq) (resp cluster.TakeoverClusterResp, err error) {
	if len(req.ClusterNames) == 0 {
		err = errors.NewError(errors.TIEM_PARAMETER_INVALID, "cluster names required")
		return
	}

	client, sftpClient, err := openSftpClient(ctx, req)
	sftpClient.Close()
	client.Close()

	if err != nil {
		return
	}

	resp.Clusters = make(map[string]string)
	resp.FailedErrors = make(map[string]string)

	for _, clusterName := range req.ClusterNames {
		meta := &handler.ClusterMeta{}
		if err = meta.BuildForTakeover(ctx, clusterName); err != nil {
			errMsg := fmt.Sprintf("takeover cluser %s error: %s", clusterName, err.Error())
			framework.LogWithContext(ctx).Errorf(errMsg)
			resp.FailedErrors[clusterName] = errMsg
			continue
		}

		data := map[string]interface{}{
			ContextClusterMeta:     meta,
			ContextTakeoverRequest: req,
		}
		flowID, startError := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceCreating, createClusterFlow.FlowName, data)
		if startError != nil {
			errMsg := fmt.Sprintf("cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
			framework.LogWithContext(ctx).Errorf(errMsg)
			resp.FailedErrors[clusterName] = errMsg
			continue
		}
		resp.Clusters[clusterName] = flowID
	}

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
func asyncMaintenance(ctx context.Context, meta *handler.ClusterMeta,
	status constants.ClusterMaintenanceStatus, flowName string, data map[string]interface{}) (string, error) {
	if err := meta.StartMaintenance(ctx, status); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"start maintenance failed, cluster %s, status %s, error: %s", meta.Cluster.ID, status, err.Error())
		return "", err
	}

	flow, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, flowName)
	if err != nil {
		meta.EndMaintenance(ctx, status)
		framework.LogWithContext(ctx).Errorf(
			"create flow %s failed, cluster %s, error: %s", flow.Flow.ID, meta.Cluster.ID, err.Error())
		return "", err
	}

	for key, value := range data {
		flow.Context.SetData(key, value)
	}
	if err = workflow.GetWorkFlowService().AsyncStart(ctx, flow); err != nil {
		meta.EndMaintenance(ctx, status)
		framework.LogWithContext(ctx).Errorf(
			"start flow %s failed, cluster %s, error: %s", flow.Flow.ID, meta.Cluster.ID, err.Error())
		return "", err
	}
	framework.LogWithContext(ctx).Infof(
		"create flow %s succeed, cluster %s", flow.Flow.ID, meta.Cluster.ID)

	return flow.Flow.ID, nil
}

func (p *Manager) QueryCluster(ctx context.Context, req cluster.QueryClustersReq) (resp cluster.QueryClusterResp, total int, err error) {
	return handler.Query(ctx, req)
}

func (p *Manager) DetailCluster(ctx context.Context, req cluster.QueryClusterDetailReq) (resp cluster.QueryClusterDetailResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster failed, clusterId = %s", req.ClusterID)
		return
	}

	resp.Info = meta.DisplayClusterInfo(ctx)
	resp.ClusterTopologyInfo, resp.ClusterResourceInfo = meta.DisplayInstanceInfo(ctx)
	return
}

func (manager *Manager) GetClusterDashboardInfo(ctx context.Context, request cluster.GetDashboardInfoReq) (resp cluster.GetDashboardInfoResp, err error) {
	return GetDashboardInfo(ctx, request)
}

func (p *Manager) GetMonitorInfo(ctx context.Context, req cluster.QueryMonitorInfoReq) (cluster.QueryMonitorInfoResp, error) {
	resp := cluster.QueryMonitorInfoResp{}
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser %s meta from db error: %s", req.ClusterID, err.Error())
		return resp, err
	}

	alertServers := clusterMeta.GetAlertManagerAddresses()
	grafanaServers := clusterMeta.GetGrafanaAddresses()
	if len(alertServers) <= 0 || len(grafanaServers) <= 0 {
		errMsg := fmt.Sprintf("cluster %s alert server or grafana server not available", clusterMeta.Cluster.ID)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return resp, errors.NewError(errors.TIEM_CLUSTER_RESOURCE_NOT_ENOUGH, errMsg)
	}

	alertPort := alertServers[0].Port
	if alertPort <= 0 {
		errMsg := fmt.Sprintf("cluster %s alert port %d not available", clusterMeta.Cluster.ID, alertPort)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return resp, errors.NewError(errors.TIEM_CLUSTER_GET_CLUSTER_PORT_ERROR, errMsg)
	}
	grafanaPort := grafanaServers[0].Port
	if grafanaPort <= 0 {
		errMsg := fmt.Sprintf("cluster %s grafana port %d not available", clusterMeta.Cluster.ID, grafanaPort)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return resp, errors.NewError(errors.TIEM_CLUSTER_GET_CLUSTER_PORT_ERROR, errMsg)
	}

	alertUrl := fmt.Sprintf("http://%s:%d", alertServers[0].IP, alertPort)
	grafanaUrl := fmt.Sprintf("http://%s:%d", grafanaServers[0].IP, grafanaPort)

	resp = cluster.QueryMonitorInfoResp{
		ClusterID:  clusterMeta.Cluster.ID,
		AlertUrl:   alertUrl,
		GrafanaUrl: grafanaUrl,
	}
	return resp, nil
}
