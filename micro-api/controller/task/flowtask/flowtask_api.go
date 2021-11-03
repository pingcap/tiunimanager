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
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
	"strconv"
)

// Query query flow works
// @Summary query flow works
// @Description query flow works
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query QueryReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=[]FlowWorkDisplayInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /flowworks [get]
func Query(c *gin.Context) {
	var queryReq QueryReq
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		_ = c.Error(err)
		return
	}

	reqDTO := &clusterpb.ListFlowsRequest{
		BizId:   queryReq.ClusterId,
		Keyword: queryReq.Keyword,
		Status:  int64(queryReq.Status),
		Page:    queryReq.PageRequest.ConvertToDTO(),
	}

	respDTO, err := client.ClusterClient.ListFlows(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetStatus()

		flows := make([]FlowWorkDisplayInfo, len(respDTO.Flows))

		for i, v := range respDTO.Flows {
			flows[i] = ParseFlowFromDTO(v)
		}

		result := controller.BuildResultWithPage(int(status.Code), status.Message, controller.ParsePageFromDTO(respDTO.Page), flows)

		c.JSON(http.StatusOK, result)
	}
}

// Detail show details of a flow work
// @Summary show details of a flow work
// @Description show details of a flow work
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param flowWorkId path int true "flow work id"
// @Success 200 {object} controller.CommonResult{data=FlowWorkDetailInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /flowworks/{flowWorkId} [get]
func Detail(c *gin.Context) {
	//operator := controller.GetOperator(c)
	flowWorkId, err := strconv.Atoi(c.Param("flowWorkId"))

	reqDTO := &clusterpb.DetailFlowRequest {
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
}
