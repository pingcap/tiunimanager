package clusterapi

import "github.com/pingcap/ticp/micro-api/controller"

type CreateClusterRsp struct {
	ClusterId 			string
	ClusterBaseInfo
	controller.StatusInfo
}

type DeleteClusterRsp struct {
	ClusterId 			string
	controller.StatusInfo
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterMaintenanceInfo
	Components []ComponentInstance
}
