package auth

import (
	"github.com/pingcap/tcp/addon/logger"
	"github.com/pingcap/tcp/client"
	dbPb "github.com/pingcap/tcp/proto/db"

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

			var req dbPb.CheckUserRequest
			req.Name = userName
			req.Passwd = userPass
			resp, err := client.DbClient.CheckUser(c, &req)

			sp.Finish()

			if err == nil && resp.ErrCode == 0 {
				return basicAuth.AuthResult{Success: true, Text: "authorized"}
			} else {
				if err != nil {
					logger.WithContext(c).Errorf("client.DbClient.CheckUser met an error:%v", err)
				}
				return basicAuth.AuthResult{Success: false, Text: "not authorized"}
			}
		},
	)
}
