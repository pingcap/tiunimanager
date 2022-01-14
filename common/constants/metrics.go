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
 * limitations under the License                                              *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: metrics.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/22
*******************************************************************************/

package constants

type MetricsType string

const (
	OpenAPIMetricsPrefix MetricsType = "/api/v1/"
	// MetricsClusterCreate define cluster metrics
	MetricsClusterCreate                MetricsType = "cluster/create"
	MetricsClusterDelete                MetricsType = "cluster/delete"
	MetricsClusterStop                  MetricsType = "cluster/stop"
	MetricsClusterStart                 MetricsType = "cluster/start"
	MetricsClusterRestart               MetricsType = "cluster/restart"
	MetricsClusterScaleIn               MetricsType = "cluster/scale_in"
	MetricsClusterPreviewScaleOut       MetricsType = "cluster/preview_scale_out"
	MetricsClusterScaleOut              MetricsType = "cluster/scale_out"
	MetricsClusterClone                 MetricsType = "cluster/clone"
	MetricsClusterRestore               MetricsType = "cluster/restore"
	MetricsClusterTakeover              MetricsType = "cluster/takeover"
	MetricsClusterPreview               MetricsType = "cluster/preview"
	MetricsClusterQuery                 MetricsType = "cluster/query"
	MetricsClusterDetail                MetricsType = "cluster/detail"
	MetricsClusterQueryMonitorAddress   MetricsType = "cluster/query_monitor_address"
	MetricsClusterQueryDashboardAddress MetricsType = "cluster/query_dashboard_address"
	MetricsClusterQueryParameter        MetricsType = "cluster/query_parameter"
	MetricsClusterModifyParameter       MetricsType = "cluster/modify_parameter"
	MetricsClusterInspectParameter      MetricsType = "cluster/inspect_parameter"
	MetricsClusterQueryLogParameter     MetricsType = "cluster/query_logs"

	// MetricsBackupCreate define backup metrics
	MetricsBackupCreate         MetricsType = "backup/create"
	MetricsBackupDelete         MetricsType = "backup/delete"
	MetricsBackupQuery          MetricsType = "backup/query"
	MetricsBackupQueryStrategy  MetricsType = "backup/query_strategy"
	MetricsBackupModifyStrategy MetricsType = "backup/modify_strategy"

	// MetricsDataExport define data export & import metrics
	MetricsDataExport             MetricsType = "data/export"
	MetricsDataImport             MetricsType = "data/import"
	MetricsDataExportImportQuery  MetricsType = "data/query_export_import_record"
	MetricsDataExportImportDelete MetricsType = "data/delete_export_import_record"

	// MetricsPlatformQueryKnowledge define knowledge metrics
	MetricsPlatformQueryKnowledge MetricsType = "platform/query_knowledge"

	// MetricsCDCTaskCreate define cdc metrics
	MetricsCDCTaskCreate MetricsType = "cdc/create"
	MetricsCDCTaskDelete MetricsType = "cdc/delete"
	MetricsCDCTaskPause  MetricsType = "cdc/pause"
	MetricsCDCTaskResume MetricsType = "cdc/resume"
	MetricsCDCTaskUpdate MetricsType = "cdc/update"
	MetricsCDCTaskQuery  MetricsType = "cdc/query"
	MetricsCDCTaskDetail MetricsType = "cdc/detail"
	MetricsCDCDownstream MetricsType = "cdc/downstream/delete"

	// MetricsParameterGroupCreate define parameter group metrics
	MetricsParameterGroupCreate MetricsType = "parameter_group/create"
	MetricsParameterGroupDelete MetricsType = "parameter_group/delete"
	MetricsParameterGroupApply  MetricsType = "parameter_group/apply"
	MetricsParameterGroupCopy   MetricsType = "parameter_group/copy"
	MetricsParameterGroupQuery  MetricsType = "parameter_group/query"
	MetricsParameterGroupDetail MetricsType = "parameter_group/detail"
	MetricsParameterGroupUpdate MetricsType = "parameter_group/update"

	// MetricsUserLogin define user metrics
	MetricsUserLogin   MetricsType = "user/login"
	MetricsUserLogout  MetricsType = "user/logout"
	MetricsUserProfile MetricsType = "user/profile"

	// MetricsWorkFlowQuery define workflow metrics
	MetricsWorkFlowQuery  MetricsType = "workflow/query"
	MetricsWorkFlowDetail MetricsType = "workflow/detail"

	// MetricsResourceQueryHierarchy define resource metrics
	MetricsResourceQueryHierarchy           MetricsType = "resource/query_hierarchy"
	MetricsResourceQueryStocks              MetricsType = "resource/query_stocks"
	MetricsResourceDownloadHostTemplateFile MetricsType = "resource/download_host_template_file"
	MetricsResourceReservedHost             MetricsType = "resource/reserved_host"
	MetricsResourceModifyHostStatus         MetricsType = "resource/modify_host_status"
	MetricsResourceImportHosts              MetricsType = "resource/import_host"
	MetricsResourceDeleteHost               MetricsType = "resource/delete_host"
	MetricsResourceQueryHosts               MetricsType = "resource/query"
)

