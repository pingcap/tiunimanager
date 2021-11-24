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
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	groups := make([]QueryParamGroupResp, 1)
	groups[0] = QueryParamGroupResp{}
	result := controller.BuildResultWithPage(0, "OK", &controller.Page{Page: 1, PageSize: 1, Total: 1}, groups)
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
	//param := c.Param("paramGroupId")
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", QueryParamGroupResp{})
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
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", CommonParamGroupResp{})
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
	//param := c.Param("paramGroupId")
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", CommonParamGroupResp{})
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
	var req UpdateParamGroupReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", CommonParamGroupResp{})
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
	var req CopyParamGroupReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", CommonParamGroupResp{})
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
	var req ApplyParamGroupReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	// todo: invoke micro-cluster service

	result := controller.BuildCommonResult(0, "OK", ApplyParamGroupResp{})
	c.JSON(http.StatusOK, result)
}
