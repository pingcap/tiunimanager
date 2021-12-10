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

package parametergroup

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Query query param group
// @Summary query param group
// @Description query param group
// @Tags param group
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

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryParameterGroup, make([]message.QueryParameterGroupResp, 0),
			requestBody,
			controller.DefaultTimeout)
	}
}

const paramNameOfParameterGroupId = "paramGroupId"

// Detail show details of a param group
// @Summary show details of a param group
// @Description show details of a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "param group id"
// @Success 200 {object} controller.CommonResult{data=message.DetailParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [get]
func Detail(c *gin.Context) {
	requestBody, err := controller.HandleJsonRequestWithBuiltReq(c, &message.DetailParameterGroupReq{
		ID: c.Param(paramNameOfParameterGroupId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.DetailParameterGroup, &message.DetailParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Create create a param group
// @Summary create a param group
// @Description create a param group
// @Tags param group
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

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateParameterGroup, &message.CreateParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Delete delete a param group
// @Summary delete a param group
// @Description delete a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param paramGroupId path string true "param group id"
// @Success 200 {object} controller.CommonResult{data=message.DeleteParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [delete]
func Delete(c *gin.Context) {
	requestBody, err := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteParameterGroupReq{
		ID: c.Param(paramNameOfParameterGroupId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteParameterGroup, &message.DeleteParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Update update a param group
// @Summary update a param group
// @Description update a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body message.UpdateParameterGroupReq true "update param group request"
// @Success 200 {object} controller.CommonResult{data=message.UpdateParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId} [put]
func Update(c *gin.Context) {
	var req message.UpdateParameterGroupReq

	requestBody, err := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.UpdateParameterGroupReq).ID = c.Param(paramNameOfParameterGroupId)
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateParameterGroup, &message.UpdateParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Copy copy a param group
// @Summary copy a param group
// @Description copy a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param copyReq body message.CopyParameterGroupReq true "copy param group request"
// @Success 200 {object} controller.CommonResult{data=message.CopyParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/copy [post]
func Copy(c *gin.Context) {
	var req message.CopyParameterGroupReq

	requestBody, err := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.CopyParameterGroupReq).ID = c.Param(paramNameOfParameterGroupId)
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CopyParameterGroup, &message.CopyParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Apply apply a param group
// @Summary apply a param group
// @Description apply a param group
// @Tags param group
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param applyReq body message.ApplyParameterGroupReq true "apply param group request"
// @Success 200 {object} controller.CommonResult{data=message.CopyParameterGroupResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /param-groups/{paramGroupId}/apply [post]
func Apply(c *gin.Context) {
	var req message.ApplyParameterGroupReq

	requestBody, err := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.ApplyParameterGroupReq).ID = c.Param(paramNameOfParameterGroupId)
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.ApplyParameterGroup, &message.CopyParameterGroupResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
