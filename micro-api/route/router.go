
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/clusterapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/databaseapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/hostapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/instanceapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/logapi"
	"github.com/pingcap-inc/tiem/micro-api/controller/taskapi"
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
			cluster.POST("/restore", clusterapi.Restore)
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
			backup.DELETE("/:backupId", instanceapi.DeleteBackup)
			//backup.GET("/:backupId", instanceapi.DetailsBackup)
		}

		flowworks := apiV1.Group("/flowworks")
		{
			flowworks.Use(interceptor.VerifyIdentity)
			flowworks.Use(interceptor.AuditLog())
			flowworks.GET("/", taskapi.Query)
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

		log := apiV1.Group("/logs")
		{
			log.Use(interceptor.VerifyIdentity)
			log.GET("/tidb/:clusterId", logapi.SearchTiDBLog)
		}
	}

}
