package port

import "github.com/pingcap/ticp/cluster/domain/entity"

var clusterOperator ClusterOperator
var clusterMonitor ClusterMonitor

// ClusterOperator 集群操作
type ClusterOperator interface {
	// DeployCluster 部署一个集群
	DeployCluster(cluster *entity.Cluster, bizId string)

	// CheckProgress 查看处理过程
	CheckProgress(bizId string)
}

// ClusterMonitor 集群监控
type ClusterMonitor interface {

}