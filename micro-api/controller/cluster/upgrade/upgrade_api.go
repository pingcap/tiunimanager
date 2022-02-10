/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

const paramNameOfClusterID = "clusterId"

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
// @Router /clusters/:clusterId/upgrade/path [get]
func QueryUpgradePaths(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &cluster.QueryUpgradePathReq{
		ClusterID: c.Param(paramNameOfClusterID),
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
// @Router /clusters/:clusterId/upgrade/diff [get]
func QueryUpgradeVersionDiffInfo(c *gin.Context) {
	var req cluster.QueryUpgradeVersionDiffInfoReq
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.QueryUpgradeVersionDiffInfoReq).ClusterID = c.Param(paramNameOfClusterID)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryUpgradeVersionDiffInfo, &cluster.QueryUpgradeVersionDiffInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// ClusterUpgrade request for upgrade
// @Summary request for upgrade
// @Description request for upgrade
// @Tags cluster upgrade
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param upgradeReq body cluster.ClusterUpgradeReq true "upgrade request"
// @Success 200 {object} controller.CommonResult{data=cluster.ClusterUpgradeResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/:clusterId/upgrade [post]
func ClusterUpgrade(c *gin.Context) {
	var req cluster.ClusterUpgradeReq
	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.ClusterUpgradeReq).ClusterID = c.Param(paramNameOfClusterID)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ClusterUpgrade, &cluster.ClusterUpgradeResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
