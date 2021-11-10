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
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
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
		Cluster: ConvertClusterToDTO(cluster),
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
	cluster := aggregation.Cluster

	if aggregation.StatusModified || aggregation.FlowModified {
		resp, err := client.DBClient.UpdateClusterStatus(ctx, &dbpb.DBUpdateClusterStatusRequest{
			ClusterId:    cluster.Id,
			Status:       int32(cluster.Status),
			UpdateStatus: aggregation.StatusModified,
			FlowId:       int64(aggregation.Cluster.WorkFlowId),
			UpdateFlow:   aggregation.FlowModified,
		})

		if err != nil {
			// todo
			return err
		}

		cluster.Status = domain.ClusterStatusFromValue(int(resp.Cluster.Status))
		cluster.WorkFlowId = uint(resp.Cluster.WorkFlowId)
	}

	if aggregation.ConfigModified {
		resp, err := client.DBClient.UpdateClusterTopologyConfig(ctx, &dbpb.DBUpdateTopologyConfigRequest{
			ClusterId: aggregation.Cluster.Id,
			Content:   aggregation.CurrentTopologyConfigRecord.Content(),
			TenantId:  aggregation.Cluster.TenantId,
		})

		if err != nil {
			// todo
			return err
		}
		aggregation.CurrentTopologyConfigRecord = parseConfigRecordDTO(resp.TopologyConfigRecord)
	}
	/*
		if aggregation.LastBackupRecord != nil && aggregation.LastBackupRecord.Id == 0 {
			record := aggregation.LastBackupRecord
			resp, err :=  client.DBClient.SaveBackupRecord(context.TODO(), &db.DBSaveBackupRecordRequest{
				BackupRecord: &db.DBBackupRecordDTO{
					TenantId:    cluster.TenantId,
					ClusterId:   record.ClusterId,
					BackupType: string(record.BackupType),
					BackupRange: string(record.Range),
					OperatorId:  record.OperatorId,
					FilePath:    record.FilePath,
					FlowId:      int64(aggregation.CurrentWorkFlow.Id),
				},
			})
			if err != nil {
				// todo
				return err
			}
			record.Id = resp.BackupRecord.Id
		}

		if aggregation.LastRecoverRecord != nil && aggregation.LastRecoverRecord.Id == 0 {
			record := aggregation.LastRecoverRecord
			resp, err :=  client.DBClient.SaveRecoverRecord(context.TODO(), &db.DBSaveRecoverRecordRequest{
				RecoverRecord: &db.DBRecoverRecordDTO{
					TenantId:       cluster.TenantId,
					ClusterId:      record.ClusterId,
					OperatorId:     record.OperatorId,
					BackupRecordId: record.BackupRecord.Id,
					FlowId:         int64(aggregation.CurrentWorkFlow.Id),
				},
			})
			if err != nil {
				// todo
				return err
			}
			aggregation.LastRecoverRecord.Id = uint(resp.RecoverRecord.Id)
		}*/

	if aggregation.LastParameterRecord != nil && aggregation.LastParameterRecord.Id == 0 {
		record := aggregation.LastParameterRecord
		resp, err := client.DBClient.SaveParametersRecord(ctx, &dbpb.DBSaveParametersRequest{
			Parameters: &dbpb.DBParameterRecordDTO{
				TenantId:   cluster.TenantId,
				ClusterId:  record.ClusterId,
				OperatorId: record.OperatorId,
				Content:    record.Content,
				FlowId:     int64(aggregation.CurrentWorkFlow.Id),
			},
		})
		if err != nil {
			// todo
			return err
		}
		aggregation.LastParameterRecord.Id = uint(resp.Parameters.Id)
	}

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

		if err != nil {
			// todo
			return err
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
		return nil, framework.WrapError(common.TIEM_METADB_SERVER_CALL_ERROR, common.TiEMErrMsg[common.TIEM_METADB_SERVER_CALL_ERROR], err)
	}
	if resp.Status.Code != 0 {
		framework.LogWithContext(ctx).Errorf("LoadFlowWork failed, error: %s", resp.Status.Message)
		return nil, framework.CustomizeMessageError(common.TIEM_ERROR_CODE(resp.Status.Code), resp.Status.Message)
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

func ConvertClusterToDTO(cluster *domain.Cluster) (dto *dbpb.DBClusterDTO) {
	if cluster == nil {
		return
	}
	dto = &dbpb.DBClusterDTO{
		Id:          cluster.Id,
		Code:        cluster.Code,
		Name:        cluster.ClusterName,
		TenantId:    cluster.TenantId,
		DbPassword:  cluster.DbPassword,
		ClusterType: cluster.ClusterType.Code,
		VersionCode: cluster.ClusterVersion.Code,
		Status:      int32(cluster.Status),
		Tls:         cluster.Tls,
		WorkFlowId:  int32(cluster.WorkFlowId),
		OwnerId:     cluster.OwnerId,
	}

	if len(cluster.Tags) > 0 {
		bytes, _ := json.Marshal(cluster.Tags)
		dto.Tags = string(bytes)
	}

	if len(cluster.Demands) > 0 {
		bytes, _ := json.Marshal(cluster.Demands)
		dto.Demands = string(bytes)
	}

	return
}

func ParseFromClusterDTO(dto *dbpb.DBClusterDTO) (cluster *domain.Cluster) {
	if dto == nil {
		return nil
	}
	cluster = &domain.Cluster{
		Id:             dto.Id,
		Code:           dto.Code,
		TenantId:       dto.TenantId,
		ClusterName:    dto.Name,
		DbPassword:     dto.DbPassword,
		ClusterType:    *knowledge.ClusterTypeFromCode(dto.ClusterType),
		ClusterVersion: *knowledge.ClusterVersionFromCode(dto.VersionCode),
		Tls:            dto.Tls,
		Status:         domain.ClusterStatusFromValue(int(dto.Status)),
		WorkFlowId:     uint(dto.WorkFlowId),
		OwnerId:        dto.OwnerId,
		CreateTime:     time.Unix(dto.CreateTime, 0),
		UpdateTime:     time.Unix(dto.UpdateTime, 0),
		DeleteTime:     time.Unix(dto.DeleteTime, 0),
	}

	json.Unmarshal([]byte(dto.Tags), &cluster.Tags)
	json.Unmarshal([]byte(dto.Demands), &cluster.Demands)

	return
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
