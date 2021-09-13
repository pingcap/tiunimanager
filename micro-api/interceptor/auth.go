package interceptor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	utils "github.com/pingcap-inc/tiem/library/util/stringutil"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
)

const VisitorIdentityKey = "VisitorIdentity"

type VisitorIdentity struct {
	AccountId   string
	AccountName string
	TenantId    string
}

func VerifyIdentity(c *gin.Context) {

	bearerTokenStr := c.GetHeader("Authorization")

	tokenString, err := utils.GetTokenFromBearer(bearerTokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
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
			AccountId:   result.AccountId,
			AccountName: result.AccountName,
			TenantId:    result.TenantId,
		})
		c.Next()
	}
}
