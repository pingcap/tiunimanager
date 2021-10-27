package adapt

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
)

type DefaultTopologyPlanner struct {
}

func (d DefaultTopologyPlanner) BuildComponents(cluster *domain.Cluster, demands []*domain.ClusterComponentDemand) ([]domain.ComponentGroup, error) {
	panic("implement me")
}

func (d DefaultTopologyPlanner) AnalysisResourceRequest(components []domain.ComponentGroup) (clusterpb.BatchAllocRequest, error) {
	panic("implement me")
}

func (d DefaultTopologyPlanner) ApplyResourceToComponents(components []domain.ComponentGroup, response clusterpb.BatchAllocResponse) error {
	panic("implement me")
}

