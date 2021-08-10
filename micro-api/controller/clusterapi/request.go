package clusterapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type CreateReq struct {
	ClusterBaseInfo
	NodeDemandList  []ClusterNodeDemand	`json:"nodeDemandList"`
}


type QueryReq struct {
	controller.PageRequest
	ClusterId 		string	`json:"clusterId"`
	ClusterName 	string	`json:"clusterName"`
	ClusterType 	string	`json:"clusterType"`
	ClusterStatus 	string	`json:"clusterStatus"`
	ClusterTag 		string	`json:"clusterTag"`
}

