package clusterapi

import "github.com/pingcap/ticp/micro-api/controller"

type CreateRequest struct {
	ClusterBaseInfo
	NodeDemandList  []ClusterNodeDemand
}

type QueryRequest struct {
	controller.PageRequest
}

