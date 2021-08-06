package adapt

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiem/micro-metadb/client"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"time"
)

func InjectionMetaDbRepo() {
	domain.TaskRepo = TaskRepoAdapter{}
	domain.ClusterRepo = ClusterRepoAdapter{}
	domain.InstanceRepo = InstanceRepoAdapter{}
}

type ClusterRepoAdapter struct {}

func (c ClusterRepoAdapter) Query(clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*domain.ClusterAggregation, int, error) {
	req := &db.DBListClusterRequest {
		ClusterName: clusterName,
		ClusterId: clusterId,
		ClusterTag: clusterTag,
		ClusterStatus: clusterStatus,
		ClusterType: clusterType,
		PageReq: &db.DBPageDTO{
			Page: int32(page),
			PageSize: int32(pageSize),
		},
	}

	resp, err := client.DBClient.ListCluster(context.TODO(), req)

	if err != nil {
		return nil, 0, err
	}
	if resp.Status.Code != 0  {
		err = errors.New(resp.Status.Message)
		return nil, 0, err
	}

	clusters := make([]*domain.ClusterAggregation, len(resp.Clusters), len(resp.Clusters))
	for i,v := range resp.Clusters {
		cluster := &domain.ClusterAggregation{}
		cluster.Cluster = ParseFromClusterDTO(v.Cluster)
		cluster.CurrentTiUPConfigRecord = parseConfigRecordDTO(v.TiupConfigRecord)
		cluster.CurrentWorkFlow = parseFlowFromDTO(v.Flow)
		cluster.MaintainCronTask = domain.GetDefaultMaintainTask() // next_version get from db

		clusters[i] = cluster
	}
	return clusters, 0, err

}

