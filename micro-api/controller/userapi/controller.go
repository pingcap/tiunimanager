package userapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-manager/client"
	"github.com/pingcap/ticp/micro-manager/proto"
	"net/http"
)

// Login 登录接口
// @Summary 登录接口
// @Description 登录
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param loginInfo body LoginInfo true "登录用户信息"
// @Header 200 {string} Token "DUISAFNDHIGADS"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Router /user/login [post]
func Login(c *gin.Context) {
	var req LoginInfo

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	loginReq := manager.LoginRequest{AccountName: req.UserName, Password: req.UserPassword}
	result, err := client.ManagerClient.Login(c, &loginReq)

	if err == nil {
		c.Header("Token", result.TokenString)
		c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: "peijin"}))
	} else {
		c.JSON(http.StatusOK, controller.Fail(03, "账号或密码错误"))
	}
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param logoutInfo body LogoutInfo false "退出登录信息"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	var req LogoutInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	logoutReq := manager.LogoutRequest{TokenString: c.GetHeader("Token")}
	result, err := client.ManagerClient.Logout(c, &logoutReq)

	if err == nil {
		c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: result.GetAccountName()}))
	} else {
		c.JSON(http.StatusOK, controller.Fail(03, err.Error()))
	}
}