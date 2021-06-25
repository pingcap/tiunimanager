package service

import (
	"context"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

var ClusterSuccessResponseStatus =  &dbPb.DBClusterResponseStatus{Code: 0}

func (*DBServiceHandler) AddCluster(ctx context.Context, req *dbPb.DBCreateClusterRequest, resp *dbPb.DBCreateClusterResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AddCluster"})

	dto := req.Cluster
	clusterModel, err := models.CreateCluster(
		uint(dto.TenantId),
		dto.Name,
		dto.DbPassword,
		dto.Version,
		int(dto.Status),
		int(dto.TidbCount),
		int(dto.TikvCount),
		int(dto.PdCount))

	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ClusterSuccessResponseStatus
		clusterDTO := new(dbPb.DBClusterDTO)
		copyClusterModelToDTO(&clusterModel, clusterDTO)
		resp.Cluster = clusterDTO
	}
	return nil
}

func (*DBServiceHandler) FindCluster(ctx context.Context, req *dbPb.DBFindClusterRequest, resp *dbPb.DBFindClusterResponse) error {
	clusterModel, err := models.FetchCluster(uint(req.ClusterId))
	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
		return nil
	}

	clusterDTO := new(dbPb.DBClusterDTO)
	copyClusterModelToDTO(clusterModel, clusterDTO)
	resp.Cluster = clusterDTO
	if clusterModel.ConfigID <= 0 {
		return nil
	}

	configModel, err := models.FetchTiUPConfig(clusterModel.ConfigID)
	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
		return nil
	} else {
		configDTO := new(dbPb.DBTiUPConfigDTO)
		copyConfigModelToDTO(configModel, configDTO)
		resp.Config = configDTO

		return nil
	}
}

func (*DBServiceHandler) UpdateTiUPConfig(ctx context.Context, req *dbPb.DBUpdateTiUPConfigRequest, resp *dbPb.DBUpdateTiUPConfigResponse) error {
	clusterModel, configModel, err := models.UpdateClusterTiUPConfig(uint(req.GetClusterId()), req.GetConfigContent())

	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
		return nil
	} else {
		clusterDTO := new(dbPb.DBClusterDTO)
		copyClusterModelToDTO(&clusterModel, clusterDTO)
		resp.Cluster = clusterDTO

		configDTO := new(dbPb.DBTiUPConfigDTO)
		copyConfigModelToDTO(&configModel, configDTO)
		resp.Config = configDTO

		return nil
	}
}

func (*DBServiceHandler) ListCluster(ctx context.Context, req *dbPb.DBListClusterRequest, resp *dbPb.DBListClusterResponse) error {
	clusters, _ := models.ListClusters(int(req.Page), int(req.PageSize))
	for _, v := range clusters {
		var clusterDTO dbPb.DBClusterDTO
		copyClusterModelToDTO(&v, &clusterDTO)
		resp.Clusters = append(resp.Clusters, &clusterDTO)
	}
	return nil
}

func copyClusterModelToDTO(model *models.Cluster, dto *dbPb.DBClusterDTO) {
	dto.Id = int32( model.ID)
	dto.TenantId = int32(model.TenantId)
	dto.Name = model.Name
	dto.DbPassword = model.DbPassword
	dto.Version = model.Version
	dto.Status = int32(model.Status)
	dto.TidbCount = int32(model.TidbCount)
	dto.TikvCount = int32(model.TikvCount)
	dto.PdCount = 	int32(model.PdCount)
}

func copyConfigModelToDTO(model *models.TiUPConfig, dto *dbPb.DBTiUPConfigDTO) {
	dto.Id = int32(model.ID)
	dto.TenantId = int32(model.TenantId)
	dto.ClusterId = int32(model.ClusterId)
	dto.Latest = model.Latest
	dto.Content = model.Content
}

