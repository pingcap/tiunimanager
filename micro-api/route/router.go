package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-api/controller/clusterapi"
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
		// 替换成静态文件
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

		user := apiV1.Group("/user")
		{
			user.POST("/login", userapi.Login)
			user.POST("/logout", userapi.Logout)
		}

		cluster := apiV1.Group("/cluster")
		{
			//cluster.Use(security.VerifyIdentity)
			cluster.GET("/:clusterId", clusterapi.Detail)
			cluster.POST("", clusterapi.Create)
			cluster.GET("/query", clusterapi.Query)
			cluster.DELETE("/:clusterId", clusterapi.Delete)

			cluster.GET("/knowledge", clusterapi.ClusterKnowledge)
		}

		param := apiV1.Group("/params")
		{
			//param.Use(security.VerifyIdentity)
			param.GET("/:clusterId", instanceapi.QueryParams)
			param.POST("/submit", instanceapi.SubmitParams)
		}

		demoHost := apiV1.Group("/host")
		{
			demoHost.Use(security.VerifyIdentity)
			demoHost.POST("/query", hostapi.Query)
		}

		demoInstance := apiV1.Group("/instance")
		{
			demoInstance.Use(security.VerifyIdentity)
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
