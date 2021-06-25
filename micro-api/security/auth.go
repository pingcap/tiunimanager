package security

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"net/http"
)

const AccountNameFromToken = "AccountNameFromToken"
const TenantIdFromToken = "TenantIdFromToken"

func VerifyIdentity(c *gin.Context) {

	tokenString := c.GetHeader("Token")

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, "")
	}

	path := c.Request.URL
	req := manager.VerifyIdentityRequest{TokenString: tokenString, Path: path.String()}

	result, err := client.ManagerClient.VerifyIdentity(c, &req)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if result.Status.Code != 0 {
		c.Status(int(result.Status.Code))
		c.Abort()
	} else {
		c.Set(AccountNameFromToken, result.AccountName)
		c.Set(TenantIdFromToken, int(result.TenantId))
		c.Next()
	}
}
