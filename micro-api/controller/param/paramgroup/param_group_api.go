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
	var queryReq ListParamGroupReq

	if err := c.ShouldBindQuery(&queryReq); err != nil {
		_ = c.Error(err)
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
	resp, err := client.ClusterClient.ListParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus
	page := &controller.Page{Page: int(resp.Page.Page), PageSize: int(resp.Page.PageSize), Total: int(resp.Page.Total)}

	pgs := make([]QueryParamGroupResp, 0)
	if len(resp.ParamGroups) > 0 {
		pgs = make([]QueryParamGroupResp, len(resp.ParamGroups))
		err = convertObj(resp.ParamGroups, &pgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
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
	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	reqDTO := &clusterpb.DetailParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
	}
	resp, err := client.ClusterClient.DetailParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus

	pg := QueryParamGroupResp{}
	err = convertObj(resp.ParamGroup, &pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
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
	var req CreateParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		err = c.Error(err)
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}

	reqDTO := &clusterpb.CreateParamGroupRequest{}
	err := convertObj(req, reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	resp, err := client.ClusterClient.CreateParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus

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
	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	reqDTO := &clusterpb.DeleteParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
	}
	resp, err := client.ClusterClient.DeleteParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus

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
	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	var req UpdateParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}

	reqDTO := &clusterpb.UpdateParamGroupRequest{}
	err = convertObj(req, reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	// set param group id
	reqDTO.ParamGroupId = int64(paramGroupId)
	resp, err := client.ClusterClient.UpdateParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus
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
	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	var req CopyParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	reqDTO := &clusterpb.CopyParamGroupRequest{}
	err = convertObj(req, reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	// set param group id
	reqDTO.ParamGroupId = int64(paramGroupId)
	resp, err := client.ClusterClient.CopyParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus
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
	paramGroupId, err := strconv.Atoi(c.Param("paramGroupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	var req ApplyParamGroupReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}

	reqDTO := &clusterpb.ApplyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		ClusterId:    req.ClusterId,
	}
	resp, err := client.ClusterClient.ApplyParamGroup(framework.NewMicroCtxFromGinCtx(c), reqDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	status := resp.RespStatus
	result := controller.BuildCommonResult(int(status.Code), status.Message, CommonParamGroupResp{ParamGroupId: resp.GetParamGroupId()})
	c.JSON(http.StatusOK, result)
}
