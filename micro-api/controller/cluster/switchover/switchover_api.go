/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package switchover

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// Switchover master/slave switchover
// @Summary master/slave switchover
// @Description master/slave switchover
// @Tags switchover
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param masterSlaveClusterSwitchoverReq body cluster.MasterSlaveClusterSwitchoverReq true "switchover request"
// @Success 200 {object} controller.CommonResult{data=cluster.MasterSlaveClusterSwitchoverResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/switchover [post]
func Switchover(c *gin.Context) {
	var req cluster.MasterSlaveClusterSwitchoverReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.MasterSlaveSwitchover, &cluster.MasterSlaveClusterSwitchoverResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