var EMMetrics = []MetricsType{
	MetricsClusterCreate,
	MetricsClusterDelete,
	MetricsClusterStop,
	MetricsClusterStart,
	MetricsClusterRestart,
	MetricsClusterScaleIn,
	MetricsClusterScaleOut,
	MetricsClusterClone,
	MetricsClusterRestore,
	MetricsClusterTakeover,
	MetricsClusterPreview,
	MetricsClusterQuery,
	MetricsClusterDetail,
	MetricsClusterQueryMonitorAddress,
	MetricsClusterQueryDashboardAddress,
	MetricsClusterQueryParameter,
	MetricsClusterModifyParameter,
	MetricsClusterInspectParameter,
	MetricsClusterQueryLogParameter,

	// MetricsBackupCreate define backup metrics
	MetricsBackupCreate,
	MetricsBackupDelete,
	MetricsBackupQuery,
	MetricsBackupQueryStrategy,
	MetricsBackupModifyStrategy,

	// MetricsDataExport define data export & import metrics
	MetricsDataExport,
	MetricsDataImport,
	MetricsDataExportImportQuery,
	MetricsDataExportImportDelete,

	// MetricsPlatformQueryKnowledge define knowledge metrics
	MetricsPlatformQueryKnowledge,

	// MetricsCDCTaskCreate define cdc metrics
	MetricsCDCTaskCreate,
	MetricsCDCTaskDelete,
	MetricsCDCTaskPause,
	MetricsCDCTaskResume,
	MetricsCDCTaskUpdate,
	MetricsCDCTaskQuery,
	MetricsCDCTaskDetail,
	MetricsCDCDownstream,

	// MetricsParameterGroupCreate define parameter group metrics
	MetricsParameterGroupCreate,
	MetricsParameterGroupDelete,
	MetricsParameterGroupApply,
	MetricsParameterGroupCopy,
	MetricsParameterGroupQuery,
	MetricsParameterGroupDetail,
	MetricsParameterGroupUpdate,

	// MetricsUserLogin define user metrics
	MetricsUserLogin,
	MetricsUserLogout,
	MetricsUserProfile,

	// MetricsWorkFlowQuery define workflow metrics
	MetricsWorkFlowQuery,
	MetricsWorkFlowDetail,

	// MetricsResourceQueryHierarchy define resource metrics
	MetricsResourceQueryHierarchy,
	MetricsResourceQueryStocks,
	MetricsResourceDownloadHostTemplateFile,
	MetricsResourceModifyHostStatus,
	MetricsResourceImportHosts,
	MetricsResourceDeleteHost,
	MetricsResourceQueryHosts,
	MetricsResourceReservedHost,
}
