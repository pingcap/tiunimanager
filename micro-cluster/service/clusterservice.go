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
	changeFeedApi "github.com/pingcap-inc/tiem/apimodels/datatransfer/changefeed"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
	changeFeedManager "github.com/pingcap-inc/tiem/micro-cluster/service/datatransfer/changefeed"
	"net/http"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/library/util/convert"

	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/labstack/gommon/bytes"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/micro-cluster/service/resource"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/adapt"
	user "github.com/pingcap-inc/tiem/micro-cluster/service/user/application"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	userDomain "github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"

	log "github.com/sirupsen/logrus"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &clusterpb.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &clusterpb.ResponseStatusDTO{Code: 500}

type ClusterServiceHandler struct {
	resourceManager   *resource.ResourceManager
	authManager       *user.AuthManager
	tenantManager     *user.TenantManager
	userManager       *user.UserManager
	changeFeedManager *changeFeedManager.ChangeFeedManager
}

func handleResponse(resp *clusterpb.RpcResponse, err error, getData func() ([]byte, error)) {
	if err == nil {
		respData, getDataError := getData()
		if getDataError != nil {
			err = framework.WrapError(common.TIEM_UNRECOGNIZED_ERROR, "", getDataError)
		} else {
			resp.Code = int32(common.TIEM_SUCCESS)
			resp.Response = string(respData)
		}
	}

	if _, ok := err.(framework.TiEMError); !ok {
		err = framework.WrapError(common.TIEM_UNRECOGNIZED_ERROR, "", err)
	}

	resp.Code = int32(err.(framework.TiEMError).GetCode())
	resp.Message = err.(framework.TiEMError).GetMsg()
}

func (handler *ClusterServiceHandler) CreateChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	request.GetOperator()
	reqData := request.GetRequest()

	task := &changeFeedApi.ChangeFeedTask{}

	err := json.Unmarshal([]byte(reqData), task)

	if err != nil {
		handleResponse(response, framework.SimpleError(common.TIEM_PARAMETER_INVALID), nil)
		return nil
	}

	result := handler.changeFeedManager.Create(task.Name)

	handleResponse(response, err, func() ([]byte, error) {
		// todo build api response data
		task.Id = result
		return json.Marshal(task)
	})

	return nil
}

func (handler *ClusterServiceHandler) PauseChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) ResumeChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) DeleteChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) UpdateChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func (handler *ClusterServiceHandler) QueryChangeFeedTasks(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	panic("implement me")
}

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	handler := new(ClusterServiceHandler)
	handler.SetResourceManager(resource.NewResourceManager())
	handler.userManager = user.NewUserManager(adapt.MicroMetaDbRepo{})
	handler.tenantManager = user.NewTenantManager(adapt.MicroMetaDbRepo{})
	handler.authManager = user.NewAuthManager(handler.userManager, adapt.MicroMetaDbRepo{})
	handler.changeFeedManager = changeFeedManager.NewChangeFeedManager()

	domain.InitFlowMap()
	return handler
}

func (handler *ClusterServiceHandler) SetResourceManager(resourceManager *resource.ResourceManager) {
	handler.resourceManager = resourceManager
}

func (handler *ClusterServiceHandler) ResourceManager() *resource.ResourceManager {
	return handler.resourceManager
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

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterpb.ClusterCreateReqDTO, resp *clusterpb.ClusterCreateRespDTO) (err error) {
	framework.LogWithContext(ctx).Info("create cluster")
	clusterAggregation, err := domain.CreateCluster(ctx, req.GetOperator(), req.GetCluster(), req.GetCommonDemand(), req.GetDemands())

	if err != nil {
		framework.LogWithContext(ctx).Info(err)
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		return nil
	}
}

func (c ClusterServiceHandler) ScaleOutCluster(ctx context.Context, req *clusterpb.ScaleOutRequest, resp *clusterpb.ScaleOutResponse) (err error) {
	framework.LogWithContext(ctx).Info("scale out cluster")
	_, err = domain.ScaleOutCluster(ctx, req.GetOperator(), req.GetClusterId(), req.GetDemands())

	if err != nil {
		framework.LogWithContext(ctx).Info(err)
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
	} else {
		resp.RespStatus = SuccessResponseStatus
	}

	return nil
}

