package api

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(logger.GenGinLogger(), gin.Recovery())
	g.Use(tracer.GinOpenTracing())
	//g.Use(auth.GenBasicAuth())
	g.GET("/api/hello", Hello)
	g.GET("/api/user/login", Hello)
	return g
}
