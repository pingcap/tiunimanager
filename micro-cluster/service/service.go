package service

import (
	"context"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.TiDBClusterResponseStatus {Code:0}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateTiDBCluster(ctx context.Context, request *cluster.CreateTiDBClusterRequest, response *cluster.CreateTiDBClusterResponse) error {
	panic("implement me")
}

