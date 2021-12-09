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
	"github.com/pingcap-inc/tiem/apimodels/cluster/upgrade"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

const paramNameOfClusterID = "clusterId"

// QueryUpgradePaths query upgrade path for given cluster
// @Summary query upgrade path for given cluster
// @Description query upgrade path for given cluster
// @Tags upgrade
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=upgrade.QueryUpgradePathRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/:clusterId/upgrade/path [get]
func QueryUpgradePaths(c *gin.Context) {
	requestBody, err := controller.HandleJsonRequestWithBuiltReq(c, &upgrade.QueryUpgradePathReq{
		ClusterID: c.Param(paramNameOfClusterID),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProductUpgradePath, &upgrade.QueryUpgradePathRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryUpgradeVersionDiffInfo query upgrade params diff between current cluster and dst version
// @Summary query upgrade params diff between current cluster and dst version
// @Description query upgrade params diff between current cluster and dst version
// @Tags upgrade
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param task body upgrade.QueryUpgradeVersionDiffInfoReq true "upgrade version diff info"
// @Success 200 {object} controller.CommonResult{data=upgrade.QueryUpgradeVersionDiffInfoRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/:clusterId/upgrade/diff [get]
func QueryUpgradeVersionDiffInfo(c *gin.Context) {
	var req upgrade.QueryUpgradeVersionDiffInfoReq

	requestBody, err := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*upgrade.QueryUpgradeVersionDiffInfoReq).ClusterID = c.Param(paramNameOfClusterID)
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryUpgradeVersionDiffInfo, &upgrade.QueryUpgradeVersionDiffInfoRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
