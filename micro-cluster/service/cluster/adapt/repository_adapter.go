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

package adapt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
	"strconv"
	"time"
)

func InjectionMetaDbRepo() {
	domain.TaskRepo = TaskRepoAdapter{}
	domain.ClusterRepo = ClusterRepoAdapter{}
	domain.RemoteClusterProxy = RemoteClusterProxy{}
	domain.MetadataMgr = NewTiUPTiDBMetadataManager()
	domain.TopologyPlanner = DefaultTopologyPlanner{}
}

type ClusterRepoAdapter struct{}

func (c ClusterRepoAdapter) Query(ctx context.Context, clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*domain.ClusterAggregation, int, error) {
	req := &dbpb.DBListClusterRequest{
		ClusterName:   clusterName,
		ClusterId:     clusterId,
		ClusterTag:    clusterTag,
		ClusterStatus: clusterStatus,
		ClusterType:   clusterType,
		PageReq: &dbpb.DBPageDTO{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}

	resp, err := client.DBClient.ListCluster(ctx, req)

	if err != nil {
		return nil, 0, err
	}
	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
		return nil, 0, err
	}

	clusters := make([]*domain.ClusterAggregation, len(resp.Clusters))
	for i, v := range resp.Clusters {
		cluster := &domain.ClusterAggregation{}
		cluster.Cluster = ParseFromClusterDTO(v.Cluster)
		cluster.CurrentTopologyConfigRecord = parseConfigRecordDTO(v.TopologyConfigRecord)
		cluster.CurrentWorkFlow = parseFlowFromDTO(v.Flow)
		cluster.MaintainCronTask = domain.GetDefaultMaintainTask() // next_version get from db

		clusters[i] = cluster
	}
	return clusters, int(resp.Page.Total), err

}

func (c ClusterRepoAdapter) AddCluster(ctx context.Context, cluster *domain.Cluster) error {
	req := &dbpb.DBCreateClusterRequest{
		Cluster: convertClusterToDTO(cluster),
	}

	resp, err := client.DBClient.CreateCluster(ctx, req)
	if err != nil {
		return err
	}

	if resp.Status.Code != 0 {
		return errors.New(resp.Status.Message)
	} else {
		dto := resp.Cluster
		cluster.Id = dto.Id
		cluster.Code = dto.Code
		cluster.Status = domain.ClusterStatusFromValue(int(dto.Status))
		cluster.CreateTime = time.Unix(dto.CreateTime, 0)
		cluster.UpdateTime = time.Unix(dto.UpdateTime, 0)
	}
	return nil
}

func (c ClusterRepoAdapter) Persist(ctx context.Context, aggregation *domain.ClusterAggregation) error {

	if err := c.PersistStatus(ctx, aggregation); err != nil {
		framework.LogWithContext(ctx).Errorf("persist cluster error, module %s, error %v", "status", err)
		return err
	}

	if err := persisClusterParameter(ctx, aggregation); err != nil {
		framework.LogWithContext(ctx).Errorf("persist cluster error, module %s, error %v", "parameter", err)
		return err
	}

	if err := persisClusterTopologyConfig(ctx, aggregation); err != nil {
		framework.LogWithContext(ctx).Errorf("persist cluster error, module %s, error %v", "topology", err)
		return err
	}

	if err := persistClusterComponents(ctx, aggregation); err != nil {
		framework.LogWithContext(ctx).Errorf("persist cluster error, module %s, error %v", "components", err)
		return err
	}

	if err := persistClusterBaseInfo(ctx, aggregation); err != nil {
		framework.LogWithContext(ctx).Errorf("persist cluster error, module %s, error %v", "baseinfo", err)
		return err
	}

	return nil
}

