package userapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
)

// Login 登录接口
// @Summary 登录接口
// @Description 登录
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param userName body string true "账户名"
// @Param userPassword body string true "账户密码"
// @Header 200 {string} Token "DUISAFNDHIGADS"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Router /user/login [post]
func Login(c *gin.Context) {
	var req LoginInfo

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: "peijin"}))
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param userName body string false "账户名"
// @Success 200 {object} controller.CommonResult{data=UserIdentity}
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	//var req LogoutInfo
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	_ = c.Error(err)
	//	return
	//}

	c.JSON(http.StatusOK, controller.Success(UserIdentity{UserName: "peijin"}))
}