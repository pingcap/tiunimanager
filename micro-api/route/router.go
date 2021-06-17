package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/micro-api/controller"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(logger.GenGinLogger(), gin.Recovery())
	g.Use(tracer.GinOpenTracing())
	//g.Use(auth.GenBasicAuth())
	g.GET("/api/hello", controller.Hello)
	g.GET("/api/user/login", controller.Hello)
	return g
}
