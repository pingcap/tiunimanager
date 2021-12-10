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
)

type ClusterManager struct{}

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

func NewClusterManager() *ClusterManager {
	workflowManager := workflow.GetWorkFlowManager()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)
	return &ClusterManager{}
}

func (manager *ClusterManager) load(ctx context.Context, clusterID string) (*handler.ClusterMeta, error) {
	// TODO: read db
	return nil, nil
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	operator
// @Parameter	request
// @Return		*cluster.ScaleOutClusterResp
// @Return		error
func (manager *ClusterManager) ScaleOut(ctx context.Context, request *cluster.ScaleOutClusterReq) (*cluster.ScaleOutClusterResp, error) {
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := manager.load(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluser[%s] meta from db error: %s", request.ClusterID, err.Error())
		return nil, err
	}

	// Add instance into cluster topology
	if err = clusterMeta.AddInstances(ctx, request.Compute); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"add instances into cluster[%s] topology error: %s", clusterMeta.Cluster.Name, err.Error())
		return nil, err
	}

	// Update cluster maintenance status into scale out
	if err = clusterMeta.UpdateClusterMaintenanceStatus(ctx, constants.ClusterMaintenanceScaleOut); err != nil {
		framework.LogWithContext(ctx).Errorf(
			"update cluster[%s] maintenance status error: %s", clusterMeta.Cluster.Name, err.Error())
		return nil, err
	}

	// Create the workflow to scale out a cluster
	workflowManager := workflow.GetWorkFlowManager()
	flow, err := workflowManager.CreateWorkFlow(ctx, clusterMeta.Cluster.ID, constants.FlowScaleOutCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create workflow error: %s", err.Error())
		return nil, err
	}
	workflowManager.AddContext(flow, ContextClusterMeta, clusterMeta)

	if err = workflowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start workflow[%s] error: %s", flow.Flow.ID, err.Error())
		return nil, err
	}

	// Handle response
	response := &cluster.ScaleOutClusterResp{}
	response.ClusterID = clusterMeta.Cluster.ID
	response.WorkFlowID = flow.Flow.ID

	return response, nil
}

func (manager *ClusterManager) ScaleIn(ctx context.Context, request *cluster.ScaleInClusterReq) (*cluster.ScaleInClusterResp, error) {
	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := manager.load(ctx, request.ClusterID)
	if err != nil {
		return nil, err
	}

	// Set cluster maintenance status into scale in
	if err = clusterMeta.UpdateClusterMaintenanceStatus(ctx, constants.ClusterMaintenanceScaleIn); err != nil {
		return nil, err
	}

	// Start the workflow to scale in a cluster

	return nil, nil
}

func (manager *ClusterManager) Clone(ctx context.Context, request *cluster.CloneClusterReq) (*cluster.CloneClusterResp, error) {
	return nil, nil
}

/*
type Manager struct {}

func (p *Manager) CreateCluster(ctx context.Context, req cluster.CreateClusterReq) (resp cluster.CreateClusterResp, err error) {
	meta, err := handler.Create(ctx, buildClusterForCreate(ctx, req.CreateClusterParameter))
	meta.ScaleOut(ctx, buildInstances(req.ResourceParameter))

	if err != nil {
		// wrap error
		return resp, err
	}

	// start flow of creating, and get flowID
	flowID := ""
	resp.WorkFlowID = flowID
	return
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

func buildClusterForCreate(ctx context.Context, p structs.CreateClusterParameter) management.Cluster {
	return management.Cluster{
		Entity: common.Entity{
			// todo replace
			TenantId: ctx.Value("TENANT_ID").(string),
		},
		Name:            p.Name,
		DBUser:          p.DBUser,
		DBPassword:      p.DBPassword,
		Type:            p.Type,
		Version:         p.Version,
		TLS:             p.TLS,
		Tags:            p.Tags,
		// todo replace
		OwnerId:          ctx.Value("OPERATOR_ID").(string),
		Exclusive:       p.Exclusive,
		Region:          p.Region,
		CpuArchitecture: constants.ArchType(p.CpuArchitecture),
	}
}

func buildInstances(p []structs.ClusterResourceParameter) []*management.ClusterInstance {
	// todo
	return nil
}
*/
