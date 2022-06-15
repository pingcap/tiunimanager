/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package platformconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// UpdateSystemConfig
// @Summary update system config
// @Description UpdateSystemConfig
// @Tags UpdateSystemConfig
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param systemConfig body message.UpdateSystemConfigReq true "system config for update"
// @Success 200 {object} controller.CommonResult{data=message.UpdateSystemConfigResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /config/update [post]
func UpdateSystemConfig(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.UpdateSystemConfigReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateSystemConfig, &message.UpdateSystemConfigResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetSystemConfig
// @Summary get system config
// @Description get system config
// @Tags get system config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param getSystemConfigReq query message.GetSystemConfigReq true "system config key"
// @Success 200 {object} controller.CommonResult{data=message.GetSystemConfigResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /config/ [get]
func GetSystemConfig(c *gin.Context) {
	var request message.GetSystemConfigReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &request); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetSystemConfig, &message.GetSystemConfigResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
