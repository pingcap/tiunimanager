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
	"net"
	"strconv"

	"github.com/pingcap-inc/tiem/models"

	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	ContextClusterMeta       = "ClusterMeta"
	ContextTopology          = "Topology"
	ContextAllocResource     = "AllocResource"
	ContextInstanceID        = "InstanceID"
	ContextSourceClusterMeta = "SourceClusterMeta"
	ContextCloneStrategy     = "CloneStrategy"
	ContextBackupID          = "BackupID"
	ContextUpgradeVersion    = "UpgradeVersion"
	ContextUpgradeWay        = "UpgradeWay"
	ContextWorkflowID        = "WorkflowID"
	ContextTopologyConfig    = "TopologyConfig"
	ContextPublicKey         = "PublicKey"
	ContextPrivateKey        = "PrivateKey"
	ContextDeleteRequest     = "DeleteRequest"
	ContextTakeoverRequest   = "TakeoverRequest"
	ContextGCLifeTime        = "GCLifeTime"
	ContextInstanceTypes     = "InstanceTypes"
)

type Manager struct{}

func NewClusterManager() *Manager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, &createClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, &deleteClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, &restartClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, &stopClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowOnlineInPlaceUpgradeCluster, &onlineInPlaceUpgradeClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowOfflineInPlaceUpgradeCluster, &offlineInPlaceUpgradeClusterFlow)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCloneCluster, &cloneDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowTakeoverCluster, &takeoverClusterFlow)

	return &Manager{}
}

var scaleOutDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleOutCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"scaleOutCluster", "scaleOutDone", "fail", workflow.PollingNode, scaleOutCluster},
		"scaleOutDone":     {"syncTopology", "syncTopologyDone", "fail", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"getTypes", "getTypesDone", "fail", workflow.SyncFuncNode, getFirstScaleOutTypes},
		"getTypesDone":     {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"updateClusterParameters", "updateDone", "failAfterScale", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, updateClusterParameters)},
		"updateDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance, asyncBuildLog)},
		"fail":             {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(revertResourceAfterFailure, endMaintenance)},
		"failAfterScale":   {"end", "", "", workflow.SyncFuncNode, endMaintenance},
	},
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	request
// @Return		cluster.ScaleOutClusterResp
// @Return		error
func (p *Manager) ScaleOut(ctx context.Context, request cluster.ScaleOutClusterReq) (resp cluster.ScaleOutClusterResp, err error) {
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", request.ClusterID, err.Error())
		return
	}

	// When scale out TiFlash, Judge whether enable-placement-rules is true
	err = meta.ScaleOutPreCheck(ctx, clusterMeta, request.InstanceResource)
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
		"scaleInDone": {"checkInstanceStatus", "checkDone", "fail", workflow.SyncFuncNode, checkInstanceStatus},
		"checkDone":   {"freeInstanceResource", "freeDone", "fail", workflow.SyncFuncNode, freeInstanceResource},
		"freeDone":    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":        {"end", "", "", workflow.SyncFuncNode, endMaintenance},
	},
}

