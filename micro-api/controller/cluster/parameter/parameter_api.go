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
	"net/http"
	"time"

	"github.com/pingcap-inc/tiem/library/util/convert"

	"github.com/pingcap-inc/tiem/micro-api/interceptor"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// QueryParams query params of a cluster
// @Summary query params of a cluster
// @Description query params of a cluster
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query ParamQueryReq false "page" default(1)
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.ResultWithPage{data=[]ListParamsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [get]
func QueryParams(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "QueryParams", int(status.GetCode()))

	var req ParamQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		err = c.Error(err)
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	clusterId := c.Param("clusterId")

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	operator := controller.GetOperator(c)
	reqDTO := &clusterpb.ListClusterParamsRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
		Page:      &clusterpb.PageDTO{Page: int32(req.Page), PageSize: int32(req.PageSize)},
	}
	resp, err := client.ClusterClient.ListClusterParams(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	parameters := make([]ListParamsResp, 0)
	if resp.Params != nil {
		parameters = make([]ListParamsResp, len(resp.Params))
		err := convert.ConvertObj(resp.Params, &parameters)
		if err != nil {
			status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
			c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
			return
		}
	}
	status = resp.RespStatus
	page := &controller.Page{Page: int(resp.Page.Page), PageSize: int(resp.Page.PageSize), Total: int(resp.Page.Total)}

	c.JSON(http.StatusOK, controller.BuildResultWithPage(int(status.Code), status.Message, page, parameters))
}

// UpdateParams update params
// @Summary submit params
// @Description submit params
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body UpdateParamsReq true "update params request"
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=ParamUpdateRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [put]
func UpdateParams(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "UpdateParams", int(status.GetCode()))

	var req UpdateParamsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	clusterId := c.Param("clusterId")

	params := make([]*clusterpb.UpdateClusterParamDTO, len(req.Params))
	err := convert.ConvertObj(req.Params, &params)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	operator := controller.GetOperator(c)
	reqDTO := &clusterpb.UpdateClusterParamsRequest{
		ClusterId:  clusterId,
		Operator:   operator.ConvertToDTO(),
		Params:     params,
		NeedReboot: req.NeedReboot,
	}
	resp, err := client.ClusterClient.UpdateClusterParams(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus
	c.JSON(http.StatusOK, controller.BuildCommonResult(int(status.Code), status.Message, ParamUpdateRsp{
		ClusterId: resp.ClusterId,
		StatusInfo: &controller.StatusInfo{
			StatusCode:      resp.DisplayInfo.StatusCode,
			StatusName:      resp.DisplayInfo.StatusName,
			InProcessFlowId: int(resp.DisplayInfo.InProcessFlowId),
			CreateTime:      time.Unix(resp.DisplayInfo.CreateTime, 0),
			UpdateTime:      time.Unix(resp.DisplayInfo.UpdateTime, 0),
			DeleteTime:      time.Unix(resp.DisplayInfo.DeleteTime, 0),
		},
	}))
}

// InspectParams inspect params
// @Summary inspect params
// @Description inspect params
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=InspectParamsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params/inspect [post]
func InspectParams(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "InspectParams", int(status.GetCode()))

	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)

	reqDTO := &clusterpb.InspectClusterParamsRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
	}
	resp, err := client.ClusterClient.InspectClusterParams(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	params := make([]InspectParamsResp, 0)
	if len(resp.Params) > 0 {
		params = make([]InspectParamsResp, len(resp.Params))
		err := convert.ConvertObj(resp.Params, &params)
		if err != nil {
			status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
			c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
			return
		}
	}
	status = resp.RespStatus
	result := controller.BuildCommonResult(int(status.Code), status.Message, params)
	c.JSON(http.StatusOK, result)
}
