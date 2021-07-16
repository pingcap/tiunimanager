package domain

import (
	expert "github.com/pingcap/ticp/knowledge"
	proto "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service/clustermanage"
)

type ClusterAggregation struct {
	Cluster 			*Cluster

	DemandRecords       []*ClusterDemandRecord

	InstanceProxy 		*ClusterInstanceProxy

	WorkFlow 			*clustermanage.FlowWorkEntity
	HistoryWorkFLows    []*clustermanage.FlowWorkEntity
}

func ParseNodeDemandFromDTO(dto *proto.ClusterNodeDemandDTO) *ClusterComponentDemand {
	
}

func CreateCluster(operator Operator, clusterInfo *proto.ClusterBaseInfoDTO, demandDTOs []*proto.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	cluster := &Cluster{
		ClusterName: clusterInfo.ClusterName,
		DbPassword: clusterInfo.DbPasswordName,
		ClusterType: *expert.ClusterTypeFromCode(clusterInfo.ClusterType.Code),
		ClusterVersion: *expert.ClusterVersionFromCode(clusterInfo.ClusterVersion.Code),
		Tls: clusterInfo.Tls,
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs), len(demandDTOs))

	for i,v := range demandDTOs {
		demands[i] = ParseNodeDemandFromDTO(v)
	}

	cluster.Demand = demands

	ClusterRepo.AddCluster(cluster)

	clusterAggregation := &ClusterAggregation{
		Cluster: cluster,
	}

	ClusterRepo.Persist(clusterAggregation)
	return nil, nil
}

func DeleteCluster(clusterId string, operator Operator) (string, error) {
	return "", nil
}

func ListCluster(operator Operator) ([]*ClusterAggregation, error) {
	return nil, nil
}

func GetClusterDetail(clusterId string, operator Operator) (*ClusterAggregation, error) {
	return nil, nil
}
