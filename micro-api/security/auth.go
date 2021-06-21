package security

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"net/http"
)

func VerifyIdentity(c *gin.Context) {

	tokenString := c.GetHeader("Token")

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, "")
	}

	path := c.Request.URL
	req := manager.VerifyIdentityRequest{TokenString: tokenString, Path: path.String()}

	result, err := client.ManagerClient.VerifyIdentity(c, &req)
	if err != nil {
		c.JSON(http.StatusOK, controller.Fail(03, err.Error()))
	} else {
		c.Set("accountName", result.AccountName)
		c.Set("tenantId", result.TenantId)
	}
}
