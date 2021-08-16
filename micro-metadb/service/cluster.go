package service

import (
	"context"
	"github.com/pingcap/tiem/micro-metadb/models"
	dbPb "github.com/pingcap/tiem/micro-metadb/proto"
	"gorm.io/gorm"
)

var ClusterSuccessResponseStatus =  &dbPb.DBClusterResponseStatus{Code: 0}
var ClusterNoResultResponseStatus =  &dbPb.DBClusterResponseStatus{Code: 1}

func (*DBServiceHandler) CreateCluster(ctx context.Context, req *dbPb.DBCreateClusterRequest, resp *dbPb.DBCreateClusterResponse) (err error) {

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

func (*DBServiceHandler) DeleteCluster(ctx context.Context, req *dbPb.DBDeleteClusterRequest, resp *dbPb.DBDeleteClusterResponse)  (err error)  {

	cluster, err := models.DeleteCluster(req.ClusterId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = convertToClusterDTO(cluster, nil)
	return nil
}

func (*DBServiceHandler) UpdateClusterTiupConfig(ctx context.Context, req *dbPb.DBUpdateTiupConfigRequest, resp *dbPb.DBUpdateTiupConfigResponse)  (err error)  {

	do, err := models.UpdateTiUPConfig(req.ClusterId, req.Content, req.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)
	return nil
}

func (*DBServiceHandler) UpdateClusterStatus(ctx context.Context, req *dbPb.DBUpdateClusterStatusRequest, resp *dbPb.DBUpdateClusterStatusResponse)  (err error)  {

	var do *models.ClusterDO
	if req.UpdateStatus {
		do, err = models.UpdateClusterStatus(req.ClusterId, int8(req.Status))
		if err != nil {
			// todo
			return nil
		}
	}

	if req.UpdateFlow {
		do, err = models.UpdateClusterFlowId(req.ClusterId, uint(req.FlowId))
		if err != nil {
			// todo
			return nil
		}
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = convertToClusterDTO(do, nil)

	return nil
}

func (*DBServiceHandler) LoadCluster(ctx context.Context, req *dbPb.DBLoadClusterRequest, resp *dbPb.DBLoadClusterResponse)  (err error)  {

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

func (*DBServiceHandler) ListCluster (ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) (err error)  {

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

func (*DBServiceHandler) SaveBackupRecord(ctx context.Context, req *dbPb.DBSaveBackupRecordRequest, resp *dbPb.DBSaveBackupRecordResponse) (err error) {

	dto := req.BackupRecord
	result, err := models.SaveBackupRecord(dto.TenantId, dto.ClusterId, dto.OperatorId, dto.BackupRange, dto.BackupType, dto.FilePath, uint(dto.FlowId))

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecord = ConvertToBackupRecordDTO(result)
	return nil
}

func (*DBServiceHandler) UpdateBackupRecord(ctx context.Context, req *dbPb.DBUpdateBackupRecordRequest, resp *dbPb.DBUpdateBackupRecordResponse) (err error) {

	dto := req.BackupRecord
	err = models.UpdateBackupRecord(dto)
	if err != nil {
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	return nil
}

func (*DBServiceHandler) DeleteBackupRecord(ctx context.Context, req *dbPb.DBDeleteBackupRecordRequest, resp *dbPb.DBDeleteBackupRecordResponse) (err error) {

	result, err := models.DeleteBackupRecord(uint(req.Id))
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecord = ConvertToBackupRecordDTO(result)
	return nil
}

func (*DBServiceHandler) QueryBackupRecord(ctx context.Context, req *dbPb.DBQueryBackupRecordRequest, resp *dbPb.DBQueryBackupRecordResponse) (err error) {
	result, err := models.QueryBackupRecord(req.ClusterId, req.RecordId)
	if err != nil {
		return err
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecords = ConvertToBackupRecordDisplayDTO(result.BackupRecordDO, result.Flow)
	return nil
}

func (*DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbPb.DBListBackupRecordsRequest, resp *dbPb.DBListBackupRecordsResponse) (err error) {

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

func (*DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbPb.DBSaveRecoverRecordRequest, resp *dbPb.DBSaveRecoverRecordResponse)  (err error) {

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

func (*DBServiceHandler) SaveBackupStrategy(ctx context.Context, req *dbPb.DBSaveBackupStrategyRequest, resp *dbPb.DBSaveBackupStrategyResponse)  (err error) {

	dto := req.Strategy
	result, err := models.SaveBackupStrategy(dto)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = ConvertToBackupStrategyDTO(result)
	return nil
}

func (*DBServiceHandler) QueryBackupStrategy(ctx context.Context, req *dbPb.DBQueryBackupStrategyRequest, resp *dbPb.DBQueryBackupStrategyResponse)  (err error) {

	clusterId := req.ClusterId
	result, err := models.QueryBackupStartegy(clusterId)

	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = ConvertToBackupStrategyDTO(result)
	return nil
}

func (*DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbPb.DBSaveParametersRequest, resp *dbPb.DBSaveParametersResponse)  (err error) {

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

func (*DBServiceHandler) GetCurrentParametersRecord(ctx context.Context, req *dbPb.DBGetCurrentParametersRequest, resp *dbPb.DBGetCurrentParametersResponse) (err error) {

	result, err := models.GetCurrentParameters(req.GetClusterId())

	if err != nil {
		resp.Status = ClusterNoResultResponseStatus
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Parameters = ConvertToParameterRecordDTO(result)
	return nil
}

func ConvertToBackupRecordDTO(do *models.BackupRecordDO) (dto *dbPb.DBBackupRecordDTO) {
	if do == nil {
		return nil
	}
	dto =  &dbPb.DBBackupRecordDTO{
		Id:          int64(do.ID),
		TenantId:    do.TenantId,
		ClusterId:   do.ClusterId,
		CreateTime:  do.CreatedAt.Unix(),
		BackupRange: do.BackupRange,
		BackupType:  do.BackupType,
		OperatorId:  do.OperatorId,
		FilePath:    do.FilePath,
		FlowId: int64(do.FlowId),
	}
	return
}

func ConvertToBackupRecordDisplayDTO(do *models.BackupRecordDO, flow *models.FlowDO) (dto *dbPb.DBDBBackupRecordDisplayDTO){
	if do == nil {
		return nil
	}

	dto = &dbPb.DBDBBackupRecordDisplayDTO{
		BackupRecord: ConvertToBackupRecordDTO(do),
		Flow: convertFlowToDTO(flow),
	}

	return
}

func ConvertToRecoverRecordDTO(do *models.RecoverRecordDO) (dto *dbPb.DBRecoverRecordDTO) {
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

func ConvertToBackupStrategyDTO(do *models.BackupStrategyDO) (dto *dbPb.DBBackupStrategyDTO) {
	if do == nil {
		return nil
	}
	dto = &dbPb.DBBackupStrategyDTO{
		Id:             int64(do.ID),
		TenantId:       do.TenantId,
		ClusterId:      do.ClusterId,
		CreateTime:     do.CreatedAt.Unix(),
		UpdateTime:  	do.UpdatedAt.Unix(),
		BackupRange:  	do.BackupRange,
		BackupType:  	do.BackupType,
		BackupDate:  	do.BackupDate,
		Period:  		do.Period,
		FilePath:  		do.FilePath,
	}
	return
}

func ConvertToParameterRecordDTO(do *models.ParametersRecordDO) (dto *dbPb.DBParameterRecordDTO) {
	if do == nil {
		return nil
	}
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
	if do == nil {
		return nil
	}
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
	if do == nil {
		return nil
	}
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
