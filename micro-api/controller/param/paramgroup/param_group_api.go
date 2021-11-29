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

/*******************************************************************************
 * @File: param_group_api.go
 * @Description: param group api
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/17 10:24
*******************************************************************************/

package paramgroup

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/library/util/convert"

	"github.com/pingcap-inc/tiem/micro-api/interceptor"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/gin-gonic/gin/binding"

	"github.com/pingcap-inc/tiem/micro-api/controller"

	"github.com/gin-gonic/gin"
)

// Query query param group
// @Summary query param group
// @Description query param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query ListParamGroupReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=[]QueryParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/ [get]
func Query(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "QueryParamGroup", int(status.GetCode()))

	var queryReq ListParamGroupReq

	if err := c.ShouldBindQuery(&queryReq); err != nil {
		err = c.Error(err)
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	if queryReq.Page <= 0 {
		queryReq.Page = 1
	}
	if queryReq.PageSize <= 0 {
		queryReq.PageSize = 10
	}
	reqDTO := &clusterpb.ListParamGroupRequest{
		Name:       queryReq.Name,
		Version:    queryReq.Version,
		Spec:       queryReq.Spec,
		HasDetail:  queryReq.HasDetail,
		DbType:     queryReq.DbType,
		HasDefault: queryReq.HasDefault,
		Page:       &clusterpb.PageDTO{Page: int32(queryReq.Page), PageSize: int32(queryReq.PageSize)},
	}
	resp, err := client.ClusterClient.ListParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus
	page := &controller.Page{Page: int(resp.Page.Page), PageSize: int(resp.Page.PageSize), Total: int(resp.Page.Total)}

	pgs := make([]QueryParamGroupResp, 0)
	if len(resp.ParamGroups) > 0 {
		pgs = make([]QueryParamGroupResp, len(resp.ParamGroups))
		err = convert.ConvertObj(resp.ParamGroups, &pgs)
		if err != nil {
			status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
			c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
			return
		}
	}
	result := controller.BuildResultWithPage(int(status.Code), status.Message, page, pgs)
	c.JSON(http.StatusOK, result)
}

// Detail show details of a param group
// @Summary show details of a param group
// @Description show details of a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "param group id"
// @Success 200 {object} controller.CommonResult{data=QueryParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [get]
func Detail(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "DetailParamGroup", int(status.GetCode()))

	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	reqDTO := &clusterpb.DetailParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
	}
	resp, err := client.ClusterClient.DetailParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus

	pg := QueryParamGroupResp{}
	err = convert.ConvertObj(resp.ParamGroup, &pg)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	result := controller.BuildCommonResult(int(status.Code), status.Message, pg)
	c.JSON(http.StatusOK, result)
}

// Create create a param group
// @Summary create a param group
// @Description create a param group
// @Tags param group
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param createReq body CreateParamGroupReq true "create request"
// @Success 200 {object} controller.CommonResult{data=CommonParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/ [post]
func Create(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "CreateParamGroup", int(status.GetCode()))

	var req CreateParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		err = c.Error(err)
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	reqDTO := &clusterpb.CreateParamGroupRequest{}
	err := convert.ConvertObj(req, reqDTO)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	resp, err := client.ClusterClient.CreateParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus

	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.ParamGroupId})
	c.JSON(http.StatusOK, result)
}

// Delete delete a param group
// @Summary delete a param group
// @Description delete a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "param group id"
// @Success 200 {object} controller.CommonResult{data=CommonParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [delete]
func Delete(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "DeleteParamGroup", int(status.GetCode()))

	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	reqDTO := &clusterpb.DeleteParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
	}
	resp, err := client.ClusterClient.DeleteParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus

	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.ParamGroupId})
	c.JSON(http.StatusOK, result)
}

// Update update a param group
// @Summary update a param group
// @Description update a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body UpdateParamGroupReq true "update param group request"
// @Success 200 {object} controller.CommonResult{data=CommonParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [put]
func Update(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "UpdateParamGroup", int(status.GetCode()))

	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	var req UpdateParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	reqDTO := &clusterpb.UpdateParamGroupRequest{}
	err = convert.ConvertObj(req, reqDTO)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	// set param group id
	reqDTO.ParamGroupId = int64(paramGroupId)
	resp, err := client.ClusterClient.UpdateParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus
	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.GetParamGroupId()})
	c.JSON(http.StatusOK, result)
}

// Copy copy a param group
// @Summary copy a param group
// @Description copy a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param copyReq body CopyParamGroupReq true "copy param group request"
// @Success 200 {object} controller.CommonResult{data=CommonParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/copy [post]
func Copy(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "CopyParamGroup", int(status.GetCode()))

	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	var req CopyParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	reqDTO := &clusterpb.CopyParamGroupRequest{}
	err = convert.ConvertObj(req, reqDTO)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	// set param group id
	reqDTO.ParamGroupId = int64(paramGroupId)
	resp, err := client.ClusterClient.CopyParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus
	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.GetParamGroupId()})
	c.JSON(http.StatusOK, result)
}

// Apply apply a param group
// @Summary apply a param group
// @Description apply a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param applyReq body ApplyParamGroupReq true "apply param group request"
// @Success 200 {object} controller.CommonResult{data=ApplyParamGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/apply [post]
func Apply(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "ApplyParamGroup", int(status.GetCode()))

	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	var req ApplyParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}

	operator := controller.GetOperator(c)
	reqDTO := &clusterpb.ApplyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		ClusterId:    req.ClusterId,
		Operator:     operator.ConvertToDTO(),
		NeedReboot:   req.NeedReboot,
	}
	resp, err := client.ClusterClient.ApplyParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(status.Code), status.Message))
		return
	}
	status = resp.RespStatus
	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.GetParamGroupId()})
	c.JSON(http.StatusOK, result)
}
