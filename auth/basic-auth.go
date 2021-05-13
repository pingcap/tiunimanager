package auth

import (
	"tcp/models"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/gin-gonic/gin"
	basicAuth "github.com/jackdoe/gin-basic-auth-dynamic"
	"github.com/opentracing/opentracing-go"
)

func GenBasicAuth() gin.HandlerFunc {
	return basicAuth.BasicAuth(
		func(context *gin.Context, realm, userName, userPass string) basicAuth.AuthResult {
			c := context
			var sp opentracing.Span

			v, existFlag := c.Get("ParentSpan")
			var err error
			var parentSpan opentracing.Span
			var ok bool
			if existFlag {
				parentSpan, ok = v.(opentracing.Span)
				//ctx := opentracing.ContextWithSpan(c, parentSpan)
			} else {
			}
			if existFlag && ok {
				sp = opentracing.StartSpan(c.Request.URL.Path+"|mysql", opentracing.ChildOf(parentSpan.Context()))
			} else {
				sp = opentracing.StartSpan(c.Request.URL.Path + "|mysql")
			}

			user, err := models.FindUserByName(userName)

			sp.Finish()

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
