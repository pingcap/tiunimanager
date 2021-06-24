package clustermanage

import (
	"context"
	"time"

	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	spec "github.com/pingcap/tiup/pkg/cluster/spec"
)

// Cluster 集群
type Cluster struct {
	Id         uint
	TenantId   uint
	Name       string
	Status     ClusterStatus
	Version    ClusterVersion
	CreateTime time.Time

	demand     ClusterDemand
	TiUPConfig spec.Specification
}

type ClusterStatus int8
type ClusterVersion string

// ClusterDemand 集群配置要求
type ClusterDemand struct {
	pdNodeQuantity   int8
	tiDBNodeQuantity int8
	tiKVNodeQuantity int8
}

func CreateCluster() {
	var cluster Cluster
	// 创建cluster，并保存到db，当前TiUPConfig 为空
	flowWork := ClusterInitFlowWork()
	flowWork.context.Put("cluster", &cluster)
	flowWork.moveOn("start")
}

// AllocTask 申请主机的同步任务，还待抽象
func (cluster *Cluster) AllocTask(f *FlowWork) {
	req := manager.AllocHostsRequest{
		PdCount:   int32(cluster.demand.pdNodeQuantity),
		TidbCount: int32(cluster.demand.tiDBNodeQuantity),
		TikvCount: int32(cluster.demand.tiKVNodeQuantity),
	}
	resp, err := client.ManagerClient.AllocHosts(context.TODO(), &req)

	if err != nil {
		// 处理远程异常
	}

	f.context.Put("hosts", resp.Hosts)
	f.moveOn("allocDone")
}

// BuildConfig 根据要求和申请到的主机，生成一份TiUP的配置
func (cluster *Cluster) BuildConfig(f *FlowWork) {

	hosts := f.context.Value("hosts").([]*manager.AllocHost)
	// Deal with Global Settings
	cluster.TiUPConfig.GlobalOptions.User = "tidb"
	cluster.TiUPConfig.GlobalOptions.SSHPort = 22
	cluster.TiUPConfig.GlobalOptions.Arch = "amd64"
	cluster.TiUPConfig.GlobalOptions.LogDir = "/tidb-log"
	// Deal with Promethus, AlertManger, Grafana
	cluster.TiUPConfig.Monitors = append(cluster.TiUPConfig.Monitors, &spec.PrometheusSpec{
		Host:      hosts[0].Ip,
		DataDir:   hosts[0].Disk.Path,
		DeployDir: hosts[0].Disk.Path,
	})
	cluster.TiUPConfig.Alertmanagers = append(cluster.TiUPConfig.Alertmanagers, &spec.AlertmanagerSpec{
		Host:      hosts[0].Ip,
		DataDir:   hosts[0].Disk.Path,
		DeployDir: hosts[0].Disk.Path,
	})
	cluster.TiUPConfig.Grafanas = append(cluster.TiUPConfig.Grafanas, &spec.GrafanaSpec{
		Host:      hosts[0].Ip,
		DeployDir: hosts[0].Disk.Path,
	})
	// Deal with PDServers, TiDBServers, TiKVServers
	for _, v := range hosts {
		cluster.TiUPConfig.PDServers = append(cluster.TiUPConfig.PDServers, &spec.PDSpec{
			Host:      v.Ip,
			DataDir:   v.Disk.Path,
			DeployDir: v.Disk.Path,
		})
		cluster.TiUPConfig.TiDBServers = append(cluster.TiUPConfig.TiDBServers, &spec.TiDBSpec{
			Host:      v.Ip,
			DeployDir: v.Disk.Path,
		})
		cluster.TiUPConfig.TiKVServers = append(cluster.TiUPConfig.TiKVServers, &spec.TiKVSpec{
			Host:      v.Ip,
			DataDir:   v.Disk.Path,
			DeployDir: v.Disk.Path,
		})
	}

	f.moveOn("configDone")
}

func (cluster *Cluster) ExecuteTiUP(f *FlowWork) {

	// todo 从韩森提供的接口去调用
	f.moveOn("tiUPStart")
}

func (cluster *Cluster) CheckTiUPResult(f *FlowWork) {

	// todo 异步轮询韩森提供的接口？
	f.moveOn("tiUPDone")
}
