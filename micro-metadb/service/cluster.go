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

	do, err := models.UpdateClusterDemand(cluster.ID, req.Cluster.Demands, cluster.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = ConvertToClusterDTO(do)
	return nil
}

func (*DBServiceHandler) DeleteCluster(ctx context.Context, req *dbPb.DBDeleteClusterRequest, resp *dbPb.DBDeleteClusterResponse) error {
	cluster, err := models.DeleteCluster(req.ClusterId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus

	resp.Cluster = ConvertToClusterDTO(cluster)
	return nil
}

func ConvertToDisplayDTO(do *models.ClusterDO, flow *models.FlowDO) (dto *dbPb.DBClusterDisplayDTO) {
	return &dbPb.DBClusterDisplayDTO{
		Cluster: ConvertToClusterDTO(do),
		Flow: convertFlowToDTO(flow),
	}
}

func ConvertToClusterDTO(do *models.ClusterDO) (dto *dbPb.DBClusterDTO) {
	return &dbPb.DBClusterDTO {
		Id:          do.ID,
		Code:        do.Code,
		Name:        do.ClusterName,
		TenantId:    do.TenantId,
		DbPassword:  do.DbPassword,
		ClusterType: do.ClusterType,
		VersionCode: do.ClusterVersion,
		Status: 	 int32(do.Status),
		Tags:        do.Tags,
		Tls:         do.Tls,
		WorkFlowId:  int32(do.CurrentFlowId),
		OwnerId:     do.OwnerId,
		CreateTime:  do.CreatedAt.Unix(),
		UpdateTime:  do.UpdatedAt.Unix(),
		DeleteTime:  DeletedAtUnix(do.DeletedAt),
	}
}

func DeletedAtUnix(at gorm.DeletedAt) (unix int64) {
	if at.Valid {
		return at.Time.Unix()
	}
	return
}

func (*DBServiceHandler) UpdateClusterTiupConfig(ctx context.Context, req *dbPb.DBUpdateTiupConfigRequest, resp *dbPb.DBUpdateTiupConfigResponse) error {
	var err error
	do, err := models.UpdateTiUPConfig(req.ClusterId, req.Content, req.TenantId)
	if err != nil {
		// todo
		return nil
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Cluster = ConvertToClusterDTO(do)
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
	resp.Cluster = ConvertToClusterDTO(do)

	return nil
}
func (*DBServiceHandler) LoadCluster(ctx context.Context, req *dbPb.DBLoadClusterRequest, resp *dbPb.DBLoadClusterResponse) error {
	return nil
}
func (*DBServiceHandler) ListCluster(ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) error {
	return nil
}

func ParseDBClusterDTO(dto *dbPb.DBClusterDTO) (do *models.ClusterDO){
	return
}