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

package platformdignose

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiunimanager/common/client"
	"github.com/pingcap-inc/tiunimanager/message"
	"github.com/pingcap-inc/tiunimanager/micro-api/controller"
)

// QueryPlatformLog
// @Summary query platform log
// @Description query platform log
// @Tags platform log
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param searchReq query message.QueryPlatformLogReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=message.QueryPlatformLogResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /platform/log [get]
func QueryPlatformLog(c *gin.Context) {
	var req message.QueryPlatformLogReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req); ok {
		// default value valid
		if req.Page <= 1 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryPlatformLog, &message.QueryPlatformLogResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