func (c ClusterServiceHandler) ScaleInCluster(ctx context.Context, req *clusterpb.ScaleInRequest, resp *clusterpb.ScaleInResponse) (err error) {
	framework.LogWithContext(ctx).Info("scale in cluster")
	_, err = domain.ScaleInCluster(ctx, req.GetOperator(), req.GetClusterId(), req.GetNodeId())

	if err != nil {
		framework.LogWithContext(ctx).Info(err)
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
	} else {
		resp.RespStatus = SuccessResponseStatus
	}

	return nil
}

func (c ClusterServiceHandler) TakeoverClusters(ctx context.Context, req *clusterpb.ClusterTakeoverReqDTO, resp *clusterpb.ClusterTakeoverRespDTO) (err error) {
	framework.LogWithContext(ctx).Info("takeover clusters")
	clusters, err := domain.TakeoverClusters(ctx, req.Operator, req)
	if err != nil {
		framework.LogWithContext(ctx).Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.Clusters = make([]*clusterpb.ClusterDisplayDTO, len(clusters))

		return nil
	}
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	framework.LogWithContext(ctx).Info("query cluster")
	request := &management.QueryReq{}
	err = json.Unmarshal([]byte(req.Request), request)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		return
	}
	clusters, _, err := domain.ListCluster(ctx, request)
	if err != nil {
		resp.Code = int32(err.(framework.TiEMError).GetCode())
		resp.Message = err.(framework.TiEMError).GetMsg()
	} else {
		response := &management.QueryClusterRsp{}
		for _, cluster := range clusters {
			response.Data = append(response.Data, cluster.ExtractDisplayInfo())
		}
		body, err := json.Marshal(response)
		if err != nil {
			resp.Code = http.StatusInternalServerError
			resp.Message = err.Error()
		} else {
			resp.Code = http.StatusOK
			resp.Response = string(body)
		}
	}
	return
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterpb.ClusterDeleteReqDTO, resp *clusterpb.ClusterDeleteRespDTO) (err error) {
	framework.LogWithContext(ctx).Info("delete cluster")

	clusterAggregation, err := domain.DeleteCluster(ctx, req.GetOperator(), req.GetClusterId())
	if err != nil {
		// todo
		framework.LogWithContext(ctx).Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		return nil
	}
}

func (c ClusterServiceHandler) RestartCluster(ctx context.Context, req *clusterpb.ClusterRestartReqDTO, resp *clusterpb.ClusterRestartRespDTO) (err error) {
	framework.LogWithContext(ctx).Info("restart cluster")
	start := time.Now()
	defer handleMetrics(start, "RestartCluster", int(resp.GetRespStatus().GetCode()))

	clusterAggregation, err := domain.RestartCluster(ctx, req.GetOperator(), req.GetClusterId())
	if err != nil {
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
		framework.LogWithContext(ctx).Error(err)
		return nil
	}
	resp.RespStatus = SuccessResponseStatus
	resp.ClusterId = clusterAggregation.Cluster.Id
	return nil
}

