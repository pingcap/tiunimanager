package management

import "github.com/pingcap-inc/tiem/micro-api/controller"

type CreateClusterRsp struct {
	ClusterId string `json:"clusterId"`
	ClusterBaseInfo
	controller.StatusInfo
}

type DeleteClusterRsp struct {
	ClusterId string `json:"clusterId"`
	controller.StatusInfo
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterMaintenanceInfo
	Components []ComponentInstance `json:"components"`
}

type DescribeDashboardRsp struct {
	ClusterId string `json:"clusterId"`
	Url       string `json:"url"`
	Token     string `json:"token"`
}
