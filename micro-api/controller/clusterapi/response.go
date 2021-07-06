package clusterapi

import "github.com/pingcap/ticp/micro-api/controller"

type CreateClusterRsp struct {
	ClusterId 			string
	ClusterBaseInfo
	controller.StatusInfo
}

type ClusterKnowledgeRsp struct {
	ClusterTypes 		[]ClusterTypeKnowledge
}

type DeleteClusterRsp struct {
	ClusterId 			string
	controller.StatusInfo
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterMaintenanceInfo
	components []ComponentInstance
}
