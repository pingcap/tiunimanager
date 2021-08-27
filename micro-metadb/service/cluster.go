package service

import (
	"context"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"gorm.io/gorm"
)

var ClusterSuccessResponseStatus = &dbPb.DBClusterResponseStatus{Code: 0}
var ClusterNoResultResponseStatus = &dbPb.DBClusterResponseStatus{Code: 1}

func (handler *DBServiceHandler) CreateCluster(ctx context.Context, req *dbPb.DBCreateClusterRequest, resp *dbPb.DBCreateClusterResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	dto := req.Cluster
	cluster, err := clusterManager.CreateCluster(dto.Name, dto.DbPassword, dto.ClusterType, dto.VersionCode, dto.Tls, dto.Tags, dto.OwnerId, dto.TenantId)
	if err != nil {
		// todo
		return nil
	}

	do, demand, err := clusterManager.UpdateClusterDemand(cluster.ID, req.Cluster.Demands, cluster.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = convertToClusterDTO(do, demand)
	return nil
}

func (handler *DBServiceHandler) DeleteCluster(ctx context.Context, req *dbPb.DBDeleteClusterRequest, resp *dbPb.DBDeleteClusterResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.DeleteCluster(req.ClusterId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = convertToClusterDTO(cluster, nil)
	return nil
}

func (handler *DBServiceHandler) UpdateClusterTiupConfig(ctx context.Context, req *dbPb.DBUpdateTiupConfigRequest, resp *dbPb.DBUpdateTiupConfigResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	do, err := clusterManager.UpdateTiUPConfig(req.ClusterId, req.Content, req.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)
	return nil
}

func (handler *DBServiceHandler) UpdateClusterStatus(ctx context.Context, req *dbPb.DBUpdateClusterStatusRequest, resp *dbPb.DBUpdateClusterStatusResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	var do *models.Cluster
	if req.UpdateStatus {
		do, err = clusterManager.UpdateClusterStatus(req.ClusterId, int8(req.Status))
		if err != nil {
			// todo
			return nil
		}
	}

	if req.UpdateFlow {
		do, err = clusterManager.UpdateClusterFlowId(req.ClusterId, uint(req.FlowId))
		if err != nil {
			// todo
			return nil
		}
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)

	return nil
}

func (handler *DBServiceHandler) LoadCluster(ctx context.Context, req *dbPb.DBLoadClusterRequest, resp *dbPb.DBLoadClusterResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.FetchCluster(req.ClusterId)
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

func (handler *DBServiceHandler) ListCluster(ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) (err error) {
	clusterManager := handler.Dao().ClusterManager()
	clusters, total, err := clusterManager.ListClusterDetails(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int((req.PageReq.Page-1)*req.PageReq.PageSize), int(req.PageReq.PageSize))

	if err != nil {
		// todo
	}

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
			TiupConfigRecord: convertToConfigDTO(v.TiUPConfig),
			Flow:             convertFlowToDTO(v.Flow),
		}
	}

	resp.Clusters = clusterDetails

	return nil
}

func (handler *DBServiceHandler) SaveBackupRecord(ctx context.Context, req *dbPb.DBSaveBackupRecordRequest, resp *dbPb.DBSaveBackupRecordResponse) (err error) {
	pt := handler.Dao().Tables()[models.TABLE_NAME_BACKUP_RECORD].(*models.BackupRecord)
	db := handler.Dao().Db()
	dto := req.BackupRecord
	result, err := pt.SaveBackupRecord(db, dto.TenantId, dto.ClusterId, dto.OperatorId, int8(dto.BackupRange), int8(dto.BackupType), uint(dto.FlowId), dto.FilePath)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecord = ConvertToBackupRecordDTO(result)
	return nil
}

func (handler *DBServiceHandler) DeleteBackupRecord(ctx context.Context, req *dbPb.DBDeleteBackupRecordRequest, resp *dbPb.DBDeleteBackupRecordResponse) (err error) {

	pt := handler.Dao().Tables()[models.TABLE_NAME_BACKUP_RECORD].(*models.BackupRecord)
	db := handler.Dao().Db()
	result, err := pt.DeleteBackupRecord(db, uint(req.Id))
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecord = ConvertToBackupRecordDTO(result)
	return nil
}

func (handler *DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbPb.DBListBackupRecordsRequest, resp *dbPb.DBListBackupRecordsResponse) (err error) {
	pt := handler.Dao().Tables()[models.TABLE_NAME_BACKUP_RECORD].(*models.BackupRecord)
	db := handler.Dao().Db()
	backupRecords, total, err := pt.ListBackupRecords(db, req.ClusterId,
		int((req.Page.Page-1)*req.Page.PageSize), int(req.Page.PageSize))

	if err != nil {
		// todo
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Page = &dbPb.DBPageDTO{
		Page:     req.Page.Page,
		PageSize: req.Page.PageSize,
		Total:    int32(total),
	}

	backupRecordDTOs := make([]*dbPb.DBDBBackupRecordDisplayDTO, len(backupRecords), len(backupRecords))

	for i, v := range backupRecords {
		backupRecordDTOs[i] = ConvertToBackupRecordDisplayDTO(v.BackupRecord, v.Flow)
	}

	resp.BackupRecords = backupRecordDTOs
	return nil
}

func (handler *DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbPb.DBSaveRecoverRecordRequest, resp *dbPb.DBSaveRecoverRecordResponse) (err error) {

	pt := handler.Dao().Tables()[models.TABLE_NAME_RECOVER_RECORD].(*models.RecoverRecord)
	db := handler.Dao().Db()
	dto := req.RecoverRecord
	result, err := pt.SaveRecoverRecord(db, dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.BackupRecordId), uint(dto.FlowId))

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.RecoverRecord = ConvertToRecoverRecordDTO(result)
	return nil
}

func (handler *DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbPb.DBSaveParametersRequest, resp *dbPb.DBSaveParametersResponse) (err error) {

	pt := handler.Dao().Tables()[models.TABLE_NAME_PARAMETERS_RECORD].(*models.ParametersRecord)
	db := handler.Dao().Db()
	dto := req.Parameters
	result, err := pt.SaveParameters(db, dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.FlowId), dto.Content)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Parameters = ConvertToParameterRecordDTO(result)
	return nil
}