func (c ClusterRepoAdapter) PersistStatus(ctx context.Context, aggregation *domain.ClusterAggregation) error {
	if aggregation.StatusModified || aggregation.FlowModified {
		if aggregation.Cluster.Status == domain.ClusterStatusDeleted {
			_, err := client.DBClient.DeleteCluster(ctx,  &dbpb.DBDeleteClusterRequest{
				ClusterId: aggregation.Cluster.Id,
			})
			return err
		} else {
			resp, err := client.DBClient.UpdateClusterStatus(ctx, &dbpb.DBUpdateClusterStatusRequest{
				ClusterId:    aggregation.Cluster.Id,
				Status:       int32(aggregation.Cluster.Status),
				UpdateStatus: aggregation.StatusModified,
				FlowId:       int64(aggregation.Cluster.WorkFlowId),
				UpdateFlow:   aggregation.FlowModified,
			})

			if err != nil {
				return err
			}

			aggregation.Cluster.Status = domain.ClusterStatusFromValue(int(resp.Cluster.Status))
			aggregation.Cluster.WorkFlowId = uint(resp.Cluster.WorkFlowId)
		}
	}
	return nil
}

func persisClusterParameter(ctx context.Context, aggregation *domain.ClusterAggregation) error {
	if aggregation.LastParameterRecord != nil && aggregation.LastParameterRecord.Id == 0 {
		record := aggregation.LastParameterRecord
		resp, err := client.DBClient.SaveParametersRecord(ctx, &dbpb.DBSaveParametersRequest{
			Parameters: &dbpb.DBParameterRecordDTO{
				TenantId:   aggregation.Cluster.TenantId,
				ClusterId:  record.ClusterId,
				OperatorId: record.OperatorId,
				Content:    record.Content,
				FlowId:     int64(aggregation.CurrentWorkFlow.Id),
			},
		})
		if err != nil {
			return err
		}
		aggregation.LastParameterRecord.Id = uint(resp.Parameters.Id)
	}
	return nil
}

func persisClusterTopologyConfig(ctx context.Context, aggregation *domain.ClusterAggregation) error {
	if aggregation.ConfigModified {
		resp, err := client.DBClient.UpdateClusterTopologyConfig(ctx, &dbpb.DBUpdateTopologyConfigRequest{
			ClusterId: aggregation.Cluster.Id,
			Content:   aggregation.CurrentTopologyConfigRecord.Content(),
			TenantId:  aggregation.Cluster.TenantId,
		})

		if err != nil {
			return err
		}
		aggregation.CurrentTopologyConfigRecord = parseConfigRecordDTO(resp.TopologyConfigRecord)
	}
	return nil
}

