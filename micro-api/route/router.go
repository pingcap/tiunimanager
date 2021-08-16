package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-api/controller/clusterapi"
	"github.com/pingcap/tiem/micro-api/controller/databaseapi"
	"github.com/pingcap/tiem/micro-api/controller/hostapi"
	"github.com/pingcap/tiem/micro-api/controller/instanceapi"
	"github.com/pingcap/tiem/micro-api/controller/userapi"
	"github.com/pingcap/tiem/micro-api/security"
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

		cluster := apiV1.Group("/clusters")
		{
			cluster.Use(security.VerifyIdentity)
			cluster.GET("/:clusterId", clusterapi.Detail)
			cluster.POST("/", clusterapi.Create)
			cluster.GET("/", clusterapi.Query)
			cluster.DELETE("/:clusterId", clusterapi.Delete)

			// Params
			cluster.GET("/:clusterId/params", instanceapi.QueryParams)
			cluster.POST("/:clusterId/params", instanceapi.SubmitParams)

			// Backup Strategy
			cluster.GET("/:clusterId/strategy", instanceapi.QueryBackupStrategy)
			cluster.PUT("/:clusterId/strategy", instanceapi.SaveBackupStrategy)
			// cluster.DELETE("/:clusterId/strategy", instanceapi.DeleteBackupStrategy)
		}

		knowledge := apiV1.Group("/knowledges")
		{
			// api/v1/knowledges?type=cluster
			knowledge.GET("/", clusterapi.ClusterKnowledge)
		}

		backup := apiV1.Group("/backups")
		{
			backup.Use(security.VerifyIdentity)
			backup.POST("/", instanceapi.Backup)
			backup.GET("/", instanceapi.QueryBackup)
			backup.POST("/:backupId/restore", instanceapi.RecoverBackup)
			backup.DELETE("/:backupId", instanceapi.DeleteBackup)
			//backup.GET("/:backupId", instanceapi.DetailsBackup)
		}

		database := apiV1.Group("/database")
		{
			database.Use(security.VerifyIdentity)
			database.POST("/import", databaseapi.ImportData)
			database.POST("/export", databaseapi.ExportData)
			database.GET("/describeDataTransport", databaseapi.DescribeDataTransport)
		}

		host := apiV1.Group("/resources")
		{
			host.Use(security.VerifyIdentity)
			host.POST("host", hostapi.ImportHost)
			host.POST("hosts", hostapi.ImportHosts)
			host.GET("hosts", hostapi.ListHost)
			host.GET("hosts/:hostId", hostapi.HostDetails)
			host.DELETE("hosts/:hostId", hostapi.RemoveHost)
			host.DELETE("hosts", hostapi.RemoveHosts)

			host.GET("hosts-template", hostapi.DownloadHostTemplateFile)

			host.GET("failuredomains", hostapi.GetFailureDomain)

			// Add allochosts API for debugging, not release.
			host.POST("allochosts", hostapi.AllocHosts)
		}
	}

}
