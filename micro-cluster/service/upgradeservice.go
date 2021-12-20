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

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/message/cluster"
)

func (handler *ClusterServiceHandler) CreateProductUpgradePath(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) DeleteProductUpgradePath(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) UpdateProductUpgradePath(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) QueryProductUpgradePath(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	request := cluster.QueryUpgradePathReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterManager.QueryProductUpdatePath(ctx, request.ClusterID)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryUpgradeVersionDiffInfo(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	request := cluster.QueryUpgradeVersionDiffInfoReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterManager.QueryUpgradeVersionDiffInfo(ctx, request.ClusterID, request.Version)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ClusterUpgrade(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	request := cluster.ClusterUpgradeReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterManager.InPlaceUpgradeCluster(ctx, &request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
