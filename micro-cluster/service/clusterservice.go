
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

package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"net/http"
	"strconv"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap-inc/tiem/micro-cluster/service/host"
	domain2 "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/domain"
	log "github.com/sirupsen/logrus"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &clusterpb.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &clusterpb.ResponseStatusDTO{Code: 500}

type ClusterServiceHandler struct {
	resourceManager *host.ResourceManager
}

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	handler := new(ClusterServiceHandler)
	resourceManager := host.NewResourceManager()
	handler.SetResourceManager(resourceManager)
	return handler
}

func (handler *ClusterServiceHandler) SetResourceManager(resourceManager *host.ResourceManager) {
	handler.resourceManager = resourceManager
}

func (handler *ClusterServiceHandler) ResourceManager() *host.ResourceManager {
	return handler.resourceManager
}

func getLogger() *log.Entry {
	return framework.Log()
}

func getLoggerWithContext(ctx context.Context) *log.Entry {
	return framework.LogWithContext(ctx)
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterpb.ClusterCreateReqDTO, resp *clusterpb.ClusterCreateRespDTO) (err error) {
	getLogger().Info("create cluster")
	clusterAggregation, err := domain.CreateCluster(req.GetOperator(), req.GetCluster(), req.GetDemands())

	if err != nil {
		getLogger().Info(err)
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.BaseInfo = clusterAggregation.ExtractBaseInfoDTO()
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterpb.ClusterQueryReqDTO, resp *clusterpb.ClusterQueryRespDTO) (err error) {
	getLogger().Info("query cluster")
	clusters, total, err := domain.ListCluster(req.Operator, req)
	if err != nil {
		getLogger().Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.Clusters = make([]*clusterpb.ClusterDisplayDTO, len(clusters), len(clusters))
		for i, v := range clusters {
			resp.Clusters[i] = v.ExtractDisplayDTO()
		}
		resp.Page = &clusterpb.PageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterpb.ClusterDeleteReqDTO, resp *clusterpb.ClusterDeleteRespDTO) (err error) {
	getLogger().Info("delete cluster")

	clusterAggregation, err := domain.DeleteCluster(req.GetOperator(), req.GetClusterId())
	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterpb.ClusterDetailReqDTO, resp *clusterpb.ClusterDetailRespDTO) (err error) {
	getLogger().Info("detail cluster")

	cluster, err := domain.GetClusterDetail(req.Operator, req.ClusterId)

	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.DisplayInfo = cluster.ExtractDisplayDTO()
		resp.Components = cluster.ExtractComponentDTOs()
		resp.MaintenanceInfo = cluster.ExtractMaintenanceDTO()

		return nil
	}
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, req *clusterpb.DataExportRequest, resp *clusterpb.DataExportResponse) error {
	if err := domain.ExportDataPreCheck(req); err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_EXPORT_PARAM_INVALID, Message: err.Error()}
		return nil
	}

	recordId, err := domain.ExportData(ctx, req)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_EXPORT_PROCESS_FAILED, Message: common.TiEMErrMsg[common.TIEM_EXPORT_PROCESS_FAILED]}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.RecordId = recordId
	}

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, req *clusterpb.DataImportRequest, resp *clusterpb.DataImportResponse) error {
	if err := domain.ImportDataPreCheck(req); err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_IMPORT_PARAM_INVALID, Message: err.Error()}
		return nil
	}

	recordId, err := domain.ImportData(ctx, req)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_IMPORT_PARAM_INVALID, Message: common.TiEMErrMsg[common.TIEM_IMPORT_PARAM_INVALID]}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.RecordId = recordId
	}

	return nil
}