func (handler *DBServiceHandler) GetCurrentParametersRecord(ctx context.Context, req *dbPb.DBGetCurrentParametersRequest, resp *dbPb.DBGetCurrentParametersResponse) (err error) {

	pt := handler.Dao().Tables()[models.TABLE_NAME_PARAMETERS_RECORD].(*models.ParametersRecord)
	db := handler.Dao().Db()
	result, err := pt.GetCurrentParameters(db, req.GetClusterId())

	if err != nil {
		resp.Status = ClusterNoResultResponseStatus
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Parameters = ConvertToParameterRecordDTO(result)
	return nil
}

func ConvertToBackupRecordDTO(do *models.BackupRecord) (dto *dbPb.DBBackupRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBBackupRecordDTO{
		Id:          int64(do.ID),
		TenantId:    do.TenantId,
		ClusterId:   do.ClusterId,
		CreateTime:  do.CreatedAt.Unix(),
		BackupRange: int32(do.Range),
		BackupType:  int32(do.Type),
		OperatorId:  do.OperatorId,
		FilePath:    do.FilePath,
		FlowId:      int64(do.FlowId),
	}
	return
}

func ConvertToBackupRecordDisplayDTO(do *models.BackupRecord, flow *models.FlowDO) (dto *dbPb.DBDBBackupRecordDisplayDTO) {
	if do == nil {
		return nil
	}

	dto = &dbPb.DBDBBackupRecordDisplayDTO{
		BackupRecord: ConvertToBackupRecordDTO(do),
		Flow:         convertFlowToDTO(flow),
	}

	return
}

func ConvertToRecoverRecordDTO(do *models.RecoverRecord) (dto *dbPb.DBRecoverRecordDTO) {
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

func ConvertToParameterRecordDTO(do *models.ParametersRecord) (dto *dbPb.DBParameterRecordDTO) {
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

func convertToConfigDTO(do *models.TiUPConfig) (dto *dbPb.DBTiUPConfigDTO) {
	if do == nil {
		return nil
	}
	return &dbPb.DBTiUPConfigDTO{
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
