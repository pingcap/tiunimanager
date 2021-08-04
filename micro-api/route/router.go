package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/library/thirdparty/logger"
	"github.com/pingcap/ticp/library/thirdparty/tracer"
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
			cluster.Use(security.VerifyIdentity)
			cluster.GET("/:clusterId", clusterapi.Detail)
			cluster.POST("", clusterapi.Create)
			cluster.POST("/query", clusterapi.Query)
			cluster.DELETE("/:clusterId", clusterapi.Delete)

			cluster.GET("/knowledge", clusterapi.ClusterKnowledge)
		}

		param := apiV1.Group("/params")
		{
			param.Use(security.VerifyIdentity)
			param.POST("/:clusterId", instanceapi.QueryParams)
			param.POST("/submit", instanceapi.SubmitParams)
		}

		backup := apiV1.Group("/backup")
		{
			backup.Use(security.VerifyIdentity)
			backup.POST("/:clusterId", instanceapi.Backup)
			backup.GET("/strategy/:clusterId", instanceapi.QueryBackupStrategy)
			backup.POST("/strategy", instanceapi.SaveBackupStrategy)
			backup.POST("/records/:clusterId", instanceapi.QueryBackup)
			backup.POST("/record/recover", instanceapi.RecoverBackup)

			backup.DELETE("/record/:recordId", instanceapi.DeleteBackup)
		}

		demoInstance := apiV1.Group("/instance")
		{
			demoInstance.Use(security.VerifyIdentity)
		}

		host := apiV1.Group("/")
		{
			host.Use(security.VerifyIdentity)
			host.POST("host", hostapi.ImportHost)
			host.POST("hosts", hostapi.ImportHosts)
			host.GET("hosts", hostapi.ListHost)
			host.GET("host/:hostId", hostapi.HostDetails)
			host.DELETE("host/:hostId", hostapi.RemoveHost)
			host.DELETE("hosts", hostapi.RemoveHosts)
			host.GET("download_template", hostapi.DownloadHostTemplateFile)
			host.GET("failuredomains", hostapi.GetFailureDomain)
			// Add allochosts API for debugging, not release.
			host.POST("allochosts", hostapi.AllocHosts)
		}
	}

}
