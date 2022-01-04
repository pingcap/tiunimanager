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
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"time"

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
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.CreateCluster(ctx, models.Cluster{Entity: models.Entity{TenantId: dto.TenantId},
		Name:            dto.Name,
		DbPassword:      dto.DbPassword,
		Type:            dto.ClusterType,
		Version:         dto.VersionCode,
		Tls:             dto.Tls,
		Tags:            dto.Tags,
		OwnerId:         dto.OwnerId,
		Region:          dto.Region,
		Exclusive:       dto.Exclusive,
		CpuArchitecture: dto.CpuArchitecture,
	})
	if nil == err {
		do, demand, newErr := clusterManager.UpdateComponentDemand(ctx, cluster.ID, req.Cluster.Demands, cluster.TenantId)
		if newErr == nil {
			resp.Status = ClusterSuccessResponseStatus
			resp.Cluster = convertToClusterDTO(do, demand)
		} else {
			err = fmt.Errorf("CreateCluster failed, update cluster demand failed, clusterId: %s, tenantId: %s, errors: %v", cluster.ID, cluster.TenantId, newErr)
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

func (handler *DBServiceHandler) UpdateClusterDemand(ctx context.Context, req *dbpb.DBUpdateDemandRequest, resp *dbpb.DBUpdateDemandResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateComponentDemand has invalid parameter")
	}
	clusterManager := handler.Dao().ClusterManager()
	log := framework.LogWithContext(ctx)
	do, demand, err := clusterManager.UpdateComponentDemand(ctx, req.ClusterId, req.Demands, req.TenantId)
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, demand)
		log.Infof("UpdateComponentDemand successful, clusterId: %s, tenantId: %s", req.ClusterId, req.TenantId)
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("CreateCluster failed, clusterId: %s, tenantId: %s, error: %v", req.ClusterId, req.TenantId, err)
	}

	return nil
}

