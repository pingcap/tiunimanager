package adapt

import "github.com/pingcap/ticp/cluster/domain/entity"

// TiUPOperator 需要实现drivers里的 ClusterOperator
type TiUPOperator struct {}

func (*TiUPOperator) DeployCluster(cluster *entity.Cluster, bizId string) error {
	// todo 韩森
	return nil
}

// CheckProgress 查看处理过程
func (*TiUPOperator) CheckProgress(bizId string) error {
	// todo 韩森
	return nil
}