func (c ClusterRepoAdapter) AddCluster(cluster *domain.Cluster) error {
	req := &db.DBCreateClusterRequest{
		Cluster: ConvertClusterToDTO(cluster),
	}

	resp, err := client.DBClient.CreateCluster(context.TODO(), req)
	if err != nil {
		return err
	}

	if resp.Status.Code != 0  {
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
		resp, err := client.DBClient.UpdateClusterStatus(context.TODO(), &db.DBUpdateClusterStatusRequest{
			ClusterId:    cluster.Id,
			Status:       int32(cluster.Status),
			UpdateStatus: aggregation.StatusModified,
			FlowId: int64(aggregation.CurrentWorkFlow.Id),
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
		resp, err := client.DBClient.UpdateClusterTiupConfig(context.TODO(), &db.DBUpdateTiupConfigRequest{
			ClusterId: aggregation.Cluster.Id,
			Content: aggregation.CurrentTiUPConfigRecord.Content(),
			TenantId: aggregation.Cluster.TenantId,
		})

		if err != nil {
			// todo
			return err
		}
		aggregation.CurrentTiUPConfigRecord = parseConfigRecordDTO(resp.TiupConfigRecord)
	}

	if aggregation.LastBackupRecord != nil && aggregation.LastBackupRecord.Id == 0 {
		record := aggregation.LastBackupRecord
		resp, err :=  client.DBClient.SaveBackupRecord(context.TODO(), &db.DBSaveBackupRecordRequest{
			BackupRecord: &db.DBBackupRecordDTO{
				TenantId:    cluster.TenantId,
				ClusterId:   record.ClusterId,
				BackupType: int32(record.BackupType),
				BackupRange: int32(record.Range),
				OperatorId:  record.OperatorId,
				FilePath:    record.FilePath,
				FlowId:      int64(aggregation.CurrentWorkFlow.Id),
			},
		})
		if err != nil {
			// todo
			return err
		}
		aggregation.LastBackupRecord.Id = uint(resp.BackupRecord.Id)
	}

	if aggregation.LastRecoverRecord != nil && aggregation.LastRecoverRecord.Id == 0 {
		record := aggregation.LastRecoverRecord
		resp, err :=  client.DBClient.SaveRecoverRecord(context.TODO(), &db.DBSaveRecoverRecordRequest{
			RecoverRecord: &db.DBRecoverRecordDTO{
				TenantId:       cluster.TenantId,
				ClusterId:      record.ClusterId,
				OperatorId:     record.OperatorId,
				BackupRecordId: int64(record.BackupRecord.Id),
				FlowId:         int64(aggregation.CurrentWorkFlow.Id),
			},
		})
		if err != nil {
			// todo
			return err
		}
		aggregation.LastRecoverRecord.Id = uint(resp.RecoverRecord.Id)
	}

	if aggregation.LastParameterRecord != nil && aggregation.LastParameterRecord.Id == 0 {
		record := aggregation.LastParameterRecord
		resp, err :=  client.DBClient.SaveParametersRecord(context.TODO(), &db.DBSaveParametersRequest{
			Parameters: &db.DBParameterRecordDTO{
				TenantId:       cluster.TenantId,
				ClusterId:      record.ClusterId,
				OperatorId:     record.OperatorId,
				Content: 		record.Content,
				FlowId:         int64(aggregation.CurrentWorkFlow.Id),
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
	req := &db.DBLoadClusterRequest{
		ClusterId: id,
	}

	resp, err := client.DBClient.LoadCluster(context.TODO(), req)

	if err != nil {
		return
	}

	if resp.Status.Code != 0  {
		err = errors.New(resp.Status.Message)
		return
	} else {
		cluster = &domain.ClusterAggregation{}
		cluster.Cluster = ParseFromClusterDTO(resp.ClusterDetail.Cluster)
		cluster.CurrentTiUPConfigRecord = parseConfigRecordDTO(resp.ClusterDetail.TiupConfigRecord)
		cluster.CurrentWorkFlow = parseFlowFromDTO(resp.ClusterDetail.Flow)
		cluster.MaintainCronTask = domain.GetDefaultMaintainTask() // next_version get from db
		return
	}
}

type InstanceRepoAdapter struct {}

func (c InstanceRepoAdapter) QueryParameterJson(clusterId string) (content string, err error) {
	resp, err := client.DBClient.GetCurrentParametersRecord(context.TODO(), &db.DBGetCurrentParametersRequest{
		ClusterId: clusterId,
	})

	if err != nil {
		return
	}

	if resp.Status.Code != 0 {
		err = errors.New(resp.Status.Message)
		return
	} else {
		resp.GetParameters().GetContent()
	}
	return
}
type TaskRepoAdapter struct {}

func (t TaskRepoAdapter) QueryCronTask(bizId string, cronTaskType int) (cronTask *domain.CronTaskEntity, err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) PersistCronTask(cronTask *domain.CronTaskEntity) (err error) {
	panic("implement me")
}

func (t TaskRepoAdapter) AddFlowWork(flowWork *domain.FlowWorkEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) AddTask(task *domain.TaskEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) AddCronTask(cronTask *domain.CronTaskEntity) error {
	panic("implement me")
}

func (t TaskRepoAdapter) Persist(flowWork *domain.FlowWorkAggregation) error {
	panic("implement me")
}

func (t TaskRepoAdapter) LoadFlowWork(id uint) (*domain.FlowWorkEntity, error) {
	panic("implement me")
}

func (t TaskRepoAdapter) Load(id uint) (flowWork *domain.FlowWorkAggregation, err error) {
	panic("implement me")
}

func ConvertClusterToDTO(cluster *domain.Cluster) (dto *db.DBClusterDTO) {
	if cluster == nil {
		return
	}
	dto = &db.DBClusterDTO{
		Id :         cluster.Id,
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

func ParseFromClusterDTO(dto *db.DBClusterDTO) (cluster *domain.Cluster) {
	cluster = &domain.Cluster{
		Id: dto.Id,
		Code:           dto.Code,
		TenantId:       dto.TenantId,
		ClusterName:    dto.Name,
		DbPassword:     dto.DbPassword,
		ClusterType:    *knowledge.ClusterTypeFromCode(dto.ClusterType),
		ClusterVersion: *knowledge.ClusterVersionFromCode(dto.VersionCode),
		Tls:            dto.Tls,
		Status:         domain.ClusterStatusFromValue(int(dto.Status)),
		WorkFlowId: uint(dto.WorkFlowId),
		OwnerId:        dto.OwnerId,
		CreateTime:     time.Unix(dto.CreateTime, 0),
		UpdateTime:     time.Unix(dto.UpdateTime, 0),
		DeleteTime:     time.Unix(dto.DeleteTime, 0),
	}

	json.Unmarshal([]byte(dto.Tags), cluster.Tags)
	json.Unmarshal([]byte(dto.Demands), cluster.Demands)

	return
}

func parseConfigRecordDTO(dto *db.DBTiUPConfigDTO) (record *domain.TiUPConfigRecord) {
	record = &domain.TiUPConfigRecord{
		Id: uint(dto.Id),
		TenantId:   dto.TenantId,
		ClusterId:  dto.ClusterId,
		CreateTime: time.Unix(dto.CreateTime, 0),
	}

	spec := &spec.Specification{}
	json.Unmarshal([]byte(dto.Content), spec)

	return
}

func parseFlowFromDTO(dto *db.DBFlowDTO) (flow *domain.FlowWorkEntity) {
	flow = &domain.FlowWorkEntity {
		Id: uint(dto.Id),
		FlowName:    dto.FlowName,
		StatusAlias: dto.StatusAlias,
		BizId:       dto.BizId,
		Status:      domain.TaskStatusFromValue(int(dto.Status)),
	}
	// todo
	return
}
