package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	transportApi "github.com/pingcap-inc/tiem/micro-api/controller/business-data/import-export"
	brApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/backup-restore"
	logApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/log"
	clusterApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
	parameterApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/parameter"
	resourceApi "github.com/pingcap-inc/tiem/micro-api/controller/resource/management"
	warehouseApi "github.com/pingcap-inc/tiem/micro-api/controller/resource/warehouse"
	flowtaskApi "github.com/pingcap-inc/tiem/micro-api/controller/task/flowtask"
	idApi "github.com/pingcap-inc/tiem/micro-api/controller/user/identification"
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
			user.POST("/login", idApi.Login)
			user.POST("/logout", idApi.Logout)
		}

		profile := user.Group("")
		{
			profile.Use(interceptor.VerifyIdentity)
			profile.Use(interceptor.AuditLog())
			profile.GET("/profile", idApi.Profile)
		}

		cluster := apiV1.Group("/clusters")
		{
			cluster.Use(interceptor.VerifyIdentity)
			cluster.Use(interceptor.AuditLog())
			cluster.GET("/:clusterId", clusterApi.Detail)
			cluster.POST("/", clusterApi.Create)
			cluster.GET("/", clusterApi.Query)
			cluster.DELETE("/:clusterId", clusterApi.Delete)
			cluster.POST("/restore", brApi.Restore)
			cluster.GET("/:clusterId/dashboard", clusterApi.DescribeDashboard)
			// Params
			cluster.GET("/:clusterId/params", parameterApi.QueryParams)
			cluster.POST("/:clusterId/params", parameterApi.SubmitParams)

			// Backup Strategy
			cluster.GET("/:clusterId/strategy", brApi.QueryBackupStrategy)
			cluster.PUT("/:clusterId/strategy", brApi.SaveBackupStrategy)
			// cluster.DELETE("/:clusterId/strategy", instanceapi.DeleteBackupStrategy)

			//Import and Export
			cluster.POST("/import", transportApi.ImportData)
			cluster.POST("/export", transportApi.ExportData)
			cluster.GET("/:clusterId/transport", transportApi.DescribeDataTransport)
		}

		knowledge := apiV1.Group("/knowledges")
		{
			knowledge.GET("/", clusterApi.ClusterKnowledge)
		}

		backup := apiV1.Group("/backups")
		{
			backup.Use(interceptor.VerifyIdentity)
			backup.Use(interceptor.AuditLog())
			backup.POST("/", brApi.Backup)
			backup.GET("/", brApi.QueryBackup)
			backup.DELETE("/:backupId", brApi.DeleteBackup)
			//backup.GET("/:backupId", instanceapi.DetailsBackup)
		}

		flowworks := apiV1.Group("/flowworks")
		{
			flowworks.Use(interceptor.VerifyIdentity)
			flowworks.Use(interceptor.AuditLog())
			flowworks.GET("/", flowtaskApi.Query)
		}

		host := apiV1.Group("/resources")
		{
			host.Use(interceptor.VerifyIdentity)
			host.Use(interceptor.AuditLog())
			host.POST("host", resourceApi.ImportHost)
			host.POST("hosts", resourceApi.ImportHosts)
			host.GET("hosts", resourceApi.ListHost)
			host.GET("hosts/:hostId", resourceApi.HostDetails)
			host.DELETE("hosts/:hostId", resourceApi.RemoveHost)
			host.DELETE("hosts", resourceApi.RemoveHosts)

			host.GET("hosts-template", resourceApi.DownloadHostTemplateFile)

			host.GET("failuredomains", warehouseApi.GetFailureDomain)

			// Add allochosts API for debugging, not release.
			host.POST("allochosts", resourceApi.AllocHosts)
		}

		log := apiV1.Group("/logs")
		{
			log.Use(interceptor.VerifyIdentity)
			log.GET("/tidb/:clusterId", logApi.SearchTiDBLog)
		}
	}

}
