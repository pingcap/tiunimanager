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
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
)

type ClusterManager struct {}

func NewClusterManager() *ClusterManager {
	return &ClusterManager{}
}


func (manager *ClusterManager) load(ctx context.Context, clusterID string) (*handler.ClusterMeta, error) {
	return nil, nil
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	operator
// @Parameter	request
// @Return		*cluster.ScaleOutClusterResp
// @Return		error
func (manager *ClusterManager) ScaleOut(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.ScaleOutClusterReq) (*cluster.ScaleOutClusterResp, error) {
	// Get cluster info and topology from db based by clusterId
	clusterMeta, err := manager.load(ctx, request.ClusterID)
	if err != nil {
		return nil, err
	}

	// Add instance into cluster topology
	clusterMeta.AddInstances(ctx, request.Compute)

	// Start the workflow to scale a cluster
	//flow := workflow.GetWorkFlowManager()
	//flow.RegisterWorkFlow(ctx, )

	// Handle response
	response := &cluster.ScaleOutClusterResp{}
	response.ClusterID = clusterMeta.GetClusterID()

	return response, nil
}

func (manager *ClusterManager) ScaleIn(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.ScaleInClusterReq) (*cluster.ScaleInClusterResp, error) {
	return nil, nil
}

func (manager *ClusterManager) Clone(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.CloneClusterReq) (*cluster.CloneClusterResp, error) {
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