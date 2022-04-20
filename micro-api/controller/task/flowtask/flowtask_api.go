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

package flowtask

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Query query flow works
// @Summary query flow works
// @Description query flow works
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query message.QueryWorkFlowsReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=message.QueryWorkFlowsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/ [get]
func Query(c *gin.Context) {
	var request message.QueryWorkFlowsReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &request); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ListFlows, &message.QueryWorkFlowsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Detail show details of a flow work
// @Summary show details of a flow work
// @Description show details of a flow work
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param workFlowId path string true "flow work id"
// @Success 200 {object} controller.CommonResult{data=message.QueryWorkFlowDetailResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/{workFlowId} [get]
func Detail(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.QueryWorkFlowDetailReq{
		WorkFlowID: c.Param("workFlowId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DetailFlow, &message.QueryWorkFlowDetailResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Start
// @Summary start workflow
// @Description start workflow
// @Tags start workflow
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param startReq body message.StartWorkFlowReq true "start workflow"
// @Success 200 {object} controller.CommonResult{data=message.StartWorkFlowResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/start [post]
func Start(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.StartWorkFlowReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.StartFlow, &message.StartWorkFlowResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Stop
// @Summary stop workflow
// @Description stop workflow
// @Tags stop workflow
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param stopReq body message.StopWorkFlowReq true "stop workflow"
// @Success 200 {object} controller.CommonResult{data=message.StopWorkFlowResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/stop [post]
func Stop(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.StopWorkFlowReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.StopFlow, &message.StopWorkFlowResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
