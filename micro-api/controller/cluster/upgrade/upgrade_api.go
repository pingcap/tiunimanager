/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: upgrade_api
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/7
*******************************************************************************/

package upgrade

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiunimanager/common/client"
	"github.com/pingcap-inc/tiunimanager/message/cluster"
	"github.com/pingcap-inc/tiunimanager/micro-api/controller"
)

const ParamClusterID = "clusterId"

// QueryUpgradePaths query upgrade path for given cluster id
// @Summary query upgrade path for given cluster id
// @Description query upgrade path for given cluster id
// @Tags cluster upgrade
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=cluster.QueryUpgradePathRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/upgrade/path [get]
func QueryUpgradePaths(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &cluster.QueryUpgradePathReq{
		ClusterID: c.Param(ParamClusterID),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProductUpgradePath, &cluster.QueryUpgradePathRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryUpgradeVersionDiffInfo query config diff between current cluster and target upgrade version
// @Summary query config diff between current cluster and target upgrade version
// @Description query config diff between current cluster and target upgrade version
// @Tags cluster upgrade
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param upgradeVersionDiffQuery query cluster.QueryUpgradeVersionDiffInfoReq true "upgrade version diff query"
// @Success 200 {object} controller.CommonResult{data=cluster.QueryUpgradeVersionDiffInfoResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/upgrade/diff [get]
func QueryUpgradeVersionDiffInfo(c *gin.Context) {
	var req cluster.QueryUpgradeVersionDiffInfoReq
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.QueryUpgradeVersionDiffInfoReq).ClusterID = c.Param(ParamClusterID)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryUpgradeVersionDiffInfo, &cluster.QueryUpgradeVersionDiffInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Upgrade a cluster
// @Summary request for upgrade TiDB cluster
// @Description request for upgrade TiDB cluster
// @Tags cluster upgrade
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param upgradeReq body cluster.UpgradeClusterReq true "upgrade request"
// @Success 200 {object} controller.CommonResult{data=cluster.UpgradeClusterResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/upgrade [post]
func Upgrade(c *gin.Context) {
	// handle upgrade request and call rpc method
	if body, ok := controller.HandleJsonRequestFromBody(c,
		&cluster.UpgradeClusterReq{},
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.UpgradeClusterReq).ClusterID = c.Param(ParamClusterID)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpgradeCluster,
			&cluster.UpgradeClusterResp{}, body, controller.DefaultTimeout)
	}
}
