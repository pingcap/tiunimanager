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
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/workflow"
)

type Manager struct {}

var createClusterFlow = &workflow.WorkFlowDefine {
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
	meta, err := handler.Create(ctx, buildClusterForCreate(ctx, req.CreateClusterParameter))
	meta.ScaleOut(ctx, buildInstances(req.ResourceParameter))

	if err != nil {
		return resp, err
	}

	// start flow of creating, and get flowID
	flow, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.GetID(), createClusterFlow.FlowName)

	if err != nil {
		return resp, err
	}

	err = workflow.GetWorkFlowService().AsyncStart(ctx, flow)
	if err != nil {
		return resp, err
	}

	resp.ClusterID = meta.GetID()
	resp.WorkFlowID = flow.Flow.ID
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

func Init() {
	f := workflow.GetWorkFlowService()
	f.RegisterWorkFlow(context.TODO(), createClusterFlow.FlowName, createClusterFlow)
}

func buildClusterForCreate(ctx context.Context, p structs.CreateClusterParameter) management.Cluster {
	return management.Cluster{
		Entity: common.Entity{
			// todo get from context
			TenantId: "111",
		},
		Name:            p.Name,
		DBUser:          p.DBUser,
		DBPassword:      p.DBPassword,
		Type:            p.Type,
		Version:         p.Version,
		TLS:             p.TLS,
		Tags:            p.Tags,
		// todo get from context
		OwnerId:         "111",
		Exclusive:       p.Exclusive,
		Region:          p.Region,
		CpuArchitecture: constants.ArchType(p.CpuArchitecture),
	}
}

func buildInstances(p []structs.ClusterResourceParameter) []*management.ClusterInstance {
	// todo
	return nil
}