package service

import (
	"context"
	"github.com/pingcap/ticp/micro-metadb/models"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
	"gorm.io/gorm"
)

var ClusterSuccessResponseStatus =  &dbPb.DBClusterResponseStatus{Code: 0}

func (*DBServiceHandler) CreateCluster(ctx context.Context, req *dbPb.DBCreateClusterRequest, resp *dbPb.DBCreateClusterResponse) error {
	dto := req.Cluster
	cluster, err := models.CreateCluster(dto.Name, dto.DbPassword, dto.ClusterType, dto.VersionCode, dto.Tls, dto.Tags, dto.OwnerId, dto.TenantId)
	if err != nil {
		// todo
		return nil
	}

	do, demand, err := models.UpdateClusterDemand(cluster.ID, req.Cluster.Demands, cluster.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = convertToClusterDTO(do, demand)
	return nil
}

func (*DBServiceHandler) DeleteCluster(ctx context.Context, req *dbPb.DBDeleteClusterRequest, resp *dbPb.DBDeleteClusterResponse) error {
	cluster, err := models.DeleteCluster(req.ClusterId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = convertToClusterDTO(cluster, nil)
	return nil
}

func (*DBServiceHandler) UpdateClusterTiupConfig(ctx context.Context, req *dbPb.DBUpdateTiupConfigRequest, resp *dbPb.DBUpdateTiupConfigResponse) error {
	var err error
	do, err := models.UpdateTiUPConfig(req.ClusterId, req.Content, req.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)
	return nil
}

func (*DBServiceHandler) UpdateClusterStatus(ctx context.Context, req *dbPb.DBUpdateClusterStatusRequest, resp *dbPb.DBUpdateClusterStatusResponse) error {
	var err error
	var do *models.ClusterDO
	if req.UpdateStatus {
		do, err = models.UpdateClusterFlowId(req.ClusterId, uint(req.FlowId))
		if err != nil {
			// todo
			return nil
		}
	}

	if req.UpdateFlow {
		do, err = models.UpdateClusterStatus(req.ClusterId, int8(req.Status))
		if err != nil {
			// todo
			return nil
		}
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)

	return nil
}

func (*DBServiceHandler) LoadCluster(ctx context.Context, req *dbPb.DBLoadClusterRequest, resp *dbPb.DBLoadClusterResponse) error {
	result, err := models.FetchCluster(req.ClusterId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.ClusterDetail = &dbPb.DBClusterDetailDTO{
		Cluster:          convertToClusterDTO(result.Cluster, result.DemandRecord),
		TiupConfigRecord: convertToConfigDTO(result.TiUPConfig),
		Flow:             convertFlowToDTO(result.Flow),
	}
	return nil
}

func (*DBServiceHandler) ListCluster (ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) error {
	clusters, total ,err  := models.ListClusterDetails(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int((req.PageReq.Page - 1) * req.PageReq.PageSize), int(req.PageReq.PageSize))

	if err != nil {
		// todo
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Page = &dbPb.DBPageDTO{
		Page:     req.PageReq.Page,
		PageSize: req.PageReq.PageSize,
		Total: int32(total),
	}

	clusterDetails :=  make([]*dbPb.DBClusterDetailDTO, len(clusters), len(clusters))

	for i, v := range clusters {
		clusterDetails[i] = &dbPb.DBClusterDetailDTO{
			Cluster:          convertToClusterDTO(v.Cluster, v.DemandRecord),
			TiupConfigRecord: convertToConfigDTO(v.TiUPConfig),
			Flow:             convertFlowToDTO(v.Flow),
		}
	}

	resp.Clusters = clusterDetails

	return nil
}

func (*DBServiceHandler) SaveBackupRecord(ctx context.Context, req *dbPb.DBSaveBackupRecordRequest, resp *dbPb.DBSaveBackupRecordResponse) error{

	dto := req.BackupRecord
	result, err := models.SaveBackupRecord(dto.TenantId, dto.ClusterId, dto.OperatorId, int8(dto.BackupRange), int8(dto.BackupType), uint(dto.FlowId), dto.FilePath)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecord = ConvertToBackupRecordDTO(result)
	return nil
}

func (*DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbPb.DBListBackupRecordsRequest, resp *dbPb.DBListBackupRecordsResponse) error{
	backupRecords, total ,err  := models.ListBackupRecords(req.ClusterId,
		int((req.Page.Page - 1) * req.Page.PageSize), int(req.Page.PageSize))

	if err != nil {
		// todo
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Page = &dbPb.DBPageDTO{
		Page:     req.Page.Page,
		PageSize: req.Page.PageSize,
		Total: int32(total),
	}

	backupRecordDTOs :=  make([]*dbPb.DBDBBackupRecordDisplayDTO, len(backupRecords), len(backupRecords))

	for i, v := range backupRecords {
		backupRecordDTOs[i] = ConvertToBackupRecordDisplayDTO(v.BackupRecordDO, v.Flow)
	}

	resp.BackupRecords = backupRecordDTOs
	return nil
}

func (*DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbPb.DBSaveRecoverRecordRequest, resp *dbPb.DBSaveRecoverRecordResponse) error{

	dto := req.RecoverRecord
	result, err := models.SaveRecoverRecord(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.BackupRecordId), uint(dto.FlowId))

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.RecoverRecord = ConvertToRecoverRecordDTO(result)
	return nil
}

func (*DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbPb.DBSaveParametersRequest, resp *dbPb.DBSaveParametersResponse) error{

	dto := req.Parameters
	result, err := models.SaveParameters(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.FlowId), dto.Content)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Parameters = ConvertToParameterRecordDTO(result)
	return nil
}

func (*DBServiceHandler) GetCurrentParametersRecord(ctx context.Context, req *dbPb.DBGetCurrentParametersRequest, resp *dbPb.DBGetCurrentParametersResponse) error{
	result, err := models.GetCurrentParameters(req.GetClusterId())

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Parameters = ConvertToParameterRecordDTO(result)
	return nil
}

func ConvertToBackupRecordDTO(do *models.BackupRecordDO) (dto *dbPb.DBBackupRecordDTO) {
	dto =  &dbPb.DBBackupRecordDTO{
		Id:          int64(do.ID),
		TenantId:    do.TenantId,
		ClusterId:   do.ClusterId,
		CreateTime:  do.CreatedAt.Unix(),
		BackupRange: int32(do.BackupRange),
		BackupType:  int32(do.BackupType),
		OperatorId:  do.OperatorId,
		FilePath:    do.FilePath,
		FlowId: int64(do.FlowId),
	}
	return
}

func ConvertToBackupRecordDisplayDTO(do *models.BackupRecordDO, flow *models.FlowDO) (dto *dbPb.DBDBBackupRecordDisplayDTO){
	dto = &dbPb.DBDBBackupRecordDisplayDTO{
		BackupRecord: ConvertToBackupRecordDTO(do),
		Flow: convertFlowToDTO(flow),
	}

	return
}

func ConvertToRecoverRecordDTO(do *models.RecoverRecordDO) (dto *dbPb.DBRecoverRecordDTO) {
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

func ConvertToParameterRecordDTO(do *models.ParametersRecordDO) (dto *dbPb.DBParameterRecordDTO) {
	dto = &dbPb.DBParameterRecordDTO{
		Id:             int64(do.ID),
		TenantId:       do.TenantId,
		ClusterId:      do.ClusterId,
		CreateTime:     do.CreatedAt.Unix(),
		OperatorId:     do.OperatorId,
		FlowId:         int64(do.FlowId),
		Content: 		do.Content,
	}
	return
}

func convertToClusterDTO(do *models.ClusterDO, demand *models.DemandRecordDO) (dto *dbPb.DBClusterDTO) {
	dto = &dbPb.DBClusterDTO {
		Id:          do.ID,
		Code:        do.Code,
		Name:        do.ClusterName,
		TenantId:    do.TenantId,
		DbPassword:  do.DbPassword,
		ClusterType: do.ClusterType,
		VersionCode: do.ClusterVersion,
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

func convertToConfigDTO(do *models.TiUPConfigDO) (dto *dbPb.DBTiUPConfigDTO) {

	return &dbPb.DBTiUPConfigDTO{
		Id: int32(do.ID),
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
