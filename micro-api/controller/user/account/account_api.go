package account

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	"net/http"
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
	v, _ := c.Get(interceptor.VisitorIdentityKey)

	visitor, _ := v.(*interceptor.VisitorIdentity)
	c.JSON(http.StatusOK, controller.Success(UserProfile{UserName: visitor.AccountName, TenantId: visitor.TenantId}))
}
