package domain

import (
	"github.com/pingcap/ticp/micro-cluster/service/clustermanage"
)

type ClusterAggregation struct {
	Cluster 			*Cluster

	DemandRecords       []*ClusterDemandRecord

	InstanceProxy 		*ClusterInstanceProxy

	WorkFlow 			*clustermanage.FlowWorkEntity
	HistoryWorkFLows    []*clustermanage.FlowWorkEntity
}

func CreateCluster(operator Operator) (*ClusterAggregation, error) {
	cluster := &Cluster{}

	ClusterRepo.AddCluster(cluster)

	clusterAggregation := &ClusterAggregation{
		Cluster: cluster,
	}

	// start the workflow
	// persist workflow
	// persist ClusterAggregation

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

func GetClusterKnowledge() []*ClusterTypeSpec {
	if Knowledge == nil {
		loadKnowledge()
	}

	return Knowledge.SortedTypesKnowledge()
}