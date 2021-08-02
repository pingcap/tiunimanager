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
