package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
)

const VisitorIdentityKey = "VisitorIdentity"

type VisitorIdentity struct {
	AccountId   string
	AccountName string
	TenantId    string
}

func VerifyIdentity(c *gin.Context) {

	bearerTokenString := c.GetHeader("Authorization")
	bearerPrefix := "Bearer "

	if bearerTokenString == "" {
		errMsg := "authorization token empty"
		c.AbortWithStatusJSON(http.StatusUnauthorized, errMsg)
	}

	var tokenString string
	if strings.HasPrefix(bearerTokenString, bearerPrefix) {
		tokenString = bearerTokenString[len(bearerPrefix):]
	} else {
		errMsg := fmt.Sprintf("bad authorization token: %s", bearerTokenString)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errMsg)
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
