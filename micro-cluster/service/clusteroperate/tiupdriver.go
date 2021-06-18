package clusteroperate

import "github.com/pingcap/ticp/micro-cluster/service/clustermanage"

var clusterOperator ClusterOperator
var clusterMonitor ClusterMonitor

// ClusterOperator 集群操作
type ClusterOperator interface {
	// DeployCluster 部署一个集群
	DeployCluster(cluster *clustermanage.Cluster, bizId string)

	// CheckProgress 查看处理过程
	CheckProgress(bizId string)
}

// ClusterMonitor 集群监控
type ClusterMonitor interface {

}

// TiUPOperator 需要实现drivers里的 ClusterOperator
type TiUPOperator struct {}

func (*TiUPOperator) DeployCluster(cluster *clustermanage.Cluster, bizId string) error {
	// todo 韩森
	return nil
}

// CheckProgress 查看处理过程
func (*TiUPOperator) CheckProgress(bizId string) error {
	// todo 韩森
	return nil
}