func (c ClusterServiceHandler) StopCluster(ctx context.Context, req *clusterpb.ClusterStopReqDTO, resp *clusterpb.ClusterStopRespDTO) (err error) {
	framework.LogWithContext(ctx).Info("stop cluster")
	start := time.Now()
	defer handleMetrics(start, "StopCluster", int(resp.GetRespStatus().GetCode()))

	clusterAggregation, err := domain.StopCluster(ctx, req.GetOperator(), req.GetClusterId())
	if err != nil {
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
		framework.LogWithContext(ctx).Error(err)
		return nil
	}
	resp.RespStatus = SuccessResponseStatus
	resp.ClusterId = clusterAggregation.Cluster.Id
	return nil
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	framework.LogWithContext(ctx).Info("detail cluster")
	type DetailRequest struct {
		clusterID string `json:"clusterId"`
	}
	request := &DetailRequest{}
	err = json.Unmarshal([]byte(req.Request), request)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		return
	}

	cluster, err := domain.GetClusterDetail(ctx, request.clusterID)

	if err != nil {
		resp.Code = int32(err.(framework.TiEMError).GetCode())
		resp.Message = err.(framework.TiEMError).GetMsg()
	} else {
		resp.Code = http.StatusOK
		resp.Response = domain.ExtractClusterInfo(cluster)
	}
	return
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, req *clusterpb.DataExportRequest, resp *clusterpb.DataExportResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ExportData", int(resp.GetRespStatus().GetCode()))
	if err := domain.ExportDataPreCheck(req); err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_EXPORT_PARAM_INVALID), Message: err.Error()}
		return nil
	}

	recordId, err := domain.ExportData(ctx, req)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_EXPORT_PROCESS_FAILED), Message: common.TIEM_EXPORT_PROCESS_FAILED.Explain()}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.RecordId = recordId
	}

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, req *clusterpb.DataImportRequest, resp *clusterpb.DataImportResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ImportData", int(resp.GetRespStatus().GetCode()))
	if err := domain.ImportDataPreCheck(ctx, req); err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_IMPORT_PARAM_INVALID), Message: err.Error()}
		return nil
	}

	recordId, err := domain.ImportData(ctx, req)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_IMPORT_PARAM_INVALID), Message: common.TIEM_IMPORT_PARAM_INVALID.Explain()}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.RecordId = recordId
	}

	return nil
}

func (c ClusterServiceHandler) DescribeDataTransport(ctx context.Context, req *clusterpb.DataTransportQueryRequest, resp *clusterpb.DataTransportQueryResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DescribeDataTransport", int(resp.GetRespStatus().GetCode()))
	infos, page, err := domain.DescribeDataTransportRecord(ctx, req.GetOperator(), req.GetRecordId(), req.GetClusterId(), req.GetReImport(), req.GetStartTime(), req.GetEndTime(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_TRANSPORT_RECORD_NOT_FOUND), Message: common.TIEM_TRANSPORT_RECORD_NOT_FOUND.Explain()}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.PageReq = &clusterpb.PageDTO{
			Page:     page.GetPage(),
			PageSize: page.GetPageSize(),
			Total:    page.GetTotal(),
		}
		resp.TransportInfos = make([]*clusterpb.DataTransportInfo, len(infos))
		for i, v := range infos {
			resp.TransportInfos[i] = &clusterpb.DataTransportInfo{
				RecordId:      infos[i].GetRecord().GetRecordId(),
				ClusterId:     infos[i].GetRecord().GetClusterId(),
				TransportType: infos[i].GetRecord().GetTransportType(),
				StorageType:   infos[i].GetRecord().GetStorageType(),
				FilePath:      infos[i].GetRecord().GetFilePath(),
				ZipName:       infos[i].GetRecord().GetZipName(),
				StartTime:     infos[i].GetRecord().GetStartTime(),
				EndTime:       infos[i].GetRecord().GetEndTime(),
				Comment:       infos[i].GetRecord().GetComment(),
				DisplayStatus: &clusterpb.DisplayStatusDTO{
					StatusCode: strconv.Itoa(int(v.Flow.Status)),
					//StatusName:      v.Flow.StatusAlias,
					StatusName:      domain.TaskStatus(int(v.Flow.Status)).Display(),
					InProcessFlowId: int32(v.Flow.Id),
				},
			}
		}
	}
	return nil
}

