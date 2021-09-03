package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/clusterapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/databaseapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/hostapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/instanceapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/userapi"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Route(g *gin.Engine) {
	// system check
	check := g.Group("/system")
	{
		check.GET("/check", controller.Hello)
	}

	// support swagger
	swagger := g.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// web
	web := g.Group("/web")
	{
		web.Use(interceptor.AccessLog(), gin.Recovery())
		// 替换成静态文件
		web.GET("/*any", controller.HelloPage)
	}

	// api
	apiV1 := g.Group("/api/v1")
	{
		apiV1.Use(interceptor.GinOpenTracing())
		apiV1.Use(interceptor.GinTraceIDHandler())
		apiV1.Use(interceptor.AccessLog(), gin.Recovery())

		user := apiV1.Group("/user")
		{
			user.POST("/login", userapi.Login)
			user.POST("/logout", userapi.Logout)
		}

		profile := user.Group("")
		{
			profile.Use(interceptor.VerifyIdentity)
			profile.Use(interceptor.AuditLog())
			profile.GET("/profile", userapi.Profile)
		}

		cluster := apiV1.Group("/clusters")
		{
			cluster.Use(interceptor.VerifyIdentity)
			cluster.Use(interceptor.AuditLog())
			cluster.GET("/:clusterId", clusterapi.Detail)
			cluster.POST("/", clusterapi.Create)
			cluster.GET("/", clusterapi.Query)
			cluster.DELETE("/:clusterId", clusterapi.Delete)
			cluster.GET("/:clusterId/dashboard", clusterapi.DescribeDashboard)
			// Params
			cluster.GET("/:clusterId/params", instanceapi.QueryParams)
			cluster.POST("/:clusterId/params", instanceapi.SubmitParams)

			// Backup Strategy
			cluster.GET("/:clusterId/strategy", instanceapi.QueryBackupStrategy)
			cluster.PUT("/:clusterId/strategy", instanceapi.SaveBackupStrategy)
			// cluster.DELETE("/:clusterId/strategy", instanceapi.DeleteBackupStrategy)

			//Import and Export
			cluster.POST("/import", databaseapi.ImportData)
			cluster.POST("/export", databaseapi.ExportData)
			cluster.GET("/:clusterId/transport", databaseapi.DescribeDataTransport)
		}

		knowledge := apiV1.Group("/knowledges")
		{
			knowledge.GET("/", clusterapi.ClusterKnowledge)
		}

		backup := apiV1.Group("/backups")
		{
			backup.Use(interceptor.VerifyIdentity)
			backup.Use(interceptor.AuditLog())
			backup.POST("/", instanceapi.Backup)
			backup.GET("/", instanceapi.QueryBackup)
			backup.POST("/:backupId/restore", instanceapi.RecoverBackup)
			backup.DELETE("/:backupId", instanceapi.DeleteBackup)
			//backup.GET("/:backupId", instanceapi.DetailsBackup)
		}

		host := apiV1.Group("/resources")
		{
			host.Use(interceptor.VerifyIdentity)
			host.Use(interceptor.AuditLog())
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
