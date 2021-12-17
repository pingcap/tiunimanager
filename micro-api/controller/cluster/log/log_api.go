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

package log

import (
	"github.com/pingcap-inc/tiem/library/client"

	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/micro-api/controller"

	"github.com/gin-gonic/gin"
)

const paramNameOfClusterId = "clusterId"

// QueryClusterLog
// @Summary search cluster log
// @Description search cluster log
// @Tags logs
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param searchReq query cluster.QueryClusterLogReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=cluster.QueryClusterLogResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /logs/tidb/{clusterId} [get]
func QueryClusterLog(c *gin.Context) {
	var req cluster.QueryClusterLogReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.QueryClusterLogReq).ClusterID = c.Param(paramNameOfClusterId)
			return nil
		}); ok {
		// default value valid
		if req.Page <= 1 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryClusterLog, &cluster.QueryClusterLogResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
