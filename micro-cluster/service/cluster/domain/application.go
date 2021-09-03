package domain

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"

	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap-inc/tiem/micro-cluster/service/host"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
)

type ClusterAggregation struct {
	Cluster *Cluster

	CurrentTiUPConfigRecord *TiUPConfigRecord
	CurrentWorkFlow         *FlowWorkEntity

	CurrentOperator *Operator

	MaintainCronTask *CronTaskEntity
	HistoryWorkFLows []*FlowWorkEntity

	UsedResources interface{}

	AvailableResources *proto.AllocHostResponse

	StatusModified bool
	FlowModified   bool

	ConfigModified bool

	LastBackupRecord *BackupRecord

	LastRecoverRecord *RecoverRecord

	LastParameterRecord *ParameterRecord
}

var contextClusterKey = "clusterAggregation"

func CreateCluster(ope *proto.OperatorDTO, clusterInfo *proto.ClusterBaseInfoDTO, demandDTOs []*proto.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	cluster := &Cluster{
		ClusterName:    clusterInfo.ClusterName,
		DbPassword:     clusterInfo.DbPassword,
		ClusterType:    *knowledge.ClusterTypeFromCode(clusterInfo.ClusterType.Code),
		ClusterVersion: *knowledge.ClusterVersionFromCode(clusterInfo.ClusterVersion.Code),
		Tls:            clusterInfo.Tls,
		TenantId:       operator.TenantId,
		OwnerId:        operator.Id,
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs), len(demandDTOs))

	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	cluster.Demands = demands

	// persist the cluster into database
	ClusterRepo.AddCluster(cluster)

	clusterAggregation := &ClusterAggregation{
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to create a cluster instance

	flow, err := CreateFlowWork(cluster.Id, FlowCreateCluster)
	if err != nil {
		// todo
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func (clusterAggregation *ClusterAggregation) updateWorkFlow(flow *FlowWorkEntity) {
	clusterAggregation.CurrentWorkFlow = flow
	clusterAggregation.FlowModified = true
}

func DeleteCluster(ope *proto.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	clusterAggregation.CurrentOperator = operator

	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}

	flow, err := CreateFlowWork(clusterAggregation.Cluster.Id, FlowDeleteCluster)
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	TaskRepo.Persist(flow)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func ListCluster(ope *proto.OperatorDTO, req *proto.ClusterQueryReqDTO) ([]*ClusterAggregation, int, error) {
	return ClusterRepo.Query(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int(req.PageReq.Page), int(req.PageReq.PageSize))
}

func GetClusterDetail(ope *proto.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	cluster, err := ClusterRepo.Load(clusterId)
	// todo 补充其他的信息
	return cluster, err
}

func ModifyParameters(ope *proto.OperatorDTO, clusterId string, content string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	clusterAggregation.CurrentOperator = operator
	clusterAggregation.LastParameterRecord = &ParameterRecord{
		ClusterId:  clusterId,
		OperatorId: operator.Id,
		Content:    content,
	}
	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}

	//currentFlow := clusterAggregation.CurrentWorkFlow
	//if currentFlow != nil && !currentFlow.Finished(){
	//	return clusterAggregation, errors.New("incomplete processing flow")
	//}

	flow, err := CreateFlowWork(clusterId, FlowModifyParameters)
	if err != nil {
		// todo
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func GetParameters(ope *proto.OperatorDTO, clusterId string) (parameterJson string, err error) {
	return InstanceRepo.QueryParameterJson(clusterId)
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

func prepareResource(task *TaskEntity, flowContext *FlowContext) bool {
	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)

	demands := clusterAggregation.Cluster.Demands

	clusterAggregation.AvailableResources = &proto.AllocHostResponse{}
	err := host.NewResourceManager().AllocHosts(context.TODO(), convertAllocHostsRequest(demands), clusterAggregation.AvailableResources)

	if err != nil {
		// todo
	}

	task.Success(nil)
	return true
}

func buildConfig(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)

	config := &TiUPConfigRecord{
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: convertConfig(clusterAggregation.AvailableResources, clusterAggregation.Cluster),
	}

	clusterAggregation.CurrentTiUPConfigRecord = config
	clusterAggregation.ConfigModified = true
	task.Success(config.Id)
	return true
}

func deployCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	spec := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel

	if true {
		bs, err := yaml.Marshal(spec)
		if err != nil {
			task.Fail(err)
			return false
		}

		cfgYamlStr := string(bs)
		go func() {
			_, err = libtiup.MicroSrvTiupDeploy(
				cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr, 0, []string{"--user", "root", "-i", "/root/.ssh/tiup_rsa"}, uint64(task.Id),
			)
		}()
	}

	return true
}

func startupCluster(task *TaskEntity, context *FlowContext) bool {
	// todo
	task.Success(nil)
	return true
}

func modifyParameters(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func deleteCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.Delete()
	clusterAggregation.StatusModified = true

	task.Success(nil)
	return true
}

func destroyCluster(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func freedResource(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func destroyTasks(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func (aggregation *ClusterAggregation) ExtractStatusDTO() *proto.DisplayStatusDTO {
	cluster := aggregation.Cluster

	dto := &proto.DisplayStatusDTO{
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
		BaseInfo:  aggregation.ExtractBaseInfoDTO(),
		Status:    aggregation.ExtractStatusDTO(),
		Instances: aggregation.ExtractInstancesDTO(),
	}
	return dto
}

func (aggregation *ClusterAggregation) ExtractMaintenanceDTO() *proto.ClusterMaintenanceDTO {
	dto := &proto.ClusterMaintenanceDTO{}
	if aggregation.MaintainCronTask != nil {
		dto.MaintainTaskCron = aggregation.MaintainCronTask.Cron
	} else {
		// default maintain ?
	}

	return dto
}

func (aggregation *ClusterAggregation) ExtractBaseInfoDTO() *proto.ClusterBaseInfoDTO {
	cluster := aggregation.Cluster
	return &proto.ClusterBaseInfoDTO{
		ClusterName: cluster.ClusterName,
		DbPassword:  cluster.DbPassword,
		ClusterType: &proto.ClusterTypeDTO{
			Code: cluster.ClusterType.Code,
			Name: cluster.ClusterType.Name,
		},
		ClusterVersion: &proto.ClusterVersionDTO{
			Code: cluster.ClusterVersion.Code,
			Name: cluster.ClusterVersion.Name,
		},
		Tags: cluster.Tags,
		Tls:  cluster.Tls,
	}
}

func (aggregation *ClusterAggregation) ExtractBackupRecordDTO() *proto.BackupRecordDTO {
	record := aggregation.LastBackupRecord
	currentFlow := aggregation.CurrentWorkFlow

	return &proto.BackupRecordDTO{
		Id:         record.Id,
		ClusterId:  record.ClusterId,
		Range:      string(record.Range),
		BackupType: string(record.BackupType),
		Size:       record.Size,
		StartTime:  record.StartTime,
		EndTime:    record.EndTime,
		FilePath:   record.FilePath,
		DisplayStatus: &proto.DisplayStatusDTO{
			InProcessFlowId: int32(currentFlow.Id),
			StatusCode:      strconv.Itoa(int(currentFlow.Status)),
			StatusName:      currentFlow.Status.Display(),
		},
		Operator: &proto.OperatorDTO{
			Id:       aggregation.CurrentOperator.Id,
			Name:     aggregation.CurrentOperator.Name,
			TenantId: aggregation.CurrentOperator.TenantId,
		},
	}
}

func (aggregation *ClusterAggregation) ExtractRecoverRecordDTO() *proto.BackupRecoverRecordDTO {
	record := aggregation.LastRecoverRecord
	currentFlow := aggregation.CurrentWorkFlow
	return &proto.BackupRecoverRecordDTO{
		Id:        int64(record.Id),
		ClusterId: record.ClusterId,
		DisplayStatus: &proto.DisplayStatusDTO{
			InProcessFlowId: int32(currentFlow.Id),
			StatusCode:      strconv.Itoa(int(currentFlow.Status)),
			StatusName:      currentFlow.Status.Display(),
		},
		BackupRecordId: record.BackupRecord.Id,
	}
}

func parseOperatorFromDTO(dto *proto.OperatorDTO) (operator *Operator) {
	operator = &Operator{
		Id:       dto.Id,
		Name:     dto.Name,
		TenantId: dto.TenantId,
	}
	return
}

func parseDistributionItemFromDTO(dto *proto.DistributionItemDTO) (item *ClusterNodeDistributionItem) {
	item = &ClusterNodeDistributionItem{
		ZoneCode: dto.ZoneCode,
		SpecCode: dto.SpecCode,
		Count:    int(dto.Count),
	}
	return
}

func parseRecoverInFoFromDTO(dto *proto.RecoverInfoDTO) (info RecoverInfo) {
	info = RecoverInfo{
		SourceClusterId: dto.SourceClusterId,
		BackupRecordId:  dto.BackupRecordId,
	}
	return
}

func parseNodeDemandFromDTO(dto *proto.ClusterNodeDemandDTO) (demand *ClusterComponentDemand) {
	items := make([]*ClusterNodeDistributionItem, len(dto.Items), len(dto.Items))

	for i, v := range dto.Items {
		items[i] = parseDistributionItemFromDTO(v)
	}

	demand = &ClusterComponentDemand{
		ComponentType:     knowledge.ClusterComponentFromCode(dto.ComponentType),
		TotalNodeCount:    int(dto.TotalNodeCount),
		DistributionItems: items,
	}

	return demand
}

func convertAllocHostsRequest(demands []*ClusterComponentDemand) (req *proto.AllocHostsRequest) {
	req = &proto.AllocHostsRequest{}

	for _, d := range demands {
		switch d.ComponentType.ComponentType {
		case "tidb":
			req.TidbReq = make([]*proto.AllocationReq, len(d.DistributionItems), len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.TidbReq[i] = convertAllocationReq(v)
			}
		case "tikv":
			req.TikvReq = make([]*proto.AllocationReq, len(d.DistributionItems), len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.TikvReq[i] = convertAllocationReq(v)
			}
		case "pd":
			req.PdReq = make([]*proto.AllocationReq, len(d.DistributionItems), len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.PdReq[i] = convertAllocationReq(v)
			}
		}
	}
	return
}

func convertAllocationReq(item *ClusterNodeDistributionItem) *proto.AllocationReq {
	return &proto.AllocationReq{
		FailureDomain: item.ZoneCode,
		CpuCores:      int32(knowledge.ParseCpu(item.SpecCode)),
		Memory:        int32(knowledge.ParseMemory(item.SpecCode)),
		Count:         int32(item.Count),
	}
}

func convertConfig(resource *proto.AllocHostResponse, cluster *Cluster) *spec.Specification {

	tidbHosts := resource.TidbHosts
	tikvHosts := resource.TikvHosts
	pdHosts := resource.PdHosts

	tiupConfig := new(spec.Specification)

	// Deal with Global Settings
	tiupConfig.GlobalOptions.DataDir = filepath.Join(cluster.Id, "tidb-data")
	tiupConfig.GlobalOptions.DeployDir = filepath.Join(cluster.Id, "tidb-deploy")
	tiupConfig.GlobalOptions.User = "tidb"
	tiupConfig.GlobalOptions.SSHPort = 22
	tiupConfig.GlobalOptions.Arch = "amd64"
	tiupConfig.GlobalOptions.LogDir = filepath.Join(cluster.Id, "tidb-log")
	// Deal with Promethus, AlertManger, Grafana
	tiupConfig.Monitors = append(tiupConfig.Monitors, &spec.PrometheusSpec{
		Host:      pdHosts[0].Ip,
		DataDir:   filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "prometheus-data"),
		DeployDir: filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "prometheus-deploy"),
	})
	tiupConfig.Alertmanagers = append(tiupConfig.Alertmanagers, &spec.AlertmanagerSpec{
		Host:      pdHosts[0].Ip,
		DataDir:   filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "alertmanagers-data"),
		DeployDir: filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "alertmanagers-deploy"),
	})
	tiupConfig.Grafanas = append(tiupConfig.Grafanas, &spec.GrafanaSpec{
		Host:      pdHosts[0].Ip,
		DeployDir: filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "grafanas-deploy"),
	})
	// Deal with PDServers, TiDBServers, TiKVServers
	for _, v := range pdHosts {
		tiupConfig.PDServers = append(tiupConfig.PDServers, &spec.PDSpec{
			Host:      v.Ip,
			DataDir:   filepath.Join(v.Disk.Path, cluster.Id, "pd-data"),
			DeployDir: filepath.Join(v.Disk.Path, cluster.Id, "pd-deploy"),
		})
	}
	for _, v := range tidbHosts {
		tiupConfig.TiDBServers = append(tiupConfig.TiDBServers, &spec.TiDBSpec{
			Host:      v.Ip,
			DeployDir: filepath.Join(v.Disk.Path, cluster.Id, "tidb-deploy"),
		})
	}
	for _, v := range tikvHosts {
		tiupConfig.TiKVServers = append(tiupConfig.TiKVServers, &spec.TiKVSpec{
			Host:      v.Ip,
			DataDir:   filepath.Join(v.Disk.Path, cluster.Id, "tikv-data"),
			DeployDir: filepath.Join(v.Disk.Path, cluster.Id, "tikv-deploy"),
		})
	}

	return tiupConfig
}
