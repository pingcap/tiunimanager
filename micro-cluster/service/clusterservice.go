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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/micro-cluster/user/identification"
	"github.com/pingcap-inc/tiem/micro-cluster/user/tenant"
	"github.com/pingcap-inc/tiem/micro-cluster/user/userinfo"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pingcap-inc/tiem/micro-cluster/platform/config"

	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/changefeed"
	clusterLog "github.com/pingcap-inc/tiem/micro-cluster/cluster/log"
	clusterManager "github.com/pingcap-inc/tiem/micro-cluster/cluster/management"
	clusterParameter "github.com/pingcap-inc/tiem/micro-cluster/cluster/parameter"
	"github.com/pingcap-inc/tiem/micro-cluster/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/micro-cluster/parametergroup"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"

	"github.com/pingcap-inc/tiem/library/framework"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

type ClusterServiceHandler struct {
	resourceManager         *resourcemanager.ResourceManager
	changeFeedManager       *changefeed.Manager
	parameterGroupManager   *parametergroup.Manager
	clusterParameterManager *clusterParameter.Manager
	clusterManager          *clusterManager.Manager
	systemConfigManager     *config.SystemConfigManager
	brManager               backuprestore.BRService
	importexportManager     importexport.ImportExportService
	clusterLogManager       *clusterLog.Manager
	tenantManager           *tenant.Manager
	accountManager          *userinfo.Manager
	authManager             *identification.Manager
}

func handleRequest(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse, requestBody interface{}) bool {
	err := json.Unmarshal([]byte(req.GetRequest()), requestBody)
	if err != nil {
		errMsg := fmt.Sprintf("unmarshal request error, request = %s, err = %s", req.GetRequest(), err.Error())
		handleResponse(ctx, resp, framework.NewTiEMErrorf(common.TIEM_UNMARSHAL_ERROR, errMsg), nil, nil)
		return false
	} else {
		return true
	}
}

func handleResponse(ctx context.Context, resp *clusterpb.RpcResponse, err error, responseData interface{}, page *clusterpb.RpcPage) {
	if err == nil {
		data, getDataError := json.Marshal(responseData)
		if getDataError != nil {
			// deal with err uniformly later
			err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, fmt.Sprintf("marshal request data error, data = %v", responseData), getDataError)
		} else {
			// handle data and page
			resp.Code = int32(common.TIEM_SUCCESS)
			resp.Response = string(data)
			if page != nil {
				resp.Page = page
			}
			return
		}
	}

	if err != nil {
		if finalError, ok := err.(framework.TiEMError); ok {
			framework.LogWithContext(ctx).Errorf("rpc method failed with error, %s", err.Error())
			resp.Code = int32(finalError.GetCode())
			resp.Message = finalError.GetMsg()
			return
		}
		if finalError, ok := err.(errors.EMError); ok {
			framework.LogWithContext(ctx).Errorf("rpc method failed with error, %s", err.Error())
			resp.Code = int32(finalError.GetCode())
			resp.Message = finalError.GetMsg()
			return
		} else {
			resp.Code = int32(errors.TIEM_UNRECOGNIZED_ERROR)
			resp.Message = err.Error()
		}

		return
	}
}

func getLoggerWithContext(ctx context.Context) *log.Entry {
	return framework.LogWithContext(ctx)
}

func handleMetrics(start time.Time, funcName string, code int) {

	duration := time.Since(start)
	framework.Current.GetMetrics().MicroDurationHistogramMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Observe(duration.Seconds())
	framework.Current.GetMetrics().MicroRequestsCounterMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Inc()
}

/*
// handlePanic
// @Description: recover from any panic from user request
// @Parameter ctx
// @Parameter funcName
// @Parameter resp
func handlePanic(ctx context.Context, funcName string, resp *clusterpb.RpcResponse)  {
	if r := recover(); r != nil {
		framework.LogWithContext(ctx).Errorf("recover from %s", funcName)
		resp.Code = int32(errors.TIEM_PANIC)
		resp.Message = fmt.Sprintf("%v", r)
	}
}
*/

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	handler := new(ClusterServiceHandler)
	handler.resourceManager = resourcemanager.NewResourceManager()
	handler.changeFeedManager = changefeed.NewManager()
	handler.parameterGroupManager = parametergroup.NewManager()
	handler.clusterParameterManager = clusterParameter.NewManager()
	handler.clusterManager = clusterManager.NewClusterManager()
	handler.systemConfigManager = config.NewSystemConfigManager()
	handler.brManager = backuprestore.GetBRService()
	handler.importexportManager = importexport.GetImportExportService()
	handler.clusterLogManager = clusterLog.NewManager()
	handler.tenantManager = tenant.NewTenantManager()
	handler.accountManager = userinfo.NewAccountManager()
	handler.authManager = identification.NewIdentificationManager()

	return handler
}

func (handler *ClusterServiceHandler) CreateChangeFeedTask(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateChangeFeedTask", int(resp.GetCode()))

	request := cluster.CreateChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Create(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) PauseChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "PauseChangeFeedTask", int(response.GetCode()))

	panic("implement me")
}

func (handler *ClusterServiceHandler) ResumeChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ResumeChangeFeedTask", int(response.GetCode()))

	panic("implement me")
}

