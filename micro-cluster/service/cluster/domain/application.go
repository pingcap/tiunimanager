package domain

import (
	"errors"
	"fmt"
	"github.com/pingcap/ticp/knowledge"
	proto "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service/clusteroperate/libtiup"
	mngPb "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strconv"
)

type ClusterAggregation struct {
	Cluster 				*Cluster

	CurrentTiUPConfigRecord *TiUPConfigRecord
	CurrentWorkFlow  		*FlowWorkEntity

	MaintainCronTask 		*CronTaskEntity
	HistoryWorkFLows 		[]*FlowWorkEntity

	UsedResources 			interface{}
	
	AvailableResources		interface{}

	StatusModified 			bool
	FlowModified 			bool

	ConfigModified 			bool
}

var contextClusterKey = "clusterAggregation"

func CreateCluster(ope *proto.OperatorDTO, clusterInfo *proto.ClusterBaseInfoDTO, demandDTOs []*proto.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	cluster := &Cluster{
		ClusterName: clusterInfo.ClusterName,
		DbPassword: clusterInfo.DbPassword,
		ClusterType: *knowledge.ClusterTypeFromCode(clusterInfo.ClusterType.Code),
		ClusterVersion: *knowledge.ClusterVersionFromCode(clusterInfo.ClusterVersion.Code),
		Tls: clusterInfo.Tls,

		OwnerId: operator.Id,
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs), len(demandDTOs))

	for i,v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	cluster.Demands = demands

	// persist the cluster into database
	ClusterRepo.AddCluster(cluster)

	clusterAggregation := &ClusterAggregation{
		Cluster: cluster,
	}

	// Start the workflow to create a cluster instance

	flow, err := CreateFlowWork(cluster.Id, FlowCreateCluster)
	if err != nil {
		// todo
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func DeleteCluster(ope *proto.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)
	log.Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)

	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}


	flow, err := CreateFlowWork(clusterAggregation.Cluster.Id, FlowDeleteCluster)
	flow.AddContext(contextClusterKey, clusterAggregation)

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	TaskRepo.Persist(flow)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func ListCluster(operator Operator) ([]*ClusterAggregation, error) {
	return nil, nil
}

func GetClusterDetail(clusterId string, ope *proto.OperatorDTO) (*ClusterAggregation, error) {
	cluster, err := ClusterRepo.Load(clusterId)
	// todo 补充其他的信息
	return cluster, err
}

func (aggregation *ClusterAggregation) loadWorkFlow() error {
	if aggregation.Cluster.WorkFlowId > 0 && aggregation.CurrentWorkFlow == nil {
		flowWork, err := TaskRepo.LoadFlowWork(aggregation.Cluster.WorkFlowId)
		if err != nil {
			return err
		} else {
			aggregation.CurrentWorkFlow = flowWork
			return nil
		}
	}

	return nil
}

func prepareResource(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	if true {
		// todo
	}

	clusterAggregation.AvailableResources = nil
	task.Success(nil)
	return true
}

func convertConfig(resource interface{}, cluster *Cluster) *spec.Specification {
	hosts := resource.([]*mngPb.AllocHost)

	tiupConfig := new(spec.Specification)

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
	return tiupConfig
}

func buildConfig(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)

	resources := clusterAggregation.AvailableResources
	// todo
	fmt.Println(resources)

	config := &TiUPConfigRecord {
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: convertConfig(resources, clusterAggregation.Cluster),
	}

	clusterAggregation.CurrentTiUPConfigRecord = config
	task.Success(config.Id)
	return true
}

func deployCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	cluster := clusterAggregation.Cluster
	spec := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel

	if true {
		bs, err := yaml.Marshal(spec)
		if err != nil {
			task.Fail(err)
			return false
		}

		cfgYamlStr := string(bs)
		_, err = libtiup.MicroSrvTiupDeploy(
			cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr, 0, []string{"-i", "/root/.ssh/id_rsa_tiup_test"}, uint64(task.Id),
		)
	}

	return true
}

