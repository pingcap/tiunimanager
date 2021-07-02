package clusterapi

type CreateClusterRsp struct {
	ClusterId 			string
	ClusterBaseInfo
	ClusterStatusInfo
}

type ClusterKnowledgeRsp struct {
	ClusterTypes 		[]ClusterTypeInfo
}

type DeleteClusterRsp struct {
	ClusterId 			string
	ClusterStatusInfo
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterMaintenanceInfo
	components []ComponentInstance
}