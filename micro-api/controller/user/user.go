/*
 * Copyright (c)  2022 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: user.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/7
*******************************************************************************/

package user

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// CreateUser create user interface
// @Summary created  user
// @Description created user
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.CreateUserReq true "create user request parameter"
// @Success 200 {object} controller.CommonResult{data=message.CreateUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user [post]
func CreateUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.CreateUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateUser, &message.CreateUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteUser delete user interface
// @Summary delete user
// @Description delete user
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.DeleteUserReq true "delete user request parameter"
// @Success 200 {object} controller.CommonResult{data=message.DeleteUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user [delete]
func DeleteUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteUser, &message.DeleteUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetUser get user profile interface
// @Summary get user profile
// @Description get user profile
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.GetUserReq true "get user profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.GetUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user [get]
func GetUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.GetUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetUser, &message.GetUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryUsers query all user profile interface
// @Summary queries all user profile
// @Description query all user profile
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.QueryUserReq true "query user profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user [get]
func QueryUsers(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.QueryUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryUsers, &message.QueryUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UpdateUserProfile update user profile interface
// @Summary update user profile
// @Description update user profile
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.UpdateUserProfileReq true "query user profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateUserProfileResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user [post]
func UpdateUserProfile(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.UpdateUserProfileReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateUserProfile, &message.UpdateUserProfileResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