func (c ClusterServiceHandler) DeleteDataTransportRecord(ctx context.Context, req *clusterpb.DataTransportDeleteRequest, resp *clusterpb.DataTransportDeleteResponse) error {
	err := domain.DeleteDataTransportRecord(ctx, req.GetOperator(), req.GetClusterId(), req.GetRecordId())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_TRANSPORT_RECORD_DEL_FAIL), Message: common.TIEM_TRANSPORT_RECORD_DEL_FAIL.Explain()}
	} else {
		resp.Status = SuccessResponseStatus
	}

	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterpb.CreateBackupRequest, response *clusterpb.CreateBackupResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "CreateBackup", int(response.GetStatus().GetCode()))
	clusterAggregation, err := domain.Backup(ctx, request.Operator, request.ClusterId, request.BackupMethod, request.BackupType, domain.BackupModeManual, request.FilePath)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_PROCESS_FAILED), Message: common.TIEM_BACKUP_PROCESS_FAILED.Explain()}
	} else {
		response.Status = SuccessResponseStatus
		response.BackupRecord = clusterAggregation.ExtractBackupRecordDTO()
	}
	return nil
}

func (c ClusterServiceHandler) RecoverCluster(ctx context.Context, req *clusterpb.RecoverRequest, resp *clusterpb.RecoverResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "RecoverCluster", int(resp.GetRespStatus().GetCode()))
	if err = domain.RecoverPreCheck(ctx, req); err != nil {
		getLoggerWithContext(ctx).Errorf("recover cluster pre check failed, %s", err.Error())
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_RECOVER_PARAM_INVALID), Message: err.Error()}
		return nil
	}

	clusterAggregation, err := domain.Recover(ctx, req.GetOperator(), req.GetCluster(), req.GetCommonDemand(), req.GetDemands())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_RECOVER_PROCESS_FAILED), Message: common.TIEM_RECOVER_PROCESS_FAILED.Explain()}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
	}
	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *clusterpb.DeleteBackupRequest, response *clusterpb.DeleteBackupResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DeleteBackupRecord", int(response.GetStatus().GetCode()))
	err = domain.DeleteBackup(ctx, request.Operator, request.GetClusterId(), request.GetBackupRecordId())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_RECORD_DELETE_FAILED), Message: common.TIEM_BACKUP_RECORD_DELETE_FAILED.Explain()}
	} else {
		response.Status = SuccessResponseStatus
	}
	return nil
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterpb.SaveBackupStrategyRequest, response *clusterpb.SaveBackupStrategyResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "SaveBackupStrategy", int(response.GetStatus().GetCode()))
	err = domain.SaveBackupStrategyPreCheck(request.GetOperator(), request.GetStrategy())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_STRATEGY_PARAM_INVALID), Message: err.Error()}
		return nil
	}

	err = domain.SaveBackupStrategy(ctx, request.GetOperator(), request.GetStrategy())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_STRATEGY_SAVE_FAILED), Message: common.TIEM_BACKUP_STRATEGY_SAVE_FAILED.Explain()}
	} else {
		response.Status = SuccessResponseStatus
	}
	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterpb.GetBackupStrategyRequest, response *clusterpb.GetBackupStrategyResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "GetBackupStrategy", int(response.GetStatus().GetCode()))
	strategy, err := domain.QueryBackupStrategy(ctx, request.GetOperator(), request.GetClusterId())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_STRATEGY_QUERY_FAILED), Message: common.TIEM_BACKUP_STRATEGY_QUERY_FAILED.Explain()}
	} else {
		response.Status = SuccessResponseStatus
		response.Strategy = strategy
	}
	return nil
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *clusterpb.QueryBackupRequest, response *clusterpb.QueryBackupResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "QueryBackupRecord", int(response.GetStatus().GetCode()))
	result, err := client.DBClient.ListBackupRecords(ctx, &dbpb.DBListBackupRecordsRequest{
		ClusterId: request.ClusterId,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		Page: &dbpb.DBPageDTO{
			Page:     request.Page.Page,
			PageSize: request.Page.PageSize,
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_BACKUP_RECORD_QUERY_FAILED), Message: common.TIEM_BACKUP_RECORD_QUERY_FAILED.Explain()}
	} else {
		response.Status = SuccessResponseStatus
		response.Page = &clusterpb.PageDTO{
			Page:     result.Page.Page,
			PageSize: result.Page.PageSize,
			Total:    result.Page.Total,
		}
		response.BackupRecords = make([]*clusterpb.BackupRecordDTO, len(result.BackupRecords))
		for i, v := range result.BackupRecords {
			response.BackupRecords[i] = &clusterpb.BackupRecordDTO{
				Id:           v.BackupRecord.Id,
				ClusterId:    v.BackupRecord.ClusterId,
				BackupMethod: v.BackupRecord.BackupMethod,
				BackupType:   v.BackupRecord.BackupType,
				BackupMode:   v.BackupRecord.BackupMode,
				FilePath:     v.BackupRecord.FilePath,
				StartTime:    v.Flow.CreateTime,
				EndTime:      v.Flow.UpdateTime,
				Size:         float32(v.BackupRecord.Size) / bytes.MB, //Byte to MByte
				BackupTso:    v.BackupRecord.BackupTso,
				Operator: &clusterpb.OperatorDTO{
					Id: v.BackupRecord.OperatorId,
				},
				DisplayStatus: &clusterpb.DisplayStatusDTO{
					StatusCode: strconv.Itoa(int(v.Flow.Status)),
					//StatusName:      v.Flow.StatusAlias,
					StatusName:      domain.TaskStatus(int(v.Flow.Status)).Display(),
					InProcessFlowId: int32(v.Flow.Id),
				},
			}
		}
	}
	return nil
}

