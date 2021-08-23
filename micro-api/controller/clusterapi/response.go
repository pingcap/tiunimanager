package clusterapi

import "github.com/pingcap/tiem/micro-api/controller"

type CreateClusterRsp struct {
	ClusterId 			string	`json:"clusterId"`
	ClusterBaseInfo
	controller.StatusInfo
}

type DeleteClusterRsp struct {
	ClusterId 			string	`json:"clusterId"`
	controller.StatusInfo
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterMaintenanceInfo
	Components []ComponentInstance	`json:"components"`
}

type DescribeDashboardRsp struct {
	ClusterId		string		`json:"clusterId"`
	Url 			string 		`json:"url"`
	ShareCode 		string 		`json:"shareCode"`
}