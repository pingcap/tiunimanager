package service

import (
	"context"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ResponseStatus {Code:0}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, request *cluster.CreateClusterRequest, response *cluster.CreateClusterResponse) error {
	panic("implement me")
}

