package management

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type CreateReq struct {
	ClusterBaseInfo
	NodeDemandList []ClusterNodeDemand `json:"nodeDemandList"`
}

type RestoreReq struct {
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
