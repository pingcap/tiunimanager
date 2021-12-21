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

package identification

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Login login
// @Summary login
// @Description login
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param loginInfo body message.LoginReq true "login info"
// @Header 200 {string} Token "DUISAFNDHIGADS"
// @Success 200 {object} controller.CommonResult{data=message.LoginResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/login [post]
func Login(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.LoginReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.Login, &message.LoginResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Logout logout
// @Summary logout
// @Description logout
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param logoutInfo body message.LogoutReq true "login info"
// @Security ApiKeyAuth
// @Success 200 {object} controller.CommonResult{data=message.LogoutResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.LogoutReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.Logout, &message.LogoutResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
