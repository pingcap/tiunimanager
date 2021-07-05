package clusterapi

type CreateReq struct {
	ClusterBaseInfo
	NodeDemandList  []ClusterNodeDemand
}
