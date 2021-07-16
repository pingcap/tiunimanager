package clusteroperate

import (
	"github.com/pingcap/ticp/micro-cluster/service/clustermanage"
	"github.com/pingcap/ticp/micro-cluster/service/clusteroperate/libtiup"
	mngPb "github.com/pingcap/ticp/micro-manager/proto"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
	"gopkg.in/yaml.v2"
	"path/filepath"

	spec "github.com/pingcap/tiup/pkg/cluster/spec"
)

// TiUPOperator 需要实现drivers里的 ClusterOperator
type TiUPOperator struct{}

func (TiUPOperator) DeployCluster(cluster *clustermanage.Cluster, bizId uint64) error {
	bs, err := yaml.Marshal(&cluster.TiUPConfig)
	if err != nil {
		return err
	}
	cfgYamlStr := string(bs)
	_, err = libtiup.MicroSrvTiupDeploy(
		cluster.Name, cluster.Version, cfgYamlStr, 0, []string{"-i", "/root/.ssh/id_rsa_tiup_test"}, bizId,
	)
	return err
}

func (TiUPOperator) BuildConfig(cluster *clustermanage.Cluster, bizId uint64) error {
	tiupConfig := new(spec.Specification)
	hosts := f.context.Value("hosts").([]*mngPb.AllocHost)
	dataDir := filepath.Join(hosts[0].Disk.Path, "data")
	deployDir := filepath.Join(hosts[0].Disk.Path, "deploy")
	// Deal with Global Settings
	tiupConfig.GlobalOptions.DataDir = dataDir
	tiupConfig.GlobalOptions.DeployDir = deployDir
	tiupConfig.GlobalOptions.User = "tidb"
	tiupConfig.GlobalOptions.SSHPort = 22
	tiupConfig.GlobalOptions.Arch = "amd64"
	tiupConfig.GlobalOptions.LogDir = "/tidb-log"
	// Deal with Promethus, AlertManger, Grafana
	tiupConfig.Monitors = append(tiupConfig.Monitors, &spec.PrometheusSpec{
		Host: hosts[0].Ip,
	})
	tiupConfig.Alertmanagers = append(tiupConfig.Alertmanagers, &spec.AlertmanagerSpec{
		Host: hosts[0].Ip,
	})
	tiupConfig.Grafanas = append(tiupConfig.Grafanas, &spec.GrafanaSpec{
		Host: hosts[0].Ip,
	})
	// Deal with PDServers, TiDBServers, TiKVServers
	for _, v := range hosts {
		tiupConfig.PDServers = append(tiupConfig.PDServers, &spec.PDSpec{
			Host: v.Ip,
		})
		tiupConfig.TiDBServers = append(tiupConfig.TiDBServers, &spec.TiDBSpec{
			Host: v.Ip,
		})
		tiupConfig.TiKVServers = append(tiupConfig.TiKVServers, &spec.TiKVSpec{
			Host: v.Ip,
		})
	}

}

// CheckProgress 查看处理过程
func (TiUPOperator) CheckProgress(bizId uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
	stat, statErrStr, err = libtiup.MicroSrvTiupGetTaskStatusByBizID(bizId)
	return
}
