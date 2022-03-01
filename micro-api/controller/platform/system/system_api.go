/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/*******************************************************************************
 * @File: system_api.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/22
*******************************************************************************/

package system

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// GetSystemInfo get system info
// @Summary get system info
// @Description get system info
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param req query message.GetSystemInfoReq true "request"
// @Success 200 {object} controller.CommonResult{data=message.GetSystemInfoResp}
// @Failure 500 {object} controller.CommonResult
// @Router /system/info [get]
func GetSystemInfo(c *gin.Context) {
	var req message.GetSystemInfoReq
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetSystemInfo, &message.GetSystemInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

