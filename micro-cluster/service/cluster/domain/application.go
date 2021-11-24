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
	"fmt"
	resourceType "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/gommon/bytes"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/service/resource"

	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
)

type ClusterAggregation struct {
	Cluster         *Cluster
	ClusterMetadata spec.Metadata

	AddedComponentDemand   []*ClusterComponentDemand
	AddedClusterComponents []*ComponentGroup

	CurrentComponentInstances []*ComponentInstance

	CurrentWorkFlow *FlowWorkEntity
	CurrentOperator *Operator

	CurrentTopologyConfigRecord *TopologyConfigRecord
	AlteredTopology             *spec.Specification

	MaintainCronTask *CronTaskEntity
	HistoryWorkFLows []*FlowWorkEntity

	AddedAllocResources *clusterpb.BatchAllocResponse

	BaseInfoModified bool
	StatusModified   bool
	FlowModified     bool

	ConfigModified  bool
	DemandsModified bool

	LastBackupRecord *BackupRecord

	LastRecoverRecord *RecoverRecord

	LastParameterRecord *ParameterRecord
}

var contextClusterKey = "clusterAggregation"
var contextTakeoverReqKey = "takeoverRequest"
var contextDeleteNodeKey = "deleteNode"
var contextAllocRequestKey = "allocResourceRequest"

func (cluster *ClusterAggregation) tryStartFlow(ctx ctx.Context, flow *FlowWorkAggregation) error {
	if cluster.CurrentWorkFlow != nil {
		flow.Destroy("task conflicts")
		return framework.NewTiEMErrorf(common.TIEM_TASK_CONFLICT, "task conflicts, current task %s, expected task %s", cluster.CurrentWorkFlow.FlowName, flow.Define.FlowName)
	}

	cluster.updateWorkFlow(flow.FlowWork)
	cluster.Cluster.WorkFlowId = flow.FlowWork.Id

	err := ClusterRepo.PersistStatus(ctx, cluster)

	if err != nil {
		return err
	} else {
		flow.AsyncStart()
		return nil
	}
}

func CreateCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterInfo *clusterpb.ClusterBaseInfoDTO, commonDemand *clusterpb.ClusterCommonDemandDTO, demandDTOs []*clusterpb.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	cluster := &Cluster{
		ClusterName:     clusterInfo.ClusterName,
		DbPassword:      clusterInfo.DbPassword,
		ClusterType:     *knowledge.ClusterTypeFromCode(clusterInfo.ClusterType.Code),
		ClusterVersion:  *knowledge.ClusterVersionFromCode(clusterInfo.ClusterVersion.Code),
		Tls:             clusterInfo.Tls,
		TenantId:        operator.TenantId,
		OwnerId:         operator.Id,
		Region:          commonDemand.Region,
		CpuArchitecture: commonDemand.CpuArchitecture,
		Exclusive:       commonDemand.Exclusive,
	}

	if cluster.CpuArchitecture == "" {
		cluster.CpuArchitecture = string(resourceType.X86)
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs))

	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	// add default parasite components
	demands = append(demands,
		&ClusterComponentDemand{
			ComponentType:     knowledge.ClusterComponentFromCode("Grafana"),
			TotalNodeCount:    1,
			DistributionItems: []*ClusterNodeDistributionItem{},
		},
		&ClusterComponentDemand{ComponentType: knowledge.ClusterComponentFromCode("Prometheus"),
			TotalNodeCount:    1,
			DistributionItems: []*ClusterNodeDistributionItem{},
		},
		&ClusterComponentDemand{ComponentType: knowledge.ClusterComponentFromCode("AlertManger"),
			TotalNodeCount:    1,
			DistributionItems: []*ClusterNodeDistributionItem{},
		},
	)

	// persist the cluster into database
	err := ClusterRepo.AddCluster(ctx, cluster)

	if err != nil {
		return nil, err
	}
	clusterAggregation := &ClusterAggregation{
		Cluster:              cluster,
		MaintainCronTask:     GetDefaultMaintainTask(),
		CurrentOperator:      operator,
		AddedComponentDemand: demands,
	}

	// Start the workflow to create a cluster instance
	flow, err := CreateFlowWork(ctx, cluster.Id, FlowCreateCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(ctx, clusterAggregation)
	return clusterAggregation, nil
}

