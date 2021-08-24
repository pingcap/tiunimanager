package service

import (
	"context"
	domain2 "github.com/pingcap/tiem/micro-cluster/service/tenant/domain"
	"net/http"
	"strconv"

	dbClient "github.com/pingcap/tiem/library/firstparty/client"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	clusterPb "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/pingcap/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiem/micro-cluster/service/host"
	dbPb "github.com/pingcap/tiem/micro-metadb/proto"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &clusterPb.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &clusterPb.ResponseStatusDTO{Code: 1}

type ClusterServiceHandler struct{}

var log *logger.LogRecord

func InitClusterLogger(key config.Key) {
	log = logger.GetLogger(key)
	domain.InitDomainLogger(key)
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterPb.ClusterCreateReqDTO, resp *clusterPb.ClusterCreateRespDTO) (err error) {
	log.Info("create cluster")
	clusterAggregation, err := domain.CreateCluster(req.GetOperator(), req.GetCluster(), req.GetDemands())

	if err != nil {
		// todo
		log.Info(err)
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
	log.Info("query cluster")
	clusters, total, err := domain.ListCluster(req.Operator, req)
	if err != nil {
		log.Info(err)

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
	log.Info("delete cluster")

	clusterAggregation, err := domain.DeleteCluster(req.GetOperator(), req.GetClusterId())
	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterPb.ClusterDetailReqDTO, resp *clusterPb.ClusterDetailRespDTO) (err error) {
	log.Info("detail cluster")

	cluster, err := domain.GetClusterDetail(req.Operator, req.ClusterId)

	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.DisplayInfo = cluster.ExtractDisplayDTO()
		resp.Components = cluster.ExtractComponentDTOs()
		resp.MaintenanceInfo = cluster.ExtractMaintenanceDTO()

		return nil
	}
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterPb.CreateBackupRequest, response *clusterPb.CreateBackupResponse) (err error) {
	log.Info("backup cluster")

	cluster, err := domain.Backup(request.Operator, request.ClusterId)

	if err != nil {
		log.Info(err)
		// todo
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.BackupRecord = cluster.ExtractBackupRecordDTO()
		return nil
	}
}

func (c ClusterServiceHandler) RecoverBackupRecord(ctx context.Context, request *clusterPb.RecoverBackupRequest, response *clusterPb.RecoverBackupResponse) (err error) {
	log.Info("recover cluster")

	cluster, err := domain.Recover(request.Operator, request.ClusterId, request.BackupRecordId)

	if err != nil {
		log.Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.RecoverRecord = cluster.ExtractRecoverRecordDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *clusterPb.DeleteBackupRequest, response *clusterPb.DeleteBackupResponse) (err error) {

	_, err = dbClient.DBClient.DeleteBackupRecord(context.TODO(), &dbPb.DBDeleteBackupRecordRequest{Id: request.GetBackupRecordId()})
	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		return nil
	}
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterPb.SaveBackupStrategyRequest, response *clusterPb.SaveBackupStrategyResponse) (err error) {

	cronEntity, err := domain.TaskRepo.QueryCronTask(request.ClusterId, int(domain.CronBackup))

	if err != nil {
		// todo
		log.Info(err)
		return err
	}

	if cronEntity == nil {
		cronEntity = &domain.CronTaskEntity{
			Cron:         request.Cron,
			BizId:        request.ClusterId,
			CronTaskType: domain.CronBackup,
			Status:       domain.CronStatusValid,
		}

		domain.TaskRepo.AddCronTask(cronEntity)
	} else {
		cronEntity.Cron = request.Cron
	}

	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterPb.GetBackupStrategyRequest, response *clusterPb.GetBackupStrategyResponse) (err error) {

	cronEntity, err := domain.TaskRepo.QueryCronTask(request.ClusterId, int(domain.CronBackup))

	if err != nil {
		// todo
		log.Info(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.Cron = cronEntity.Cron

	return nil
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *clusterPb.QueryBackupRequest, response *clusterPb.QueryBackupResponse) (err error) {

	result, err := dbClient.DBClient.ListBackupRecords(context.TODO(), &dbPb.DBListBackupRecordsRequest{
		ClusterId: request.ClusterId,
		Page: &dbPb.DBPageDTO{
			Page:     request.Page.Page,
			PageSize: request.Page.PageSize,
		},
	})
	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.Page = &clusterPb.PageDTO{
			Page:     result.Page.Page,
			PageSize: result.Page.PageSize,
			Total:    result.Page.Total,
		}
		response.BackupRecords = make([]*clusterPb.BackupRecordDTO, len(result.BackupRecords), len(result.BackupRecords))
		for i, v := range result.BackupRecords {
			response.BackupRecords[i] = &clusterPb.BackupRecordDTO{
				Id:        v.BackupRecord.Id,
				ClusterId: v.BackupRecord.ClusterId,
				Range:     v.BackupRecord.BackupRange,
				Way:       v.BackupRecord.BackupType,
				FilePath:  v.BackupRecord.FilePath,
				StartTime: v.Flow.CreateTime,
				EndTime:   v.Flow.UpdateTime,
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
		log.Info(err)
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
		log.Info(err)
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
	info, err := domain.DecribeDashboard(request.Operator, request.ClusterId)
	if err != nil {
		log.Error(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.ClusterId = info.ClusterId
	response.Url = info.Url
	response.ShareCode = info.ShareCode

	return nil
}


var ManageSuccessResponseStatus = &clusterPb.ManagerResponseStatus{
	Code: 0,
}

func (*ClusterServiceHandler) Login(ctx context.Context, req *clusterPb.LoginRequest, resp *clusterPb.LoginResponse) error {

	token, err := domain2.Login(req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &clusterPb.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TokenString = token
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

func InitHostLogger(key config.Key) {
	host.InitLogger(key)
}

func (*ClusterServiceHandler) ImportHost(ctx context.Context, in *clusterPb.ImportHostRequest, out *clusterPb.ImportHostResponse) error {
	return host.ImportHost(ctx, in, out)
}

func (*ClusterServiceHandler) ImportHostsInBatch(ctx context.Context, in *clusterPb.ImportHostsInBatchRequest, out *clusterPb.ImportHostsInBatchResponse) error {
	return host.ImportHostsInBatch(ctx, in, out)
}

func (*ClusterServiceHandler) RemoveHost(ctx context.Context, in *clusterPb.RemoveHostRequest, out *clusterPb.RemoveHostResponse) error {
	return host.RemoveHost(ctx, in, out)
}

func (*ClusterServiceHandler) RemoveHostsInBatch(ctx context.Context, in *clusterPb.RemoveHostsInBatchRequest, out *clusterPb.RemoveHostsInBatchResponse) error {
	return host.RemoveHostsInBatch(ctx, in, out)
}

func (*ClusterServiceHandler) ListHost(ctx context.Context, in *clusterPb.ListHostsRequest, out *clusterPb.ListHostsResponse) error {
	return host.ListHost(ctx, in, out)
}

func (*ClusterServiceHandler) CheckDetails(ctx context.Context, in *clusterPb.CheckDetailsRequest, out *clusterPb.CheckDetailsResponse) error {
	return host.CheckDetails(ctx, in, out)
}

func (*ClusterServiceHandler) AllocHosts(ctx context.Context, in *clusterPb.AllocHostsRequest, out *clusterPb.AllocHostResponse) error {
	return host.AllocHosts(ctx, in, out)
}

func (*ClusterServiceHandler) GetFailureDomain(ctx context.Context, in *clusterPb.GetFailureDomainRequest, out *clusterPb.GetFailureDomainResponse) error {
	return host.GetFailureDomain(ctx, in, out)
}