// ScaleIn
// @Description scale in a cluster
// @Parameter	request
// @Return		cluster.ScaleInClusterResp
// @Return		error
func (p *Manager) ScaleIn(ctx context.Context, request cluster.ScaleInClusterReq) (resp cluster.ScaleInClusterResp, err error) {
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", request.ClusterID, err.Error())
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
	err = meta.ScaleInPreCheck(ctx, clusterMeta, instance)
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
		"start":                   {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":            {"modifySourceClusterGCTime", "modifyGCTimeDone", "fail", workflow.SyncFuncNode, modifySourceClusterGCTime},
		"modifyGCTimeDone":        {"backupSourceCluster", "backupDone", "fail", workflow.SyncFuncNode, backupSourceCluster},
		"backupDone":              {"waitBackup", "waitBackupDone", "fail", workflow.SyncFuncNode, waitWorkFlow},
		"waitBackupDone":          {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":              {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":              {"syncConnectionKey", "syncConnectionKeyDone", "failAfterDeploy", workflow.SyncFuncNode, syncConnectionKey},
		"syncConnectionKeyDone":   {"syncTopology", "syncTopologyDone", "failAfterDeploy", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone":        {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":               {"setClusterOnline", "onlineDone", "failAfterDeploy", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":              {"initRootAccount", "initRootAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initRootAccount},
		"initRootAccountDone":     {"initAccount", "initAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
		"initAccountDone":         {"applyParameterGroup", "applyParameterGroupDone", "failAfterDeploy", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, applyParameterGroup)},
		"applyParameterGroupDone": {"syncBackupStrategy", "syncBackupStrategyDone", "failAfterDeploy", workflow.SyncFuncNode, syncBackupStrategy},
		"syncBackupStrategyDone":  {"syncParameters", "syncParametersDone", "failAfterDeploy", workflow.SyncFuncNode, syncParameters},
		"syncParametersDone":      {"waitSyncParam", "waitSyncParamDone", "failAfterDeploy", workflow.SyncFuncNode, waitWorkFlow},
		"waitSyncParamDone":       {"adjustParameters", "initParametersDone", "failAfterDeploy", workflow.SyncFuncNode, adjustParameters},
		"initParametersDone":      {"restoreCluster", "restoreClusterDone", "failAfterDeploy", workflow.SyncFuncNode, restoreCluster},
		"restoreClusterDone":      {"waitRestore", "waitRestoreDone", "failAfterDeploy", workflow.SyncFuncNode, waitWorkFlow},
		"waitRestoreDone":         {"syncIncrData", "syncIncrDataDone", "failAfterDeploy", workflow.SyncFuncNode, syncIncrData},
		"syncIncrDataDone":        {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, persistCluster, endMaintenance, asyncBuildLog)},
		"fail":                    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, setClusterFailure, revertResourceAfterFailure, endMaintenance)},
		"failAfterDeploy":         {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, setClusterFailure, endMaintenance)},
	},
}