func (c ClusterServiceHandler) DescribeDataTransport(ctx context.Context, req *clusterpb.DataTransportQueryRequest, resp *clusterpb.DataTransportQueryResponse) error {
	infos, page, err := domain.DescribeDataTransportRecord(ctx, req.GetOperator(), req.GetRecordId(), req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_TRANSPORT_RECORD_NOT_FOUND, Message: common.TiEMErrMsg[common.TIEM_TRANSPORT_RECORD_NOT_FOUND]}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.PageReq = &clusterpb.PageDTO{
			Page:     page.GetPage(),
			PageSize: page.GetPageSize(),
			Total:    page.GetTotal(),
		}
		resp.TransportInfos = make([]*clusterpb.DataTransportInfo, len(infos))
		for index := 0; index < len(infos); index++ {
			resp.TransportInfos[index] = &clusterpb.DataTransportInfo{
				RecordId:      infos[index].ID,
				ClusterId:     infos[index].ClusterId,
				TransportType: infos[index].TransportType,
				FilePath:      infos[index].FilePath,
				Status:        infos[index].Status,
				StartTime:     infos[index].StartTime,
				EndTime:       infos[index].EndTime,
			}
		}
	}
	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterpb.CreateBackupRequest, response *clusterpb.CreateBackupResponse) (err error) {
	clusterAggregation, err := domain.Backup(ctx, request.Operator, request.ClusterId, request.BackupMethod, request.BackupType, domain.BackupModeManual, request.FilePath)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_PROCESS_FAILED, Message: common.TiEMErrMsg[common.TIEM_BACKUP_PROCESS_FAILED]}
	} else {
		response.Status = SuccessResponseStatus
		response.BackupRecord = clusterAggregation.ExtractBackupRecordDTO()
	}
	return nil
}

func (c ClusterServiceHandler) RecoverCluster(ctx context.Context, req *clusterpb.RecoverRequest, resp *clusterpb.RecoverResponse) (err error) {
	if err = domain.RecoverPreCheck(req); err != nil {
		getLoggerWithContext(ctx).Errorf("recover cluster pre check failed, %s", err.Error())
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_RECOVER_PARAM_INVALID, Message: err.Error()}
		return nil
	}

	clusterAggregation, err := domain.Recover(ctx, req.GetOperator(), req.GetCluster(), req.GetDemands())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: common.TIEM_RECOVER_PROCESS_FAILED, Message: common.TiEMErrMsg[common.TIEM_RECOVER_PROCESS_FAILED]}
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.BaseInfo = clusterAggregation.ExtractBaseInfoDTO()
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
	}
	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *clusterpb.DeleteBackupRequest, response *clusterpb.DeleteBackupResponse) (err error) {
	err = domain.DeleteBackup(ctx, request.Operator, request.GetClusterId(), request.GetBackupRecordId())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_RECORD_DELETE_FAILED, Message: common.TiEMErrMsg[common.TIEM_BACKUP_RECORD_DELETE_FAILED]}
	} else {
		response.Status = SuccessResponseStatus
	}
	return nil
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterpb.SaveBackupStrategyRequest, response *clusterpb.SaveBackupStrategyResponse) (err error) {
	err = domain.SaveBackupStrategyPreCheck(request.GetOperator(), request.GetStrategy())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_STRATEGY_PARAM_INVALID, Message: err.Error()}
		return nil
	}

	err = domain.SaveBackupStrategy(ctx, request.GetOperator(), request.GetStrategy())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_STRATEGY_SAVE_FAILED, Message: common.TiEMErrMsg[common.TIEM_BACKUP_STRATEGY_SAVE_FAILED]}
	} else {
		response.Status = SuccessResponseStatus
	}
	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterpb.GetBackupStrategyRequest, response *clusterpb.GetBackupStrategyResponse) (err error) {
	strategy, err := domain.QueryBackupStrategy(ctx, request.GetOperator(), request.GetClusterId())
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_STRATEGY_QUERY_FAILED, Message: common.TiEMErrMsg[common.TIEM_BACKUP_STRATEGY_QUERY_FAILED]}
	} else {
		response.Status = SuccessResponseStatus
		response.Strategy = strategy
	}
	return nil
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *clusterpb.QueryBackupRequest, response *clusterpb.QueryBackupResponse) (err error) {
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
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_BACKUP_RECORD_QUERY_FAILED, Message: common.TiEMErrMsg[common.TIEM_BACKUP_RECORD_QUERY_FAILED]}
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
				Size:         v.BackupRecord.Size,
				Operator: &clusterpb.OperatorDTO{
					Id: v.BackupRecord.OperatorId,
				},
				DisplayStatus: &clusterpb.DisplayStatusDTO{
					StatusCode:      strconv.Itoa(int(v.Flow.Status)),
					StatusName:      v.Flow.StatusAlias,
					InProcessFlowId: int32(v.Flow.Id),
				},
			}
		}
	}
	return nil
}

