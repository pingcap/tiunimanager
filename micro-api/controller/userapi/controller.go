package userapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
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
	result, err := client.ClusterClient.Login(c, &loginReq)

	if err == nil {
		if result.Status.Code != 0 {
			c.JSON(http.StatusOK, controller.Fail(int(result.GetStatus().GetCode()), result.GetStatus().GetMessage()))
		} else {
			c.Header("Token", result.TokenString)
			c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: req.UserName}))
		}
	} else {
		c.JSON(http.StatusOK, controller.Fail(401, "账号或密码错误"))
	}
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param Token header string true "token"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	logoutReq := cluster.LogoutRequest{TokenString: c.GetHeader("Token")}
	result, err := client.ClusterClient.Logout(c, &logoutReq)

	if err == nil {
		c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: result.GetAccountName()}))
	} else {
		c.JSON(http.StatusOK, controller.Fail(03, err.Error()))
	}
}
