package clusterapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type CreateReq struct {
	ClusterBaseInfo
	NodeDemandList  []ClusterNodeDemand
}


type QueryReq struct {
	controller.PageRequest
	ClusterId 		string
	ClusterName 	string
	ClusterType 	string
	ClusterStatus 	string
	ClusterTag 		string
}

