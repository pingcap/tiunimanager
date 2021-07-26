package domain

var ClusterRepo ClusterRepository

var TaskRepo TaskRepository

type ClusterRepository interface {
	AddCluster (cluster *Cluster) error

	Persist(aggregation *ClusterAggregation) error
	Load (id string) (cluster *ClusterAggregation, err error)
}

type TaskRepository interface {
	AddFlowWork(flowWork *FlowWorkEntity) error
	AddTask(task *TaskEntity) error
	AddCronTask(cronTask *CronTaskEntity) error

	Persist(flowWork *FlowWorkAggregation) error

	LoadFlowWork(id uint) (*FlowWorkEntity, error)
	Load(id uint) (flowWork *FlowWorkAggregation, err error)
}
