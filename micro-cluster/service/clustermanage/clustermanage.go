package clustermanage

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"

	mngClient "github.com/pingcap/ticp/micro-manager/client"
	dbClient "github.com/pingcap/ticp/micro-metadb/client"

	mngPb "github.com/pingcap/ticp/micro-manager/proto"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"

	spec "github.com/pingcap/tiup/pkg/cluster/spec"
)

// Cluster 集群
type Cluster struct {
	Id         	uint
	TenantId   	uint
	Name       	string
	Status     	int
	Version    	string
	CreateTime 	time.Time

	Demand          ClusterDemand
	CurrentConfigId uint
	TiUPConfig      spec.Specification
}

// ClusterDemand 集群配置要求
type ClusterDemand struct {
	pdNodeQuantity   int
	tiDBNodeQuantity int
	tiKVNodeQuantity int
}

func CreateCluster(name, dbPassword, version string,
	tikvCount, tidbCount, pdCount int32,
	operatorName string, tenantId uint) (*Cluster, error) {
	log.Info("create cluster by ", operatorName)
	req := dbPb.DBCreateClusterRequest{
		Cluster: &dbPb.DBClusterDTO{
			Name:      name,
			TenantId: int32(tenantId),
			DbPassword: dbPassword,
			Version:   version,
			TikvCount: tikvCount,
			TidbCount: tidbCount,
			PdCount:   pdCount,
		},
	}

	resp, err :=  dbClient.DBClient.AddCluster(context.TODO(), &req)
	if err != nil {
		// 处理异常
	}
	cluster := new(Cluster)
	copyClusterDbDtoToDomain(resp.Cluster, cluster)

	flowWork := ClusterInitFlowWork()
	flowWork.context.Put("cluster", cluster)
	flowWork.moveOn("start")
	return cluster, nil
}

func copyClusterDbDtoToDomain(dto *dbPb.DBClusterDTO, domain *Cluster) {
	// 模型转换
	domain.Id = uint(dto.Id)
	domain.TenantId = uint(dto.TenantId)
	domain.Name = dto.Name
	domain.Status = int(dto.Status)
	domain.Version = dto.Version
	domain.Demand = ClusterDemand{
		tiDBNodeQuantity: int(dto.TidbCount),
		tiKVNodeQuantity: int(dto.TikvCount),
		pdNodeQuantity: int(dto.PdCount),
	}
}

// PrepareResource 申请主机的同步任务，还待抽象
func (cluster *Cluster) PrepareResource(f *FlowWork) {
	req := mngPb.AllocHostsRequest{
		PdCount:   int32(cluster.Demand.pdNodeQuantity),
		TidbCount: int32(cluster.Demand.tiDBNodeQuantity),
		TikvCount: int32(cluster.Demand.tiKVNodeQuantity),
	}
	resp, err := mngClient.ManagerClient.AllocHosts(context.TODO(), &req)

	if err != nil {
		// 处理远程异常
	}

	f.context.Put("hosts", resp.Hosts)
	f.moveOn("allocDone")
}

// BuildConfig 根据要求和申请到的主机，生成一份TiUP的配置
func (cluster *Cluster) BuildConfig(f *FlowWork) {

	hosts := f.context.Value("hosts").([]*mngPb.AllocHost)
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

	cluster.persistCurrentConfig()
	f.moveOn("configDone")
}

func (cluster *Cluster) persistCurrentConfig() {
	configByte, err := json.Marshal(cluster.TiUPConfig)
	if err != nil {
		// 处理json序列化异常
	}
	resp, err := dbClient.DBClient.UpdateTiUPConfig(context.TODO(), &dbPb.DBUpdateTiUPConfigRequest{
		ClusterId: int32(cluster.Id),
		ConfigContent: string(configByte),
	})

	cluster.CurrentConfigId = uint(resp.Config.Id)
}

func (cluster *Cluster) ExecuteTiUP(f *FlowWork) {
	f.currentTask = CreateTask()

	Operator.DeployCluster(cluster, uint64(f.currentTask.id))

	f.moveOn("tiUPStart")
}

func (cluster *Cluster) CheckTiUPResult(f *FlowWork) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for _ = range ticker.C {
			status, s, err := Operator.CheckProgress(uint64(f.currentTask.id))
			if err != nil {
				log.Error(err)
				continue
			}

			switch status {
			case dbPb.TiupTaskStatus_Init:
				log.Info(s)
			case dbPb.TiupTaskStatus_Processing:
				log.Info(s)
			case dbPb.TiupTaskStatus_Finished:
				log.Info(s)
				f.moveOn("tiUPDone")
				ticker.Stop()
			case dbPb.TiupTaskStatus_Error:
				log.Error(s)
				f.moveOn("tiUPDone")
				ticker.Stop()
			}
		}
	}()


}

func QueryCluster(page, pageSize int) (clusters []*Cluster, err error) {
	resp, err := dbClient.DBClient.ListCluster(context.TODO(), &dbPb.DBListClusterRequest{
		Page: 		int32(page),
		PageSize: 	int32(pageSize),
	})

	if err != nil {

	}

	for _, c := range resp.Clusters {
		cluster := new(Cluster)
		copyClusterDbDtoToDomain(c, cluster)
		clusters = append(clusters, cluster)
	}
	return
}

var Operator ClusterOperator
var Monitor ClusterMonitor

// ClusterOperator 集群操作
type ClusterOperator interface {
	// DeployCluster 部署一个集群
	DeployCluster(cluster *Cluster, bizId uint64) error
	// CheckProgress 查看处理过程
	CheckProgress(bizId uint64) (dbPb.TiupTaskStatus, string, error)
}

// ClusterMonitor 集群监控
type ClusterMonitor interface {
}