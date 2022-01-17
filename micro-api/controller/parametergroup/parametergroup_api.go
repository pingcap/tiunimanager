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
 * @Description: parameter group api
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/17 10:24
*******************************************************************************/

package parametergroup

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Query query parameter group
// @Summary query parameter group
// @Description query parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query message.QueryParameterGroupReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=[]message.QueryParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/ [get]
func Query(c *gin.Context) {
	var req message.QueryParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req); ok {
		resp := make([]message.QueryParameterGroupResp, 0)
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryParameterGroup, &resp,
			requestBody,
			controller.DefaultTimeout)
	}
}

const paramNameOfParameterGroupId = "paramGroupId"

// Detail show details of a parameter group
// @Summary show details of a parameter group
// @Description show details of a parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "parameter group id"
// @Success 200 {object} controller.CommonResult{data=message.DetailParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [get]
func Detail(c *gin.Context) {
	var req message.DetailParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.DetailParameterGroupReq).ParamGroupID = c.Param(paramNameOfParameterGroupId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DetailParameterGroup, &message.DetailParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Create create a parameter group
// @Summary create a parameter group
// @Description create a parameter group
// @Tags parameter group
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param createReq body message.CreateParameterGroupReq true "create request"
// @Success 200 {object} controller.CommonResult{data=message.CreateParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/ [post]
func Create(c *gin.Context) {
	var req message.CreateParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateParameterGroup, &message.CreateParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Delete delete a parameter group
// @Summary delete a parameter group
// @Description delete a parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "parameter group id"
// @Success 200 {object} controller.CommonResult{data=message.DeleteParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [delete]
func Delete(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteParameterGroupReq{
		ParamGroupID: c.Param(paramNameOfParameterGroupId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteParameterGroup, &message.DeleteParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Update update a parameter group
// @Summary update a parameter group
// @Description update a parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "parameter group id"
// @Param updateReq body message.UpdateParameterGroupReq true "update parameter group request"
// @Success 200 {object} controller.CommonResult{data=message.UpdateParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [put]
func Update(c *gin.Context) {
	var req message.UpdateParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.UpdateParameterGroupReq).ParamGroupID = c.Param(paramNameOfParameterGroupId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateParameterGroup, &message.UpdateParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Copy copy a parameter group
// @Summary copy a parameter group
// @Description copy a parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "parameter group id"
// @Param copyReq body message.CopyParameterGroupReq true "copy parameter group request"
// @Success 200 {object} controller.CommonResult{data=message.CopyParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/copy [post]
func Copy(c *gin.Context) {
	var req message.CopyParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.CopyParameterGroupReq).ParamGroupID = c.Param(paramNameOfParameterGroupId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CopyParameterGroup, &message.CopyParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Apply apply a parameter group
// @Summary apply a parameter group
// @Description apply a parameter group
// @Tags parameter group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "parameter group id"
// @Param applyReq body message.ApplyParameterGroupReq true "apply parameter group request"
// @Success 200 {object} controller.CommonResult{data=message.ApplyParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/apply [post]
func Apply(c *gin.Context) {
	var req message.ApplyParameterGroupReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.ApplyParameterGroupReq).ParamGroupId = c.Param(paramNameOfParameterGroupId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ApplyParameterGroup, &message.ApplyParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