// ScaleOutCluster
// @Description: scale out a cluster
// @Parameter ope
// @Parameter clusterId
// @Parameter demands
// @return *ClusterAggregation
// @return error
func ScaleOutCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string, demandDTOs []*clusterpb.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	// Get cluster info from db based by clusterId
	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}
	operator := parseOperatorFromDTO(ope)
	clusterAggregation.CurrentOperator = operator

	// Parse resources for scaling out
	demands := make([]*ClusterComponentDemand, len(demandDTOs))
	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	// Merge multi demands
	clusterAggregation.AddedComponentDemand = demands
	clusterAggregation.DemandsModified = true

	// Start the workflow to scale out a cluster
	flow, err := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowScaleOutCluster, operator)
	if err != nil {
		return nil, err
	}
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	// Update workflow and persist clusterAggregation
	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(ctx, clusterAggregation)
	return clusterAggregation, nil
}

// ScaleInCluster
// @Description: scale in a cluster
// @Parameter ope
// @Parameter clusterId
// @Parameter nodeId
// @return *ClusterAggregation
// @return error
func ScaleInCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId, nodeId string) (*ClusterAggregation, error) {
	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}
	operator := parseOperatorFromDTO(ope)
	clusterAggregation.CurrentOperator = operator

	// Start the workflow to scale in a cluster
	flow, err := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowScaleInCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDeleteNodeKey, nodeId)
	flow.Start()

	// Update workflow and persist clusterAggregation
	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(ctx, clusterAggregation)
	return clusterAggregation, nil
}

// TakeoverClusters
// @Description:
// @Parameter ope
// @Parameter req
// @return []*ClusterAggregation
// @return error
func TakeoverClusters(ctx ctx.Context, ope *clusterpb.OperatorDTO, req *clusterpb.ClusterTakeoverReqDTO) ([]*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	if len(req.ClusterNames) != 1 {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	clusterName := req.ClusterNames[0]
	cluster := &Cluster{
		ClusterName: clusterName,
		TenantId:    operator.TenantId,
		OwnerId:     operator.Id,
	}

	// persist the cluster into database
	err := ClusterRepo.AddCluster(ctx, cluster)

	if err != nil {
		return nil, err
	}

	clusterAggregation := &ClusterAggregation{
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to takeover a cluster instance
	flow, err := CreateFlowWork(ctx, cluster.Id, FlowTakeoverCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextTakeoverReqKey, req)

	flow.Start()

	clusterAggregation.Cluster.Online()
	clusterAggregation.StatusModified = true
	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(ctx, clusterAggregation)
	return []*ClusterAggregation{clusterAggregation}, nil
}

func (clusterAggregation *ClusterAggregation) updateWorkFlow(flow *FlowWorkEntity) {
	clusterAggregation.CurrentWorkFlow = flow
	clusterAggregation.FlowModified = true
}

func DeleteCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	clusterAggregation.CurrentOperator = operator

	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}

	flow, _ := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowDeleteCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	err = clusterAggregation.tryStartFlow(ctx, flow)

	return clusterAggregation, err
}

func RestartCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil {
		return clusterAggregation, framework.SimpleError(common.TIEM_CLUSTER_NOT_FOUND)
	}
	clusterAggregation.CurrentOperator = operator

	flow, err := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowRestartCluster, operator)
	if err != nil {
		return nil, err
	}

	clusterAggregation.Cluster.Restart()
	clusterAggregation.StatusModified = true
	flow.AddContext(contextClusterKey, clusterAggregation)

	err = clusterAggregation.tryStartFlow(ctx, flow)

	return clusterAggregation, err
}

func StopCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil {
		return clusterAggregation, errors.New("cluster not exist")
	}
	clusterAggregation.CurrentOperator = operator

	flow, err := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowStopCluster, operator)
	if err != nil {
		return nil, err
	}

	clusterAggregation.Cluster.Status = ClusterStatusStopping
	clusterAggregation.StatusModified = true
	flow.AddContext(contextClusterKey, clusterAggregation)

	err = clusterAggregation.tryStartFlow(ctx, flow)

	return clusterAggregation, err
}

func ListCluster(ctx ctx.Context, ope *clusterpb.OperatorDTO, req *clusterpb.ClusterQueryReqDTO) ([]*ClusterAggregation, int, error) {
	return ClusterRepo.Query(ctx, req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int(req.PageReq.Page), int(req.PageReq.PageSize))
}

func GetClusterDetail(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string) (*ClusterAggregation, error) {
	cluster, err := ClusterRepo.Load(ctx, clusterId)

	return cluster, err
}

func ModifyParameters(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string, content string) (*ClusterAggregation, error) {
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
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

	flow, err := CreateFlowWork(ctx, clusterId, FlowModifyParameters, operator)
	if err != nil {
		// todo
		getLogger().Errorf("modify parameters clusterid = %s, content = %s, errStr: %s", clusterId, content, err.Error())
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(ctx, clusterAggregation)
	return clusterAggregation, nil
}

func GetParameters(ctx ctx.Context, ope *clusterpb.OperatorDTO, clusterId string) (parameterJson string, err error) {
	return RemoteClusterProxy.QueryParameterJson(ctx, clusterId)
}

func BuildClusterLogConfig(ctx ctx.Context, clusterId string) error {
	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_METADB_SERVER_CALL_ERROR, "load cluster %s failed", clusterId)
	}
	flow, err := CreateFlowWork(ctx, clusterAggregation.Cluster.Id, FlowBuildLogConfig, BuildSystemOperator())
	if err != nil {
		return err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()
	return nil
}

func collectorTiDBLogConfig(task *TaskEntity, ctx *FlowContext) bool {
	aggregation := ctx.GetData(contextClusterKey).(*ClusterAggregation)

	clusters, total, err := ClusterRepo.Query(ctx, "", "", "", "", "", 1, 10000)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("invoke cluster repo list cluster err： %v", err)
		task.Fail(err)
		return false
	}
	getLoggerWithContext(ctx).Infof("list cluster total count: %d", total)
	hosts := listClusterHosts(aggregation)
	getLoggerWithContext(ctx).Infof("cluster %s list host: %v", aggregation.Cluster.Id, hosts)
	go func() {
		for _, host := range hosts {
			collectorConfigs, err := buildCollectorTiDBLogConfig(ctx, host, clusters)
			if err != nil {
				getLoggerWithContext(ctx).Errorf("build collector tidb log config err： %v", err)
				break
			}
			bs, err := yaml.Marshal(collectorConfigs)
			if err != nil {
				getLoggerWithContext(ctx).Errorf("marshal yaml err： %v", err)
				break
			}
			collectorYaml := string(bs)
			// todo: When the tiem scale-out and scale-in is complete, change to take the filebeat deployDir from the tiem topology
			deployDir := "/tiem-test/filebeat"
			transferTaskId, err := secondparty.SecondParty.MicroSrvTiupTransfer(ctx, secondparty.ClusterComponentTypeStr,
				aggregation.Cluster.ClusterName, collectorYaml, deployDir+"/conf/input_tidb.yml",
				0, []string{"-N", host}, uint64(task.Id))
			getLoggerWithContext(ctx).Infof("got transferTaskId %d", transferTaskId)
			if err != nil {
				getLoggerWithContext(ctx).Errorf("collectorTiDBLogConfig invoke tiup transfer err： %v", err)
				break
			}
		}
	}()
	task.Success(nil)
	return true
}

