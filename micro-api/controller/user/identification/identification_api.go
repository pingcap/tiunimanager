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
	"net/http"
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	utils "github.com/pingcap-inc/tiem/library/util/stringutil"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
)

// Login login
// @Summary login
// @Description login
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param loginInfo body LoginInfo true "login info"
// @Header 200 {string} Token "DUISAFNDHIGADS"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/login [post]
func Login(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "Login", int(status.GetCode()))

	var req LoginInfo

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	loginReq := clusterpb.LoginRequest{AccountName: req.UserName, Password: req.UserPassword}
	result, err := client.ClusterClient.Login(framework.NewMicroCtxFromGinCtx(c), &loginReq)

	if err == nil {
		status = result.GetStatus()
		if result.Status.Code != 0 {
			status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
			c.JSON(http.StatusBadRequest, controller.Fail(int(result.GetStatus().GetCode()), result.GetStatus().GetMessage()))
		} else {
			c.Header("Token", result.TokenString)
			c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: req.UserName, Token: result.TokenString}))
		}
	} else {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	}
}

// Logout logout
// @Summary logout
// @Description logout
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "Logout", int(status.GetCode()))

	bearerStr := c.GetHeader("Authorization")
	tokenStr, err := utils.GetTokenFromBearer(bearerStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
	}
	logoutReq := clusterpb.LogoutRequest{TokenString: tokenStr}
	result, err := client.ClusterClient.Logout(c, &logoutReq)

	if err == nil {
		status = result.GetStatus()
		c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: result.GetAccountName()}))
	} else {
		c.JSON(http.StatusOK, controller.Fail(03, err.Error()))
	}
}
