package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-api/controller/hostapi"
	"github.com/pingcap/ticp/micro-api/controller/instanceapi"
	"github.com/pingcap/ticp/micro-api/controller/userapi"
)

func Route(g *gin.Engine){
	g.GET("/api/hello", controller.Hello)
	g.GET("/api/user/login", userapi.Login)
	g.GET("/api/user/logout", userapi.Logout)

	g.GET("/api/instance/create", instanceapi.Create)
	g.GET("/api/instance/query", instanceapi.Query)

	g.GET("/api/host/query", hostapi.Query)
}
