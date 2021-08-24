package service

import (
	"context"
	"strconv"

	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/pingcap/tiem/micro-metadb/client"

	"github.com/pingcap/tiem/micro-cluster/service/cluster/domain"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &cluster.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &cluster.ResponseStatusDTO{Code: 1}

type ClusterServiceHandler struct{}

var log *logger.LogRecord

func InitClusterLogger(key config.Key) {
	log = logger.GetLogger(key)
	domain.InitDomainLogger(key)
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *cluster.ClusterCreateReqDTO, resp *cluster.ClusterCreateRespDTO) (err error) {
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

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *cluster.ClusterQueryReqDTO, resp *cluster.ClusterQueryRespDTO) (err error) {
	log.Info("query cluster")
	clusters, total, err := domain.ListCluster(req.Operator, req)
	if err != nil {
		log.Info(err)

		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.Clusters = make([]*cluster.ClusterDisplayDTO, len(clusters), len(clusters))
		for i, v := range clusters {
			resp.Clusters[i] = v.ExtractDisplayDTO()
		}
		resp.Page = &cluster.PageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *cluster.ClusterDeleteReqDTO, resp *cluster.ClusterDeleteRespDTO) (err error) {
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

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *cluster.ClusterDetailReqDTO, resp *cluster.ClusterDetailRespDTO) (err error) {
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

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *cluster.CreateBackupRequest, response *cluster.CreateBackupResponse) (err error) {
	log.Info("backup cluster")

	if err = domain.BackupPreCheck(request); err != nil {
		return err
	}

	clusterAggregation, err := domain.Backup(request.Operator, request.ClusterId, request.BackupRange, request.BackupType, request.FilePath)
	if err != nil {
		log.Info(err)
		// todo
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.BackupRecord = clusterAggregation.ExtractBackupRecordDTO()
		return nil
	}
}

func (c ClusterServiceHandler) RecoverBackupRecord(ctx context.Context, request *cluster.RecoverBackupRequest, response *cluster.RecoverBackupResponse) (err error) {
	log.Info("recover cluster")

	//todo: param precheck
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

func (c ClusterServiceHandler) DeleteBackupRecord(ctx context.Context, request *cluster.DeleteBackupRequest, response *cluster.DeleteBackupResponse) (err error) {
	log.Info("delete backup")

	err = domain.DeleteBackup(request.Operator, request.GetClusterId(), request.GetBackupRecordId())
	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		return nil
	}
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *cluster.SaveBackupStrategyRequest, response *cluster.SaveBackupStrategyResponse) (err error) {
	_, err = client.DBClient.SaveBackupStrategy(context.TODO(), &db.DBSaveBackupStrategyRequest{
		Strategy: &db.DBBackupStrategyDTO{
			TenantId: request.Operator.TenantId,
			ClusterId: request.Strategy.ClusterId,
			BackupDate: request.Strategy.BackupDate,
			FilePath: request.Strategy.FilePath,
			BackupRange: request.Strategy.BackupRange,
			BackupType: request.Strategy.BackupType,
			Period: request.Strategy.Period,
		},
	})
	if err != nil {
		log.Error(err)
		return err
	} else {
		response.Status = SuccessResponseStatus
		return nil
	}
	/*
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
	*/
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *cluster.GetBackupStrategyRequest, response *cluster.GetBackupStrategyResponse) (err error) {
	resp, err := client.DBClient.QueryBackupStrategy(context.TODO(), &db.DBQueryBackupStrategyRequest{
		ClusterId: request.ClusterId,
	})
	if err != nil {
		log.Error(err)
		return err
	} else {
		response.Status = SuccessResponseStatus
		response.Strategy = &cluster.BackupStrategy{
			ClusterId: resp.GetStrategy().GetClusterId(),
			BackupDate: resp.GetStrategy().GetBackupDate(),
			FilePath: resp.GetStrategy().GetFilePath(),
			BackupRange: resp.GetStrategy().GetBackupRange(),
			BackupType: resp.GetStrategy().GetBackupType(),
			Period: resp.GetStrategy().GetPeriod(),
		}
		return nil
	}

	/*
	cronEntity, err := domain.TaskRepo.QueryCronTask(request.ClusterId, int(domain.CronBackup))

	if err != nil {
		// todo
		log.Info(err)
		return err
	}

	response.Status = SuccessResponseStatus
	response.Cron = cronEntity.Cron

	return nil

	 */
}

func (c ClusterServiceHandler) QueryBackupRecord(ctx context.Context, request *cluster.QueryBackupRequest, response *cluster.QueryBackupResponse) (err error) {
	result, err := client.DBClient.ListBackupRecords(context.TODO(), &db.DBListBackupRecordsRequest{
		ClusterId: request.ClusterId,
		Page: &db.DBPageDTO{
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
		response.Page = &cluster.PageDTO{
			Page:     result.Page.Page,
			PageSize: result.Page.PageSize,
			Total:    result.Page.Total,
		}
		response.BackupRecords = make([]*cluster.BackupRecordDTO, len(result.BackupRecords), len(result.BackupRecords))
		for i, v := range result.BackupRecords {
			response.BackupRecords[i] = &cluster.BackupRecordDTO{
				Id:        v.BackupRecord.Id,
				ClusterId: v.BackupRecord.ClusterId,
				Range: v.BackupRecord.BackupRange,
				BackupType: v.BackupRecord.BackupType,
				FilePath: v.BackupRecord.FilePath,
				StartTime: v.Flow.CreateTime,
				EndTime: v.Flow.UpdateTime,
				Size: v.BackupRecord.Size,
				Operator: &cluster.OperatorDTO {
					Id: v.BackupRecord.OperatorId,
				},
				DisplayStatus: &cluster.DisplayStatusDTO{
					StatusCode:      strconv.Itoa(int(v.Flow.Status)),
					StatusName:      v.Flow.StatusAlias,
					InProcessFlowId: int32(v.Flow.Id),
				},
			}
		}
		return nil
	}
}

func (c ClusterServiceHandler) QueryParameters(ctx context.Context, request *cluster.QueryClusterParametersRequest, response *cluster.QueryClusterParametersResponse) (err error) {

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

func (c ClusterServiceHandler) SaveParameters(ctx context.Context, request *cluster.SaveClusterParametersRequest, response *cluster.SaveClusterParametersResponse) (err error) {

	clusterAggregation, err := domain.ModifyParameters(request.Operator, request.ClusterId, request.ParametersJson)

	if err != nil {
		// todo
		log.Info(err)
		return nil
	} else {
		response.Status = SuccessResponseStatus
		response.DisplayInfo = &cluster.DisplayStatusDTO{
			InProcessFlowId: int32(clusterAggregation.CurrentWorkFlow.Id),
		}
		return nil
	}
}

func (c ClusterServiceHandler) DescribeDashboard(ctx context.Context, request *cluster.DescribeDashboardRequest, response *cluster.DescribeDashboardResponse) (err error) {
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
