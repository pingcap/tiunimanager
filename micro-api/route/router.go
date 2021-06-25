package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-api/controller/hostapi"
	"github.com/pingcap/ticp/micro-api/controller/instanceapi"
	"github.com/pingcap/ticp/micro-api/controller/userapi"
	"github.com/pingcap/ticp/micro-api/security"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Route(g *gin.Engine) {
	// 系统检查
	check := g.Group("/system")
	{
		check.GET("/check", controller.Hello)
	}

	// web静态资源
	web := g.Group("/web")
	{
		web.GET("/*any", controller.HelloPage)
	}

	swagger := g.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// api
	apiV1 := g.Group("/api/v1")
	{

		apiV1.Use(logger.GenGinLogger(), gin.Recovery())
		apiV1.Use(tracer.GinOpenTracing())

		demoUser := apiV1.Group("/user")
		{
			demoUser.POST("/login", userapi.Login)
			demoUser.POST("/logout", userapi.Logout)
		}

		demoHost := apiV1.Group("/host")
		{
			demoHost.Use(security.VerifyIdentity)
			demoHost.POST("/query", hostapi.Query)
		}

		demoInstance := apiV1.Group("/instance")
		{
			demoInstance.Use(security.VerifyIdentity)
			demoInstance.POST("/create", instanceapi.Create)
			demoInstance.POST("/query", instanceapi.Query)
		}

		host := apiV1.Group("/")
		{
			demoInstance.Use(security.VerifyIdentity)
			host.POST("host", hostapi.ImportHost)
			host.GET("hosts", hostapi.ListHost)
			host.GET("host/:hostId", hostapi.HostDetails)
			host.DELETE("host/:hostId", hostapi.RemoveHost)
		}
	}

}
