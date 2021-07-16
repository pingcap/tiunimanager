package service

import (
	"context"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ResponseStatusDTO {Code:0}
var BizErrorResponseStatus = &cluster.ResponseStatusDTO {Code:1}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *cluster.ClusterCreateReqDTO, resp *cluster.ClusterCreateRespDTO) error {
	return nil
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *cluster.ClusterQueryReqDTO, resp *cluster.ClusterQueryRespDTO) error {
	return nil
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *cluster.ClusterDeleteReqDTO, resp *cluster.ClusterDeleteRespDTO) error {
	return nil
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *cluster.ClusterDetailReqDTO, resp *cluster.ClusterDetailRespDTO) error {
	return nil
}
