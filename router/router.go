package router

import (
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/api"
	"github.com/pingcap/ticp/auth"

	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(logger.GenGinLogger(), gin.Recovery())
	g.Use(tracer.GinOpenTracing())
	g.Use(auth.GenBasicAuth())
	g.GET("/api/hello", api.Hello)
	return g
}
