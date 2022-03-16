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

package platform

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Check platform check
// @Summary platform check
// @Description platform check
// @Tags platform
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param checkPlatformReq query message.CheckPlatformReq false "check platform"
// @Success 200 {object} controller.ResultWithPage{data=message.CheckPlatformRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /platform/check [post]
func Check(c *gin.Context) {
	var request message.CheckPlatformReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &request); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CheckPlatform, &message.CheckPlatformRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// CheckCluster platform check cluster
// @Summary platform check cluster
// @Description platform check cluster
// @Tags platform
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param checkClusterReq query message.CheckClusterReq false "check cluster"
// @Success 200 {object} controller.ResultWithPage{data=message.CheckClusterRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /platform/check/{clusterId} [post]
func CheckCluster(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.CheckClusterReq{
		ClusterID: c.Param("clusterId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CheckCluster, &message.CheckClusterRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetCheckReport get check report based on checkId
// @Summary get check report based on checkId
// @Description get check report based on checkId
// @Tags platform
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param getCheckReportReq query message.GetCheckReportReq false "get check report"
// @Success 200 {object} controller.ResultWithPage{data=message.GetCheckReportRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /platform/report/{checkId} [get]
func GetCheckReport(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.GetCheckReportReq{
		ID: c.Param("checkId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetCheckReport, &message.GetCheckReportRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryCheckReports query check reports
// @Summary query check reports
// @Description get query check reports
// @Tags platform
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryCheckReportsReq query message.QueryCheckReportsReq false "query check reports"
// @Success 200 {object} controller.ResultWithPage{data=message.QueryCheckReportsRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /platform/reports [get]
func QueryCheckReports(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryCheckReportsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryCheckReports, &message.QueryCheckReportsRsp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