var resourceManager = resource.NewResourceManager()

func prepareResource(task *TaskEntity, flowContext *FlowContext) bool {
	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	demands := clusterAggregation.AddedComponentDemand

	components, err := TopologyPlanner.BuildComponents(flowContext.Context, demands, clusterAggregation.Cluster)
	if err != nil {
		getLoggerWithContext(flowContext).Error(err)
		task.Fail(err)
		return false
	}
	clusterAggregation.AddedClusterComponents = components

	// build resource request
	req, err := TopologyPlanner.AnalysisResourceRequest(flowContext.Context, clusterAggregation.Cluster, clusterAggregation.AddedClusterComponents, false)
	if err != nil {
		getLoggerWithContext(flowContext).Error(err)
		task.Fail(err)
		return false
	}

	// alloc resource
	clusterAggregation.AddedAllocResources = &clusterpb.BatchAllocResponse{}
	err = resourceManager.AllocResourcesInBatch(flowContext.Context, req, clusterAggregation.AddedAllocResources)
	if err != nil {
		getLoggerWithContext(flowContext).Error(err)
		task.Fail(err)
		return false
	} else if clusterAggregation.AddedAllocResources.Rs.Code != 0 {
		err = framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, clusterAggregation.AddedAllocResources.Rs.Message)
		getLoggerWithContext(flowContext).Error(err)
		task.Fail(err)
		return false
	}

	flowContext.SetData(contextAllocRequestKey, req)

	prepareResourceSucceed(task, clusterAggregation.AddedAllocResources)
	return true
}

func prepareResourceSucceed(task *TaskEntity, resource *clusterpb.BatchAllocResponse) {
	allHost := make([]string, 0)
	if resource != nil && resource.Rs.Code == 0 {
		for _, r := range resource.BatchResults {
			if r != nil && r.Rs.Code != 0 {
				continue
			}
			for _, k := range r.Results {
				allHost = append(allHost, k.HostIp)
			}
		}
	}
	task.Success(fmt.Sprintf("alloc succeed with hosts: %s", allHost))
}

func buildConfig(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)

	// update cluster components
	err := TopologyPlanner.ApplyResourceToComponents(context.Context, clusterAggregation.Cluster, clusterAggregation.AddedAllocResources, clusterAggregation.AddedClusterComponents)
	if err != nil {
		getLoggerWithContext(context).Error(err)
		task.Fail(err)
		return false
	}

	// build TopologyConfigRecord
	configModel, err := TopologyPlanner.GenerateTopologyConfig(context.Context, clusterAggregation.AddedClusterComponents, clusterAggregation.Cluster)
	if err != nil {
		getLoggerWithContext(context).Error(err)
		task.Fail(err)
		return false
	}

	clusterAggregation.AlteredTopology = configModel
	bytes, err := yaml.Marshal(configModel)
	if err != nil {
		task.Fail(err)
	} else {
		task.Success(string(bytes))
	}
	return true
}

func deployCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	spec := clusterAggregation.AlteredTopology

	bs, err := yaml.Marshal(spec)
	if err != nil {
		task.Fail(err)
		return false
	}

	cfgYamlStr := string(bs)
	getLoggerWithContext(context).Infof("deploy cluster %s, version = %s, cfgYamlStr = %s", cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr)
	deployTaskId, err := secondparty.SecondParty.MicroSrvTiupDeploy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr, 0, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, uint64(task.Id),
	)

	if err != nil {
		getLoggerWithContext(context).Errorf("call tiup api start cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	getLoggerWithContext(context).Infof("got deployTaskId %s", strconv.Itoa(int(deployTaskId)))

	return true
}

func scaleOutCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	spec := clusterAggregation.AlteredTopology

	bs, err := yaml.Marshal(spec)
	if err != nil {
		task.Fail(err)
		return false
	}

	cfgYamlStr := string(bs)
	getLoggerWithContext(context).Infof("scale out cluster %s, version = %s, cfgYamlStr = %s", cluster.ClusterName, cluster.ClusterVersion.Code, cfgYamlStr)
	scaleOutTaskId, err := secondparty.SecondParty.MicroSrvTiupScaleOut(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, cfgYamlStr, 0, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, uint64(task.Id))

	if err != nil {
		getLoggerWithContext(context).Errorf("call tiup api scale out cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	getLoggerWithContext(context).Infof("get scaleOutTaskId %s", strconv.Itoa(int(scaleOutTaskId)))
	return true
}

func getInstance(instances []*ComponentInstance, nodeId string) *ComponentInstance {
	results := strings.Split(nodeId, ":")
	if len(results) != 2 {
		return nil
	}
	for _, instance := range instances {
		if instance.Host == results[0] {
			for _, port := range instance.PortList {
				if strconv.Itoa(port) == results[1] {
					return instance
				}
			}
		}
	}
	return nil
}

func scaleInCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	nodeId := context.GetData(contextDeleteNodeKey).(string)

	componentInstance := getInstance(clusterAggregation.CurrentComponentInstances, nodeId)
	if componentInstance == nil {
		getLoggerWithContext(context).Errorf("node: %s is not exist in %s", nodeId, cluster.ClusterName)
		task.Fail(fmt.Errorf("node: %s is not exist", nodeId))
		return false
	}

	if knowledge.GetComponentSpec(cluster.ClusterType.Code,
		cluster.ClusterVersion.Code, componentInstance.ComponentType.ComponentType).ComponentConstraint.ComponentRequired {
		nodeCount := 0
		for _, instance := range clusterAggregation.CurrentComponentInstances {
			if instance.ComponentType.ComponentType == componentInstance.ComponentType.ComponentType {
				nodeCount += 1
			}
		}
		if nodeCount <= 1 {
			getLoggerWithContext(context).Errorf("node: %s is unique in %s, can not delete it", nodeId, cluster.ClusterName)
			task.Fail(fmt.Errorf("node: %s can not be deleted", nodeId))
			return false
		}
	}

	getLoggerWithContext(context).Infof("scale in cluster %s, delete node: %s", cluster.ClusterName, nodeId)
	scaleInTaskId, err := secondparty.SecondParty.MicroSrvTiupScaleIn(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, nodeId, 0, []string{"--yes"}, uint64(task.Id))

	if err != nil {
		getLoggerWithContext(context).Errorf("call tiup api scale in cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	getLoggerWithContext(context).Infof("get scaleInTaskId %s", strconv.Itoa(int(scaleInTaskId)))
	return true
}

func freeNodeResource(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	nodeId := context.GetData(contextDeleteNodeKey).(string)

	componentInstance := getInstance(clusterAggregation.CurrentComponentInstances, nodeId)
	if componentInstance == nil {
		getLoggerWithContext(context).Errorf("node: %s is not exist in %s", nodeId, cluster.ClusterName)
		task.Fail(fmt.Errorf("node: %s is not exist", nodeId))
		return false
	}

	componentInstance.Status = ClusterStatusDeleted

	ports := make([]int32, 0)
	for _, port := range componentInstance.PortList {
		ports = append(ports, int32(port))
	}
	request := &clusterpb.RecycleRequest{
		RecycleReqs: []*clusterpb.RecycleRequire{
			{
				RecycleType: int32(resourceType.RecycleHost),
				HolderId:    componentInstance.ClusterId,
				RequestId:   componentInstance.AllocRequestId,
				HostId:      componentInstance.HostId,
				ComputeReq: &clusterpb.ComputeRequirement{
					CpuCores: componentInstance.Compute.CpuCores,
					Memory:   componentInstance.Compute.Memory,
				},
				DiskReq: []*clusterpb.DiskResource{
					{DiskId: componentInstance.DiskId},
				},
				PortReq: []*clusterpb.PortResource{
					{Ports: ports},
				},
			},
		},
	}
	response := &clusterpb.RecycleResponse{}
	err := resource.NewResourceManager().RecycleResources(context, request, response)

	if err != nil {
		task.Fail(err)
		return false
	}

	task.Success(nil)
	return true
}

func startupCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	startTaskId, err := secondparty.SecondParty.MicroSrvTiupStart(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, 0, []string{}, uint64(task.Id),
	)

	if err != nil {
		getLoggerWithContext(context).Errorf("call tiup api start cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	getLoggerWithContext(context).Infof("got startTaskId %s", strconv.Itoa(int(startTaskId)))

	return true
}

func setClusterOffline(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.StatusModified = true
	clusterAggregation.Cluster.Offline()

	task.Success(nil)
	return true
}

func setClusterOnline(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.StatusModified = true
	clusterAggregation.Cluster.Online()
	// set instance status into online
	for _, component := range clusterAggregation.AddedClusterComponents {
		for _, instance := range component.Nodes {
			if instance.Status == ClusterStatusUnlined {
				instance.Status = ClusterStatusOnline
			}
		}
	}

	task.Success(nil)
	return true
}

func syncTopology(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)

	metadata, err := MetadataMgr.FetchFromLocal(ctx.TODO(), "/root/.tiup/", clusterAggregation.Cluster.ClusterName)

	if err != nil {
		task.Fail(err)
		return false
	}

	clusterAggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: metadata.GetTopology().(*spec.Specification),
	}
	clusterAggregation.ConfigModified = true
	task.Success(nil)

	return true
}

func modifyParameters(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}

func fetchTopologyFile(task *TaskEntity, context *FlowContext) bool {
	req := context.GetData(contextTakeoverReqKey).(*clusterpb.ClusterTakeoverReqDTO)

	metadata, err := MetadataMgr.FetchFromRemoteCluster(ctx.TODO(), req)
	if err != nil {
		task.Fail(err)
		return false
	}

	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.ClusterMetadata = metadata

	clusterAggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    clusterAggregation.Cluster.TenantId,
		ClusterId:   clusterAggregation.Cluster.Id,
		ConfigModel: metadata.GetTopology().(*spec.Specification),
	}

	clusterType, _, _, version := MetadataMgr.ParseClusterInfoFromMetaData(context, *metadata.GetBaseMeta())
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
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)

	components, err := MetadataMgr.ParseComponentsFromMetaData(context, clusterAggregation.ClusterMetadata)
	if err != nil {
		task.Fail(err)
		return false
	}

	clusterAggregation.AddedClusterComponents = components
	task.Success(nil)
	return true
}

