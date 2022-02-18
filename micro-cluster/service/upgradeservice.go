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

/*******************************************************************************
 * @File: upgradeservice
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/20
*******************************************************************************/

package service

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/metrics"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/proto/clusterservices"

	"github.com/pingcap-inc/tiem/message/cluster"
)

func (handler *ClusterServiceHandler) CreateProductUpgradePath(context.Context, *clusterservices.RpcRequest, *clusterservices.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) DeleteProductUpgradePath(context.Context, *clusterservices.RpcRequest, *clusterservices.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) UpdateProductUpgradePath(context.Context, *clusterservices.RpcRequest, *clusterservices.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) QueryProductUpgradePath(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProductUpgradePath", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryProductUpgradePath", resp)

	request := &cluster.QueryUpgradePathReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.clusterManager.QueryProductUpdatePath(ctx, request.ClusterID)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryUpgradeVersionDiffInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryUpgradeVersionDiffInfo", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryUpgradeVersionDiffInfo", resp)

	request := &cluster.QueryUpgradeVersionDiffInfoReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.clusterManager.QueryUpgradeVersionDiffInfo(ctx, request.ClusterID, request.TargetVersion)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpgradeCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpgradeCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "UpgradeCluster", resp)

	request := &cluster.UpgradeClusterReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.clusterManager.InPlaceUpgradeCluster(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