// Clone
// @Description clone a cluster
// @Parameter	request
// @Return		cluster.CloneClusterResp
// @Return		error
func (p *Manager) Clone(ctx context.Context, request cluster.CloneClusterReq) (resp cluster.CloneClusterResp, err error) {
	// Get source cluster info and topology from db based by SourceClusterID
	sourceClusterMeta, err := meta.Get(ctx, request.SourceClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load source cluster %s meta from db error: %s", request.SourceClusterID, err.Error())
		return
	}

	// Clone source cluster meta to get cluster topology
	clusterMeta, err := sourceClusterMeta.CloneMeta(ctx, request.CreateClusterParameter)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"clone cluster %s meta error: %s", sourceClusterMeta.Cluster.ID, err.Error())
		return
	}

	// When use CDCSyncClone strategy to clone cluster, source cluster must have CDC
	err = meta.ClonePreCheck(ctx, sourceClusterMeta, clusterMeta, request.CloneStrategy)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check cluster %s clone error: %s", sourceClusterMeta.Cluster.ID, err.Error())
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
		"start":                   {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":            {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":              {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
		"deployDone":              {"syncConnectionKey", "syncConnectionKeyDone", "failAfterDeploy", workflow.SyncFuncNode, syncConnectionKey},
		"syncConnectionKeyDone":   {"syncTopology", "syncTopologyDone", "failAfterDeploy", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone":        {"startupCluster", "startupDone", "failAfterDeploy", workflow.PollingNode, startCluster},
		"startupDone":             {"setClusterOnline", "onlineDone", "failAfterDeploy", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":              {"initRootAccount", "initRootAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initRootAccount},
		"initRootAccountDone":     {"initDatabaseAccount", "initDatabaseAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
		"initDatabaseAccountDone": {"applyParameterGroup", "applyParameterGroupDone", "failAfterDeploy", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, applyParameterGroup)},
		"applyParameterGroupDone": {"adjustParameters", "initParametersDone", "failAfterDeploy", workflow.SyncFuncNode, adjustParameters},
		"initParametersDone":      {"testConnectivity", "testConnectivityDone", "failAfterDeploy", workflow.SyncFuncNode, testConnectivity},
		"testConnectivityDone":    {"initDatabaseData", "initDataDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseData},
		"initDataDone":            {"waitInitDatabaseData", "success", "failAfterDeploy", workflow.SyncFuncNode, waitInitDatabaseData},
		"success":                 {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance, asyncBuildLog)},
		"fail":                    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, revertResourceAfterFailure, endMaintenance)},
		"failAfterDeploy":         {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
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
	err = validator(ctx, &req)
	if err != nil {
		return
	}
	meta := &meta.ClusterMeta{}
	if err = meta.BuildCluster(ctx, req.CreateClusterParameter); err != nil {
		framework.LogWithContext(ctx).Errorf("build cluster %s error: %s", req.Name, err.Error())
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

// RestoreNewCluster
// @Description: restore a new cluster by backup record
// @Receiver m
// @Parameter ctx
// @Parameter request
// @Return cluster.RestoreNewClusterResp
// @Return error
func (p *Manager) RestoreNewCluster(ctx context.Context, req cluster.RestoreNewClusterReq) (resp cluster.RestoreNewClusterResp, err error) {
	meta := &meta.ClusterMeta{}

	if err = p.restoreNewClusterPreCheck(ctx, req); err != nil {
		framework.LogWithContext(ctx).Errorf("restore new cluster precheck failed: %s", err.Error())
		return
	}

	if err = meta.BuildCluster(ctx, req.CreateClusterParameter); err != nil {
		framework.LogWithContext(ctx).Errorf("build cluster %s error: %s", req.Name, err.Error())
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

// PreviewCluster
// @Description: preview cluster
// @Receiver p
// @Parameter ctx
// @Parameter req
// @return resp
// @return err
func (p *Manager) PreviewCluster(ctx context.Context, req cluster.CreateClusterReq) (resp cluster.PreviewClusterResp, err error) {
	_, total, _ := meta.Query(ctx, cluster.QueryClustersReq{
		Name: req.Name,
		PageRequest: structs.PageRequest{
			Page:     1,
			PageSize: 1,
		},
	})
	if total > 0 {
		err = errors.Error(errors.TIEM_DUPLICATED_NAME)
		return
	}

	err = validator(ctx, &req)
	if err != nil {
		return
	}
	resp = cluster.PreviewClusterResp{
		Region:            req.Region,
		CpuArchitecture:   req.CpuArchitecture,
		ClusterType:       req.Type,
		ClusterVersion:    req.Version,
		ClusterName:       req.Name,
		CapabilityIndexes: []structs.Index{},
	}

	checkResult, err := preCheckStock(ctx, req.Region, req.CpuArchitecture, req.ResourceParameter.InstanceResource)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check stocks failed, err = %s", err.Error())
		return
	} else {
		resp.StockCheckResult = checkResult
	}

	return
}

func preCheckStock(ctx context.Context, region string, arch string, instanceResource []structs.ClusterResourceParameterCompute) ([]structs.ResourceStockCheckResult, error) {
	result := make([]structs.ResourceStockCheckResult, 0)

	stocks, err := resourcepool.GetResourcePool().GetStocks(ctx, &structs.Location{
		Region: region,
	}, &structs.HostFilter{
		Arch: arch,
	}, &structs.DiskFilter{})

	if err != nil {
		return result, err
	}

	for _, instance := range instanceResource {
		for _, resource := range instance.Resource {
			enough := true
			if zoneResource, ok := stocks[resource.Zone]; ok &&
				zoneResource.FreeHostCount >= int32(resource.Count) &&
				zoneResource.FreeDiskCount >= int32(resource.Count) &&
				zoneResource.FreeCpuCores >= int32(structs.ParseCpu(resource.Spec)*resource.Count) &&
				zoneResource.FreeMemory >= int32(structs.ParseMemory(resource.Spec)*resource.Count) {

				enough = true
				// deduction
				zoneResource.FreeHostCount = zoneResource.FreeHostCount - int32(resource.Count)
				zoneResource.FreeDiskCount = zoneResource.FreeDiskCount - int32(resource.Count)
				zoneResource.FreeCpuCores = zoneResource.FreeCpuCores - int32(structs.ParseCpu(resource.Spec)*resource.Count)
				zoneResource.FreeMemory = zoneResource.FreeMemory - int32(structs.ParseMemory(resource.Spec)*resource.Count)
			} else {
				framework.LogWithContext(ctx).Warnf("stock is not enough, instance: %v, stock %v", resource, stocks)
				enough = false
			}

			result = append(result, structs.ResourceStockCheckResult{
				Type:                                    instance.Type,
				Name:                                    instance.Type,
				ClusterResourceParameterComputeResource: resource,
				Enough:                                  enough,
			})
		}
	}
	return result, nil
}

// PreviewScaleOutCluster
// @Description: preview
// @Receiver p
// @Parameter ctx
// @Parameter req
// @return resp
// @return err
func (p *Manager) PreviewScaleOutCluster(ctx context.Context, req cluster.ScaleOutClusterReq) (resp cluster.PreviewClusterResp, err error) {
	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		return
	}

	// todo validate
	resp = cluster.PreviewClusterResp{
		Region:            clusterMeta.Cluster.Region,
		CpuArchitecture:   string(clusterMeta.Cluster.CpuArchitecture),
		ClusterType:       clusterMeta.Cluster.Type,
		ClusterVersion:    clusterMeta.Cluster.Version,
		ClusterName:       clusterMeta.Cluster.Name,
		CapabilityIndexes: []structs.Index{},
	}
	checkResult, err := preCheckStock(ctx, clusterMeta.Cluster.Region, string(clusterMeta.Cluster.CpuArchitecture), req.InstanceResource)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check stocks failed, err = %s", err.Error())
		return
	} else {
		resp.StockCheckResult = checkResult
	}

	return
}

var stopClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowStopCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":       {"clusterStop", "stopDone", "fail", workflow.PollingNode, stopCluster},
		"stopDone":    {"setClusterOffline", "offlineDone", "fail", workflow.SyncFuncNode, setClusterOffline},
		"offlineDone": {"end", "", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":        {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
	},
}

func (p *Manager) StopCluster(ctx context.Context, req cluster.StopClusterReq) (resp cluster.StopClusterResp, err error) {
	meta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
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
		"start":              {"backupBeforeDelete", "backupDone", "revert", workflow.SyncFuncNode, backupBeforeDelete},
		"backupDone":         {"destroyCluster", "destroyClusterDone", "fail", workflow.PollingNode, destroyCluster},
		"destroyClusterDone": {"freedClusterResource", "freedResourceDone", "fail", workflow.SyncFuncNode, freedClusterResource},
		"freedResourceDone":  {"clearBackupData", "clearDone", "fail", workflow.SyncFuncNode, clearBackupData},
		"clearDone":          {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(deleteCluster)},
		"fail":               {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
		"revert":               {"end", "", "", workflow.SyncFuncNode, endMaintenance},
	},
}

func (p *Manager) DeleteCluster(ctx context.Context, req cluster.DeleteClusterReq) (resp cluster.DeleteClusterResp, err error) {
	meta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	resp.ClusterID = meta.Cluster.ID

	if len(meta.Cluster.MaintenanceStatus) > 0 && !req.Force {
		msg := fmt.Sprintf("cluster maintenance status is '%s'", string(meta.Cluster.MaintenanceStatus))
		err = errors.NewError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, msg)
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

	resp.WorkFlowID = flowID
	return
}

var restartClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowRestartCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":      {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
		"startDone":  {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone": {"end", "", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
		"fail":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
	},
}