// persistClusterComponents
// @Description: create
// @Parameter ctx
// @Parameter aggregation
// @return error
func persistClusterComponents(ctx context.Context, aggregation *domain.ClusterAggregation) error {
	if len(aggregation.AddedClusterComponents) > 0  {
		toCreate := make([]*domain.ComponentInstance, 0 )

		for _, g := range aggregation.AddedClusterComponents {
			for _, c := range g.Nodes {
				if c.ID == "" {
					toCreate = append(toCreate, c)
				}
			}
		}
		if len(toCreate) >= 0 {
			request := &dbpb.DBCreateInstanceRequest{
				ClusterId: aggregation.Cluster.Id,
				TenantId: aggregation.Cluster.TenantId,
				ComponentInstances: batchConvertComponentToDTOs(toCreate),
			}

			_, err := client.DBClient.CreateInstance(ctx, request)

			if err != nil {
				return err
			}
		}
	}

	// Delete instance which status is ClusterStatusDeleted
	for _, instance := range aggregation.CurrentComponentInstances {
		if instance.Status == domain.ClusterStatusDeleted {
			request := &dbpb.DBDeleteInstanceRequest {
				Id: instance.ID,
			}
			_, err := client.DBClient.DeleteInstance(ctx, request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func persistClusterBaseInfo(ctx context.Context, aggregation *domain.ClusterAggregation) error {
	cluster := aggregation.Cluster
	if aggregation.BaseInfoModified {
		tagBytes, err := json.Marshal(cluster.Tags)
		if err != nil {
			return err
		}
		client.DBClient.UpdateClusterInfo(ctx, &dbpb.DBUpdateClusterInfoRequest{
			ClusterId:   aggregation.Cluster.Id,
			Name:        cluster.ClusterName,
			ClusterType: cluster.ClusterType.Code,
			VersionCode: cluster.ClusterVersion.Code,
			Tags:        string(tagBytes),
			Tls:         cluster.Tls,
		})

		return err
	}

	if aggregation.DemandsModified {
		var demands string
		if len(aggregation.CurrentComponentDemand) > 0 {
			bytes, err := json.Marshal(aggregation.CurrentComponentDemand)
			if err != nil {
				return err
			}
			demands = string(bytes)
		}
		resp, err := client.DBClient.UpdateClusterDemand(ctx, &dbpb.DBUpdateDemandRequest{
			ClusterId: cluster.Id,
			Demands: demands,
			TenantId: cluster.TenantId,
		})
		if err != nil {
			return err
		}
		if resp.Status.Code != 0 {
			return fmt.Errorf(resp.Status.Message)
		}
	}
	return nil
}

func (c ClusterRepoAdapter) Load(ctx context.Context, id string) (cluster *domain.ClusterAggregation, err error) {
	req := &dbpb.DBLoadClusterRequest{
		ClusterId: id,
	}

	resp, err := client.DBClient.LoadCluster(ctx, req)

	if err != nil {
		return
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
		return
	} else {
		cluster = &domain.ClusterAggregation{}
		cluster.Cluster = ParseFromClusterDTO(resp.ClusterDetail.Cluster)
		cluster.CurrentTopologyConfigRecord = parseConfigRecordDTO(resp.ClusterDetail.TopologyConfigRecord)
		cluster.CurrentWorkFlow = parseFlowFromDTO(resp.ClusterDetail.Flow)
		cluster.CurrentComponentInstances = parseInstancesFromDTO(resp.ClusterDetail.ComponentInstances)
		cluster.MaintainCronTask = domain.GetDefaultMaintainTask() // next_version get from db
		return
	}
}

type RemoteClusterProxy struct{}

func (c RemoteClusterProxy) QueryParameterJson(ctx context.Context, clusterId string) (content string, err error) {
	resp, err := client.DBClient.GetCurrentParametersRecord(ctx, &dbpb.DBGetCurrentParametersRequest{
		ClusterId: clusterId,
	})

	if err != nil {
		return
	}

	if resp.Status.Code != 0 {
		if resp.Status.Code == 1 {
			content = "[]"
			return
		}
		err = errors.New(resp.Status.Message)
		return
	} else {
		content = resp.GetParameters().GetContent()
	}
	return
}

type TaskRepoAdapter struct{}

func (t TaskRepoAdapter) ListFlows(ctx context.Context, bizId, keyword string, status int, page int, pageSize int) ([]*domain.FlowWorkEntity, int, error) {
	resp, err := client.DBClient.ListFlows(ctx, &dbpb.DBListFlowsRequest{
		BizId:   bizId,
		Keyword: keyword,
		Status:  int64(status),
		Page: &dbpb.DBTaskPageDTO{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	})

	if err != nil {
		framework.LogWithContext(ctx).Errorf("AddFlowWork error = %s", err.Error())
		return nil, 0, err
	}
	flows := make([]*domain.FlowWorkEntity, len(resp.Flows))
	for i, v := range resp.Flows {
		flows[i] = &domain.FlowWorkEntity{
			Id:          uint(v.Id),
			FlowName:    v.FlowName,
			StatusAlias: v.StatusAlias,
			BizId:       v.BizId,
			Status:      domain.TaskStatus(v.Status),
			Operator:    domain.GetOperatorFromName(v.Operator),
			CreateTime:  time.Unix(v.CreateTime, 0),
			UpdateTime:  time.Unix(v.UpdateTime, 0),
		}
	}

	return flows, int(resp.Page.Total), err
}

func (t TaskRepoAdapter) QueryCronTask(ctx context.Context, bizId string, cronTaskType int) (cronTask *domain.CronTaskEntity, err error) {
	cronTask = domain.GetDefaultMaintainTask()
	return
}

func (t TaskRepoAdapter) PersistCronTask(ctx context.Context, cronTask *domain.CronTaskEntity) (err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) AddFlowWork(ctx context.Context, flowWork *domain.FlowWorkEntity) error {
	resp, err := client.DBClient.CreateFlow(ctx, &dbpb.DBCreateFlowRequest{
		Flow: &dbpb.DBFlowDTO{
			FlowName:    flowWork.FlowName,
			StatusAlias: flowWork.StatusAlias,
			BizId:       flowWork.BizId,
			Operator:    flowWork.Operator.Name,
		},
	})

	if err != nil {
		framework.LogWithContext(ctx).Errorf("AddFlowWork error = %s", err.Error())
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
	} else {
		flowWork.Id = uint(resp.Flow.Id)
	}
	return err
}

func (t TaskRepoAdapter) AddFlowTask(ctx context.Context, task *domain.TaskEntity, flowId uint) error {
	resp, err := client.DBClient.CreateTask(ctx, &dbpb.DBCreateTaskRequest{
		Task: &dbpb.DBTaskDTO{
			TaskName:       task.TaskName,
			TaskReturnType: strconv.Itoa(int(task.TaskReturnType)),
			BizId:          task.BizId,
			Parameters:     task.Parameters,
			ParentId:       strconv.Itoa(int(flowId)),
			ParentType:     0,
			StartTime:      task.StartTime,
		},
	})

	if err != nil {
		// todo
		framework.LogWithContext(ctx).Errorf("addflowtask flowid = %d, errStr: %s", flowId, err.Error())
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
	} else {
		task.Id = uint(resp.Task.Id)
	}
	return err
}

func (t TaskRepoAdapter) AddCronTask(ctx context.Context, cronTask *domain.CronTaskEntity) error {
	return nil
}

func (t TaskRepoAdapter) Persist(ctx context.Context, flowWork *domain.FlowWorkAggregation) error {
	req := &dbpb.DBUpdateFlowRequest{
		FlowWithTasks: &dbpb.DBFlowWithTaskDTO{
			Flow: &dbpb.DBFlowDTO{
				Id:          int64(flowWork.FlowWork.Id),
				BizId:       flowWork.FlowWork.BizId,
				Status:      int32(flowWork.FlowWork.Status),
				FlowName:    flowWork.FlowWork.FlowName,
				StatusAlias: flowWork.FlowWork.StatusAlias,
			},
		},
	}

	tasks := make([]*dbpb.DBTaskDTO, len(flowWork.Tasks))
	req.FlowWithTasks.Tasks = tasks

	for i, v := range flowWork.Tasks {
		tasks[i] = &dbpb.DBTaskDTO{
			Id:             int64(v.Id),
			Status:         int32(v.Status),
			Result:         v.Result,
			TaskReturnType: strconv.Itoa(int(v.TaskReturnType)),
			TaskName:       v.TaskName,
			BizId:          v.BizId,
			Parameters:     v.Parameters,
			ParentId:       strconv.Itoa(int(flowWork.FlowWork.Id)),
			ParentType:     0,
			StartTime:      v.StartTime,
			EndTime:        v.EndTime,
		}
	}

	_, err := client.DBClient.UpdateFlow(ctx, req)
	return err
}

func (t TaskRepoAdapter) LoadFlowWork(ctx context.Context, id uint) (flow *domain.FlowWorkEntity, err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) Load(ctx context.Context, id uint) (flowWork *domain.FlowWorkAggregation, err error) {
	resp, err := client.DBClient.LoadFlow(ctx, &dbpb.DBLoadFlowRequest{
		Id: int64(id),
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("Call metadb rpc method [%s] failed, error: %s", "LoadFlow", err.Error())
		return nil, framework.WrapError(common.TIEM_METADB_SERVER_CALL_ERROR, common.TIEM_METADB_SERVER_CALL_ERROR.Explain(), err)
	}
	if resp.Status.Code != 0 {
		framework.LogWithContext(ctx).Errorf("LoadFlowWork failed, error: %s", resp.Status.Message)
		return nil, framework.NewTiEMError(common.TIEM_ERROR_CODE(resp.Status.Code), resp.Status.Message)
	} else {
		flowDTO := resp.FlowWithTasks
		flowEntity := &domain.FlowWorkEntity{
			Id:          uint(flowDTO.Flow.Id),
			FlowName:    flowDTO.Flow.FlowName,
			StatusAlias: flowDTO.Flow.StatusAlias,
			BizId:       flowDTO.Flow.BizId,
			Status:      domain.TaskStatus(flowDTO.Flow.Status),
			Operator:    domain.GetOperatorFromName(flowDTO.Flow.Operator),
			CreateTime:  time.Unix(flowDTO.Flow.CreateTime, 0),
			UpdateTime:  time.Unix(flowDTO.Flow.UpdateTime, 0),
		}

		return &domain.FlowWorkAggregation{
			FlowWork: flowEntity,
			Tasks: ParseTaskDTOInBatch(resp.FlowWithTasks.Tasks),
			Define: domain.FlowWorkDefineMap[flowEntity.FlowName],
		}, nil
	}
}

func ParseTaskDTOInBatch(dtoList []*dbpb.DBTaskDTO) []*domain.TaskEntity {
	if dtoList == nil {
		return nil
	}

	entities := make([]*domain.TaskEntity, 0)
	for _, dto := range dtoList {
		entities = append(entities, ParseTaskDTO(dto))
	}

	return entities
}

func ParseTaskDTO(dto *dbpb.DBTaskDTO) *domain.TaskEntity {
	return &domain.TaskEntity{
		Id:         uint(dto.Id),
		Status:     domain.TaskStatusFromValue(int(dto.Status)),
		TaskName:   dto.TaskName,
		BizId:      dto.BizId,
		Parameters: dto.Parameters,
		Result:     dto.Result,
		StartTime:  dto.StartTime,
		EndTime:    dto.EndTime,
	}
}

func convertClusterToDTO(cluster *domain.Cluster) (dto *dbpb.DBClusterDTO) {
	if cluster == nil {
		return
	}
	dto = &dbpb.DBClusterDTO{
		Id:              cluster.Id,
		Code:            cluster.Code,
		Name:            cluster.ClusterName,
		TenantId:        cluster.TenantId,
		DbPassword:      cluster.DbPassword,
		ClusterType:     cluster.ClusterType.Code,
		VersionCode:     cluster.ClusterVersion.Code,
		Status:          int32(cluster.Status),
		Exclusive:       cluster.Exclusive,
		Region:          cluster.Region,
		CpuArchitecture: cluster.CpuArchitecture,
		Tls:             cluster.Tls,
		WorkFlowId:      int32(cluster.WorkFlowId),
		OwnerId:         cluster.OwnerId,
	}

	if len(cluster.Tags) > 0 {
		bytes, _ := json.Marshal(cluster.Tags)
		dto.Tags = string(bytes)
	}

	return
}

func ParseFromClusterDTO(dto *dbpb.DBClusterDTO) (cluster *domain.Cluster) {
	if dto == nil {
		return nil
	}
	cluster = &domain.Cluster{
		Id:              dto.Id,
		Code:            dto.Code,
		TenantId:        dto.TenantId,
		ClusterName:     dto.Name,
		DbPassword:      dto.DbPassword,
		ClusterType:     *knowledge.ClusterTypeFromCode(dto.ClusterType),
		ClusterVersion:  *knowledge.ClusterVersionFromCode(dto.VersionCode),
		Tls:             dto.Tls,
		Status:          domain.ClusterStatusFromValue(int(dto.Status)),
		WorkFlowId:      uint(dto.WorkFlowId),
		OwnerId:         dto.OwnerId,
		Exclusive:       dto.Exclusive,
		CpuArchitecture: dto.CpuArchitecture,
		Region:          dto.Region,
		CreateTime:      time.Unix(dto.CreateTime, 0),
		UpdateTime:      time.Unix(dto.UpdateTime, 0),
		DeleteTime:      time.Unix(dto.DeleteTime, 0),
	}

	json.Unmarshal([]byte(dto.Tags), &cluster.Tags)

	return
}

func parseInstancesFromDTO(dto []*dbpb.DBComponentInstanceDTO) []*domain.ComponentInstance {
	if len(dto) == 0 {
		return nil
	}

	componentInstances := make([]*domain.ComponentInstance, 0, len(dto))
	for _, item := range dto {
		instance := &domain.ComponentInstance{
			ID: item.Id,
			Code: item.Code,
			TenantId: item.TenantId,
			Status: domain.ClusterStatus(item.Status),
			ClusterId: item.ClusterId,
			ComponentType: &knowledge.ClusterComponent{ComponentType: item.ComponentType},
			Role: item.Role,
			Version: &knowledge.ClusterVersion{Code: item.Version},
			HostId: item.HostId,
			Host: item.Host,
			DiskId: item.DiskId,
			AllocRequestId: item.AllocRequestId,
			DiskPath: item.DiskPath,
			Compute: &resource.ComputeRequirement{
				CpuCores: item.CpuCores,
				Memory: item.Memory,
			},
		}
		instance.DeserializePortInfo(item.PortInfo)
		componentInstances = append(componentInstances, instance)
	}

	return componentInstances
}

func parseConfigRecordDTO(dto *dbpb.DBTopologyConfigDTO) (record *domain.TopologyConfigRecord) {
	if dto == nil {
		return nil
	}
	record = &domain.TopologyConfigRecord{
		Id:         uint(dto.Id),
		TenantId:   dto.TenantId,
		ClusterId:  dto.ClusterId,
		CreateTime: time.Unix(dto.CreateTime, 0),
	}

	spec := &spec.Specification{}
	yaml.Unmarshal([]byte(dto.Content), spec)

	record.ConfigModel = spec
	return
}

func parseFlowFromDTO(dto *dbpb.DBFlowDTO) (flow *domain.FlowWorkEntity) {
	if dto == nil {
		return nil
	}
	flow = &domain.FlowWorkEntity{
		Id:          uint(dto.Id),
		FlowName:    dto.FlowName,
		StatusAlias: dto.StatusAlias,
		BizId:       dto.BizId,
		Status:      domain.TaskStatusFromValue(int(dto.Status)),
	}
	// todo
	return
}

func batchConvertComponentToDTOs(components []*domain.ComponentInstance) (dtoList []*dbpb.DBComponentInstanceDTO) {
	dtoList = make([]*dbpb.DBComponentInstanceDTO, 0)
	for _, c := range components {
		dtoList = append(dtoList, convertComponentToDTO(c))
	}
	return
}

func convertComponentToDTO(component *domain.ComponentInstance) (dto *dbpb.DBComponentInstanceDTO) {
	dto = &dbpb.DBComponentInstanceDTO {
		Id:       component.ID,
		Code:     component.Code,
		TenantId: component.TenantId,
		Status: int32(component.Status),
		ClusterId: component.ClusterId,
		ComponentType: component.ComponentType.ComponentType,
		Role: component.Role,
		Version: component.Version.Code,

		HostId: component.HostId,
		Host: component.Host,
		CpuCores: component.Compute.CpuCores,
		Memory: component.Compute.Memory,
		DiskId: component.DiskId,
		DiskPath: component.DiskPath,
		PortInfo: component.SerializePortInfo(),
		AllocRequestId: component.AllocRequestId,
	}

	return
}
