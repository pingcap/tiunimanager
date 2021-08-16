package clusterapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type CreateReq struct {
	ClusterBaseInfo
	NodeDemandList []ClusterNodeDemand `json:"nodeDemandList"`
}

type QueryReq struct {
	controller.PageRequest
	ClusterId     string `json:"clusterId" form:"clusterId"`
	ClusterName   string `json:"clusterName" form:"clusterName"`
	ClusterType   string `json:"clusterType" form:"clusterType"`
	ClusterStatus string `json:"clusterStatus" form:"clusterStatus"`
	ClusterTag    string `json:"clusterTag" form:"clusterTag"`
}

type DescribeDashboardReq struct {
	ClusterId		string	`json:"clusterId"`
}