func (p *Manager) RestartCluster(ctx context.Context, req cluster.RestartClusterReq) (resp cluster.RestartClusterResp, err error) {
	meta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
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
		"start":                   {"fetchTopologyFile", "fetched", "revert", workflow.SyncFuncNode, fetchTopologyFile},
		"fetched":                 {"rebuildTopologyFromConfig", "built", "revert", workflow.SyncFuncNode, rebuildTopologyFromConfig},
		"built":                   {"validateHostsStatus", "importDone", "revert", workflow.SyncFuncNode, validateHostsStatus},
		"importDone":              {"takeoverResource", "resourceDone", "revert", workflow.SyncFuncNode, takeoverResource},
		"resourceDone":            {"rebuildTiupSpaceForCluster", "workingSpaceDone", "fail", workflow.SyncFuncNode, rebuildTiupSpaceForCluster},
		"workingSpaceDone":        {"applyParameterGroup", "applyParameterGroupDone", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, applyParameterGroupForTakeover)},
		"applyParameterGroupDone": {"initDatabaseAccount", "initDatabaseAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
		"initDatabaseAccountDone": {"testConnectivity", "success", "", workflow.SyncFuncNode, testConnectivity},
		"success":                 {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance, asyncBuildLog)},
		"fail":                    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(setClusterFailure, endMaintenance)},
		"revert":                  {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(clearClusterPhysically)},
	},
}

type openSftpClientFunc func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error)

