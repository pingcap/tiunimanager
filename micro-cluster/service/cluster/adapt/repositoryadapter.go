package adapt

import "github.com/pingcap/ticp/micro-cluster/service/cluster/domain"

type ClusterRepoAdapter struct {}

func (c ClusterRepoAdapter) AddCluster(cluster *domain.Cluster) error {
	panic("implement me")
}

func (c ClusterRepoAdapter) LoadCluster(id string) (cluster *domain.Cluster, err error) {
	panic("implement me")
}

func (c ClusterRepoAdapter) Persist(aggregation *domain.ClusterAggregation) error {
	panic("implement me")
}

func (c ClusterRepoAdapter) Load(id string) (cluster *domain.ClusterAggregation, err error) {
	panic("implement me")
}

type TaskRepoAdapter struct {}

func (t TaskRepoAdapter) AddFlowWork(flowWork *domain.FlowWorkEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) AddTask(task *domain.TaskEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) AddCronTask(cronTask *domain.CronTaskEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) Persist(flowWork *domain.FlowWorkAggregation) error {
	panic("implement me")
}

func (t TaskRepoAdapter) LoadFlowWork(id uint) (*domain.FlowWorkEntity, error) {
	panic("implement me")
}

func (t TaskRepoAdapter) Load(id uint) (flowWork *domain.FlowWorkAggregation, err error) {
	panic("implement me")
}

func InjectionMetaDbRepo() {
	domain.TaskRepo = TaskRepoAdapter{}
	domain.ClusterRepo = ClusterRepoAdapter{}
}