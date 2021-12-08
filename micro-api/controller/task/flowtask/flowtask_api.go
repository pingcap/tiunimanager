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
	"github.com/pingcap-inc/tiem/library/client"
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
// @Param queryReq query QueryReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=message.QueryWorkFlowsRep}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/ [get]
func Query(c *gin.Context) {
	var req message.QueryWorkFlowsReq

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)

	if err == nil {
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
// @Param workFlowId path int true "flow work id"
// @Success 200 {object} controller.CommonResult{data=QueryWorkFlowDetailReq}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /workflow/{workFlowId} [get]
func Detail(c *gin.Context) {
	var req message.QueryWorkFlowDetailReq

	requestBody, err := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.QueryWorkFlowDetailReq).WorkFlowID = c.Param("workFlowId")
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.DetailFlow, &message.QueryWorkFlowDetailResp{},
			requestBody,
			controller.DefaultTimeout)
	}
	/*
		//operator := controller.GetOperator(c)
		flowWorkId, err := strconv.Atoi(c.Param("flowWorkId"))

		if err != nil {
			c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
			return
		}
		reqDTO := &clusterpb.DetailFlowRequest{
			FlowId: int64(flowWorkId),
		}

		respDTO, err := client.ClusterClient.DetailFlow(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)

		if err != nil {
			c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
		} else {
			status := respDTO.GetStatus()

			result := controller.BuildCommonResult(int(status.Code), status.Message, ParseFlowWorkDetailInfoFromDTO(respDTO.Flow))

			c.JSON(http.StatusOK, result)
		}

	*/
}