func (c ClusterServiceHandler) QueryParameters(ctx context.Context, request *clusterpb.QueryClusterParametersRequest, response *clusterpb.QueryClusterParametersResponse) (err error) {

	content, err := domain.GetParameters(ctx, request.Operator, request.ClusterId)

	if err != nil {
		framework.LogWithContext(ctx).Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus

		response.ClusterId = request.ClusterId
		response.ParametersJson = content
		return nil
	}
}

func (c ClusterServiceHandler) SaveParameters(ctx context.Context, request *clusterpb.SaveClusterParametersRequest, response *clusterpb.SaveClusterParametersResponse) (err error) {

	//clusterAggregation, err := domain.ModifyParameters(ctx, request.Operator, request.ClusterId, request.ParametersJson)
	//
	//if err != nil {
	//	framework.LogWithContext(ctx).Info(err)
	//	return nil
	//} else {
	//	response.Status = SuccessResponseStatus
	//	response.DisplayInfo = &clusterpb.DisplayStatusDTO{
	//		InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
	//	}
	//	return nil
	//}
	return err
}

func (c ClusterServiceHandler) DescribeDashboard(ctx context.Context, request *clusterpb.DescribeDashboardRequest, response *clusterpb.DescribeDashboardResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DescribeDashboard", int(response.GetStatus().GetCode()))
	info, err := domain.DescribeDashboard(ctx, request.Operator, request.ClusterId)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_DASHBOARD_NOT_FOUND), Message: common.TIEM_DASHBOARD_NOT_FOUND.Explain()}
	} else {
		response.Status = SuccessResponseStatus
		response.ClusterId = info.ClusterId
		response.Url = info.Url
		response.Token = info.Token
	}

	return nil
}