var openSftpClient openSftpClientFunc = func(ctx context.Context, req cluster.TakeoverClusterReq) (*ssh.Client, *sftp.Client, error) {
	conf := ssh.ClientConfig{User: req.TiUPUserName,
		Auth: []ssh.AuthMethod{ssh.Password(req.TiUPUserPassword)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * 3,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(req.TiUPIp, strconv.Itoa(req.TiUPPort)), &conf)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("connection error: %s", err.Error())
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
	if len(req.ClusterName) == 0 {
		err = errors.NewError(errors.TIEM_PARAMETER_INVALID, "cluster name required")
		return
	}

	client, sftpClient, err := openSftpClient(ctx, req)

	defer func() {
		if sftpClient != nil {
			sftpClient.Close()
		}
		if client != nil {
			client.Close()
		}
	}()
	if err != nil {
		return
	}
	meta := &meta.ClusterMeta{}
	if err = meta.BuildForTakeover(ctx, req.ClusterName, req.DBPassword); err != nil {
		framework.LogWithContext(ctx).Errorf(err.Error())
		return
	}

	data := map[string]interface{}{
		ContextClusterMeta:     meta,
		ContextTakeoverRequest: req,
	}
	flowID, err := asyncMaintenance(ctx, meta, constants.ClusterMaintenanceTakeover, takeoverClusterFlow.FlowName, data)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster %s async maintenance error: %s", meta.Cluster.ID, err.Error())
		return
	}

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
func asyncMaintenance(ctx context.Context, meta *meta.ClusterMeta,
	status constants.ClusterMaintenanceStatus, flowName string, data map[string]interface{}) (string, error) {
	if err := meta.StartMaintenance(ctx, status); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"start maintenance failed, cluster %s, status %s, error: %s", meta.Cluster.ID, status, err.Error())
		return "", err
	}

	flow, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, workflow.BizTypeCluster, flowName)
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
	return meta.Query(ctx, req)
}

func (p *Manager) DetailCluster(ctx context.Context, req cluster.QueryClusterDetailReq) (resp cluster.QueryClusterDetailResp, err error) {
	meta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	resp.Info = meta.DisplayClusterInfo(ctx)
	resp.ClusterTopologyInfo, resp.ClusterResourceInfo = meta.DisplayInstanceInfo(ctx)
	return
}

func (p *Manager) GetClusterDashboardInfo(ctx context.Context, request cluster.GetDashboardInfoReq) (resp cluster.GetDashboardInfoResp, err error) {
	return GetDashboardInfo(ctx, request)
}

func (p *Manager) GetMonitorInfo(ctx context.Context, req cluster.QueryMonitorInfoReq) (cluster.QueryMonitorInfoResp, error) {
	resp := cluster.QueryMonitorInfoResp{}
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
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

func (p *Manager) restoreNewClusterPreCheck(ctx context.Context, req cluster.RestoreNewClusterReq) error {
	if req.BackupID == "" {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("restore new cluster input backupId empty"))
	}

	brService := backuprestore.GetBRService()
	resp, _, err := brService.QueryClusterBackupRecords(ctx, cluster.QueryBackupRecordsReq{
		BackupID: req.BackupID,
		PageRequest: structs.PageRequest{
			Page:     1,
			PageSize: 1,
		},
	})
	if err != nil {
		return errors.NewErrorf(errors.TIEM_BACKUP_RECORD_QUERY_FAILED, err.Error())
	}
	if len(resp.BackupRecords) <= 0 {
		return errors.NewErrorf(errors.TIEM_BACKUP_RECORD_QUERY_FAILED, fmt.Sprintf("backup recordId %s not found", req.BackupID))
	}
	if resp.BackupRecords[0].Status != string(constants.ClusterBackupFinished) {
		return errors.NewErrorf(errors.TIEM_BACKUP_RECORD_INVALID, fmt.Sprintf("backup record status invalid"))
	}

	return nil
}

var onlineInPlaceUpgradeClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowOnlineInPlaceUpgradeCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":          {"editConfig", "editConfigDone", "fail", workflow.SyncFuncNode, editConfig},
		"editConfigDone": {"clusterUpgrade", "upgradeDone", "fail", workflow.PollingNode, upgradeCluster},
		"upgradeDone":    {"end", "", "fail", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":           {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

var offlineInPlaceUpgradeClusterFlow = workflow.WorkFlowDefine{
	FlowName: constants.FlowOfflineInPlaceUpgradeCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"editConfig", "editConfigDone", "fail", workflow.SyncFuncNode, editConfig},
		"editConfigDone":   {"clusterStop", "stopClusterDone", "fail", workflow.PollingNode, stopCluster},
		"stopClusterDone":  {"clusterUpgrade", "upgradeDone", "fail", workflow.PollingNode, upgradeCluster},
		"upgradeDone":      {"clusterStart", "startClusterDone", "fail", workflow.SyncFuncNode, startCluster},
		"startClusterDone": {"end", "", "fail", workflow.PollingNode, workflow.CompositeExecutor(endMaintenance, persistCluster)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(endMaintenance, setClusterFailure)},
	},
}

