/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package domain

import (
	ctx "context"
	"errors"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/gommon/bytes"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/micro-cluster/service/resource"

	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
)

type ClusterAggregation struct {
	Cluster *Cluster
	ClusterMetadata spec.Metadata
	ClusterComponents []*ComponentGroup

	CurrentWorkFlow             *FlowWorkEntity
	CurrentOperator *Operator

	CurrentTopologyConfigRecord *TopologyConfigRecord

	MaintainCronTask *CronTaskEntity
	HistoryWorkFLows []*FlowWorkEntity

	UsedResources interface{}

	AvailableResources *clusterpb.AllocHostResponse

	BaseInfoModified bool
	StatusModified bool
	FlowModified   bool

	ConfigModified bool

	LastBackupRecord *BackupRecord

	LastRecoverRecord *RecoverRecord

	LastParameterRecord *ParameterRecord
}

var contextClusterKey = "clusterAggregation"
var contextTakeoverReqKey = "takeoverRequest"

func CreateCluster(ope *clusterpb.OperatorDTO, clusterInfo *clusterpb.ClusterBaseInfoDTO, demandDTOs []*clusterpb.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
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

	demands := make([]*ClusterComponentDemand, len(demandDTOs))

	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	cluster.Demands = demands

	// persist the cluster into database
	err := ClusterRepo.AddCluster(cluster)

	if err != nil {
		return nil, err
	}
	clusterAggregation := &ClusterAggregation{
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to create a cluster instance

	flow, err := CreateFlowWork(cluster.Id, FlowCreateCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

// TakeoverClusters
// @Description:
// @Parameter ope
// @Parameter req
// @return []*ClusterAggregation
// @return error
func TakeoverClusters(ope *clusterpb.OperatorDTO, req *clusterpb.ClusterTakeoverReqDTO) ([]*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	if len(req.ClusterNames) != 1 {
		return nil, common.NewBizError(common.TIEM_PARAMETER_INVALID)
	}

	clusterName := req.ClusterNames[0]
	cluster := &Cluster{
		ClusterName:    clusterName,
		TenantId:       operator.TenantId,
		OwnerId:        operator.Id,
	}

	// persist the cluster into database
	err := ClusterRepo.AddCluster(cluster)

	if err != nil {
		return nil, err
	}

	clusterAggregation := &ClusterAggregation {
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to takeover a cluster instance
	flow, err := CreateFlowWork(cluster.Id, FlowTakeoverCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextTakeoverReqKey, req)

	flow.Start()

	clusterAggregation.Cluster.Online()
	clusterAggregation.StatusModified = true
	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return []*ClusterAggregation{clusterAggregation}, nil
}

func (clusterAggregation *ClusterAggregation) updateWorkFlow(flow *FlowWorkEntity) {
	clusterAggregation.CurrentWorkFlow = flow
	clusterAggregation.FlowModified = true
}

func DeleteCluster(ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	clusterAggregation.CurrentOperator = operator

	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}

	var flow *FlowWorkAggregation
	flow, err = CreateFlowWork(clusterAggregation.Cluster.Id, FlowDeleteCluster, operator)
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	TaskRepo.Persist(flow)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func RestartCluster(ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}
	clusterAggregation.CurrentOperator = operator

	flow, err := CreateFlowWork(clusterAggregation.Cluster.Id, FlowRestartCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	return clusterAggregation, nil
}

func ListCluster(ope *clusterpb.OperatorDTO, req *clusterpb.ClusterQueryReqDTO) ([]*ClusterAggregation, int, error) {
	return ClusterRepo.Query(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int(req.PageReq.Page), int(req.PageReq.PageSize))
}

func GetClusterDetail(ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	cluster, err := ClusterRepo.Load(clusterId)

	return cluster, err
}

func ModifyParameters(ope *clusterpb.OperatorDTO, clusterId string, content string) (*ClusterAggregation, error) {
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

	flow, err := CreateFlowWork(clusterId, FlowModifyParameters, operator)
	if err != nil {
		// todo
		getLogger().Errorf("modify parameters clusterid = %s, content = %s, errStr: %s", clusterId, content, err.Error())
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func GetParameters(ope *clusterpb.OperatorDTO, clusterId string) (parameterJson string, err error) {
	return RemoteClusterProxy.QueryParameterJson(clusterId)
}

//func (aggregation *ClusterAggregation) loadWorkFlow() error {
//	if aggregation.Cluster.WorkFlowId > 0 && aggregation.CurrentWorkFlow == nil {
//		flowWork, err := TaskRepo.LoadFlowWork(aggregation.Cluster.WorkFlowId)
//		if err != nil {
//			return err
//		} else {
//			aggregation.CurrentWorkFlow = flowWork
//			return nil
//		}
//	}
//
//	return nil
//}

func prepareResource(task *TaskEntity, flowContext *FlowContext) bool {
	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)

	demands := clusterAggregation.Cluster.Demands

	clusterAggregation.AvailableResources = &clusterpb.AllocHostResponse{}
	err := resource.NewResourceManager().AllocHosts(ctx.TODO(), convertAllocHostsRequest(demands), clusterAggregation.AvailableResources)

	if err != nil {
		// todo
		getLogger().Error(err)
	}

	task.Success(nil)
	return true
}

func buildConfig(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)

	config := &TopologyConfigRecord{
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: convertConfig(clusterAggregation.AvailableResources, clusterAggregation.Cluster),
	}

	clusterAggregation.CurrentTopologyConfigRecord = config
	clusterAggregation.ConfigModified = true
	task.Success(config.Id)
	return true
}

func deployCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	spec := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel

	if true {
		bs, err := yaml.Marshal(spec)
		if err != nil {
			task.Fail(err)
			return false
		}

		cfgYamlStr := string(bs)
		getLogger().Infof("deploy cluster %s, version = %s, cfgYamlStr = %s", cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr)
		deployTaskId, _ := libtiup.MicroSrvTiupDeploy(
			cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr, 0, []string{"--user", "root", "-i", "/root/.ssh/tiup_rsa"}, uint64(task.Id),
		)
		context.put("deployTaskId", deployTaskId)
		getLogger().Infof("got deployTaskId %s", strconv.Itoa(int(deployTaskId)))
	}

	task.Success(nil)
	return true
}

func startupCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	var req dbpb.FindTiupTaskByIDRequest
	req.Id = context.value("deployTaskId").(uint64)

	for i := 0; i < 30; i++ {
		time.Sleep(10 * time.Second)
		rsp, err := client.DBClient.FindTiupTaskByID(ctx.TODO(), &req)
		if err != nil {
			getLogger().Errorf("get deploy task err = %s", err.Error())
			task.Fail(err)
			return false
		}
		if rsp.TiupTask.Status == dbpb.TiupTaskStatus_Error {
			getLogger().Errorf("deploy cluster error, %s", rsp.TiupTask.ErrorStr)
			task.Fail(errors.New(rsp.TiupTask.ErrorStr))
			return false
		}
		if rsp.TiupTask.Status == dbpb.TiupTaskStatus_Finished {
			break
		}
	}
	getLogger().Infof("start cluster %s", cluster.ClusterName)
	startTaskId, err := libtiup.MicroSrvTiupStart(cluster.ClusterName, 0, []string{}, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call tiup api start cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}
	context.put("startTaskId", startTaskId)
	getLogger().Infof("got startTaskId %s", strconv.Itoa(int(startTaskId)))

	task.Success(nil)
	return true
}

func setClusterOnline(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.StatusModified = true
	clusterAggregation.Cluster.Online()

	task.Success(nil)
	return true
}

func modifyParameters(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func fetchTopologyFile(task *TaskEntity, context *FlowContext) bool {
	req := context.value(contextTakeoverReqKey).(*clusterpb.ClusterTakeoverReqDTO)

	metadata, err := MetadataMgr.FetchFromRemoteCluster(ctx.TODO(), req)
	if err != nil {
		task.Fail(err)
		return false
	}

	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.ClusterMetadata = metadata

	clusterAggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: metadata.GetTopology().(*spec.Specification),
	}

	clusterType, _, _, version := MetadataMgr.ParseClusterInfoFromMetaData(*metadata.GetBaseMeta())
	cluster := clusterAggregation.Cluster
	cluster.ClusterType = *knowledge.ClusterTypeFromCode(clusterType)
	cluster.ClusterVersion = *knowledge.ClusterVersionFromCode(version)
	cluster.Tags = []string{"takeover"}

	clusterAggregation.ConfigModified = true
	clusterAggregation.BaseInfoModified = true

	task.Success(nil)
	return true
}

func buildTopology(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)

	components, err := MetadataMgr.ParseComponentsFromMetaData(clusterAggregation.ClusterMetadata)
	if err != nil {
		task.Fail(err)
		return false
	}

	clusterAggregation.ClusterComponents = components
	task.Success(nil)
	return true
}

func takeoverResource(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)

	allocReq , err := TopologyPlanner.AnalysisResourceRequest(clusterAggregation.Cluster, clusterAggregation.ClusterComponents)
	if err != nil {
		task.Fail(err)
		return false
	}

	allocResponse := &clusterpb.BatchAllocResponse{}
	resource.NewResourceManager().AllocResourcesInBatch(ctx.TODO(), allocReq, allocResponse)

	task.Success(allocReq)
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

// clusterRestart
// @Description: restart cluster
// @Parameter task
// @Parameter context
// @return bool
func clusterRestart(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	getLogger().Infof("restart cluster %s", cluster.ClusterName)
	restartTaskId, err := libtiup.MicroSrvTiupRestart(cluster.ClusterName, 0, []string{}, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call tiup api restart cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}
	context.put("restartTaskId", restartTaskId)
	getLogger().Infof("got restartTaskId %s", strconv.Itoa(int(restartTaskId)))

	go func() {
		// get cluster restart status async
		for {
			stat, statErrStr, err := libtiup.MicroSrvTiupGetTaskStatus(restartTaskId)
			if err != nil {
				getLogger().Errorf("call tiup api get task status statErrStr = %s, err = %s", statErrStr, err.Error())
				break
			}
			if stat == dbpb.TiupTaskStatus_Finished {
				getLogger().Infof(" cluster %s restart done.", cluster.ClusterName)
				clusterAggregation.StatusModified = true
				clusterAggregation.Cluster.Online()
				err := ClusterRepo.Persist(clusterAggregation)
				if err != nil {
					getLogger().Errorf("cluster repo persist err = %v", err)
					return
				}
				break
			} else if stat == dbpb.TiupTaskStatus_Error {
				getLogger().Infof(" cluster %s restart fail.", cluster.ClusterName)
				clusterAggregation.StatusModified = true
				clusterAggregation.Cluster.Status = ClusterStatusOffline
				err := ClusterRepo.Persist(clusterAggregation)
				if err != nil {
					getLogger().Errorf("cluster repo persist err = %v", err)
					return
				}
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	clusterAggregation.Cluster.Restart()
	clusterAggregation.StatusModified = true
	err = ClusterRepo.Persist(clusterAggregation)
	if err != nil {
		return false
	}
	task.Success(nil)
	return true
}

func (aggregation *ClusterAggregation) ExtractStatusDTO() *clusterpb.DisplayStatusDTO {
	cluster := aggregation.Cluster

	dto := &clusterpb.DisplayStatusDTO{
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

func (aggregation *ClusterAggregation) ExtractDisplayDTO() *clusterpb.ClusterDisplayDTO {
	dto := &clusterpb.ClusterDisplayDTO{
		ClusterId: aggregation.Cluster.Id,
		BaseInfo:  aggregation.ExtractBaseInfoDTO(),
		Status:    aggregation.ExtractStatusDTO(),
		Instances: aggregation.ExtractInstancesDTO(),
	}
	return dto
}

func (aggregation *ClusterAggregation) ExtractMaintenanceDTO() *clusterpb.ClusterMaintenanceDTO {
	dto := &clusterpb.ClusterMaintenanceDTO{}
	if aggregation.MaintainCronTask != nil {
		dto.MaintainTaskCron = aggregation.MaintainCronTask.Cron
	}
	//else {
	//	// default maintain ?
	//}

	return dto
}

func (aggregation *ClusterAggregation) ExtractBaseInfoDTO() *clusterpb.ClusterBaseInfoDTO {
	cluster := aggregation.Cluster
	return &clusterpb.ClusterBaseInfoDTO{
		ClusterName: cluster.ClusterName,
		DbPassword:  cluster.DbPassword,
		ClusterType: &clusterpb.ClusterTypeDTO{
			Code: cluster.ClusterType.Code,
			Name: cluster.ClusterType.Name,
		},
		ClusterVersion: &clusterpb.ClusterVersionDTO{
			Code: cluster.ClusterVersion.Code,
			Name: cluster.ClusterVersion.Name,
		},
		Tags: cluster.Tags,
		Tls:  cluster.Tls,
	}
}

func (aggregation *ClusterAggregation) ExtractBackupRecordDTO() *clusterpb.BackupRecordDTO {
	record := aggregation.LastBackupRecord
	currentFlow := aggregation.CurrentWorkFlow

	return &clusterpb.BackupRecordDTO{
		Id:           record.Id,
		ClusterId:    record.ClusterId,
		BackupMethod: string(record.BackupMethod),
		BackupType:   string(record.BackupType),
		BackupMode:   string(record.BackupMode),
		Size:         float32(record.Size) / bytes.MB, //Byte to MByte
		StartTime:    record.StartTime,
		EndTime:      record.EndTime,
		FilePath:     record.FilePath,
		DisplayStatus: &clusterpb.DisplayStatusDTO{
			InProcessFlowId: int32(currentFlow.Id),
			StatusCode:      strconv.Itoa(int(currentFlow.Status)),
			StatusName:      currentFlow.Status.Display(),
		},
		Operator: &clusterpb.OperatorDTO{
			Id:       aggregation.CurrentOperator.Id,
			Name:     aggregation.CurrentOperator.Name,
			TenantId: aggregation.CurrentOperator.TenantId,
		},
	}
}

func (aggregation *ClusterAggregation) ExtractRecoverRecordDTO() *clusterpb.BackupRecoverRecordDTO {
	record := aggregation.LastRecoverRecord
	currentFlow := aggregation.CurrentWorkFlow
	return &clusterpb.BackupRecoverRecordDTO{
		Id:        int64(record.Id),
		ClusterId: record.ClusterId,
		DisplayStatus: &clusterpb.DisplayStatusDTO{
			InProcessFlowId: int32(currentFlow.Id),
			StatusCode:      strconv.Itoa(int(currentFlow.Status)),
			StatusName:      currentFlow.Status.Display(),
		},
		BackupRecordId: record.BackupRecord.Id,
	}
}

func parseOperatorFromDTO(dto *clusterpb.OperatorDTO) (operator *Operator) {
	operator = &Operator{
		Id:       dto.Id,
		Name:     dto.Name,
		TenantId: dto.TenantId,
		ManualOperator: dto.ManualOperator,
	}
	return
}

func parseDistributionItemFromDTO(dto *clusterpb.DistributionItemDTO) (item *ClusterNodeDistributionItem) {
	item = &ClusterNodeDistributionItem{
		ZoneCode: dto.ZoneCode,
		SpecCode: dto.SpecCode,
		Count:    int(dto.Count),
	}
	return
}

func parseRecoverInFoFromDTO(dto *clusterpb.RecoverInfoDTO) (info RecoverInfo) {
	info = RecoverInfo{
		SourceClusterId: dto.SourceClusterId,
		BackupRecordId:  dto.BackupRecordId,
	}
	return
}

func parseNodeDemandFromDTO(dto *clusterpb.ClusterNodeDemandDTO) (demand *ClusterComponentDemand) {
	items := make([]*ClusterNodeDistributionItem, len(dto.Items))

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

func convertAllocHostsRequest(demands []*ClusterComponentDemand) (req *clusterpb.AllocHostsRequest) {
	req = &clusterpb.AllocHostsRequest{}

	for _, d := range demands {
		switch d.ComponentType.ComponentType {
		case "TiDB":
			req.TidbReq = make([]*clusterpb.AllocationReq, len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.TidbReq[i] = convertAllocationReq(v)
			}
		case "TiKV":
			req.TikvReq = make([]*clusterpb.AllocationReq, len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.TikvReq[i] = convertAllocationReq(v)
			}
		case "PD":
			req.PdReq = make([]*clusterpb.AllocationReq, len(d.DistributionItems))
			for i, v := range d.DistributionItems {
				req.PdReq[i] = convertAllocationReq(v)
			}
		default:

		}
	}
	return
}

func convertAllocationReq(item *ClusterNodeDistributionItem) *clusterpb.AllocationReq {
	return &clusterpb.AllocationReq{
		FailureDomain: item.ZoneCode,
		CpuCores:      int32(knowledge.ParseCpu(item.SpecCode)),
		Memory:        int32(knowledge.ParseMemory(item.SpecCode)),
		Count:         int32(item.Count),
	}
}

func tidbPort() int {
	return DefaultTidbPort
}

func convertConfig(resource *clusterpb.AllocHostResponse, cluster *Cluster) *spec.Specification {

	meta := new(spec.ClusterMeta)
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
		AnonymousEnable: true,
		DefaultTheme: "light",
		OrgName: "Main Org.",
		OrgRole: "Viewer",
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
			Port:      tidbPort(),
		})
	}
	for _, v := range tikvHosts {
		tiupConfig.TiKVServers = append(tiupConfig.TiKVServers, &spec.TiKVSpec{
			Host:      v.Ip,
			DataDir:   filepath.Join(v.Disk.Path, cluster.Id, "tikv-data"),
			DeployDir: filepath.Join(v.Disk.Path, cluster.Id, "tikv-deploy"),
		})
	}

	//meta.SetUser()
	//meta.SetVersion()
	meta.SetTopology(tiupConfig)
	return tiupConfig
}
