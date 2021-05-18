package router

import (
	"tcp/addon/logger"
	"tcp/addon/tracer"
	"tcp/api"
	"tcp/auth"

	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(logger.GenGinLogger(), gin.Recovery())
	g.Use(tracer.GinOpenTracing())
	g.Use(auth.GenBasicAuth())
	g.GET("/api/greeter", api.Greeter)
	return g
}
