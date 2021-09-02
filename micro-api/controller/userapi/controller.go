package userapi

import (
	"net/http"

	"github.com/pingcap-inc/tiem/micro-api/interceptor"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	utils "github.com/pingcap-inc/tiem/library/util/stringutil"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
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
	var req LoginInfo

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	loginReq := cluster.LoginRequest{AccountName: req.UserName, Password: req.UserPassword}
	result, err := client.ClusterClient.Login(framework.NewMicroCtxFromGinCtx(c), &loginReq)

	if err == nil {
		if result.Status.Code != 0 {
			c.JSON(http.StatusOK, controller.Fail(int(result.GetStatus().GetCode()), result.GetStatus().GetMessage()))
		} else {
			c.Header("Token", result.TokenString)
			c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: req.UserName, Token: result.TokenString}))
		}
	} else {
		c.JSON(http.StatusOK, controller.Fail(401, "账号或密码错误"))
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
	bearerStr := c.GetHeader("Authorization")
	tokenStr, err := utils.GetTokenFromBearer(bearerStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
	}
	logoutReq := cluster.LogoutRequest{TokenString: tokenStr}
	result, err := client.ClusterClient.Logout(c, &logoutReq)

	if err == nil {
		c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: result.GetAccountName()}))
	} else {
		c.JSON(http.StatusOK, controller.Fail(03, err.Error()))
	}
}

// Profile user profile
// @Summary user profile
// @Description profile
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/profile [get]
func Profile(c *gin.Context) {
	v, _ := c.Get(interceptor.VisitorIdentityKey)

	visitor, _ := v.(*interceptor.VisitorIdentity)
	c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: visitor.AccountName, TenantId: visitor.TenantId}))
}