func startupCluster(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func deleteCluster(task *TaskEntity, context *FlowContext) bool {
	panic("implement me")
}

func destroyCluster(task *TaskEntity, context *FlowContext) bool {
	panic("implement me")
}

func freedResource(task *TaskEntity, context *FlowContext) bool {
	panic("implement me")
}

func destroyTasks(task *TaskEntity, context *FlowContext) bool {
	panic("implement me")
}

func (aggregation *ClusterAggregation) ExtractStatusDTO() *proto.DisplayStatusDTO{
	cluster := aggregation.Cluster

	dto := &proto.DisplayStatusDTO {
		CreateTime:      cluster.CreateTime.Unix(),
		UpdateTime:      cluster.UpdateTime.Unix(),
		DeleteTime:      cluster.DeleteTime.Unix(),
		InProcessFlowId: int32(cluster.WorkFlowId),
	}

	if cluster.WorkFlowId > 0 {
		dto.StatusCode = aggregation.CurrentWorkFlow.FlowName
		dto.StatusName = aggregation.CurrentWorkFlow.StatusAlias
	} else {
		dto.StatusCode = strconv.Itoa(int(aggregation.Cluster.Status))
		dto.StatusName = aggregation.Cluster.Status.Display()
	}

	return dto
}

func (aggregation *ClusterAggregation) GetCurrentWorkFlow() *FlowWorkEntity {
	if aggregation.CurrentWorkFlow != nil {
		return aggregation.CurrentWorkFlow
	}

	if aggregation.Cluster.WorkFlowId > 0 {
		// todo 从DB获取
		return nil
	}

	return nil
}

func (aggregation *ClusterAggregation) ExtractDisplayDTO() *proto.ClusterDisplayDTO {
	dto := &proto.ClusterDisplayDTO{
		ClusterId: aggregation.Cluster.Id,
		BaseInfo: aggregation.ExtractBaseInfoDTO(),
		Status: aggregation.ExtractStatusDTO(),
		Instances: aggregation.ExtractInstancesDTO(),
	}
	return dto
}

func (aggregation *ClusterAggregation) ExtractMaintenanceDTO() *proto.ClusterMaintenanceDTO {
	dto := &proto.ClusterMaintenanceDTO{}
	if aggregation.MaintainCronTask != nil {
		dto.MaintainTaskCron = aggregation.MaintainCronTask.cron
	} else {
		// default maintain ?
	}

	return dto
}

func (aggregation *ClusterAggregation) ExtractBaseInfoDTO() *proto.ClusterBaseInfoDTO {
	cluster :=  aggregation.Cluster
	return &proto.ClusterBaseInfoDTO {
		ClusterName: cluster.ClusterName,
		DbPassword: cluster.DbPassword,
		ClusterType: &proto.ClusterTypeDTO{
			Code: cluster.ClusterType.Code,
			Name: cluster.ClusterType.Name,
		},
		ClusterVersion: &proto.ClusterVersionDTO{
			Code: cluster.ClusterVersion.Code,
			Name: cluster.ClusterVersion.Name,
		},
		Tags: cluster.Tags,
		Tls: cluster.Tls,
	}
}

func parseOperatorFromDTO(dto *proto.OperatorDTO) (operator *Operator) {
	operator = &Operator{
		Id: dto.Id,
		Name: dto.Name,
		TenantId: dto.TenantId,
	}
	return
}

func parseDistributionItemFromDTO(dto *proto.DistributionItemDTO) (item *ClusterNodeDistributionItem) {
	item = &ClusterNodeDistributionItem{
		ZoneCode: dto.ZoneCode,
		SpecCode: dto.SpecCode,
		Count: int(dto.Count),
	}
	return
}

func parseNodeDemandFromDTO(dto *proto.ClusterNodeDemandDTO) (demand *ClusterComponentDemand) {
	items := make([]*ClusterNodeDistributionItem, len(dto.Items), len(dto.Items))

	for i,v := range dto.Items {
		items[i] = parseDistributionItemFromDTO(v)
	}

	demand = &ClusterComponentDemand{
		ComponentType:  knowledge.ClusterComponentFromCode(dto.ComponentType),
		TotalNodeCount: int(dto.TotalNodeCount),
		DistributionItems: items,
	}

	return demand
}