func (c ClusterServiceHandler) DescribeMonitor(ctx context.Context, request *clusterpb.DescribeMonitorRequest, response *clusterpb.DescribeMonitorResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DescribeMonitor", int(response.GetStatus().GetCode()))
	monitor, err := domain.DescribeMonitor(ctx, request.Operator, request.ClusterId)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_MONITOR_NOT_FOUND), Message: common.TIEM_MONITOR_NOT_FOUND.Explain()}
	} else {
		response.Status = SuccessResponseStatus
		response.ClusterId = monitor.ClusterId
		response.AlertUrl = monitor.AlertUrl
		response.GrafanaUrl = monitor.GrafanaUrl
	}

	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, req *clusterpb.ListFlowsRequest, response *clusterpb.ListFlowsResponse) (err error) {
	flows, total, err := domain.TaskRepo.ListFlows(ctx, req.BizId, req.Keyword, int(req.Status), int(req.Page.Page), int(req.Page.PageSize))
	if err != nil {
		framework.LogWithContext(ctx).Error(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.Page = &clusterpb.PageDTO{
		Page:     req.Page.Page,
		PageSize: req.Page.PageSize,
		Total:    int32(total),
	}

	response.Flows = make([]*clusterpb.FlowDTO, len(flows))
	for i, v := range flows {
		response.Flows[i] = &clusterpb.FlowDTO{
			Id:          int64(v.Id),
			FlowName:    v.FlowName,
			StatusAlias: v.StatusAlias,
			BizId:       v.BizId,
			Status:      int32(v.Status),
			StatusName:  v.Status.Display(),
			CreateTime:  v.CreateTime.Unix(),
			UpdateTime:  v.UpdateTime.Unix(),
			Operator: &clusterpb.OperatorDTO{
				Name:           v.Operator.Name,
				Id:             v.Operator.Id,
				TenantId:       v.Operator.TenantId,
				ManualOperator: v.Operator.ManualOperator,
			},
		}
	}
	return err
}

func (c *ClusterServiceHandler) DetailFlow(ctx context.Context, request *clusterpb.DetailFlowRequest, response *clusterpb.DetailFlowsResponse) error {
	flowwork, err := domain.TaskRepo.Load(ctx, uint(request.FlowId))
	if e, ok := err.(framework.TiEMError); ok {
		response.Status = &clusterpb.ResponseStatusDTO{
			Code:    int32(e.GetCode()),
			Message: e.GetMsg(),
		}
	} else {
		response.Status = SuccessResponseStatus
		response.Flow = &clusterpb.FlowWithTaskDTO{
			Flow: &clusterpb.FlowDTO{
				Id:          int64(flowwork.FlowWork.Id),
				FlowName:    flowwork.FlowWork.FlowName,
				StatusAlias: flowwork.FlowWork.StatusAlias,
				BizId:       flowwork.FlowWork.BizId,
				Status:      int32(flowwork.FlowWork.Status),
				StatusName:  flowwork.FlowWork.Status.Display(),
				CreateTime:  flowwork.FlowWork.CreateTime.Unix(),
				UpdateTime:  flowwork.FlowWork.UpdateTime.Unix(),
				Operator: &clusterpb.OperatorDTO{
					Name:           flowwork.FlowWork.Operator.Name,
					Id:             flowwork.FlowWork.Operator.Id,
					TenantId:       flowwork.FlowWork.Operator.TenantId,
					ManualOperator: flowwork.FlowWork.Operator.ManualOperator,
				},
			},
			TaskDef: flowwork.GetAllTaskDef(),
			Tasks:   flowwork.ExtractTaskDTO(),
		}
	}

	return nil
}

var ManageSuccessResponseStatus = &clusterpb.ManagerResponseStatus{
	Code: 0,
}

func (p *ClusterServiceHandler) Login(ctx context.Context, req *clusterpb.LoginRequest, resp *clusterpb.LoginResponse) error {
	log := framework.LogWithContext(ctx).WithField("fp", "ClusterServiceHandler.Login")
	log.Debug("req:", req)
	token, err := p.authManager.Login(ctx, req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &clusterpb.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
		log.Error("resp:", resp)
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TokenString = token
		log.Debug("resp:", resp)
	}
	return nil

}

func (p *ClusterServiceHandler) Logout(ctx context.Context, req *clusterpb.LogoutRequest, resp *clusterpb.LogoutResponse) error {
	accountName, err := p.authManager.Logout(ctx, req.TokenString)
	if err != nil {
		resp.Status = &clusterpb.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.AccountName = accountName
	}
	return nil

}

func (p *ClusterServiceHandler) VerifyIdentity(ctx context.Context, req *clusterpb.VerifyIdentityRequest, resp *clusterpb.VerifyIdentityResponse) error {
	tenantId, accountId, accountName, err := p.authManager.Accessible(ctx, req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*userDomain.UnauthorizedError); ok {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*userDomain.ForbiddenError); ok {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusForbidden,
				Message: "无权限",
			}
		} else {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TenantId = tenantId
		resp.AccountId = accountId
		resp.AccountName = accountName
	}

	return nil
}

func (c *ClusterServiceHandler) CreateParamGroup(ctx context.Context, req *clusterpb.CreateParamGroupRequest, resp *clusterpb.CreateParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.CreateParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) UpdateParamGroup(ctx context.Context, req *clusterpb.UpdateParamGroupRequest, resp *clusterpb.UpdateParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.UpdateParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) DeleteParamGroup(ctx context.Context, req *clusterpb.DeleteParamGroupRequest, resp *clusterpb.DeleteParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.DeleteParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) ListParamGroup(ctx context.Context, req *clusterpb.ListParamGroupRequest, resp *clusterpb.ListParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ListParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.ListParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) DetailParamGroup(ctx context.Context, req *clusterpb.DetailParamGroupRequest, resp *clusterpb.DetailParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DetailParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.DetailParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) ApplyParamGroup(ctx context.Context, req *clusterpb.ApplyParamGroupRequest, resp *clusterpb.ApplyParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ApplyParamGroup", int(resp.GetRespStatus().GetCode()))

	// query params list by param group id
	pgDetail := &clusterpb.DetailParamGroupResponse{}
	err := domain.DetailParamGroup(ctx, &clusterpb.DetailParamGroupRequest{ParamGroupId: req.ParamGroupId}, pgDetail)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply param group err: %v", err)
		return err
	}

	params := make([]*domain.ApplyParam, len(pgDetail.ParamGroup.Params))
	err = convert.ConvertObj(pgDetail.ParamGroup.Params, &params)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply param group convert obj err: %v", err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_CONVERT_OBJ_FAILED), Message: common.TIEM_CONVERT_OBJ_FAILED.Explain() + err.Error()}
		return err
	}
	// Convert the default value of the parameter group to the real value of the modified parameter
	for i, param := range pgDetail.ParamGroup.Params {
		params[i].RealValue.Cluster = param.DefaultValue
	}
	modifyParam := &domain.ModifyParam{NeedReboot: req.NeedReboot, Params: params}
	clusterAggregation, err := domain.ModifyParameters(ctx, req.Operator, req.ClusterId, modifyParam)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply param group modify parameters err: %v", err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_MODIFY_PARAM_FAILED), Message: common.TIEM_MODIFY_PARAM_FAILED.Explain() + err.Error()}
		return nil
	} else {
		resp.DisplayInfo = &clusterpb.DisplayStatusDTO{
			StatusCode:      strconv.Itoa(int(clusterAggregation.CurrentWorkFlow.Status)),
			StatusName:      clusterAggregation.CurrentWorkFlow.Status.Display(),
			InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
			CreateTime:      clusterAggregation.CurrentWorkFlow.CreateTime.Unix(),
			UpdateTime:      clusterAggregation.CurrentWorkFlow.UpdateTime.Unix(),
		}
		return domain.ApplyParamGroup(ctx, req, resp)
	}
}

