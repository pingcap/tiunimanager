package router

import (
	"tcp/addon/tracer"
	"tcp/api"
	"tcp/auth"

	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	g := gin.Default()
	g.Use(tracer.GinOpenTracing())
	g.Use(auth.GenBasicAuth())
	g.GET("/api/greeter", api.Greeter)
	return g
}
