package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-api/controller/hostapi"
	"github.com/pingcap/ticp/micro-api/controller/instanceapi"
	"github.com/pingcap/ticp/micro-api/controller/userapi"
	"github.com/pingcap/ticp/micro-api/security"
)

func Route(g *gin.Engine){
	g.POST("/api/v1/hello", security.VerifyIdentity, controller.Hello)
	g.POST("/api/v1/user/login", userapi.Login)
	g.POST("/api/v1/user/logout", userapi.Logout)

	g.POST("/api/v1/instance/create", security.VerifyIdentity, instanceapi.Create)
	g.POST("/api/v1/instance/query", security.VerifyIdentity, instanceapi.Query)

	g.POST("/api/v1/host/query", security.VerifyIdentity, hostapi.Query)
}