func takeoverResource(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)

	allocReq, err := TopologyPlanner.AnalysisResourceRequest(context.Context, clusterAggregation.Cluster, clusterAggregation.AddedClusterComponents, true)
	if err != nil {
		task.Fail(err)
		return false
	}

	allocResponse := &clusterpb.BatchAllocResponse{}
	err = resourceManager.AllocResourcesInBatch(ctx.TODO(), allocReq, allocResponse)
	if err != nil {
		task.Fail(err)
		return false
	}

	task.Success(allocReq)
	return true
}

func deleteCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.Delete()
	clusterAggregation.StatusModified = true

	task.Success(nil)
	return true
}

func destroyCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	destroyTaskId, err := secondparty.SecondParty.MicroSrvTiupDestroy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, 0, []string{}, uint64(task.Id),
	)

	if err != nil {
		getLoggerWithContext(context).Errorf("call tiup api destroy cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	getLoggerWithContext(context).Infof("got destroyTaskId %s", strconv.Itoa(int(destroyTaskId)))

	return true
}

func freedResource(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)

	request := &clusterpb.RecycleRequest{
		RecycleReqs: []*clusterpb.RecycleRequire{
			{
				RecycleType: int32(resourceType.RecycleHolder),
				HolderId:    clusterAggregation.Cluster.Id,
			},
		},
	}
	response := &clusterpb.RecycleResponse{}
	err := resourceManager.RecycleResources(context, request, response)

	if err != nil {
		task.Fail(err)
		return false
	}

	task.Success(nil)
	return true
}

