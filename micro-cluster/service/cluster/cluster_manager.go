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

package cluster

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/message/cluster"
)

type ClusterManager struct {

}

func NewClusterManager() *ClusterManager {
	return &ClusterManager{}
}

// ScaleOut
// @Description scale out a cluster
// @Parameter	operator
// @Parameter	request
// @Return		*cluster.ScaleOutClusterResp
// @Return		error
func (manager *ClusterManager) ScaleOut(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.ScaleOutClusterReq) (*cluster.ScaleOutClusterResp, error) {
	// get cluster info from db based by clusterId

	return nil, nil
}

func (manager *ClusterManager) ScaleIn(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.ScaleInClusterReq) (*cluster.ScaleInClusterResp, error) {
	return nil, nil
}

func (manager *ClusterManager) Clone(ctx context.Context, operator *clusterpb.RpcOperator, request *cluster.CloneClusterReq) (*cluster.CloneClusterResp, error) {
	return nil, nil
}