func (handler *ClusterServiceHandler) DeleteChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteChangeFeedTask", int(response.GetCode()))

	panic("implement me")
}

func (handler *ClusterServiceHandler) UpdateChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateChangeFeedTask", int(response.GetCode()))

	panic("implement me")
}

func (handler *ClusterServiceHandler) QueryChangeFeedTasks(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryChangeFeedTasks", int(response.GetCode()))

	panic("implement me")
}

func (handler *ClusterServiceHandler) CreateParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()

	defer handleMetrics(start, "CreateParameterGroup", int(resp.GetCode()))

	request := &message.CreateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CreateParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateParameterGroup", int(resp.GetCode()))

	request := &message.UpdateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.UpdateParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) DeleteParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteParameterGroup", int(resp.GetCode()))

	request := &message.DeleteParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DeleteParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryParameterGroup", int(resp.GetCode()))

	request := &message.QueryParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.parameterGroupManager.QueryParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) DetailParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DetailParameterGroup", int(resp.GetCode()))

	request := &message.DetailParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DetailParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) ApplyParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ApplyParameterGroup", int(resp.GetCode()))

	request := &message.ApplyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.ApplyParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CopyParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CopyParameterGroup", int(resp.GetCode()))

	request := &message.CopyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CopyParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryClusterParameters", int(resp.GetCode()))

	request := &cluster.QueryClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterParameterManager.QueryClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateClusterParameters", int(resp.GetCode()))

	request := &cluster.UpdateClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.UpdateClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request, true)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) InspectClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "InspectClusterParameters", int(resp.GetCode()))

	request := &cluster.InspectClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.InspectClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterLog(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryClusterLog", int(resp.GetCode()))

	request := &cluster.QueryClusterLogReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterLogManager.QueryClusterLog(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "CreateCluster", int(resp.GetCode()))

	request := cluster.CreateClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.CreateCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestoreNewCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "RestoreNewCluster", int(resp.GetCode()))

	request := cluster.RestoreNewClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestoreNewCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleOutCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ScaleOutCluster", int(resp.GetCode()))

	request := cluster.ScaleOutClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleOut(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleInCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ScaleInCluster", int(resp.GetCode()))

	request := cluster.ScaleInClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleIn(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CloneCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CloneCluster", int(resp.GetCode()))

	request := cluster.CloneClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.Clone(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler ClusterServiceHandler) TakeoverClusters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "TakeoverClusters", int(resp.GetCode()))

	request := cluster.TakeoverClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.Takeover(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "QueryCluster", int(resp.GetCode()))

	request := cluster.QueryClustersReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, total, err := c.clusterManager.QueryCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, &clusterpb.RpcPage{
			Page:     int32(request.Page),
			PageSize: int32(request.PageSize),
			Total:    int32(total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteCluster", int(resp.GetCode()))

	request := cluster.DeleteClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DeleteCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestartCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "RestartCluster", int(resp.GetCode()))

	request := cluster.RestartClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestartCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) StopCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "StopCluster", int(resp.GetCode()))

	request := cluster.StopClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.StopCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DetailCluster", int(resp.GetCode()))

	request := cluster.QueryClusterDetailReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DetailCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ExportData", int(response.GetCode()))

	exportReq := message.DataExportReq{}

	if handleRequest(ctx, request, response, &exportReq) {
		result, err := c.importexportManager.ExportData(framework.NewBackgroundMicroCtx(ctx, false), exportReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ImportData", int(response.GetCode()))

	importReq := message.DataImportReq{}

	if handleRequest(ctx, request, response, &importReq) {
		result, err := c.importexportManager.ImportData(framework.NewBackgroundMicroCtx(ctx, false), importReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryDataTransport(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryDataTransport", int(response.GetCode()))

	queryReq := message.QueryDataImportExportRecordsReq{}

	if handleRequest(ctx, request, response, &queryReq) {
		result, page, err := c.importexportManager.QueryDataTransportRecords(framework.NewBackgroundMicroCtx(ctx, false), queryReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteDataTransportRecord(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteDataTransportRecord", int(response.GetCode()))

	deleteReq := message.DeleteImportExportRecordReq{}

	if handleRequest(ctx, request, response, &deleteReq) {
		result, err := c.importexportManager.DeleteDataTransportRecord(framework.NewBackgroundMicroCtx(ctx, false), deleteReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) GetSystemConfig(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetSystemConfig", int(response.GetCode()))

	getReq := message.GetSystemConfigReq{}

	if handleRequest(ctx, request, response, &getReq) {
		result, err := c.systemConfigManager.GetSystemConfig(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateBackup", int(response.GetCode()))

	backupReq := cluster.BackupClusterDataReq{}

	if handleRequest(ctx, request, response, &backupReq) {
		result, err := c.brManager.BackupCluster(framework.NewBackgroundMicroCtx(ctx, false), backupReq, true)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecords(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteBackupRecord", int(response.GetCode()))

	deleteReq := cluster.DeleteBackupDataReq{}

	if handleRequest(ctx, request, response, &deleteReq) {
		result, err := c.brManager.DeleteBackupRecords(framework.NewBackgroundMicroCtx(ctx, false), deleteReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "SaveBackupStrategy", int(response.GetCode()))

	saveReq := cluster.SaveBackupStrategyReq{}

	if handleRequest(ctx, request, response, &saveReq) {
		result, err := c.brManager.SaveBackupStrategy(framework.NewBackgroundMicroCtx(ctx, false), saveReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetBackupStrategy", int(response.GetCode()))

	getReq := cluster.GetBackupStrategyReq{}

	if handleRequest(ctx, request, response, &getReq) {
		result, err := c.brManager.GetBackupStrategy(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryBackupRecords(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "QueryBackupRecords", int(response.GetCode()))

	queryReq := cluster.QueryBackupRecordsReq{}

	if handleRequest(ctx, request, response, &queryReq) {
		result, page, err := c.brManager.QueryClusterBackupRecords(framework.NewBackgroundMicroCtx(ctx, false), queryReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) GetDashboardInfo(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DescribeDashboard", int(response.GetCode()))

	dashboardReq := cluster.GetDashboardInfoReq{}

	if handleRequest(ctx, request, response, &dashboardReq) {
		result, err := c.clusterManager.GetClusterDashboardInfo(framework.NewBackgroundMicroCtx(ctx, false), dashboardReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetMonitorInfo(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "GetMonitorInfo", int(resp.GetCode()))

	request := &cluster.QueryMonitorInfoReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := c.clusterManager.GetMonitorInfo(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ListFlows", int(response.GetCode()))

	listReq := message.QueryWorkFlowsReq{}
	if handleRequest(ctx, request, response, &listReq) {
		manager := workflow.GetWorkFlowService()
		result, page, err := manager.ListWorkFlows(framework.NewBackgroundMicroCtx(ctx, false), listReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c *ClusterServiceHandler) DetailFlow(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DetailFlow", int(response.GetCode()))

	detailReq := message.QueryWorkFlowDetailReq{}
	if handleRequest(ctx, request, response, &detailReq) {
		manager := workflow.GetWorkFlowService()
		result, err := manager.DetailWorkFlow(framework.NewBackgroundMicroCtx(ctx, false), detailReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Login(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "Login", int(response.GetCode()))

	loginReq := message.LoginReq{}
	if handleRequest(ctx, request, response, &loginReq) {
		result, err := c.authManager.Login(framework.NewBackgroundMicroCtx(ctx, false), loginReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Logout(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "Logout", int(response.GetCode()))

	logoutReq := message.LogoutReq{}
	if handleRequest(ctx, request, response, &logoutReq) {
		result, err := c.authManager.Logout(framework.NewBackgroundMicroCtx(ctx, false), logoutReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) VerifyIdentity(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "VerifyIdentity", int(response.GetCode()))

	verReq := message.AccessibleReq{}
	if handleRequest(ctx, request, response, &verReq) {
		result, err := c.authManager.Accessible(framework.NewBackgroundMicroCtx(ctx, false), verReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ImportHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ImportHosts", int(response.GetCode()))

	reqStruct := message.ImportHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		hostIds, err := handler.resourceManager.ImportHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.Hosts)
		var rsp message.ImportHostsResp
		if err == nil {
			rsp.HostIDS = hostIds
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteHosts", int(response.GetCode()))

	reqStruct := message.DeleteHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.DeleteHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs)
		var rsp message.DeleteHostsResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryHosts", int(response.GetCode()))

	reqStruct := message.QueryHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		filter := reqStruct.GetHostFilter()
		page := reqStruct.GetPage()

		hosts, err := handler.resourceManager.QueryHosts(framework.NewBackgroundMicroCtx(ctx, false), filter, page)
		var rsp message.QueryHostsResp
		if err == nil {
			rsp.Hosts = hosts
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostReserved(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateHostReserved", int(response.GetCode()))

	reqStruct := message.UpdateHostReservedReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.UpdateHostReserved(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Reserved)
		var rsp message.UpdateHostReservedResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostStatus(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateHostStatus", int(response.GetCode()))

	reqStruct := message.UpdateHostStatusReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.UpdateHostStatus(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Status)
		var rsp message.UpdateHostStatusResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetHierarchy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetHierarchy", int(response.GetCode()))

	reqStruct := message.GetHierarchyReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		filter := reqStruct.GetHostFilter()

		root, err := handler.resourceManager.GetHierarchy(framework.NewBackgroundMicroCtx(ctx, false), filter, reqStruct.Level, reqStruct.Depth)
		var rsp message.GetHierarchyResp
		if err == nil {
			rsp.Root = *root
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetStocks(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetStocks", int(response.GetCode()))

	reqStruct := message.GetStocksReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		location := reqStruct.GetLocation()
		hostFilter := reqStruct.GetHostFilter()
		diskFilter := reqStruct.GetDiskFilter()

		stocks, err := handler.resourceManager.GetStocks(framework.NewBackgroundMicroCtx(ctx, false), location, hostFilter, diskFilter)
		var rsp message.GetStocksResp
		if err == nil {
			rsp.Stocks = *stocks
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}
