package auth

import (
	"tcp/models"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/gin-gonic/gin"
	basicAuth "github.com/jackdoe/gin-basic-auth-dynamic"
)

func GenBasicAuth() gin.HandlerFunc {
	return basicAuth.BasicAuth(
		func(context *gin.Context, realm, userName, userPass string) basicAuth.AuthResult {
			user, err := models.FindUserByName(userName)
			if err == nil {
				if user.Password == userPass {
					return basicAuth.AuthResult{Success: true, Text: "authorized"}
				} else {
					return basicAuth.AuthResult{Success: false, Text: "not authorized"}
				}
			} else {
				return basicAuth.AuthResult{Success: false, Text: "not authorized"}
			}
		},
	)
}
