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

package parameter

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// QueryParams query params of a cluster
// @Summary query params of a cluster
// @Description query params of a cluster
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query cluster.QueryClusterParametersReq false "page" default(1)
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.ResultWithPage{data=cluster.QueryClusterParametersResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [get]
func QueryParams(c *gin.Context) {
	var req cluster.QueryClusterParametersReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryClusterParameters, &cluster.QueryClusterParametersResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

const paramNameOfClusterId = "clusterId"

// UpdateParams update params
// @Summary submit params
// @Description submit params
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body cluster.UpdateClusterParametersReq true "update params request"
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=cluster.UpdateClusterParametersResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [put]
func UpdateParams(c *gin.Context) {
	var req cluster.UpdateClusterParametersReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.UpdateClusterParametersReq).ClusterID = c.Param(paramNameOfClusterId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateClusterParameters, &cluster.UpdateClusterParametersResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// InspectParams inspect params
// @Summary inspect params
// @Description inspect params
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=[]cluster.InspectClusterParametersResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params/inspect [post]
func InspectParams(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, cluster.InspectClusterParametersReq{
		ClusterID: c.Param(paramNameOfClusterId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.InspectClusterParameters, &cluster.InspectClusterParametersResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
