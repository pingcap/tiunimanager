/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/metrics"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/changefeed"
	logApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/log"
	clusterApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
	parameterApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/parameter"
	switchoverApi "github.com/pingcap-inc/tiem/micro-api/controller/cluster/switchover"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/upgrade"

	"github.com/pingcap-inc/tiem/micro-api/controller/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/micro-api/controller/parametergroup"
	"github.com/pingcap-inc/tiem/micro-api/controller/platform/product"
	resourceApi "github.com/pingcap-inc/tiem/micro-api/controller/resource/hostresource"
	warehouseApi "github.com/pingcap-inc/tiem/micro-api/controller/resource/warehouse"
	flowtaskApi "github.com/pingcap-inc/tiem/micro-api/controller/task/flowtask"
	userApi "github.com/pingcap-inc/tiem/micro-api/controller/user"
	rbacApi "github.com/pingcap-inc/tiem/micro-api/controller/user/rbac"
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
		web.GET("/*any", controller.HelloPage)
	}

	// api
	apiV1 := g.Group("/api/v1")
	{
		apiV1.Use(interceptor.GinTraceIDHandler())
		apiV1.Use(interceptor.GinOpenTracing())
		apiV1.Use(interceptor.AccessLog(), gin.Recovery())

		auth := apiV1.Group("/user")
		{
			auth.POST("/login", metrics.HandleMetrics(constants.MetricsUserLogin), userApi.Login)
			auth.POST("/logout", metrics.HandleMetrics(constants.MetricsUserLogout), userApi.Logout)
		}

		user := apiV1.Group("/users")
		{
			user.Use(interceptor.VerifyIdentity)
			user.Use(interceptor.AuditLog())
			user.POST("/", metrics.HandleMetrics(constants.MetricsUserCreate), userApi.CreateUser)
			user.DELETE("/:userId", metrics.HandleMetrics(constants.MetricsUserDelete), userApi.DeleteUser)
			user.POST("/:userId/update_profile", metrics.HandleMetrics(constants.MetricsUserUpdateProfile), userApi.UpdateUserProfile)
			user.POST("/:userId/password", metrics.HandleMetrics(constants.MetricsUserUpdatePassword), userApi.UpdateUserPassword)
			user.GET("/:userId", metrics.HandleMetrics(constants.MetricsUserGet), userApi.GetUser)
			user.GET("/", metrics.HandleMetrics(constants.MetricsUserQuery), userApi.QueryUsers)
		}

		tenant := apiV1.Group("/tenants")
		{
			tenant.Use(interceptor.VerifyIdentity)
			tenant.Use(interceptor.AuditLog())
			tenant.POST("/", metrics.HandleMetrics(constants.MetricsTenantCreate), userApi.CreateTenant)
			tenant.DELETE("/:tenantId", metrics.HandleMetrics(constants.MetricsTenantDelete), userApi.DeleteTenant)
			tenant.POST("/:tenantId/update_profile", metrics.HandleMetrics(constants.MetricsTenantUpdateProfile), userApi.UpdateTenantProfile)
			tenant.POST("/:tenantId/update_on_boarding_status", metrics.HandleMetrics(constants.MetricsTenantUpdateOnBoardingStatus), userApi.UpdateTenantOnBoardingStatus)
			tenant.GET("/:tenantId", metrics.HandleMetrics(constants.MetricsTenantGet), userApi.GetTenant)
			tenant.GET("/", metrics.HandleMetrics(constants.MetricsTenantQuery), userApi.QueryTenants)
		}

		rbac := apiV1.Group("/rbac")
		{
			rbac.Use(interceptor.VerifyIdentity)
			rbac.Use(interceptor.AuditLog())
			rbac.POST("/role/", metrics.HandleMetrics(constants.MetricsRbacCreateRole), rbacApi.CreateRbacRole)
			rbac.GET("/role/", metrics.HandleMetrics(constants.MetricsRbacQueryRole), rbacApi.QueryRbacRoles)
			rbac.POST("/role/bind", metrics.HandleMetrics(constants.MetricsRbacBindRolesForUser), rbacApi.BindRolesForUser)
			rbac.DELETE("/role/unbind", metrics.HandleMetrics(constants.MetricsRbacUnbindRoleForUser), rbacApi.UnbindRoleForUser)
			rbac.DELETE("/role/:role", metrics.HandleMetrics(constants.MetricsRbacDeleteRole), rbacApi.DeleteRbacRole)
			rbac.POST("/permission/add", metrics.HandleMetrics(constants.MetricsRbacAddPermissionForRole), rbacApi.AddPermissionsForRole)
			rbac.DELETE("/permission/delete", metrics.HandleMetrics(constants.MetricsRbacDeletePermissionForRole), rbacApi.DeletePermissionsForRole)
			rbac.GET("/permission/:userId", metrics.HandleMetrics(constants.MetricsRbacQueryPermissionForUser), rbacApi.QueryPermissionsForUser)
			rbac.POST("/permission/check", metrics.HandleMetrics(constants.MetricsRbacCheckPermissionForUser), rbacApi.CheckPermissionForUser)
		}

		cluster := apiV1.Group("/clusters")
		{
			cluster.Use(interceptor.VerifyIdentity)
			cluster.Use(interceptor.AuditLog())
			cluster.GET("/:clusterId", metrics.HandleMetrics(constants.MetricsClusterDetail), clusterApi.Detail)
			cluster.POST("/", metrics.HandleMetrics(constants.MetricsClusterCreate), clusterApi.Create)
			cluster.POST("/takeover", metrics.HandleMetrics(constants.MetricsClusterTakeover), clusterApi.Takeover)
			cluster.POST("/preview", metrics.HandleMetrics(constants.MetricsClusterPreview), clusterApi.Preview)

			cluster.GET("/", metrics.HandleMetrics(constants.MetricsClusterQuery), clusterApi.Query)
			cluster.DELETE("/:clusterId", metrics.HandleMetrics(constants.MetricsClusterDelete), clusterApi.Delete)
			cluster.POST("/:clusterId/restart", metrics.HandleMetrics(constants.MetricsClusterRestart), clusterApi.Restart)
			cluster.POST("/:clusterId/stop", metrics.HandleMetrics(constants.MetricsClusterStop), clusterApi.Stop)
			cluster.POST("/restore", metrics.HandleMetrics(constants.MetricsClusterRestore), backuprestore.Restore)
			cluster.GET("/:clusterId/dashboard", metrics.HandleMetrics(constants.MetricsClusterQueryDashboardAddress), clusterApi.GetDashboardInfo)
			cluster.GET("/:clusterId/monitor", metrics.HandleMetrics(constants.MetricsClusterQueryMonitorAddress), clusterApi.GetMonitorInfo)

			cluster.GET("/:clusterId/log", metrics.HandleMetrics(constants.MetricsClusterQueryLogParameter), logApi.QueryClusterLog)

			// Scale cluster
			cluster.POST("/:clusterId/preview-scale-out", metrics.HandleMetrics(constants.MetricsClusterPreviewScaleOut), clusterApi.ScaleOutPreview)
			cluster.POST("/:clusterId/scale-out", metrics.HandleMetrics(constants.MetricsClusterScaleOut), clusterApi.ScaleOut)
			cluster.POST("/:clusterId/scale-in", metrics.HandleMetrics(constants.MetricsClusterScaleIn), clusterApi.ScaleIn)

			// Clone cluster
			cluster.POST("/clone", metrics.HandleMetrics(constants.MetricsClusterClone), clusterApi.Clone)

			// Switchover
			cluster.POST("/switchover", metrics.HandleMetrics(constants.MetricsSwitchover), switchoverApi.Switchover)

			// Params
			cluster.GET("/:clusterId/params", metrics.HandleMetrics(constants.MetricsClusterQueryParameter), parameterApi.QueryParameters)
			cluster.PUT("/:clusterId/params", metrics.HandleMetrics(constants.MetricsClusterModifyParameter), parameterApi.UpdateParameters)
			cluster.POST("/:clusterId/params/inspect", metrics.HandleMetrics(constants.MetricsClusterInspectParameter), parameterApi.InspectParameters)

			// Backup Strategy
			cluster.GET("/:clusterId/strategy", metrics.HandleMetrics(constants.MetricsBackupQueryStrategy), backuprestore.GetBackupStrategy)
			cluster.PUT("/:clusterId/strategy", metrics.HandleMetrics(constants.MetricsBackupModifyStrategy), backuprestore.SaveBackupStrategy)

			//Import and Export
			cluster.POST("/import", metrics.HandleMetrics(constants.MetricsDataImport), importexport.ImportData)
			cluster.POST("/export", metrics.HandleMetrics(constants.MetricsDataExport), importexport.ExportData)
			cluster.GET("/transport", metrics.HandleMetrics(constants.MetricsDataExportImportQuery), importexport.QueryDataTransport)
			cluster.DELETE("/transport/:recordId", metrics.HandleMetrics(constants.MetricsDataExportImportDelete), importexport.DeleteDataTransportRecord)

			//Upgrade
			cluster.GET("/:clusterId/upgrade/path", metrics.HandleMetrics(constants.MetricsClusterUpgradePath), upgrade.QueryUpgradePaths)
			cluster.GET("/:clusterId/upgrade/diff", metrics.HandleMetrics(constants.MetricsClusterUpgradeDiff), upgrade.QueryUpgradeVersionDiffInfo)
			cluster.POST("/:clusterId/upgrade", metrics.HandleMetrics(constants.MetricsClusterUpgrade), upgrade.Upgrade)
		}

		backup := apiV1.Group("/backups")
		{
			backup.Use(interceptor.VerifyIdentity)
			backup.Use(interceptor.AuditLog())

			backup.POST("/", metrics.HandleMetrics(constants.MetricsBackupCreate), backuprestore.Backup)
			backup.GET("/", metrics.HandleMetrics(constants.MetricsBackupQuery), backuprestore.QueryBackupRecords)
			backup.DELETE("/:backupId", metrics.HandleMetrics(constants.MetricsBackupDelete), backuprestore.DeleteBackup)
		}

		changeFeeds := apiV1.Group("/changefeeds")
		{
			changeFeeds.Use(interceptor.VerifyIdentity)
			changeFeeds.Use(interceptor.AuditLog())

			changeFeeds.POST("/", metrics.HandleMetrics(constants.MetricsCDCTaskCreate), changefeed.Create)
			changeFeeds.POST("/:changeFeedTaskId/pause", metrics.HandleMetrics(constants.MetricsCDCTaskPause), changefeed.Pause)
			changeFeeds.POST("/:changeFeedTaskId/resume", metrics.HandleMetrics(constants.MetricsCDCTaskPause), changefeed.Resume)
			changeFeeds.POST("/:changeFeedTaskId/update", metrics.HandleMetrics(constants.MetricsCDCTaskUpdate), changefeed.Update)

			changeFeeds.DELETE("/:changeFeedTaskId", metrics.HandleMetrics(constants.MetricsCDCTaskDelete), changefeed.Delete)

			changeFeeds.GET("/:changeFeedTaskId/", metrics.HandleMetrics(constants.MetricsCDCTaskDetail), changefeed.Detail)
			changeFeeds.GET("/", metrics.HandleMetrics(constants.MetricsCDCTaskQuery), changefeed.Query)
		}

		flowworks := apiV1.Group("/workflow")
		{
			flowworks.Use(interceptor.VerifyIdentity)
			flowworks.Use(interceptor.AuditLog())
			flowworks.GET("/", metrics.HandleMetrics(constants.MetricsWorkFlowQuery), flowtaskApi.Query)
			flowworks.GET("/:workFlowId", metrics.HandleMetrics(constants.MetricsWorkFlowDetail), flowtaskApi.Detail)
		}

		host := apiV1.Group("/resources")
		{
			host.Use(interceptor.VerifyIdentity)
			host.Use(interceptor.AuditLog())
			host.POST("hosts", metrics.HandleMetrics(constants.MetricsResourceImportHosts), resourceApi.ImportHosts)
			host.GET("hosts", metrics.HandleMetrics(constants.MetricsResourceQueryHosts), resourceApi.QueryHosts)
			host.DELETE("hosts", metrics.HandleMetrics(constants.MetricsResourceDeleteHost), resourceApi.RemoveHosts)
			host.GET("hosts-template", metrics.HandleMetrics(constants.MetricsResourceDownloadHostTemplateFile), resourceApi.DownloadHostTemplateFile)
			host.GET("hierarchy", metrics.HandleMetrics(constants.MetricsResourceQueryHierarchy), warehouseApi.GetHierarchy)
			host.GET("stocks", metrics.HandleMetrics(constants.MetricsResourceQueryStocks), warehouseApi.GetStocks)
			host.PUT("host-reserved", metrics.HandleMetrics(constants.MetricsResourceReservedHost), resourceApi.UpdateHostReserved)
			host.PUT("host-status", metrics.HandleMetrics(constants.MetricsResourceModifyHostStatus), resourceApi.UpdateHostStatus)
		}

		paramGroups := apiV1.Group("/param-groups")
		{
			paramGroups.Use(interceptor.VerifyIdentity)
			paramGroups.Use(interceptor.AuditLog())
			paramGroups.GET("/", metrics.HandleMetrics(constants.MetricsParameterGroupQuery), parametergroup.Query)
			paramGroups.GET("/:paramGroupId", metrics.HandleMetrics(constants.MetricsParameterGroupDetail), parametergroup.Detail)
			paramGroups.POST("/", metrics.HandleMetrics(constants.MetricsParameterGroupCreate), parametergroup.Create)
			paramGroups.PUT("/:paramGroupId", metrics.HandleMetrics(constants.MetricsParameterGroupUpdate), parametergroup.Update)
			paramGroups.DELETE("/:paramGroupId", metrics.HandleMetrics(constants.MetricsParameterGroupDelete), parametergroup.Delete)
			paramGroups.POST("/:paramGroupId/copy", metrics.HandleMetrics(constants.MetricsParameterGroupCopy), parametergroup.Copy)
			paramGroups.POST("/:paramGroupId/apply", metrics.HandleMetrics(constants.MetricsParameterGroupApply), parametergroup.Apply)
		}

		productGroup := apiV1.Group("/products")
		{
			productGroup.Use(interceptor.VerifyIdentity)
			productGroup.Use(interceptor.AuditLog())
			productGroup.POST("/", product.CreateProduct)
			productGroup.DELETE("/", product.DeleteProduct)
			productGroup.GET("/", product.QueryProducts)
			productGroup.GET("/detail", product.QueryProductDetail)
		}

		zoneGroup := apiV1.Group("/zones")
		{
			zoneGroup.Use(interceptor.VerifyIdentity)
			zoneGroup.Use(interceptor.AuditLog())
			zoneGroup.POST("/", product.CreateZones)
			zoneGroup.DELETE("/", product.DeleteZones)
			zoneGroup.GET("/tree", product.QueryZonesTree)
		}

		specGroup := apiV1.Group("/specs")
		{
			specGroup.Use(interceptor.VerifyIdentity)
			specGroup.Use(interceptor.AuditLog())
			specGroup.POST("/", product.CreateSpecs)
			specGroup.DELETE("/", product.DeleteSpecs)
			specGroup.GET("/", product.QuerySpecs)
		}
	}

}