func (p *Manager) QueryProductUpdatePath(ctx context.Context, clusterID string) (*cluster.QueryUpgradePathRsp, error) {
	meta, err := meta.Get(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path while getcluster, %s", err.Error())
		return &cluster.QueryUpgradePathRsp{}, errors.WrapError(errors.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query update path while getclusterdetail", err)
	}

	version := meta.Cluster.Version
	framework.LogWithContext(ctx).Infof("query update path with version %s", version)
	productUpgradePaths, err := models.GetUpgradeReaderWriter().QueryBySrcVersion(ctx, version)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path, %s", err.Error())
		return &cluster.QueryUpgradePathRsp{}, errors.WrapError(errors.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query upgrade path", err)
	}
	framework.LogWithContext(ctx).Infof("query update path with version %s result: %d", version, len(productUpgradePaths))

	pathMap := make(map[string][]string)
	for _, productUpgradePath := range productUpgradePaths {
		framework.LogWithContext(ctx).Infof("loop type %s dstversion: %s", productUpgradePath.Type, productUpgradePath.DstVersion)
		if versions, ok := pathMap[productUpgradePath.Type]; ok {
			versions = append(versions, productUpgradePath.DstVersion)
			pathMap[productUpgradePath.Type] = versions
		} else {
			versions = []string{productUpgradePath.DstVersion}
			pathMap[productUpgradePath.Type] = versions
		}
	}

	framework.LogWithContext(ctx).Infof("pathMap %v", pathMap)
	var paths []*structs.ProductUpgradePathItem
	for k, v := range pathMap {
		path := structs.ProductUpgradePathItem{
			UpgradeType: k,
			Versions:    v,
		}
		paths = append(paths, &path)
	}
	framework.LogWithContext(ctx).Infof("paths %d", len(paths))

	resp := &cluster.QueryUpgradePathRsp{
		Paths: paths,
	}

	return resp, nil
}

func (p *Manager) QueryUpgradeVersionDiffInfo(ctx context.Context, clusterID string, version string) ([]*structs.ProductUpgradeVersionConfigDiffItem, error) {
	meta, err := meta.Get(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Error("failed to query upgrade version diff, %s", err)
		return []*structs.ProductUpgradeVersionConfigDiffItem{}, errors.WrapError(errors.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query upgrade version diff", err)
	}

	srcVersion := meta.Cluster.Version
	// TODO: get params for clusterID and dst version and check the diffs
	framework.LogWithContext(ctx).Infof("TODO: get params for current cluster(%s:%s) and dst version(%s) and get get diffs", clusterID, srcVersion, version)

	var configDiffInfos []*structs.ProductUpgradeVersionConfigDiffItem

	return configDiffInfos, nil
}

// InPlaceUpgradeCluster
// @Description: See onlineInPlaceUpgradeClusterFlow
// @Receiver p
// @Parameter ctx
// @Parameter req
// @return resp
// @return err
func (p *Manager) InPlaceUpgradeCluster(ctx context.Context, req *cluster.ClusterUpgradeReq) (resp *cluster.ClusterUpgradeResp, err error) {
	framework.LogWithContext(ctx).Infof("inplaceupgradecluster, handle request [%+v]", *req)
	meta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	if meta.Cluster.Status != string(constants.ClusterRunning) {
		errMsg := fmt.Sprintf("cannot upgrade cluster [%s] under status [%s]", meta.Cluster.Name, meta.Cluster.Status)
		framework.LogWithContext(ctx).Error(errMsg)
		err = errors.NewError(errors.TIEM_TASK_CONFLICT, errMsg)
		return &cluster.ClusterUpgradeResp{}, err
	}

	data := map[string]interface{}{
		ContextClusterMeta:    meta,
		ContextUpgradeVersion: req.TargetVersion,
		ContextUpgradeWay:     req.UpgradeWay,
	}
	var flowID string
	if req.UpgradeWay == cluster.UpgradeWayOnline {
		flowID, err = asyncMaintenance(ctx, meta, constants.ClusterMaintenanceUpgrading, onlineInPlaceUpgradeClusterFlow.FlowName, data)
	} else {
		flowID, err = asyncMaintenance(ctx, meta, constants.ClusterMaintenanceUpgrading, offlineInPlaceUpgradeClusterFlow.FlowName, data)
	}

	resp.WorkFlowID = flowID
	return
}
