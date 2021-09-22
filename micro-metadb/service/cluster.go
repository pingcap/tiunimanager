package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/pingcap/errors"
)

var ClusterSuccessResponseStatus = &dbPb.DBClusterResponseStatus{Code: 0}
var ClusterNoResultResponseStatus = &dbPb.DBClusterResponseStatus{Code: 1}
var BizErrResponseStatus = &dbPb.DBClusterResponseStatus{Code: 2}

func (handler *DBServiceHandler) CreateCluster(ctx context.Context, req *dbPb.DBCreateClusterRequest, resp *dbPb.DBCreateClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("CreateCluster has invalid parameter")
	}
	dto := req.Cluster
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.CreateCluster(dto.Name, dto.DbPassword, dto.ClusterType, dto.VersionCode, dto.Tls, dto.Tags, dto.OwnerId, dto.TenantId)
	if nil == err {
		do, demand, err := clusterManager.UpdateClusterDemand(cluster.ID, req.Cluster.Demands, cluster.TenantId)
		if err == nil {
			resp.Status = ClusterSuccessResponseStatus
			resp.Cluster = convertToClusterDTO(do, demand)
		} else {
			err = errors.New(fmt.Sprintf("CreateCluster failed, update cluster demand failed, clusterId: %s, tenantId: %s, errors: %v", cluster.ID, cluster.TenantId, err))
		}
	}
	if nil == err {
		log.Infof("CreateCluster successful, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = "CreateCluster failed"
		log.Infof("CreateCluster failed, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	}
	return nil
}

func (handler *DBServiceHandler) DeleteCluster(ctx context.Context, req *dbPb.DBDeleteClusterRequest, resp *dbPb.DBDeleteClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.DeleteCluster(req.ClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(cluster, nil)
	} else {
		err = errors.New(fmt.Sprintf("DeleteCluster failed, clusterId: %s, error: %s", req.GetClusterId(), err))
	}
	if nil == err {
		log.Infof("DeleteCluster successful, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	} else {
		log.Infof("DeleteCluster failed, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	}
	return err
}

func (handler *DBServiceHandler) CreateInstance(ctx context.Context, req *dbPb.DBCreateInstanceRequest, resp *dbPb.DBCreateInstanceResponse) (err error) {
	log := framework.Log()

	if nil == req || nil == resp {
		log.Error("CreateInstance has invalid parameter")
		// todo handle error
		return errors.Errorf("CreateInstance has invalid parameter")
	}

	clusterManager := handler.Dao().ClusterManager()

	cluster, err := clusterManager.UpdateTopologyConfig(req.ClusterId, req.TopologyContent, req.TenantId)
	if err == nil {
		componentInstances, err := clusterManager.AddClusterComponentInstance(req.ClusterId, convertToComponentInstance(req.ComponentInstances))
		if err == nil {
			resp.Status = ClusterSuccessResponseStatus
			resp.Cluster = convertToClusterDTO(cluster, nil)
			resp.ComponentInstances = convertToComponentInstanceDTO(componentInstances)
			log.Infof("CreateInstance successful, clusterId: %s, tenantId: %s, error: %v",
				req.GetClusterId(), req.GetTenantId(), err)
			return nil
		}
	}

	errMsg := fmt.Sprintf("CreateInstance failed, clusterId: %s, tenantId: %s, error: %v",
		req.GetClusterId(), req.GetTenantId(), err)
	log.Error(errMsg)

	// todo handle error
	return err
}

func (handler *DBServiceHandler) UpdateClusterTopologyConfig(ctx context.Context, req *dbPb.DBUpdateTopologyConfigRequest, resp *dbPb.DBUpdateTopologyConfigResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterTopologyConfig has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	do, err := clusterManager.UpdateTopologyConfig(req.ClusterId, req.Content, req.TenantId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, nil)
		log.Infof("UpdateClusterTopologyConfig successful, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err)
	} else {
		err = errors.New(fmt.Sprintf("UpdateClusterTopologyConfig failed, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err))
		log.Infof("UpdateClusterTopologyConfig failed, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err)
	}
	return err
}

func (handler *DBServiceHandler) UpdateClusterStatus(ctx context.Context, req *dbPb.DBUpdateClusterStatusRequest, resp *dbPb.DBUpdateClusterStatusResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterStatus has invalid parameter")
	}
	var do *models.Cluster
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()

	if req.GetUpdateStatus() {
		do, err = clusterManager.UpdateClusterStatus(req.ClusterId, int8(req.Status))
		if nil != err {
			log.Errorf("UpdateClusterStatus failed, clusterId: %s flowId: %d, ,error: %v",
				req.GetClusterId(), req.GetFlowId(), err)
			return err
		}
	}
	if req.GetUpdateFlow() {
		do, err = clusterManager.UpdateClusterFlowId(req.ClusterId, uint(req.FlowId))
	}
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, nil)
		log.Infof("UpdateClusterStatus successful, clusterId: %s flowId: %d, error: %v",
			req.GetClusterId(), req.GetFlowId(), err)
	} else {
		log.Errorf("UpdateClusterStatus failed, clusterId: %s flowId: %d, ,error: %v",
			req.GetClusterId(), req.GetFlowId(), err)
	}

	return err
}

func (handler *DBServiceHandler) LoadCluster(ctx context.Context, req *dbPb.DBLoadClusterRequest, resp *dbPb.DBLoadClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("LoadCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.FetchCluster(req.ClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.ClusterDetail = &dbPb.DBClusterDetailDTO{
			Cluster:          convertToClusterDTO(result.Cluster, result.DemandRecord),
			TopologyConfigRecord: convertToConfigDTO(result.TopologyConfig),
			Flow:             convertFlowToDTO(result.Flow),
			ComponentInstances: convertToComponentInstanceDTO(result.ComponentInstances),
		}
		log.Infof("LoadCluster successful, clusterId: %s, error: %v", req.GetClusterId(), err)
	} else {
		log.Infof("LoadCluster failed, clusterId: %s, error: %v", req.GetClusterId(), err)
	}
	return err
}

func (handler *DBServiceHandler) ListCluster(ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	clusters, total, err := clusterManager.ListClusterDetails(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int((req.PageReq.Page-1)*req.PageReq.PageSize), int(req.PageReq.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbPb.DBPageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		clusterDetails := make([]*dbPb.DBClusterDetailDTO, len(clusters), len(clusters))
		for i, v := range clusters {
			clusterDetails[i] = &dbPb.DBClusterDetailDTO{
				Cluster:          convertToClusterDTO(v.Cluster, v.DemandRecord),
				TopologyConfigRecord: convertToConfigDTO(v.TopologyConfig),
				Flow:             convertFlowToDTO(v.Flow),
			}
		}
		resp.Clusters = clusterDetails
		log.Infof("ListCluster successful, clusterId: %s, page: %d, page_size: %d, total: %d, error: %s",
			req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize(), total, err)
	} else {
		log.Infof("ListCluster failed, clusterId: %s, page: %d, page_size: %d, total: %d, error: %s",
			req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize(), total, err)
	}
	return err
}

func (handler *DBServiceHandler) SaveBackupRecord(ctx context.Context, req *dbPb.DBSaveBackupRecordRequest, resp *dbPb.DBSaveBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	dto := req.BackupRecord
	result, err := clusterManager.SaveBackupRecord(dto)
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("SaveBackupRecord successful, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath, err)
	} else {
		log.Infof("SaveBackupRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath, err)
	}
	return err
}

func (handler *DBServiceHandler) UpdateBackupRecord(ctx context.Context, req *dbPb.DBUpdateBackupRecordRequest, resp *dbPb.DBUpdateBackupRecordResponse) (err error) {

	dto := req.BackupRecord
	err = handler.Dao().ClusterManager().UpdateBackupRecord(dto)
	if err != nil {
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	return nil
}

func (handler *DBServiceHandler) DeleteBackupRecord(ctx context.Context, req *dbPb.DBDeleteBackupRecordRequest, resp *dbPb.DBDeleteBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteBackupRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.DeleteBackupRecord(uint(req.Id))
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("DeleteBackupRecord successful, Id: %d, error: %v", req.GetId(), err)
	} else {
		log.Infof("DeleteBackupRecord failed, Id: %d, error: %v", req.GetId(), err)
	}
	return err
}

func (handler *DBServiceHandler) QueryBackupRecords(ctx context.Context, req *dbPb.DBQueryBackupRecordRequest, resp *dbPb.DBQueryBackupRecordResponse) (err error) {
	result, err := handler.Dao().ClusterManager().QueryBackupRecord(req.ClusterId, req.RecordId)
	if err != nil {
		return err
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecords = convertToBackupRecordDisplayDTO(result.BackupRecord, result.Flow)
	return nil
}

func (handler *DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbPb.DBListBackupRecordsRequest, resp *dbPb.DBListBackupRecordsResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListBackupRecords has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	backupRecords, total, err := clusterManager.ListBackupRecords(req.ClusterId, req.StartTime, req.EndTime,
		int((req.Page.Page-1)*req.Page.PageSize), int(req.Page.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbPb.DBPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		backupRecordDTOs := make([]*dbPb.DBDBBackupRecordDisplayDTO, len(backupRecords), len(backupRecords))
		for i, v := range backupRecords {
			backupRecordDTOs[i] = convertToBackupRecordDisplayDTO(v.BackupRecord, v.Flow)
		}
		resp.BackupRecords = backupRecordDTOs
		log.Infof("ListBackupRecords successful, clusterId: %s, page: %d, page size: %d, error: %v",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize(), err)
	} else {
		log.Infof("ListBackupRecords failed, clusterId: %s, page: %d, page size: %d, error: %v",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize(), err)
	}
	return err
}

func (handler *DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbPb.DBSaveRecoverRecordRequest, resp *dbPb.DBSaveRecoverRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveRecoverRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	dto := req.RecoverRecord
	result, err := clusterManager.SaveRecoverRecord(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.BackupRecordId), uint(dto.FlowId))
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.RecoverRecord = convertToRecoverRecordDTO(result)
		log.Infof("SaveRecoverRecord successful, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId(), err)
	} else {
		log.Infof("SaveRecoverRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId(), err)
	}
	return err
}

func (handler *DBServiceHandler) SaveBackupStrategy(ctx context.Context, req *dbPb.DBSaveBackupStrategyRequest, resp *dbPb.DBSaveBackupStrategyResponse) (err error) {
	dto := req.Strategy
	result, err := handler.Dao().ClusterManager().SaveBackupStrategy(dto)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = convertToBackupStrategyDTO(result)
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategy(ctx context.Context, req *dbPb.DBQueryBackupStrategyRequest, resp *dbPb.DBQueryBackupStrategyResponse) (err error) {

	clusterId := req.ClusterId
	result, err := handler.Dao().ClusterManager().QueryBackupStartegy(clusterId)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = convertToBackupStrategyDTO(result)
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategyByTime(ctx context.Context, req *dbPb.DBQueryBackupStrategyByTimeRequest, resp *dbPb.DBQueryBackupStrategyByTimeResponse) (err error) {

	result, err := handler.Dao().ClusterManager().QueryBackupStartegyByTime(req.GetWeekday(), req.GetStartHour())

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	strategyList := make([]*dbPb.DBBackupStrategyDTO, len(result))
	for i, v := range result {
		strategyList[i] = convertToBackupStrategyDTO(v)
	}
	resp.Strategys = strategyList
	return nil
}

func (handler *DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbPb.DBSaveParametersRequest, resp *dbPb.DBSaveParametersResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveParametersRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	dto := req.Parameters
	result, err := clusterManager.SaveParameters(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.FlowId), dto.Content)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Parameters = convertToParameterRecordDTO(result)
		log.Infof("SaveParametersRecord successful, tenantId: %s, clusterId: %s, flowId: %d, content: %s, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetFlowId(), dto.GetContent(), err)
	} else {
		log.Infof("SaveParametersRecord failed, tenantId: %s, clusterId: %s, flowId: %d, content: %s, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetFlowId(), dto.GetContent(), err)
	}
	return err
}

func (handler *DBServiceHandler) GetCurrentParametersRecord(ctx context.Context, req *dbPb.DBGetCurrentParametersRequest, resp *dbPb.DBGetCurrentParametersResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("GetCurrentParametersRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.GetCurrentParameters(req.GetClusterId())
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.Parameters = convertToParameterRecordDTO(result)
		log.Infof("GetCurrentParametersRecord successful, clusterId: %s, error: %v",
			req.GetClusterId(), err)
		return nil
	} else {
		resp.Status = ClusterNoResultResponseStatus
		log.Warnf("GetCurrentParametersRecord failed, clusterId: %s, error: %v",
			req.GetClusterId(), err)
		return nil
	}
}

func convertToBackupRecordDTO(do *models.BackupRecord) (dto *dbPb.DBBackupRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBBackupRecordDTO{
		Id:          int64(do.ID),
		TenantId:    do.TenantId,
		ClusterId:   do.ClusterId,
		StartTime:   do.StartTime.Unix(),
		BackupMethod: do.BackupMethod,
		BackupMode:  do.BackupMode,
		BackupType:  do.BackupType,
		OperatorId:  do.OperatorId,
		FilePath:    do.FilePath,
		FlowId:      int64(do.FlowId),
	}
	return
}

func convertToComponentInstance(dtos []*dbPb.DBComponentInstanceDTO) []*models.ComponentInstance {
	if dtos == nil || len(dtos) == 0 {
		return []*models.ComponentInstance{}
	}

	result := make([]*models.ComponentInstance, len(dtos), len(dtos))

	for i, v := range dtos {
		result[i] = &models.ComponentInstance {
			Entity: models.Entity{
				ID:       v.Id,
				Code:     v.Code,
				TenantId: v.TenantId,
				Status: int8(v.Status),
			},
			ClusterId: v.ClusterId,
			ComponentType: v.ComponentType,
			Role: v.Role,
			Spec: v.Spec,
			Version: v.Version,
			HostId: v.HostId,
			DiskId: v.DiskId,
			PortInfo: v.PortInfo,
			AllocRequestId: v.AllocRequestId,
		}
	}

	return result
}

func convertToComponentInstanceDTO(models []*models.ComponentInstance) []*dbPb.DBComponentInstanceDTO {
	if models == nil || len(models) == 0 {
		return []*dbPb.DBComponentInstanceDTO{}
	}
	result := make([]*dbPb.DBComponentInstanceDTO, len(models), len(models))

	for i, v := range models {
		result[i] = &dbPb.DBComponentInstanceDTO {
			Id:       v.ID,
			Code:     v.Code,
			TenantId: v.TenantId,
			Status: int32(v.Status),

			ClusterId: v.ClusterId,
			ComponentType: v.ComponentType,
			Role: v.Role,
			Spec: v.Spec,
			Version: v.Version,
			HostId: v.HostId,
			DiskId: v.DiskId,
			PortInfo: v.PortInfo,
			AllocRequestId: v.AllocRequestId,
			CreateTime: v.CreatedAt.Unix(),
			UpdateTime: v.UpdatedAt.Unix(),
			DeleteTime: deletedAtUnix(v.DeletedAt),
		}
	}

	return result
}

func convertToBackupRecordDisplayDTO(do *models.BackupRecord, flow *models.FlowDO) (dto *dbPb.DBDBBackupRecordDisplayDTO) {
	if do == nil {
		return nil
	}

	dto = &dbPb.DBDBBackupRecordDisplayDTO{
		BackupRecord: convertToBackupRecordDTO(do),
		Flow:         convertFlowToDTO(flow),
	}
	return
}

func convertToRecoverRecordDTO(do *models.RecoverRecord) (dto *dbPb.DBRecoverRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBRecoverRecordDTO{
		Id:             int64(do.ID),
		TenantId:       do.TenantId,
		ClusterId:      do.ClusterId,
		CreateTime:     do.CreatedAt.Unix(),
		OperatorId:     do.OperatorId,
		BackupRecordId: int64(do.BackupRecordId),
		FlowId:         int64(do.FlowId),
	}
	return
}

func convertToBackupStrategyDTO(do *models.BackupStrategy) (dto *dbPb.DBBackupStrategyDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBBackupStrategyDTO{
		Id:          int64(do.ID),
		OperatorId:  do.OperatorId,
		TenantId:    do.TenantId,
		ClusterId:   do.ClusterId,
		CreateTime:  do.CreatedAt.Unix(),
		UpdateTime:  do.UpdatedAt.Unix(),
		BackupDate:  do.BackupDate,
		StartHour:   do.StartHour,
		EndHour:     do.EndHour,
	}
	return
}

func convertToParameterRecordDTO(do *models.ParametersRecord) (dto *dbPb.DBParameterRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBParameterRecordDTO{
		Id:         int64(do.ID),
		TenantId:   do.TenantId,
		ClusterId:  do.ClusterId,
		CreateTime: do.CreatedAt.Unix(),
		OperatorId: do.OperatorId,
		FlowId:     int64(do.FlowId),
		Content:    do.Content,
	}
	return
}

func convertToClusterDTO(do *models.Cluster, demand *models.DemandRecord) (dto *dbPb.DBClusterDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBClusterDTO{
		Id:          do.ID,
		Code:        do.Code,
		Name:        do.Name,
		TenantId:    do.TenantId,
		DbPassword:  do.DbPassword,
		ClusterType: do.Type,
		VersionCode: do.Version,
		Status:      int32(do.Status),
		Tags:        do.Tags,
		Tls:         do.Tls,
		WorkFlowId:  int32(do.CurrentFlowId),
		OwnerId:     do.OwnerId,
		CreateTime:  do.CreatedAt.Unix(),
		UpdateTime:  do.UpdatedAt.Unix(),
		DeleteTime:  deletedAtUnix(do.DeletedAt),
	}

	if demand != nil {
		dto.Demands = demand.Content
	}
	return
}

func convertToConfigDTO(do *models.TopologyConfig) (dto *dbPb.DBTopologyConfigDTO) {
	if do == nil {
		return nil
	}
	return &dbPb.DBTopologyConfigDTO{
		Id:         int32(do.ID),
		TenantId:   do.TenantId,
		ClusterId:  do.ClusterId,
		Content:    do.Content,
		CreateTime: do.CreatedAt.Unix(),
	}
}

func deletedAtUnix(at gorm.DeletedAt) (unix int64) {
	if at.Valid {
		return at.Time.Unix()
	}
	return
}
