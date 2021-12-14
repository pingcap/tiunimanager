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
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	ContextClusterMeta   = "ClusterMeta"
	ContextTopology      = "Topology"
	ContextAllocResource = "AllocResource"
	ContextInstanceID    = "InstanceID"
)

var scaleOutDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleOutCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":        {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone": {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":   {"scaleOutCluster", "scaleOutDone", "fail", workflow.PollingNode, scaleOutCluster},
		"scaleOutDone": {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":   {"end", "", "", workflow.SyncFuncNode, clusterEnd},
		"fail":         {"fail", "", "", workflow.SyncFuncNode, clusterFail},
	},
}

var scaleInDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleInCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":       {"scaleInCluster", "scaleInDone", "fail", workflow.PollingNode, scaleInCluster},
		"scaleInDone": {"freeInstanceResource", "freeDone", "fail", workflow.SyncFuncNode, freeInstanceResource},
		"freeDone":    {"end", "", "", workflow.SyncFuncNode, clusterEnd},
		"fail":        {"fail", "", "", workflow.SyncFuncNode, clusterFail},
	},
}

func NewClusterManager() *Manager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
	return &Manager{}
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	operator
// @Parameter	request
// @Return		*cluster.ScaleOutClusterResp
// @Return		error
func (p *Manager) ScaleOut(ctx context.Context, request cluster.ScaleOutClusterReq) (cluster.ScaleOutClusterResp, error) {
	response := cluster.ScaleOutClusterResp{}
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, request.ClusterID)

	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser[%s] meta from db error: %s", request.ClusterID, err.Error())
		return response, err
	}

	// Add instance into cluster topology
	if err = clusterMeta.AddInstances(ctx, request.Compute); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster[%s] topology error: %s", clusterMeta.Cluster.Name, err.Error())
		return response, err
	}

	// Update cluster maintenance status into scale out
	if err = clusterMeta.UpdateClusterMaintenanceStatus(ctx, constants.ClusterMaintenanceScaleOut); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"update cluster[%s] maintenance status error: %s", clusterMeta.Cluster.Name, err.Error())
		return response, err
	}

	// Create the workflow to scale out a cluster
	workflowManager := workflow.GetWorkFlowService()
	flow, err := workflowManager.CreateWorkFlow(ctx, clusterMeta.Cluster.ID, constants.FlowScaleOutCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create workflow error: %s", err.Error())
		return response, err
	}
	workflowManager.AddContext(flow, ContextClusterMeta, clusterMeta)

	if err = workflowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start workflow[%s] error: %s", flow.Flow.ID, err.Error())
		return response, err
	}

	// Handle response
	response.ClusterID = clusterMeta.Cluster.ID
	response.WorkFlowID = flow.Flow.ID

	return response, nil
}

func (manager *Manager) ScaleIn(ctx context.Context, request cluster.ScaleInClusterReq) (cluster.ScaleInClusterResp, error) {
	response := cluster.ScaleInClusterResp{}
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser[%s] meta from db error: %s", request.ClusterID, err.Error())
		return response, err
	}

	// Judge whether the instance exists
	_, err = clusterMeta.GetInstance(ctx, request.InstanceID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] has no instance[%s]", clusterMeta.Cluster.Name, request.InstanceID)
		return response, err
	}

	// Set cluster maintenance status into scale in
	if err = clusterMeta.UpdateClusterMaintenanceStatus(ctx, constants.ClusterMaintenanceScaleIn); err != nil {
		return response, err
	}

	// Start the workflow to scale in a cluster
	workflowManager := workflow.GetWorkFlowService()
	flow, err := workflowManager.CreateWorkFlow(ctx, clusterMeta.Cluster.ID, constants.FlowScaleInCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create workflow error: %s", err.Error())
		return response, err
	}
	workflowManager.AddContext(flow, ContextClusterMeta, clusterMeta)
	workflowManager.AddContext(flow, ContextInstanceID, request.InstanceID)

	if err = workflowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start workflow[%s] error: %s", flow.Flow.ID, err.Error())
		return response, err
	}

	// Handle response
	response.ClusterID = clusterMeta.Cluster.ID
	response.WorkFlowID = flow.Flow.ID
	return response, nil
}

func (manager *Manager) Clone(ctx context.Context, request *cluster.CloneClusterReq) (*cluster.CloneClusterResp, error) {
	return nil, nil
}

type Manager struct{}

var createClusterFlow = &workflow.WorkFlowDefine{
	// define
}

// CreateCluster
// @Description: See createClusterFlow
// @Receiver p
// @Parameter ctx
// @Parameter req
// @return resp
// @return err
func (p *Manager) CreateCluster(ctx context.Context, req cluster.CreateClusterReq) (resp cluster.CreateClusterResp, err error) {
	//meta, err := handler.Create(ctx, buildClusterForCreate(ctx, req.CreateClusterParameter))
	//meta.ScaleOut(ctx, buildInstances(req.ResourceParameter))

	if err != nil {
		return resp, err
	}

	// start flow of creating, and get flowID
	//flow, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.GetID(), createClusterFlow.FlowName)

	if err != nil {
		return resp, err
	}

	//err = workflow.GetWorkFlowService().AsyncStart(ctx, flow)
	if err != nil {
		return resp, err
	}

	//resp.ClusterID = meta.GetID()
	//resp.WorkFlowID = flow.Flow.ID
	return
}

var stopClusterFlow = &workflow.WorkFlowDefine{
	// define
}

func (p *Manager) StopCluster(ctx context.Context, req cluster.StopClusterReq) (resp cluster.StopClusterResp, err error) {
	meta, err := handler.Get(ctx, req.ClusterID)

	if err != nil {
		// wrap error
	}

	meta.TryMaintenance(ctx, constants.ClusterMaintenanceStopping)
	// start flow of stopping, and get flowID
	flowID := ""
	resp.WorkFlowID = flowID
	return
}

func (p *Manager) DeleteCluster(ctx context.Context, req cluster.DeleteClusterReq) (resp cluster.DeleteClusterResp, err error) {
	return
}

func (p *Manager) RestartCluster(ctx context.Context, req cluster.RestartClusterReq) (resp cluster.RestartClusterResp, err error) {
	return
}

func (manager *Manager) GetClusterDashboardInfo(ctx context.Context, request *cluster.GetDashboardInfoReq) (*cluster.GetDashboardInfoResp, error) {
	return GetDashboardInfo(ctx, request)
}