func freedResourceAfterFailure(task *TaskEntity, context *FlowContext) bool {
	allocRequest := context.GetData(contextAllocRequestKey)
	if allocRequest != nil {
		request := &clusterpb.RecycleRequest{
			RecycleReqs: []*clusterpb.RecycleRequire{},
		}

		for _, req := range allocRequest.(*clusterpb.BatchAllocRequest).BatchRequests {
			framework.LogWithContext(context).Infof("freed resource after failure, request = %s", req.Applicant.RequestId)
			require := &clusterpb.RecycleRequire{
				RecycleType: int32(resourceType.RecycleOperate),
				RequestId:   req.Applicant.RequestId,
			}
			request.RecycleReqs = append(request.RecycleReqs, require)
		}

		response := &clusterpb.RecycleResponse{}
		err := resourceManager.RecycleResources(context, request, response)
		if err != nil {
			framework.LogWithContext(context).Errorf("RecycleResources error, %s", err.Error())
		}
		return true
	}

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
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	getLoggerWithContext(context).Infof("restart cluster %s", cluster.ClusterName)
	restartTaskId, err := secondparty.SecondParty.MicroSrvTiupRestart(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, 0, []string{}, uint64(task.Id),
	)
	if err != nil {
		framework.LogWithContext(context).Errorf("call tiup api restart cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}
	getLoggerWithContext(context).Infof("got restartTaskId %s", strconv.Itoa(int(restartTaskId)))

	task.Success(nil)
	return true
}

// clusterStop
// @Description: stop cluster
// @Parameter task
// @Parameter context
// @return bool
func clusterStop(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster

	framework.LogWithContext(context).Infof("stop cluster %s", cluster.ClusterName)
	stopTaskId, err := secondparty.SecondParty.MicroSrvTiupStop(context.Context, secondparty.ClusterComponentTypeStr, cluster.ClusterName, 0, []string{}, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call tiup api stop cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}

	if err != nil {
		framework.LogWithContext(context).Errorf("call tiup api stop cluster err = %s", err.Error())
		task.Fail(err)
		return false
	}
	framework.LogWithContext(context).Infof("got stopTaskId %s", strconv.Itoa(int(stopTaskId)))

	task.Success(nil)
	return true
}

func (aggregation *ClusterAggregation) ExtractStatusDTO() *clusterpb.DisplayStatusDTO {
	cluster := aggregation.Cluster

	dto := &clusterpb.DisplayStatusDTO{
		StatusCode:      strconv.Itoa(int(aggregation.Cluster.Status)),
		StatusName:      aggregation.Cluster.Status.Display(),
		CreateTime:      cluster.CreateTime.Unix(),
		UpdateTime:      cluster.UpdateTime.Unix(),
		DeleteTime:      cluster.DeleteTime.Unix(),
		InProcessFlowId: int32(cluster.WorkFlowId),
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
		BackupTso:    record.BackupTso,
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
		Id:             dto.Id,
		Name:           dto.Name,
		TenantId:       dto.TenantId,
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
		Host:            pdHosts[0].Ip,
		DeployDir:       filepath.Join(pdHosts[0].Disk.Path, cluster.Id, "grafanas-deploy"),
		AnonymousEnable: true,
		DefaultTheme:    "light",
		OrgName:         "Main Org.",
		OrgRole:         "Viewer",
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
