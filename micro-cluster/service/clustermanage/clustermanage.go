package clustermanage

import (
	"context"
	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"time"
)

// Cluster 集群
type Cluster struct {
	Id 				uint
	TenantId 		uint
	Name 			string
	Status  		ClusterStatus
	Version 		ClusterVersion
	CreateTime  	time.Time

	demand			ClusterDemand
	TiUPConfig  	TiUPConfig
}

type ClusterStatus 		int8
type ClusterVersion 	string

// ClusterDemand 集群配置要求
type ClusterDemand struct {
	pdNodeQuantity   int8
	tiDBNodeQuantity int8
	tiKVNodeQuantity int8
}

// TiUPConfig tiup配置
type TiUPConfig struct {
	Id				uint
	TenantId		uint
	ClusterId		uint
	// 其他TiUP执行需要的内容
	// TODO 家阳 补充TiUPConfig的结构化信息
}

func CreateCluster() {
	var cluster Cluster
	// 创建cluster，并保存到db，当前TiUPConfig 为空
	flowWork := ClusterInitFlowWork()
	flowWork.context.Put("cluster", &cluster)
	flowWork.moveOn("start")
}

// AllocTask 申请主机的同步任务，还待抽象
func (cluster *Cluster) AllocTask(f *FlowWork){
	req := manager.AllocHostsRequest{
		PdCount: int32(cluster.demand.pdNodeQuantity),
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
func (cluster *Cluster) BuildConfig(f *FlowWork){
	// todo 家阳，实现从用户要求到TiUP配置的转换
	config := TiUPConfig{}

	cluster.TiUPConfig = config

	f.moveOn("configDone")
}

func (cluster *Cluster) ExecuteTiUP(f *FlowWork){

	// todo 从韩森提供的接口去调用
	f.moveOn("tiUPStart")
}

func (cluster *Cluster) CheckTiUPResult(f *FlowWork){

	// todo 异步轮询韩森提供的接口？
	f.moveOn("tiUPDone")
}
