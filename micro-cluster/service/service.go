package service

import (
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ResponseStatus {Code:0}

type ClusterServiceHandler struct {}
