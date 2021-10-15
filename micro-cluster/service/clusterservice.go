package service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	clusterPb "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap-inc/tiem/micro-cluster/service/host"
	domain2 "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/domain"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	log "github.com/sirupsen/logrus"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &clusterPb.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &clusterPb.ResponseStatusDTO{Code: 500}

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

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterPb.ClusterCreateReqDTO, resp *clusterPb.ClusterCreateRespDTO) (err error) {
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

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterPb.ClusterQueryReqDTO, resp *clusterPb.ClusterQueryRespDTO) (err error) {
	getLogger().Info("query cluster")
	clusters, total, err := domain.ListCluster(req.Operator, req)
	if err != nil {
		getLogger().Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.Clusters = make([]*clusterPb.ClusterDisplayDTO, len(clusters), len(clusters))
		for i, v := range clusters {
			resp.Clusters[i] = v.ExtractDisplayDTO()
		}
		resp.Page = &clusterPb.PageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterPb.ClusterDeleteReqDTO, resp *clusterPb.ClusterDeleteRespDTO) (err error) {
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

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterPb.ClusterDetailReqDTO, resp *clusterPb.ClusterDetailRespDTO) (err error) {
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

func (c ClusterServiceHandler) ExportData(ctx context.Context, req *clusterPb.DataExportRequest, resp *clusterPb.DataExportResponse) error {
	if err := domain.ExportDataPreCheck(req); err != nil {
		getLogger().Error(err)
		return err
	}

	recordId, err := domain.ExportData(ctx, req)

	if err != nil {
		getLogger().Error(err)
		return err
	}
	resp.RespStatus = SuccessResponseStatus
	resp.RecordId = recordId

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, req *clusterPb.DataImportRequest, resp *clusterPb.DataImportResponse) error {
	if err := domain.ImportDataPreCheck(req); err != nil {
		getLogger().Error(err)
		return err
	}

	recordId, err := domain.ImportData(ctx, req)

	if err != nil {
		getLogger().Error(err)
		return err
	}
	resp.RespStatus = SuccessResponseStatus
	resp.RecordId = recordId

	return nil
}

func (c ClusterServiceHandler) DescribeDataTransport(ctx context.Context, req *clusterPb.DataTransportQueryRequest, resp *clusterPb.DataTransportQueryResponse) error {
	infos, page, err := domain.DescribeDataTransportRecord(ctx, req.GetOperator(), req.GetRecordId(), req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize())
	if err != nil {
		//todo
		return err
	}
	resp.RespStatus = SuccessResponseStatus
	resp.PageReq = &clusterPb.PageDTO{
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
		Total:    page.GetTotal(),
	}
	resp.TransportInfos = make([]*clusterPb.DataTransportInfo, len(infos))
	for index := 0; index < len(infos); index++ {
		resp.TransportInfos[index] = &clusterPb.DataTransportInfo{
			RecordId:      infos[index].ID,
			ClusterId:     infos[index].ClusterId,
			TransportType: infos[index].TransportType,
			FilePath:      infos[index].FilePath,
			Status:        infos[index].Status,
			StartTime:     infos[index].StartTime,
			EndTime:       infos[index].EndTime,
		}
	}
	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterPb.CreateBackupRequest, response *clusterPb.CreateBackupResponse) (err error) {
	getLogger().Info("backup cluster")

	clusterAggregation, err := domain.Backup(ctx, request.Operator, request.ClusterId, request.BackupMethod, request.BackupType, domain.BackupModeManual, request.FilePath)
	if err != nil {
		getLogger().Info(err)
		// todo
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.BackupRecord = clusterAggregation.ExtractBackupRecordDTO()
		return nil
	}
}

func (c ClusterServiceHandler) RecoverCluster(ctx context.Context, req *clusterPb.RecoverRequest, resp *clusterPb.RecoverResponse) (err error) {
	getLogger().Info("recover cluster")

	if err = domain.RecoverPreCheck(req); err != nil {
		getLogger().Errorf("recover cluster pre check failed, %s", err.Error())
		return err
	}

	clusterAggregation, err := domain.Recover(ctx, req.GetOperator(), req.GetCluster(), req.GetDemands())
	if err != nil {
		getLogger().Info(err)
		resp.RespStatus = BizErrorResponseStatus
		resp.RespStatus.Message = err.Error()
		return err
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.BaseInfo = clusterAggregation.ExtractBaseInfoDTO()
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *clusterPb.DeleteBackupRequest, response *clusterPb.DeleteBackupResponse) (err error) {
	getLogger().Info("delete backup")

	err = domain.DeleteBackup(ctx, request.Operator, request.GetClusterId(), request.GetBackupRecordId())
	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		return nil
	}
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterPb.SaveBackupStrategyRequest, response *clusterPb.SaveBackupStrategyResponse) (err error) {
	getLogger().Info("save backup strategy")

	err = domain.SaveBackupStrategyPreCheck(request.GetOperator(), request.GetStrategy())
	if err != nil {
		getLogger().Error(err)
		return err
	}

	err = domain.SaveBackupStrategy(ctx, request.GetOperator(), request.GetStrategy())
	if err != nil {
		// todo
		getLogger().Error(err)
		return err
	} else {
		response.Status = SuccessResponseStatus
		return nil
	}
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterPb.GetBackupStrategyRequest, response *clusterPb.GetBackupStrategyResponse) (err error) {
	strategy, err := domain.QueryBackupStrategy(ctx, request.GetOperator(), request.GetClusterId())
	if err != nil {
		getLogger().Error(err)
		return err
	} else {
		response.Status = SuccessResponseStatus
		response.Strategy = strategy
		return nil
	}
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *clusterPb.QueryBackupRequest, response *clusterPb.QueryBackupResponse) (err error) {
	result, err := client.DBClient.ListBackupRecords(context.TODO(), &dbPb.DBListBackupRecordsRequest{
		ClusterId: request.ClusterId,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		Page: &dbPb.DBPageDTO{
			Page:     request.Page.Page,
			PageSize: request.Page.PageSize,
		},
	})
	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.Page = &clusterPb.PageDTO{
			Page:     result.Page.Page,
			PageSize: result.Page.PageSize,
			Total:    result.Page.Total,
		}
		response.BackupRecords = make([]*clusterPb.BackupRecordDTO, len(result.BackupRecords))
		for i, v := range result.BackupRecords {
			response.BackupRecords[i] = &clusterPb.BackupRecordDTO{
				Id:           v.BackupRecord.Id,
				ClusterId:    v.BackupRecord.ClusterId,
				BackupMethod: v.BackupRecord.BackupMethod,
				BackupType:   v.BackupRecord.BackupType,
				BackupMode:   v.BackupRecord.BackupMode,
				FilePath:     v.BackupRecord.FilePath,
				StartTime:    v.Flow.CreateTime,
				EndTime:      v.Flow.UpdateTime,
				Size:         v.BackupRecord.Size,
				Operator: &clusterPb.OperatorDTO{
					Id: v.BackupRecord.OperatorId,
				},
				DisplayStatus: &clusterPb.DisplayStatusDTO{
					StatusCode:      strconv.Itoa(int(v.Flow.Status)),
					StatusName:      v.Flow.StatusAlias,
					InProcessFlowId: int32(v.Flow.Id),
				},
			}
		}
		return nil
	}
}

func (c ClusterServiceHandler) QueryParameters(ctx context.Context, request *clusterPb.QueryClusterParametersRequest, response *clusterPb.QueryClusterParametersResponse) (err error) {

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

func (c ClusterServiceHandler) SaveParameters(ctx context.Context, request *clusterPb.SaveClusterParametersRequest, response *clusterPb.SaveClusterParametersResponse) (err error) {

	clusterAggregation, err := domain.ModifyParameters(request.Operator, request.ClusterId, request.ParametersJson)

	if err != nil {
		// todo
		getLogger().Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.DisplayInfo = &clusterPb.DisplayStatusDTO{
			InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DescribeDashboard(ctx context.Context, request *clusterPb.DescribeDashboardRequest, response *clusterPb.DescribeDashboardResponse) (err error) {
	info, err := domain.DescribeDashboard(ctx, request.Operator, request.ClusterId)
	if err != nil {
		getLogger().Error(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.ClusterId = info.ClusterId
	response.Url = info.Url
	response.Token = info.Token

	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, req *clusterPb.ListFlowsRequest, response *clusterPb.ListFlowsResponse) (err error) {
	flows, total, err := domain.TaskRepo.ListFlows(req.BizId, req.Keyword, int(req.Status), int(req.Page.Page), int(req.Page.PageSize))
	if err != nil {
		getLogger().Error(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.Page = &clusterPb.PageDTO{
		Page:     req.Page.Page,
		PageSize: req.Page.PageSize,
		Total:    int32(total),
	}

	response.Flows = make([]*clusterPb.FlowDTO, len(flows), len(flows))
	for i, v := range flows {
		response.Flows[i] = &clusterPb.FlowDTO{
			Id:          int64(v.Id),
			FlowName:    v.FlowName,
			StatusAlias: v.StatusAlias,
			BizId:       v.BizId,
			Status:      int32(v.Status),
			StatusName:  v.Status.Display(),
			CreateTime:  v.CreateTime.Unix(),
			UpdateTime:  v.UpdateTime.Unix(),
			Operator: &clusterPb.OperatorDTO{
				Name:     v.Operator.Name,
				Id:       v.Operator.Id,
				TenantId: v.Operator.TenantId,
			},
		}
	}
	return err
}

var ManageSuccessResponseStatus = &clusterPb.ManagerResponseStatus{
	Code: 0,
}

func (*ClusterServiceHandler) Login(ctx context.Context, req *clusterPb.LoginRequest, resp *clusterPb.LoginResponse) error {
	log := framework.LogWithContext(ctx).WithField("fp", "ClusterServiceHandler.Login")
	log.Debug("req:", req)
	token, err := domain2.Login(req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &clusterPb.ManagerResponseStatus{
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

func (*ClusterServiceHandler) Logout(ctx context.Context, req *clusterPb.LogoutRequest, resp *clusterPb.LogoutResponse) error {
	accountName, err := domain2.Logout(req.TokenString)
	if err != nil {
		resp.Status = &clusterPb.ManagerResponseStatus{
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

func (*ClusterServiceHandler) VerifyIdentity(ctx context.Context, req *clusterPb.VerifyIdentityRequest, resp *clusterPb.VerifyIdentityResponse) error {
	tenantId, accountId, accountName, err := domain2.Accessible(req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*domain2.UnauthorizedError); ok {
			resp.Status = &clusterPb.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*domain2.ForbiddenError); ok {
			resp.Status = &clusterPb.ManagerResponseStatus{
				Code:    http.StatusForbidden,
				Message: "无权限",
			}
		} else {
			resp.Status = &clusterPb.ManagerResponseStatus{
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

func (clusterManager *ClusterServiceHandler) ImportHost(ctx context.Context, in *clusterPb.ImportHostRequest, out *clusterPb.ImportHostResponse) error {
	return clusterManager.resourceManager.ImportHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) ImportHostsInBatch(ctx context.Context, in *clusterPb.ImportHostsInBatchRequest, out *clusterPb.ImportHostsInBatchResponse) error {
	return clusterManager.resourceManager.ImportHostsInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RemoveHost(ctx context.Context, in *clusterPb.RemoveHostRequest, out *clusterPb.RemoveHostResponse) error {
	return clusterManager.resourceManager.RemoveHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RemoveHostsInBatch(ctx context.Context, in *clusterPb.RemoveHostsInBatchRequest, out *clusterPb.RemoveHostsInBatchResponse) error {
	return clusterManager.resourceManager.RemoveHostsInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) ListHost(ctx context.Context, in *clusterPb.ListHostsRequest, out *clusterPb.ListHostsResponse) error {
	return clusterManager.resourceManager.ListHost(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) CheckDetails(ctx context.Context, in *clusterPb.CheckDetailsRequest, out *clusterPb.CheckDetailsResponse) error {
	return clusterManager.resourceManager.CheckDetails(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) AllocHosts(ctx context.Context, in *clusterPb.AllocHostsRequest, out *clusterPb.AllocHostResponse) error {
	return clusterManager.resourceManager.AllocHosts(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) GetFailureDomain(ctx context.Context, in *clusterPb.GetFailureDomainRequest, out *clusterPb.GetFailureDomainResponse) error {
	return clusterManager.resourceManager.GetFailureDomain(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) AllocResourcesInBatch(ctx context.Context, in *clusterPb.BatchAllocRequest, out *clusterPb.BatchAllocResponse) error {
	return clusterManager.resourceManager.AllocResourcesInBatch(ctx, in, out)
}

func (clusterManager *ClusterServiceHandler) RecycleResources(ctx context.Context, in *clusterPb.RecycleRequest, out *clusterPb.RecycleResponse) error {
	return clusterManager.resourceManager.RecycleResources(ctx, in, out)
}