func (c *ClusterServiceHandler) CopyParamGroup(ctx context.Context, req *clusterpb.CopyParamGroupRequest, resp *clusterpb.CopyParamGroupResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CopyParamGroup", int(resp.GetRespStatus().GetCode()))
	return domain.CopyParamGroup(ctx, req, resp)
}

func (c *ClusterServiceHandler) ListClusterParams(ctx context.Context, req *clusterpb.ListClusterParamsRequest, resp *clusterpb.ListClusterParamsResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ListClusterParams", int(resp.GetRespStatus().GetCode()))
	return domain.ListClusterParams(ctx, req, resp)
}

func (c *ClusterServiceHandler) UpdateClusterParams(ctx context.Context, req *clusterpb.UpdateClusterParamsRequest, resp *clusterpb.UpdateClusterParamsResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateClusterParams", int(resp.GetRespStatus().GetCode()))

	params := make([]*domain.ApplyParam, len(req.Params))
	err := convert.ConvertObj(req.Params, &params)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster params convert obj err: %v", err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_CONVERT_OBJ_FAILED), Message: common.TIEM_CONVERT_OBJ_FAILED.Explain() + err.Error()}
		return err
	}

	modifyParam := &domain.ModifyParam{NeedReboot: req.NeedReboot, Params: params}
	clusterAggregation, err := domain.ModifyParameters(ctx, req.Operator, req.ClusterId, modifyParam)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster params modify parameters err: %v", err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: int32(common.TIEM_MODIFY_PARAM_FAILED), Message: common.TIEM_MODIFY_PARAM_FAILED.Explain() + err.Error()}
		return nil
	} else {
		resp.DisplayInfo = &clusterpb.DisplayStatusDTO{
			StatusCode:      strconv.Itoa(int(clusterAggregation.CurrentWorkFlow.Status)),
			StatusName:      clusterAggregation.CurrentWorkFlow.Status.Display(),
			InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
			CreateTime:      clusterAggregation.CurrentWorkFlow.CreateTime.Unix(),
			UpdateTime:      clusterAggregation.CurrentWorkFlow.UpdateTime.Unix(),
		}
		return domain.UpdateClusterParams(ctx, req, resp)
	}
}

