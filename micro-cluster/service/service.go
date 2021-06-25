package service

import (
	"context"
	"encoding/json"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service/clustermanage"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ClusterResponseStatus {Code:0}
var BizErrorResponseStatus = &cluster.ClusterResponseStatus {Code:1}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *cluster.CreateClusterRequest, resp *cluster.CreateClusterResponse) error {
	cluster, err := clustermanage.CreateCluster(req.Name, req.DbPassword, req.Version,
		req.TikvCount, req.TidbCount, req.PdCount,
		req.Operator.AccountName, uint(req.Operator.TenantId),
	)

	if err != nil {
		resp.Status = BizErrorResponseStatus
		resp.Status.Message = err.Error()
	} else {
		resp.Status = SuccessResponseStatus
		resp.Cluster = convert(cluster)
	}

	return nil
}

func convert(cluster *clustermanage.Cluster) (dto *cluster.ClusterInfoDTO) {
	dto.Id = int32(cluster.Id)
	dto.Name = cluster.Name
	dto.Version = string(cluster.Version)
	dto.DbPassword = ""
	dto.Status = int32(cluster.Status)

	configByte, _ := json.Marshal(cluster.TiUPConfig)
	dto.Config = string(configByte)

	return
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, request *cluster.QueryClusterRequest, resp *cluster.QueryClusterResponse) error {
	panic("implement me")
}
