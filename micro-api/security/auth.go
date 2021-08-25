package security

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/firstparty/client"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"net/http"
)

const VisitorIdentityKey = "VisitorIdentity"

type VisitorIdentity struct {
	AccountId string
	AccountName string
	TenantId string
}

func VerifyIdentity(c *gin.Context) {

	tokenString := c.GetHeader("Token")

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, "")
	}

	path := c.Request.URL
	req := cluster.VerifyIdentityRequest{TokenString: tokenString, Path: path.String()}

	result, err := client.ClusterClient.VerifyIdentity(c, &req)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if result.Status.Code != 0 {
		c.Status(int(result.Status.Code))
		c.Abort()
	} else {
		c.Set(VisitorIdentityKey, &VisitorIdentity{
			AccountId: result.AccountId,
			AccountName: result.AccountName,
			TenantId: result.TenantId,
		})
		c.Next()
	}
}
