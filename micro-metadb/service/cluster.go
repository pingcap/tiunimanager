
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

package service

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"gorm.io/gorm"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"github.com/pingcap/errors"
)

var ClusterSuccessResponseStatus = &dbpb.DBClusterResponseStatus{Code: 0}
var ClusterNoResultResponseStatus = &dbpb.DBClusterResponseStatus{Code: 1}
var BizErrResponseStatus = &dbpb.DBClusterResponseStatus{Code: 2}

func (handler *DBServiceHandler) CreateCluster(ctx context.Context, req *dbpb.DBCreateClusterRequest, resp *dbpb.DBCreateClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("CreateCluster has invalid parameter")
	}
	dto := req.Cluster
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.CreateCluster(dto.Name, dto.DbPassword, dto.ClusterType, dto.VersionCode, dto.Tls, dto.Tags, dto.OwnerId, dto.TenantId)
	if nil == err {
		do, demand, err := clusterManager.UpdateClusterDemand(cluster.ID, req.Cluster.Demands, cluster.TenantId)
		if err == nil {
			resp.Status = ClusterSuccessResponseStatus
			resp.Cluster = convertToClusterDTO(do, demand)
		} else {
			err = errors.New(fmt.Sprintf("CreateCluster failed, update cluster demand failed, clusterId: %s, tenantId: %s, errors: %v", cluster.ID, cluster.TenantId, err))
		}
	}
	if nil == err {
		log.Infof("CreateCluster successful, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = "CreateCluster failed"
		log.Infof("CreateCluster failed, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	}
	return nil
}

func (handler *DBServiceHandler) DeleteCluster(ctx context.Context, req *dbpb.DBDeleteClusterRequest, resp *dbpb.DBDeleteClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.DeleteCluster(req.ClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(cluster, nil)
	} else {
		err = errors.New(fmt.Sprintf("DeleteCluster failed, clusterId: %s, error: %s", req.GetClusterId(), err))
	}
	if nil == err {
		log.Infof("DeleteCluster successful, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	} else {
		log.Infof("DeleteCluster failed, clusterId: %s, tenantId: %s, error: %v", cluster.ID, cluster.TenantId, err)
	}
	return err
}

func (handler *DBServiceHandler) CreateInstance(ctx context.Context, req *dbpb.DBCreateInstanceRequest, resp *dbpb.DBCreateInstanceResponse) (err error) {
	log := framework.Log()

	if nil == req || nil == resp {
		log.Error("CreateInstance has invalid parameter")
		// todo handle error
		return errors.Errorf("CreateInstance has invalid parameter")
	}

	clusterManager := handler.Dao().ClusterManager()

	cluster, err := clusterManager.UpdateTopologyConfig(req.ClusterId, req.TopologyContent, req.TenantId)
	if err == nil {
		componentInstances, err := clusterManager.AddClusterComponentInstance(req.ClusterId, convertToComponentInstance(req.ComponentInstances))
		if err == nil {
			resp.Status = ClusterSuccessResponseStatus
			resp.Cluster = convertToClusterDTO(cluster, nil)
			resp.ComponentInstances = convertToComponentInstanceDTO(componentInstances)
			log.Infof("CreateInstance successful, clusterId: %s, tenantId: %s, error: %v",
				req.GetClusterId(), req.GetTenantId(), err)
			return nil
		}
	}

	errMsg := fmt.Sprintf("CreateInstance failed, clusterId: %s, tenantId: %s, error: %v",
		req.GetClusterId(), req.GetTenantId(), err)
	log.Error(errMsg)

	// todo handle error
	return err
}

func (handler *DBServiceHandler) UpdateClusterTopologyConfig(ctx context.Context, req *dbpb.DBUpdateTopologyConfigRequest, resp *dbpb.DBUpdateTopologyConfigResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterTopologyConfig has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	do, err := clusterManager.UpdateTopologyConfig(req.ClusterId, req.Content, req.TenantId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, nil)
		log.Infof("UpdateClusterTopologyConfig successful, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err)
	} else {
		err = errors.New(fmt.Sprintf("UpdateClusterTopologyConfig failed, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err))
		log.Infof("UpdateClusterTopologyConfig failed, clusterId: %s, tenantId: %s, error: %v",
			req.GetClusterId(), req.GetTenantId(), err)
	}
	return err
}

func (handler *DBServiceHandler) UpdateClusterStatus(ctx context.Context, req *dbpb.DBUpdateClusterStatusRequest, resp *dbpb.DBUpdateClusterStatusResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterStatus has invalid parameter")
	}
	var do *models.Cluster
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()

	if req.GetUpdateStatus() {
		do, err = clusterManager.UpdateClusterStatus(req.ClusterId, int8(req.Status))
		if nil != err {
			log.Errorf("UpdateClusterStatus failed, clusterId: %s flowId: %d, ,error: %v",
				req.GetClusterId(), req.GetFlowId(), err)
			return err
		}
	}
	if req.GetUpdateFlow() {
		do, err = clusterManager.UpdateClusterFlowId(req.ClusterId, uint(req.FlowId))
	}
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, nil)
		log.Infof("UpdateClusterStatus successful, clusterId: %s flowId: %d, error: %v",
			req.GetClusterId(), req.GetFlowId(), err)
	} else {
		log.Errorf("UpdateClusterStatus failed, clusterId: %s flowId: %d, ,error: %v",
			req.GetClusterId(), req.GetFlowId(), err)
	}

	return err
}

func (handler *DBServiceHandler) LoadCluster(ctx context.Context, req *dbpb.DBLoadClusterRequest, resp *dbpb.DBLoadClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("LoadCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.FetchCluster(req.ClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.ClusterDetail = &dbpb.DBClusterDetailDTO{
			Cluster:              convertToClusterDTO(result.Cluster, result.DemandRecord),
			TopologyConfigRecord: convertToConfigDTO(result.TopologyConfig),
			Flow:                 convertFlowToDTO(result.Flow),
			ComponentInstances:   convertToComponentInstanceDTO(result.ComponentInstances),
		}
		log.Infof("LoadCluster successful, clusterId: %s, error: %v", req.GetClusterId(), err)
	} else {
		log.Infof("LoadCluster failed, clusterId: %s, error: %v", req.GetClusterId(), err)
	}
	return err
}

func (handler *DBServiceHandler) ListCluster(ctx context.Context, req *dbpb.DBListClusterRequest, resp *dbpb.DBListClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListCluster has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	clusters, total, err := clusterManager.ListClusterDetails(req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int((req.PageReq.Page-1)*req.PageReq.PageSize), int(req.PageReq.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbpb.DBPageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		clusterDetails := make([]*dbpb.DBClusterDetailDTO, len(clusters), len(clusters))
		for i, v := range clusters {
			clusterDetails[i] = &dbpb.DBClusterDetailDTO{
				Cluster:              convertToClusterDTO(v.Cluster, v.DemandRecord),
				TopologyConfigRecord: convertToConfigDTO(v.TopologyConfig),
				Flow:                 convertFlowToDTO(v.Flow),
			}
		}
		resp.Clusters = clusterDetails
		log.Infof("ListCluster successful, clusterId: %s, page: %d, page_size: %d, total: %d, error: %s",
			req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize(), total, err)
	} else {
		log.Infof("ListCluster failed, clusterId: %s, page: %d, page_size: %d, total: %d, error: %s",
			req.GetClusterId(), req.GetPageReq().GetPage(), req.GetPageReq().GetPageSize(), total, err)
	}
	return err
}

func (handler *DBServiceHandler) SaveBackupRecord(ctx context.Context, req *dbpb.DBSaveBackupRecordRequest, resp *dbpb.DBSaveBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	dto := req.BackupRecord
	result, err := clusterManager.SaveBackupRecord(dto)
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("SaveBackupRecord success, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath)
	} else {
		log.Errorf("SaveBackupRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath, err.Error())
	}
	return err
}

func (handler *DBServiceHandler) UpdateBackupRecord(ctx context.Context, req *dbpb.DBUpdateBackupRecordRequest, resp *dbpb.DBUpdateBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	dto := req.BackupRecord
	err = handler.Dao().ClusterManager().UpdateBackupRecord(dto)
	if err != nil {
		log.Errorf("SaveBackupRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), err.Error())
	}
	log.Infof("SaveBackupRecord success, tenantId: %s, clusterId: %s, operatorId: %s",
		dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId())
	resp.Status = ClusterSuccessResponseStatus
	return err
}

func (handler *DBServiceHandler) DeleteBackupRecord(ctx context.Context, req *dbpb.DBDeleteBackupRecordRequest, resp *dbpb.DBDeleteBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.DeleteBackupRecord(uint(req.Id))
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("DeleteBackupRecord success, Id: %d", req.GetId())
	} else {
		log.Errorf("DeleteBackupRecord failed, Id: %d, error: %s", req.GetId(), err.Error())
	}
	return err
}

func (handler *DBServiceHandler) QueryBackupRecords(ctx context.Context, req *dbpb.DBQueryBackupRecordRequest, resp *dbpb.DBQueryBackupRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupRecords has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().QueryBackupRecord(req.ClusterId, req.RecordId)
	if err != nil {
		log.Errorf("QueryBackupRecords failed, clusterId: %d, error: %s", req.GetClusterId(), err.Error())
		return err
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.BackupRecords = convertToBackupRecordDisplayDTO(result.BackupRecord, result.Flow)
	return nil
}

func (handler *DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbpb.DBListBackupRecordsRequest, resp *dbpb.DBListBackupRecordsResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListBackupRecords has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	backupRecords, total, err := clusterManager.ListBackupRecords(req.ClusterId, req.StartTime, req.EndTime,
		int((req.Page.Page-1)*req.Page.PageSize), int(req.Page.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbpb.DBPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		backupRecordDTOs := make([]*dbpb.DBDBBackupRecordDisplayDTO, len(backupRecords), len(backupRecords))
		for i, v := range backupRecords {
			backupRecordDTOs[i] = convertToBackupRecordDisplayDTO(v.BackupRecord, v.Flow)
		}
		resp.BackupRecords = backupRecordDTOs
		log.Infof("ListBackupRecords success, clusterId: %s, page: %d, page size: %d",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize())
	} else {
		log.Errorf("ListBackupRecords failed, clusterId: %s, page: %d, page size: %d, error: %s",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize(), err.Error())
	}
	return err
}

func (handler *DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbpb.DBSaveRecoverRecordRequest, resp *dbpb.DBSaveRecoverRecordResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveRecoverRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	dto := req.RecoverRecord
	result, err := clusterManager.SaveRecoverRecord(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.BackupRecordId), uint(dto.FlowId))
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.RecoverRecord = convertToRecoverRecordDTO(result)
		log.Infof("SaveRecoverRecord successful, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId())
	} else {
		log.Errorf("SaveRecoverRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId(), err.Error())
	}
	return err
}

func (handler *DBServiceHandler) SaveBackupStrategy(ctx context.Context, req *dbpb.DBSaveBackupStrategyRequest, resp *dbpb.DBSaveBackupStrategyResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveBackupStrategy has invalid parameter")
	}
	dto := req.Strategy
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().SaveBackupStrategy(dto)

	if err != nil {
		log.Errorf("SaveBackupStrategy failed, req: %+v, error: %s", dto, err.Error())
		return err
	}
	log.Infof("SaveBackupStrategy success, req: %+v", dto)
	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = convertToBackupStrategyDTO(result)
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategy(ctx context.Context, req *dbpb.DBQueryBackupStrategyRequest, resp *dbpb.DBQueryBackupStrategyResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupStrategy has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterId := req.ClusterId
	result, err := handler.Dao().ClusterManager().QueryBackupStartegy(clusterId)
	if err != nil {
		log.Errorf("QueryBackupStrategy failed, clusterId: %s, error: %s", req.GetClusterId(), err.Error())
		return err
	}

	resp.Status = ClusterSuccessResponseStatus
	resp.Strategy = convertToBackupStrategyDTO(result)
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategyByTime(ctx context.Context, req *dbpb.DBQueryBackupStrategyByTimeRequest, resp *dbpb.DBQueryBackupStrategyByTimeResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupStrategyByTime has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().QueryBackupStartegyByTime(req.GetWeekday(), req.GetStartHour())
	if err != nil {
		log.Errorf("QueryBackupStrategyByTime failed, req: %+v, error: %s", req, err.Error())
		return err
	}

	resp.Status = ClusterSuccessResponseStatus
	strategyList := make([]*dbpb.DBBackupStrategyDTO, len(result))
	for i, v := range result {
		strategyList[i] = convertToBackupStrategyDTO(v)
	}
	resp.Strategys = strategyList
	return nil
}

func (handler *DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbpb.DBSaveParametersRequest, resp *dbpb.DBSaveParametersResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveParametersRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	dto := req.Parameters
	result, err := clusterManager.SaveParameters(dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.FlowId), dto.Content)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Parameters = convertToParameterRecordDTO(result)
		log.Infof("SaveParametersRecord successful, tenantId: %s, clusterId: %s, flowId: %d, content: %s, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetFlowId(), dto.GetContent(), err)
	} else {
		log.Infof("SaveParametersRecord failed, tenantId: %s, clusterId: %s, flowId: %d, content: %s, error: %v",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetFlowId(), dto.GetContent(), err)
	}
	return err
}

func (handler *DBServiceHandler) GetCurrentParametersRecord(ctx context.Context, req *dbpb.DBGetCurrentParametersRequest, resp *dbpb.DBGetCurrentParametersResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("GetCurrentParametersRecord has invalid parameter")
	}
	log := framework.Log()
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.GetCurrentParameters(req.GetClusterId())
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.Parameters = convertToParameterRecordDTO(result)
		log.Infof("GetCurrentParametersRecord successful, clusterId: %s, error: %v",
			req.GetClusterId(), err)
		return nil
	} else {
		resp.Status = ClusterNoResultResponseStatus
		log.Warnf("GetCurrentParametersRecord failed, clusterId: %s, error: %v",
			req.GetClusterId(), err)
		return nil
	}
}

func convertToBackupRecordDTO(do *models.BackupRecord) (dto *dbpb.DBBackupRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBBackupRecordDTO{
		Id:           int64(do.ID),
		TenantId:     do.TenantId,
		ClusterId:    do.ClusterId,
		StartTime:    do.StartTime.Unix(),
		BackupMethod: do.BackupMethod,
		BackupMode:   do.BackupMode,
		BackupType:   do.BackupType,
		StorageType:  do.StorageType,
		OperatorId:   do.OperatorId,
		FilePath:     do.FilePath,
		Size:  		  do.Size,
		FlowId:       int64(do.FlowId),
	}
	return
}

func convertToComponentInstance(dtos []*dbpb.DBComponentInstanceDTO) []*models.ComponentInstance {
	if dtos == nil || len(dtos) == 0 {
		return []*models.ComponentInstance{}
	}

	result := make([]*models.ComponentInstance, len(dtos), len(dtos))

	for i, v := range dtos {
		result[i] = &models.ComponentInstance{
			Entity: models.Entity{
				ID:       v.Id,
				Code:     v.Code,
				TenantId: v.TenantId,
				Status:   int8(v.Status),
			},
			ClusterId:      v.ClusterId,
			ComponentType:  v.ComponentType,
			Role:           v.Role,
			Spec:           v.Spec,
			Version:        v.Version,
			HostId:         v.HostId,
			DiskId:         v.DiskId,
			PortInfo:       v.PortInfo,
			AllocRequestId: v.AllocRequestId,
		}
	}

	return result
}

func convertToComponentInstanceDTO(models []*models.ComponentInstance) []*dbpb.DBComponentInstanceDTO {
	if models == nil || len(models) == 0 {
		return []*dbpb.DBComponentInstanceDTO{}
	}
	result := make([]*dbpb.DBComponentInstanceDTO, len(models), len(models))

	for i, v := range models {
		result[i] = &dbpb.DBComponentInstanceDTO{
			Id:       v.ID,
			Code:     v.Code,
			TenantId: v.TenantId,
			Status:   int32(v.Status),

			ClusterId:      v.ClusterId,
			ComponentType:  v.ComponentType,
			Role:           v.Role,
			Spec:           v.Spec,
			Version:        v.Version,
			HostId:         v.HostId,
			DiskId:         v.DiskId,
			PortInfo:       v.PortInfo,
			AllocRequestId: v.AllocRequestId,
			CreateTime:     v.CreatedAt.Unix(),
			UpdateTime:     v.UpdatedAt.Unix(),
			DeleteTime:     deletedAtUnix(v.DeletedAt),
		}
	}

	return result
}

func convertToBackupRecordDisplayDTO(do *models.BackupRecord, flow *models.FlowDO) (dto *dbpb.DBDBBackupRecordDisplayDTO) {
	if do == nil {
		return nil
	}

	dto = &dbpb.DBDBBackupRecordDisplayDTO{
		BackupRecord: convertToBackupRecordDTO(do),
		Flow:         convertFlowToDTO(flow),
	}
	return
}

func convertToRecoverRecordDTO(do *models.RecoverRecord) (dto *dbpb.DBRecoverRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBRecoverRecordDTO{
		Id:             int64(do.ID),
		TenantId:       do.TenantId,
		ClusterId:      do.ClusterId,
		CreateTime:     do.CreatedAt.Unix(),
		OperatorId:     do.OperatorId,
		BackupRecordId: int64(do.BackupRecordId),
		FlowId:         int64(do.FlowId),
	}
	return
}

func convertToBackupStrategyDTO(do *models.BackupStrategy) (dto *dbpb.DBBackupStrategyDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBBackupStrategyDTO{
		Id:         int64(do.ID),
		OperatorId: do.OperatorId,
		TenantId:   do.TenantId,
		ClusterId:  do.ClusterId,
		CreateTime: do.CreatedAt.Unix(),
		UpdateTime: do.UpdatedAt.Unix(),
		BackupDate: do.BackupDate,
		StartHour:  do.StartHour,
		EndHour:    do.EndHour,
	}
	return
}

func convertToParameterRecordDTO(do *models.ParametersRecord) (dto *dbpb.DBParameterRecordDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBParameterRecordDTO{
		Id:         int64(do.ID),
		TenantId:   do.TenantId,
		ClusterId:  do.ClusterId,
		CreateTime: do.CreatedAt.Unix(),
		OperatorId: do.OperatorId,
		FlowId:     int64(do.FlowId),
		Content:    do.Content,
	}
	return
}

func convertToClusterDTO(do *models.Cluster, demand *models.DemandRecord) (dto *dbpb.DBClusterDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBClusterDTO{
		Id:          do.ID,
		Code:        do.Code,
		Name:        do.Name,
		TenantId:    do.TenantId,
		DbPassword:  do.DbPassword,
		ClusterType: do.Type,
		VersionCode: do.Version,
		Status:      int32(do.Status),
		Tags:        do.Tags,
		Tls:         do.Tls,
		WorkFlowId:  int32(do.CurrentFlowId),
		OwnerId:     do.OwnerId,
		CreateTime:  do.CreatedAt.Unix(),
		UpdateTime:  do.UpdatedAt.Unix(),
		DeleteTime:  deletedAtUnix(do.DeletedAt),
	}

	if demand != nil {
		dto.Demands = demand.Content
	}
	return
}

func convertToConfigDTO(do *models.TopologyConfig) (dto *dbpb.DBTopologyConfigDTO) {
	if do == nil {
		return nil
	}
	return &dbpb.DBTopologyConfigDTO{
		Id:         int32(do.ID),
		TenantId:   do.TenantId,
		ClusterId:  do.ClusterId,
		Content:    do.Content,
		CreateTime: do.CreatedAt.Unix(),
	}
}

func deletedAtUnix(at gorm.DeletedAt) (unix int64) {
	if at.Valid {
		return at.Time.Unix()
	}
	return
}
