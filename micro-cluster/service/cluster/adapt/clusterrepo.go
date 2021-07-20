package adapt

import "github.com/pingcap/ticp/micro-cluster/service/cluster/domain"

type ClusterRepoAdapter struct {}

func (c ClusterRepoAdapter) AddCluster(cluster *domain.Cluster) error {
	panic("implement me")
}

func (c ClusterRepoAdapter) Persist(aggregation *domain.ClusterAggregation) error {
	panic("implement me")
}

type TaskRepoAdapter struct {}

func (t TaskRepoAdapter) PersistFlowWork(flowWork *domain.FlowWorkAggregation) error {
	panic("implement me")
}

func (t TaskRepoAdapter) PersistCronTask(cronTask *domain.CronTaskEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) LoadFlowWork(id uint) (*domain.FlowWorkEntity, error) {
	panic("implement me")
}

func InjectionMetaDbRepo() {
	domain.TaskRepo = TaskRepoAdapter{}
	domain.ClusterRepo = ClusterRepoAdapter{}
}