func (c ClusterServiceHandler) QueryParameters(ctx context.Context, request *clusterpb.QueryClusterParametersRequest, response *clusterpb.QueryClusterParametersResponse) (err error) {

	content, err := domain.GetParameters(request.Operator, request.ClusterId)

	if err != nil {
		getLogger().Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus

		response.ClusterId = request.ClusterId
		response.ParametersJson = content
		return nil
	}
}

func (c ClusterServiceHandler) SaveParameters(ctx context.Context, request *clusterpb.SaveClusterParametersRequest, response *clusterpb.SaveClusterParametersResponse) (err error) {

	clusterAggregation, err := domain.ModifyParameters(request.Operator, request.ClusterId, request.ParametersJson)

	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.DisplayInfo = &clusterpb.DisplayStatusDTO{
			InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DescribeDashboard(ctx context.Context, request *clusterpb.DescribeDashboardRequest, response *clusterpb.DescribeDashboardResponse) (err error) {
	info, err := domain.DescribeDashboard(ctx, request.Operator, request.ClusterId)
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		response.Status = &clusterpb.ResponseStatusDTO{Code: common.TIEM_DASHBOARD_NOT_FOUND, Message: common.TiEMErrMsg[common.TIEM_DASHBOARD_NOT_FOUND]}
	} else {
		response.Status = SuccessResponseStatus
		response.ClusterId = info.ClusterId
		response.Url = info.Url
		response.Token = info.Token
	}

	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, req *clusterpb.ListFlowsRequest, response *clusterpb.ListFlowsResponse) (err error) {
	flows, total, err := domain.TaskRepo.ListFlows(req.BizId, req.Keyword, int(req.Status), int(req.Page.Page), int(req.Page.PageSize))
	if err != nil {
		getLogger().Error(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.Page = &clusterpb.PageDTO{
		Page:     req.Page.Page,
		PageSize: req.Page.PageSize,
		Total:    int32(total),
	}

	response.Flows = make([]*clusterpb.FlowDTO, len(flows), len(flows))
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
				Name:     v.Operator.Name,
				Id:       v.Operator.Id,
				TenantId: v.Operator.TenantId,
			},
		}
	}
	return err
}

var ManageSuccessResponseStatus = &clusterpb.ManagerResponseStatus{
	Code: 0,
}

func (*ClusterServiceHandler) Login(ctx context.Context, req *clusterpb.LoginRequest, resp *clusterpb.LoginResponse) error {
	log := framework.LogWithContext(ctx).WithField("fp", "ClusterServiceHandler.Login")
	log.Debug("req:", req)
	token, err := domain2.Login(req.GetAccountName(), req.GetPassword())

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

func (*ClusterServiceHandler) Logout(ctx context.Context, req *clusterpb.LogoutRequest, resp *clusterpb.LogoutResponse) error {
	accountName, err := domain2.Logout(req.TokenString)
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

func (*ClusterServiceHandler) VerifyIdentity(ctx context.Context, req *clusterpb.VerifyIdentityRequest, resp *clusterpb.VerifyIdentityResponse) error {
	tenantId, accountId, accountName, err := domain2.Accessible(req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*domain2.UnauthorizedError); ok {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*domain2.ForbiddenError); ok {
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
