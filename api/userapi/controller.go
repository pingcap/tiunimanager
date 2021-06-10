package userapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/api"
	"net/http"
)

// Login 登录接口
// @Summary 登录接口
// @Description 登录
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Param request formData LoginInfo true "登录信息"
// @Header 200 {string} Token "DUISAFNDHIGADS"
// @Success 200 {object} api.ResultWithPage{data=UserIdentity}
// @Router /user/login [post]
func Login(c *gin.Context) {
	//var req LoginInfo
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	_ = c.Error(err)
	//	return
	//}
	//
	c.JSON(http.StatusOK, api.Success(UserIdentity{UserName: "peijin"}))
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags 用户
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param {object} formData LogoutInfo false "登出信息"
// @Success 200 {object} api.ResultWithPage{data=UserIdentity}
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	//var req LogoutInfo
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	_ = c.Error(err)
	//	return
	//}

	c.JSON(http.StatusOK, api.Success(UserIdentity{UserName: "peijin"}))
}