func (handler *DBServiceHandler) DeleteCluster(ctx context.Context, req *dbpb.DBDeleteClusterRequest, resp *dbpb.DBDeleteClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteCluster has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	cluster, err := clusterManager.DeleteCluster(ctx, req.ClusterId)
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

func (handler *DBServiceHandler) DeleteInstance(ctx context.Context, req *dbpb.DBDeleteInstanceRequest, resp *dbpb.DBDeleteInstanceResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("DeleteInstance has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	err = clusterManager.DeleteClusterComponentInstance(ctx, req.Id)

	if err != nil {
		resp.Status = BizErrResponseStatus
		log.Infof("DeleteInstance failed, instanceId: %s， error: %v", req.Id, err)
	} else {
		resp.Status = ClusterSuccessResponseStatus
		log.Infof("DeleteInstance successfully, instanceId: %s", req.Id)
	}

	return nil
}

func (handler *DBServiceHandler) CreateInstance(ctx context.Context, req *dbpb.DBCreateInstanceRequest, resp *dbpb.DBCreateInstanceResponse) (err error) {
	log := framework.LogWithContext(ctx)

	if nil == req || nil == resp {
		log.Error("CreateInstance has invalid parameter")
		// todo handle error
		return errors.Errorf("CreateInstance has invalid parameter")
	}

	clusterManager := handler.Dao().ClusterManager()

	if err == nil {
		componentInstances, err := clusterManager.AddClusterComponentInstance(ctx, req.ClusterId, convertToComponentInstance(req.ComponentInstances))
		if err == nil {
			resp.Status = ClusterSuccessResponseStatus
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

func (handler *DBServiceHandler) UpdateClusterInfo(ctx context.Context, req *dbpb.DBUpdateClusterInfoRequest, resp *dbpb.DBUpdateClusterInfoResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterInfo has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	do, err := clusterManager.UpdateClusterInfo(ctx, req.ClusterId, req.Name, req.ClusterType, req.VersionCode, req.Tags, req.Tls)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Cluster = convertToClusterDTO(do, nil)
		log.Infof("UpdateClusterInfo successful, clusterId: %s, req: %s, error: %v",
			req.GetClusterId(), req, err)
	} else {
		err = errors.New(fmt.Sprintf("UpdateClusterInfo failed, clusterId: %s, req: %s, error: %v",
			req.GetClusterId(), req, err))
		log.Infof("UpdateClusterInfo failed, clusterId: %s, req: %s, error: %v",
			req.GetClusterId(), req, err)
	}
	return err
}

func (handler *DBServiceHandler) UpdateClusterTopologyConfig(ctx context.Context, req *dbpb.DBUpdateTopologyConfigRequest, resp *dbpb.DBUpdateTopologyConfigResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("UpdateClusterTopologyConfig has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	do, err := clusterManager.UpdateTopologyConfig(ctx, req.ClusterId, req.Content, req.TenantId)
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
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()

	if req.GetUpdateStatus() {
		do, err = clusterManager.UpdateClusterStatus(ctx, req.ClusterId, int8(req.Status))
		if nil != err {
			log.Errorf("UpdateClusterStatus failed, clusterId: %s flowId: %d, ,error: %v",
				req.GetClusterId(), req.GetFlowId(), err)
			return err
		}
	}
	if req.GetUpdateFlow() {
		do, err = clusterManager.UpdateClusterFlowId(ctx, req.ClusterId, uint(req.FlowId))
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
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.FetchCluster(ctx, req.ClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.ClusterDetail = &dbpb.DBClusterDetailDTO{
			Cluster:              convertToClusterDTO(result.Cluster, result.DemandRecord),
			TopologyConfigRecord: convertToConfigDTO(result.TopologyConfig),
			Flow:                 convertFlowToDTO(result.Flow),
			ComponentInstances:   convertToComponentInstanceDTO(result.ComponentInstances),
		}
		log.Infof("LoadCluster successfully, clusterId: %s, error: %v", req.GetClusterId(), err)
	} else {
		resp.Status = &dbpb.DBClusterResponseStatus{
			Code:    500,
			Message: err.Error(),
		}
		log.Errorf("LoadCluster failed, clusterId: %s, error: %v", req.GetClusterId(), err)
	}
	return nil
}

func (handler *DBServiceHandler) ListCluster(ctx context.Context, req *dbpb.DBListClusterRequest, resp *dbpb.DBListClusterResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("ListCluster has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	clusters, total, err := clusterManager.ListClusterDetails(ctx, req.ClusterId, req.ClusterName, req.ClusterType, req.ClusterStatus, req.ClusterTag,
		int((req.PageReq.Page-1)*req.PageReq.PageSize), int(req.PageReq.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbpb.DBPageDTO{
			Page:     req.PageReq.Page,
			PageSize: req.PageReq.PageSize,
			Total:    int32(total),
		}
		clusterDetails := make([]*dbpb.DBClusterDetailDTO, len(clusters))
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
	start := time.Now()
	defer handler.HandleMetrics(start, "SaveBackupRecord", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("SaveBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	dto := req.BackupRecord
	result, err := clusterManager.SaveBackupRecord(ctx, dto)
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("SaveBackupRecord success, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath)
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("SaveBackupRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, filepath: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.FilePath, err.Error())
	}
	return nil
}

func (handler *DBServiceHandler) UpdateBackupRecord(ctx context.Context, req *dbpb.DBUpdateBackupRecordRequest, resp *dbpb.DBUpdateBackupRecordResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "UpdateBackupRecord", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("UpdateBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	dto := req.BackupRecord
	err = handler.Dao().ClusterManager().UpdateBackupRecord(ctx, dto)
	if err != nil {
		log.Errorf("SaveBackupRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), err.Error())
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
	} else {
		log.Infof("SaveBackupRecord success, tenantId: %s, clusterId: %s, operatorId: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId())
		resp.Status = ClusterSuccessResponseStatus
	}
	return nil
}

func (handler *DBServiceHandler) DeleteBackupRecord(ctx context.Context, req *dbpb.DBDeleteBackupRecordRequest, resp *dbpb.DBDeleteBackupRecordResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "DeleteBackupRecord", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("DeleteBackupRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.DeleteBackupRecord(ctx, uint(req.Id))
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecord = convertToBackupRecordDTO(result)
		log.Infof("DeleteBackupRecord success, Id: %d", req.GetId())
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("DeleteBackupRecord failed, Id: %d, error: %s", req.GetId(), err.Error())
	}
	return nil
}

func (handler *DBServiceHandler) QueryBackupRecords(ctx context.Context, req *dbpb.DBQueryBackupRecordRequest, resp *dbpb.DBQueryBackupRecordResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "QueryBackupRecords", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupRecords has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().QueryBackupRecord(ctx, req.ClusterId, req.RecordId)
	if err != nil {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("QueryBackupRecords failed, clusterId: %s, error: %s", req.GetClusterId(), err.Error())
	} else {
		resp.Status = ClusterSuccessResponseStatus
		resp.BackupRecords = convertToBackupRecordDisplayDTO(result.BackupRecord, result.Flow)
	}
	return nil
}

func (handler *DBServiceHandler) ListBackupRecords(ctx context.Context, req *dbpb.DBListBackupRecordsRequest, resp *dbpb.DBListBackupRecordsResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "ListBackupRecords", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("ListBackupRecords has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	backupRecords, total, err := clusterManager.ListBackupRecords(ctx, req.ClusterId, req.StartTime, req.EndTime, req.BackupMode,
		int((req.Page.Page-1)*req.Page.PageSize), int(req.Page.PageSize))

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Page = &dbpb.DBPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		backupRecordDTOs := make([]*dbpb.DBDBBackupRecordDisplayDTO, len(backupRecords))
		for i, v := range backupRecords {
			backupRecordDTOs[i] = convertToBackupRecordDisplayDTO(v.BackupRecord, v.Flow)
		}
		resp.BackupRecords = backupRecordDTOs
		log.Infof("ListBackupRecords success, clusterId: %s, page: %d, page size: %d",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize())
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("ListBackupRecords failed, clusterId: %s, page: %d, page size: %d, error: %s",
			req.GetClusterId(), req.GetPage().GetPage(), req.GetPage().GetPageSize(), err.Error())
	}
	return nil
}

func (handler *DBServiceHandler) SaveRecoverRecord(ctx context.Context, req *dbpb.DBSaveRecoverRecordRequest, resp *dbpb.DBSaveRecoverRecordResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "SaveRecoverRecord", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("SaveRecoverRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	dto := req.RecoverRecord
	result, err := clusterManager.SaveRecoverRecord(ctx, dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.BackupRecordId), uint(dto.FlowId))
	if err == nil {
		resp.Status = ClusterSuccessResponseStatus
		resp.RecoverRecord = convertToRecoverRecordDTO(result)
		log.Infof("SaveRecoverRecord successful, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId())
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("SaveRecoverRecord failed, tenantId: %s, clusterId: %s, operatorId: %s, recordId: %d, error: %s",
			dto.GetTenantId(), dto.GetClusterId(), dto.GetOperatorId(), dto.GetBackupRecordId(), err.Error())
	}
	return nil
}

func (handler *DBServiceHandler) SaveBackupStrategy(ctx context.Context, req *dbpb.DBSaveBackupStrategyRequest, resp *dbpb.DBSaveBackupStrategyResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "SaveBackupStrategy", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("SaveBackupStrategy has invalid parameter")
	}
	dto := req.Strategy
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().SaveBackupStrategy(ctx, dto)

	if err != nil {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("SaveBackupStrategy failed, req: %+v, error: %s", dto, err.Error())
	} else {
		log.Infof("SaveBackupStrategy success, req: %+v", dto)
		resp.Status = ClusterSuccessResponseStatus
		resp.Strategy = convertToBackupStrategyDTO(result)
	}
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategy(ctx context.Context, req *dbpb.DBQueryBackupStrategyRequest, resp *dbpb.DBQueryBackupStrategyResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "QueryBackupStrategy", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupStrategy has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterId := req.ClusterId
	result, err := handler.Dao().ClusterManager().QueryBackupStartegy(ctx, clusterId)
	if err != nil {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("QueryBackupStrategy failed, clusterId: %s, error: %s", req.GetClusterId(), err.Error())
	} else {
		resp.Status = ClusterSuccessResponseStatus
		resp.Strategy = convertToBackupStrategyDTO(result)
	}
	return nil
}

func (handler *DBServiceHandler) QueryBackupStrategyByTime(ctx context.Context, req *dbpb.DBQueryBackupStrategyByTimeRequest, resp *dbpb.DBQueryBackupStrategyByTimeResponse) (err error) {
	start := time.Now()
	defer handler.HandleMetrics(start, "QueryBackupStrategyByTime", int(resp.GetStatus().GetCode()))
	if nil == req || nil == resp {
		return errors.Errorf("QueryBackupStrategyByTime has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	result, err := handler.Dao().ClusterManager().QueryBackupStartegyByTime(ctx, req.GetWeekday(), req.GetStartHour())
	if err != nil {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Errorf("QueryBackupStrategyByTime failed, req: %+v, error: %s", req, err.Error())
	} else {
		resp.Status = ClusterSuccessResponseStatus
		strategyList := make([]*dbpb.DBBackupStrategyDTO, len(result))
		for i, v := range result {
			strategyList[i] = convertToBackupStrategyDTO(v)
		}
		resp.Strategys = strategyList
	}

	return nil
}

func (handler *DBServiceHandler) SaveParametersRecord(ctx context.Context, req *dbpb.DBSaveParametersRequest, resp *dbpb.DBSaveParametersResponse) (err error) {
	if nil == req || nil == resp {
		return errors.Errorf("SaveParametersRecord has invalid parameter")
	}
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	dto := req.Parameters
	result, err := clusterManager.SaveParameters(ctx, dto.TenantId, dto.ClusterId, dto.OperatorId, uint(dto.FlowId), dto.Content)
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
	log := framework.LogWithContext(ctx)
	clusterManager := handler.Dao().ClusterManager()
	result, err := clusterManager.GetCurrentParameters(ctx, req.GetClusterId())
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

func (handler *DBServiceHandler) CreateClusterRelation(ctx context.Context, req *dbpb.DBCreateClusterRelationRequest, resp *dbpb.DBCreateClusterRelationResponse) error {
	log := framework.LogWithContext(ctx)
	if nil == req || nil == resp {
		log.Error("CreateClusterRelation has invalid parameters")
		return errors.Errorf("CreateClusterRelation has invalid parameters")
	}
	dto := req.Relation
	clusterManager := handler.Dao().ClusterManager()
	relation, err := clusterManager.CreateClusterRelation(ctx, models.ClusterRelation{Record: models.Record{TenantId: dto.TenantId},
		SubjectClusterId: dto.SubjectClusterId,
		ObjectClusterId:  dto.ObjectClusterId,
		RelationType:     dto.RelationType,
	})

	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		resp.Relation = convertToClusterRelationDTO(relation)
		log.Infof("CreateClusterRelation successful, clusterRelationId: %d, tenantId: %s", req.Relation.GetId(), req.Relation.GetTenantId())
		return nil
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		errMsg := fmt.Sprintf("CreateClusterRelation falied, clusterRelationId: %d, tenantId: %s, error: %s", req.Relation.GetId(), req.Relation.GetTenantId(), err)
		log.Error(errMsg)
		return err
	}
}

func (handler *DBServiceHandler) SwapClusterRelation(ctx context.Context, req *dbpb.DBSwapClusterRelationRequest, resp *dbpb.DBSwapClusterRelationResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "SwapClusterRelation", int(resp.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	if nil == req || nil == resp {
		log.Error("SwapClusterRelation has invalid parameters")
		return errors.Errorf("SwapClusterRelation has invalid parameters")
	}
	clusterManager := handler.Dao().ClusterManager()
	relation, err := clusterManager.ListClusterRelationById(ctx, uint(req.Id))
	if err != nil {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		log.Error("ListClusterRelationById has no result")
		return errors.Errorf("ListClusterRelationById has no result")
	} else {
		newRelation, err := clusterManager.UpdateClusterRelation(ctx, uint(req.Id), relation.ObjectClusterId, relation.SubjectClusterId, relation.RelationType)
		if nil == err {
			resp.Status = ClusterSuccessResponseStatus
			resp.Relation = convertToClusterRelationDTO(newRelation)
			log.Infof("SwapClusterRelationById successful, clusterrelationId: %d", req.GetId())
			return nil
		} else {
			resp.Status = BizErrResponseStatus
			resp.Status.Message = err.Error()
			errMsg := fmt.Sprintf("SwapClusterRelation falied, clusterRelationId: %d, error: %s", req.GetId(), err)
			log.Error(errMsg)
			return errors.Errorf(errMsg)
		}
	}
}

func (handler *DBServiceHandler) ListClusterRelation(ctx context.Context, req *dbpb.DBListClusterRelationRequest, resp *dbpb.DBListClusterRelationResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "ListClusterRelation", int(resp.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	if nil == req || nil == resp {
		log.Error("ListClusterRelation has invalid parameters")
		return errors.Errorf("ListClusterRelation has invalid parameters")
	}
	clusterManager := handler.Dao().ClusterManager()
	relation, err := clusterManager.ListClusterRelationBySubjectId(ctx, req.SubjectClusterId)
	if nil == err {
		resp.Status = ClusterSuccessResponseStatus
		clusterRelationDTOs := make([]*dbpb.DBClusterRelationDTO, len(relation))
		for i, v := range relation {
			clusterRelationDTOs[i] = convertToClusterRelationDTO(v)
		}
		resp.Relation = clusterRelationDTOs
		log.Infof("ListClusterRelation successful, subjectClusterId: %s", req.GetSubjectClusterId())
		return nil
	} else if len(relation) == 0 {
		resp.Status = ClusterNoResultResponseStatus
		resp.Status.Message = err.Error()
		errMsg := fmt.Sprintf("ListClusterRelation has no result, subjectClusterId: %s", req.GetSubjectClusterId())
		log.Error(errMsg)
		return errors.Errorf(errMsg)
	} else {
		resp.Status = BizErrResponseStatus
		resp.Status.Message = err.Error()
		errMsg := fmt.Sprintf("ListClusterRelation falied, subjectClusterId: %s, error: %s", req.GetSubjectClusterId(), err)
		log.Error(errMsg)
		return errors.Errorf(errMsg)
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
		Size:         do.Size,
		BackupTso:    do.BackupTso,
		FlowId:       int64(do.FlowId),
	}
	return
}

func convertToComponentInstance(dtos []*dbpb.DBComponentInstanceDTO) []*models.ComponentInstance {
	if len(dtos) == 0 {
		return []*models.ComponentInstance{}
	}

	result := make([]*models.ComponentInstance, len(dtos))

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
			Version:        v.Version,
			HostId:         v.HostId,
			Host:           v.Host,
			CpuCores:       int8(v.CpuCores),
			Memory:         int8(v.Memory),
			DiskId:         v.DiskId,
			PortInfo:       v.PortInfo,
			AllocRequestId: v.AllocRequestId,
			DiskPath:       v.DiskPath,
		}
	}

	return result
}

func convertToComponentInstanceDTO(models []*models.ComponentInstance) []*dbpb.DBComponentInstanceDTO {
	if len(models) == 0 {
		return []*dbpb.DBComponentInstanceDTO{}
	}
	result := make([]*dbpb.DBComponentInstanceDTO, len(models))

	for i, v := range models {
		result[i] = &dbpb.DBComponentInstanceDTO{
			Id:       v.ID,
			Code:     v.Code,
			TenantId: v.TenantId,
			Status:   int32(v.Status),

			ClusterId:      v.ClusterId,
			ComponentType:  v.ComponentType,
			Role:           v.Role,
			Version:        v.Version,
			HostId:         v.HostId,
			Host:           v.Host,
			CpuCores:       int32(v.CpuCores),
			Memory:         int32(v.Memory),
			DiskId:         v.DiskId,
			DiskPath:       v.DiskPath,
			PortInfo:       v.PortInfo,
			AllocRequestId: v.AllocRequestId,
			CreateTime:     v.CreatedAt.Unix(),
			UpdateTime:     v.UpdatedAt.Unix(),
			DeleteTime:     nullTime2Unix(sql.NullTime(v.DeletedAt)),
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
		Id:              do.ID,
		Code:            do.Code,
		Name:            do.Name,
		TenantId:        do.TenantId,
		DbPassword:      do.DbPassword,
		ClusterType:     do.Type,
		VersionCode:     do.Version,
		Status:          int32(do.Status),
		Tags:            do.Tags,
		Tls:             do.Tls,
		WorkFlowId:      int32(do.CurrentFlowId),
		OwnerId:         do.OwnerId,
		Region:          do.Region,
		Exclusive:       do.Exclusive,
		CpuArchitecture: do.CpuArchitecture,
		CreateTime:      do.CreatedAt.Unix(),
		UpdateTime:      do.UpdatedAt.Unix(),
		DeleteTime:      nullTime2Unix(sql.NullTime(do.DeletedAt)),
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

func nullTime2Unix(at sql.NullTime) (unix int64) {
	if at.Valid {
		return at.Time.Unix()
	}
	return
}

func unix2NullTime(unix int64) sql.NullTime {
	return sql.NullTime{
		Time:  time.Unix(unix, 0),
		Valid: unix != 0,
	}
}

func convertToClusterRelationDTO(do *models.ClusterRelation) (dto *dbpb.DBClusterRelationDTO) {
	if do == nil {
		return nil
	}
	dto = &dbpb.DBClusterRelationDTO{
		Id:               int64(do.ID),
		TenantId:         do.TenantId,
		SubjectClusterId: do.SubjectClusterId,
		ObjectClusterId:  do.ObjectClusterId,
		RelationType:     do.RelationType,
		CreateTime:       do.CreatedAt.Unix(),
		UpdateTime:       do.UpdatedAt.Unix(),
		DeleteTime:       nullTime2Unix(sql.NullTime(do.DeletedAt)),
	}
	return
}