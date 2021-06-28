package clusteroperate

import (
	"github.com/pingcap/ticp/micro-cluster/service/clustermanage"
	"github.com/pingcap/ticp/micro-cluster/service/clusteroperate/libtiup"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
	"gopkg.in/yaml.v2"
)

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
type TiUPOperator struct{}

func (*TiUPOperator) DeployCluster(cluster *clustermanage.Cluster, bizId uint64) error {
	bs, err := yaml.Marshal(&cluster.TiUPConfig)
	if err != nil {
		return err
	}
	cfgYamlStr := string(bs)
	_, err = libtiup.MicroSrvTiupDeploy(
		cluster.Name, cluster.Version, cfgYamlStr, 0, nil, bizId,
	)
	return err
}

// CheckProgress 查看处理过程
func (*TiUPOperator) CheckProgress(bizId uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
	stat, statErrStr, err = libtiup.MicroSrvTiupGetTaskStatusByBizID(bizId)
	return
}
