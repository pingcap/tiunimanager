package service

import (
	"context"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service/cluster/domain"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ResponseStatusDTO {Code:0}
var BizErrorResponseStatus = &cluster.ResponseStatusDTO {Code:1}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *cluster.ClusterCreateReqDTO, resp *cluster.ClusterCreateRespDTO) error {
	clusterAggregation, err := domain.CreateCluster(req.GetOperator(), req.GetCluster(), req.GetDemands())

	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.BaseInfo = clusterAggregation.ExtractBaseInfoDTO()
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}

}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *cluster.ClusterQueryReqDTO, resp *cluster.ClusterQueryRespDTO) error {
	clusters, total, err := domain.ListCluster(req.Operator, req)
	if err != nil {
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.Clusters = make([]*cluster.ClusterDisplayDTO, len(clusters), len(clusters))
		for i,v := range clusters {
			resp.Clusters[i] = v.ExtractDisplayDTO()
		}
		resp.Page = &cluster.PageDTO{
			Page:     req.PageReq.Page,
			PageSize: int32(len(clusters)),
			Total:    int32(total),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *cluster.ClusterDeleteReqDTO, resp *cluster.ClusterDeleteRespDTO) error {
	clusterAggregation, err := domain.DeleteCluster(req.GetOperator(), req.GetClusterId())
	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *cluster.ClusterDetailReqDTO, resp *cluster.ClusterDetailRespDTO) error {
	cluster, err := domain.GetClusterDetail(req.Operator, req.ClusterId)

	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.DisplayInfo = cluster.ExtractDisplayDTO()
		resp.Components = cluster.ExtractComponentDTOs()
		resp.MaintenanceInfo = cluster.ExtractMaintenanceDTO()

		return nil
	}
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, req *cluster.DataExportRequest, resp *cluster.DataExportResponse) error {
	flowId, err := domain.ExportData(req.Operator, req.ClusterId, req.UserName, req.Password, req.FileType)

	if err != nil {
		return err
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.FlowId = flowId
	}
	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, req *cluster.DataImportRequest, resp *cluster.DataImportResponse) error {
	flowId, err := domain.ImportData(req.Operator, req.ClusterId, req.UserName, req.Password, req.DataDir)

	if err != nil {
		return err
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.FlowId = flowId
	}
	return nil
	return nil
}

func (c ClusterServiceHandler) DescribeDataTransport(ctx context.Context, req *cluster.DataTransportQueryRequest, resp *cluster.DataTransportQueryResponse) error {
	//todo: implement
	return nil
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *cluster.QueryBackupRequest, response *cluster.QueryBackupResponse) error {
	panic("implement me")
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *cluster.CreateBackupRequest, response *cluster.CreateBackupResponse) error {
	panic("implement me")
}

func (c ClusterServiceHandler) RecoverBackupRecord(ctx context.Context, request *cluster.RecoverBackupRequest, response *cluster.RecoverBackupResponse) error {
	panic("implement me")
}

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *cluster.DeleteBackupRequest, response *cluster.DeleteBackupResponse) error {
	panic("implement me")
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *cluster.SaveBackupStrategyRequest, response *cluster.SaveBackupStrategyResponse) error {
	cronEntity, err := domain.TaskRepo.QueryCronTask(request.ClusterId, int(domain.CronBackup))

	if err != nil {
		// todo
	}

	if cronEntity == nil {
		cronEntity = &domain.CronTaskEntity{
			Cron: request.Cron,
			BizId: request.ClusterId,
			CronTaskType: domain.CronBackup,
			Status: domain.CronStatusValid,
		}

		domain.TaskRepo.AddCronTask(cronEntity)
	} else {
		cronEntity.Cron = request.Cron
	}

	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *cluster.GetBackupStrategyRequest, response *cluster.GetBackupStrategyResponse) error {
	cronEntity, err := domain.TaskRepo.QueryCronTask(request.ClusterId, int(domain.CronBackup))

	if err != nil {
		// todo
	}

	response.Status = SuccessResponseStatus
	response.Cron = cronEntity.Cron

	return nil
}