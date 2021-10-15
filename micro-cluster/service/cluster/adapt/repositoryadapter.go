package adapt

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"strconv"
	"time"
)

func InjectionMetaDbRepo() {
	domain.TaskRepo = TaskRepoAdapter{}
	domain.ClusterRepo = ClusterRepoAdapter{}
	domain.InstanceRepo = InstanceRepoAdapter{}
}

type ClusterRepoAdapter struct{}

func (c ClusterRepoAdapter) Query(clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*domain.ClusterAggregation, int, error) {
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

	resp, err := client.DBClient.ListCluster(context.TODO(), req)

	if err != nil {
		return nil, 0, err
	}
	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
		return nil, 0, err
	}

	clusters := make([]*domain.ClusterAggregation, len(resp.Clusters), len(resp.Clusters))
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

func (c ClusterRepoAdapter) AddCluster(cluster *domain.Cluster) error {
	req := &dbpb.DBCreateClusterRequest{
		Cluster: ConvertClusterToDTO(cluster),
	}

	resp, err := client.DBClient.CreateCluster(context.TODO(), req)
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

func (c ClusterRepoAdapter) Persist(aggregation *domain.ClusterAggregation) error {
	cluster := aggregation.Cluster

	if aggregation.StatusModified || aggregation.FlowModified {
		resp, err := client.DBClient.UpdateClusterStatus(context.TODO(), &dbpb.DBUpdateClusterStatusRequest{
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
		resp, err := client.DBClient.UpdateClusterTopologyConfig(context.TODO(), &dbpb.DBUpdateTopologyConfigRequest{
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
		resp, err := client.DBClient.SaveParametersRecord(context.TODO(), &dbpb.DBSaveParametersRequest{
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

	return nil
}

func (c ClusterRepoAdapter) Load(id string) (cluster *domain.ClusterAggregation, err error) {
	req := &dbpb.DBLoadClusterRequest{
		ClusterId: id,
	}

	resp, err := client.DBClient.LoadCluster(context.TODO(), req)

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

type InstanceRepoAdapter struct{}

func (c InstanceRepoAdapter) QueryParameterJson(clusterId string) (content string, err error) {
	resp, err := client.DBClient.GetCurrentParametersRecord(context.TODO(), &dbpb.DBGetCurrentParametersRequest{
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

func (t TaskRepoAdapter) ListFlows(bizId, keyword string, status int, page int, pageSize int) ([]*domain.FlowWorkEntity, int, error) {
	resp, err :=client.DBClient.ListFlows(context.TODO(), &dbpb.DBListFlowsRequest{
		BizId:   bizId,
		Keyword: keyword,
		Status: int64(status),
		Page: &dbpb.DBTaskPageDTO{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	})

	if err != nil {
		framework.Log().Errorf("AddFlowWork error = %s", err.Error())
		return nil, 0, err
	}
	flows := make([]*domain.FlowWorkEntity, len(resp.Flows), len(resp.Flows))
	for i, v := range resp.Flows {
		flows[i] = &domain.FlowWorkEntity{
			Id:          uint(v.Id),
			FlowName:    v.FlowName,
			StatusAlias: v.StatusAlias,
			BizId:       v.BizId,
			Status: domain.TaskStatus(v.Status),
			Operator: &domain.Operator{
				Name: v.Operator,
			},
			CreateTime: time.Unix(v.CreateTime, 0),
			UpdateTime: time.Unix(v.UpdateTime, 0),
		}
	}

	return flows, int(resp.Page.Total), err
}

func (t TaskRepoAdapter) QueryCronTask(bizId string, cronTaskType int) (cronTask *domain.CronTaskEntity, err error) {
	cronTask = domain.GetDefaultMaintainTask()
	return
}

func (t TaskRepoAdapter) PersistCronTask(cronTask *domain.CronTaskEntity) (err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) AddFlowWork(flowWork *domain.FlowWorkEntity) error {
	resp, err := client.DBClient.CreateFlow(context.TODO(), &dbpb.DBCreateFlowRequest{
		Flow: &dbpb.DBFlowDTO{
			FlowName:    flowWork.FlowName,
			StatusAlias: flowWork.StatusAlias,
			BizId:       flowWork.BizId,
			Operator: flowWork.Operator.Name,
		},
	})

	if err != nil {
		framework.Log().Errorf("AddFlowWork error = %s", err.Error())
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
	} else {
		flowWork.Id = uint(resp.Flow.Id)
	}
	return err
}

func (t TaskRepoAdapter) AddFlowTask(task *domain.TaskEntity, flowId uint) error {
	resp, err := client.DBClient.CreateTask(context.TODO(), &dbpb.DBCreateTaskRequest{
		Task: &dbpb.DBTaskDTO{
			TaskName:       task.TaskName,
			TaskReturnType: strconv.Itoa(int(task.TaskReturnType)),
			BizId:          task.BizId,
			Parameters:     task.Parameters,
			ParentId:       strconv.Itoa(int(flowId)),
			ParentType:     0,
		},
	})

	if err != nil {
		// todo
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
	} else {
		task.Id = uint(resp.Task.Id)
	}
	return err
}

func (t TaskRepoAdapter) AddCronTask(cronTask *domain.CronTaskEntity) error {
	return nil
}

func (t TaskRepoAdapter) Persist(flowWork *domain.FlowWorkAggregation) error {
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

	tasks := make([]*dbpb.DBTaskDTO, len(flowWork.Tasks), len(flowWork.Tasks))
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
		}
	}

	_, err := client.DBClient.UpdateFlow(context.TODO(), req)
	return err
}

func (t TaskRepoAdapter) LoadFlowWork(id uint) (flow *domain.FlowWorkEntity, err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) Load(id uint) (flowWork *domain.FlowWorkAggregation, err error) {
	panic("implement me")
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

	json.Unmarshal([]byte(dto.Tags), cluster.Tags)
	json.Unmarshal([]byte(dto.Demands), cluster.Demands)

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
	json.Unmarshal([]byte(dto.Content), spec)

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