func (c *ClusterServiceHandler) InspectClusterParams(ctx context.Context, req *clusterpb.InspectClusterParamsRequest, resp *clusterpb.InspectClusterParamsResponse) error {
	start := time.Now()
	defer handleMetrics(start, "InspectClusterParams", int(resp.GetRespStatus().GetCode()))
	return domain.InspectClusterParams(ctx, req, resp)
}

func (clusterManager *ClusterServiceHandler) ImportHost(ctx context.Context, in *clusterpb.ImportHostRequest, out *clusterpb.ImportHostResponse) error {
	return clusterManager.resourceManager.ImportHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) ImportHostsInBatch(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, out *clusterpb.ImportHostsInBatchResponse) error {
	return clusterManager.resourceManager.ImportHostsInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RemoveHost(ctx context.Context, in *clusterpb.RemoveHostRequest, out *clusterpb.RemoveHostResponse) error {
	return clusterManager.resourceManager.RemoveHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RemoveHostsInBatch(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, out *clusterpb.RemoveHostsInBatchResponse) error {
	return clusterManager.resourceManager.RemoveHostsInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) ListHost(ctx context.Context, in *clusterpb.ListHostsRequest, out *clusterpb.ListHostsResponse) error {
	return clusterManager.resourceManager.ListHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) CheckDetails(ctx context.Context, in *clusterpb.CheckDetailsRequest, out *clusterpb.CheckDetailsResponse) error {
	return clusterManager.resourceManager.CheckDetails(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) AllocHosts(ctx context.Context, in *clusterpb.AllocHostsRequest, out *clusterpb.AllocHostResponse) error {
	return clusterManager.resourceManager.AllocHosts(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) GetFailureDomain(ctx context.Context, in *clusterpb.GetFailureDomainRequest, out *clusterpb.GetFailureDomainResponse) error {
	return clusterManager.resourceManager.GetFailureDomain(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) AllocResourcesInBatch(ctx context.Context, in *clusterpb.BatchAllocRequest, out *clusterpb.BatchAllocResponse) error {
	return clusterManager.resourceManager.AllocResourcesInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RecycleResources(ctx context.Context, in *clusterpb.RecycleRequest, out *clusterpb.RecycleResponse) error {
	return clusterManager.resourceManager.RecycleResources(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) UpdateHostStatus(ctx context.Context, in *clusterpb.UpdateHostStatusRequest, out *clusterpb.UpdateHostStatusResponse) error {
	return clusterManager.resourceManager.UpdateHostStatus(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) ReserveHost(ctx context.Context, in *clusterpb.ReserveHostRequest, out *clusterpb.ReserveHostResponse) error {
	return clusterManager.resourceManager.ReserveHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) GetHierarchy(ctx context.Context, in *clusterpb.GetHierarchyRequest, out *clusterpb.GetHierarchyResponse) error {
	return clusterManager.resourceManager.GetHierarchy(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) GetStocks(ctx context.Context, in *clusterpb.GetStocksRequest, out *clusterpb.GetStocksResponse) error {
	return clusterManager.resourceManager.GetStocks(ctx, in, out)
}
