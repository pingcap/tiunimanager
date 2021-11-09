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

package account

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
)

// Profile user profile
// @Summary user profile
// @Description profile
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Success 200 {object} controller.CommonResult{data=UserProfile}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/profile [get]
func Profile(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "Profile", int(status.GetCode()))

	v, _ := c.Get(interceptor.VisitorIdentityKey)

	visitor, _ := v.(*interceptor.VisitorIdentity)
	c.JSON(http.StatusOK, controller.Success(UserProfile{UserName: visitor.AccountName, TenantId: visitor.TenantId}))
}
