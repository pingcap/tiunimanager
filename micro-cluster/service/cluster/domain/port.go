package domain

import "github.com/pingcap/ticp/micro-cluster/service/clustermanage"

var ClusterRepo ClusterRepository

var TaskRepo TaskRepository

type ClusterRepository interface {
	AddCluster (cluster *Cluster) error

	Persist (aggregation *ClusterAggregation) error
}

type TaskRepository interface {
	PersistFlowWork (flowWork *clustermanage.FlowWorkAggregation) error
	PersistCronTask (cronTask *clustermanage.CronTaskEntity) error
}
