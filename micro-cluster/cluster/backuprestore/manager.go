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
 ******************************************************************************/

package backuprestore

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/bytes"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var brService BRService
var once sync.Once

func GetBRService() BRService {
	once.Do(func() {
		if brService == nil {
			brService = NewBRManager()
		}
	})
	return brService
}

func MockBRService(service BRService) {
	brService = service
}

type BRManager struct {
	autoBackupMgr *autoBackupManager
}

func NewBRManager() *BRManager {
	flowManager := workflow.GetWorkFlowService()
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowBackupCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowBackupCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"backup", "backupDone", "fail", workflow.SyncFuncNode, backupCluster},
			"backupDone":       {"updateBackupRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateBackupRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"end", "", "", workflow.SyncFuncNode, backupFail},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestoreExistCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowRestoreExistCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":       {"restoreFromSrcCluster", "restoreDone", "fail", workflow.SyncFuncNode, restoreFromSrcCluster},
			"restoreDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":        {"end", "", "", workflow.SyncFuncNode, restoreFail},
		},
	})

	mgr := &BRManager{
		autoBackupMgr: NewAutoBackupManager(),
	}

	return mgr
}

func (mgr *BRManager) BackupCluster(ctx context.Context, request cluster.BackupClusterDataReq, maintenanceStatusChange bool) (resp cluster.BackupClusterDataResp, backupErr error) {
	framework.LogWithContext(ctx).Infof("Begin BackupCluster, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End BackupCluster")

	if err := mgr.backupClusterPreCheck(ctx, request); err != nil {
		framework.LogWithContext(ctx).Errorf("backup cluster precheck failed: %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("backup cluster precheck failed: %s", err.Error()), err)
	}
	configRW := models.GetConfigReaderWriter()
	storageTypeConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStorageType, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_SYSTEM_CONFIG_INVAILD, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyBackupStorageType, err.Error()), err)
	}
	storagePathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStoragePath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStoragePath, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_SYSTEM_CONFIG_INVAILD, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyBackupStoragePath, err.Error()), err)
	}

	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_NOT_FOUND, fmt.Sprintf("load cluster meta %s failed, %s", request.ClusterID, err.Error()), err)
	}

	if maintenanceStatusChange {
		if err := meta.StartMaintenance(ctx, constants.ClusterMaintenanceBackUp); err != nil {
			framework.LogWithContext(ctx).Errorf("start maintenance failed, %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
		}
		defer func() {
			if backupErr != nil {
				if endErr := meta.EndMaintenance(ctx, meta.Cluster.MaintenanceStatus); endErr != nil {
					framework.LogWithContext(ctx).Warnf("end maintenance failed, %s", err.Error())
				}
			}
		}()
	}

	//todo: only support full physics backup now
	record := &backuprestore.BackupRecord{
		Entity: dbModel.Entity{
			TenantId: meta.Cluster.TenantId,
			Status:   string(constants.ClusterBackupProcessing),
		},
		ClusterID:    request.ClusterID,
		StorageType:  storageTypeConfig.ConfigValue,
		BackupType:   string(constants.BackupTypeFull),
		BackupMethod: string(constants.BackupMethodPhysics),
		BackupMode:   request.BackupMode,
		FilePath:     mgr.getBackupPath(storagePathConfig.ConfigValue, request.ClusterID, time.Now(), string(constants.BackupTypeFull)),
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	brRW := models.GetBRReaderWriter()
	recordCreate, err := brRW.CreateBackupRecord(ctx, record)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("save backup record failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_RECORD_CREATE_FAILED, fmt.Sprintf("save backup record failed, %s", err.Error()), err)
	}
	defer func() {
		if backupErr != nil {
			if delErr := brRW.DeleteBackupRecord(ctx, recordCreate.ID); delErr != nil {
				framework.LogWithContext(ctx).Warnf("delete backup record %+v failed, %s", recordCreate, err.Error())
			}
		}
	}()

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, workflow.BizTypeCluster, constants.FlowBackupCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, fmt.Sprintf("create %s workflow failed, %s", constants.FlowBackupCluster, err.Error()), err)
	}

	flowManager.AddContext(flow, contextBackupRecordKey, recordCreate)
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextMaintenanceStatusChangeKey, maintenanceStatusChange)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_START_FAILED, fmt.Sprintf("async start %s workflow failed, %s", constants.FlowBackupCluster, err.Error()), err)
	}

	resp.WorkFlowID = flow.Flow.ID
	resp.BackupID = recordCreate.ID

	return resp, nil
}

func (mgr *BRManager) RestoreExistCluster(ctx context.Context, request cluster.RestoreExistClusterReq, maintenanceStatusChange bool) (resp cluster.RestoreExistClusterResp, restoreErr error) {
	framework.LogWithContext(ctx).Infof("Begin RestoreExistCluster, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End RestoreExistCluster")

	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_NOT_FOUND, fmt.Sprintf("load cluster meta %s failed, %s", request.ClusterID, err.Error()), err)
	}

	brRW := models.GetBRReaderWriter()
	record, err := brRW.GetBackupRecord(ctx, request.BackupID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get backup record %s failed, %s", request.BackupID, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_RECORD_QUERY_FAILED, fmt.Sprintf("get backup record %s failed, %s", request.BackupID, err.Error()), err)
	}

	if maintenanceStatusChange {
		if err := meta.StartMaintenance(ctx, constants.ClusterMaintenanceRestore); err != nil {
			framework.LogWithContext(ctx).Errorf("start maintenance failed, %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
		}
		defer func() {
			if restoreErr != nil {
				if endErr := meta.EndMaintenance(ctx, meta.Cluster.MaintenanceStatus); endErr != nil {
					framework.LogWithContext(ctx).Warnf("end maintenance failed, %s", err.Error())
				}
			}
		}()
	}

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, workflow.BizTypeCluster, constants.FlowRestoreExistCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, fmt.Sprintf("create %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error()), err)
	}

	flowManager.AddContext(flow, contextBackupRecordKey, record)
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextMaintenanceStatusChangeKey, maintenanceStatusChange)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_START_FAILED, fmt.Sprintf("async start %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error()), err)
	}

	resp.WorkFlowID = flow.Flow.ID
	return resp, nil
}

func (mgr *BRManager) QueryClusterBackupRecords(ctx context.Context, request cluster.QueryBackupRecordsReq) (resp cluster.QueryBackupRecordsResp, page structs.Page, err error) {
	framework.LogWithContext(ctx).Infof("Begin QueryClusterBackupRecords, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End QueryClusterBackupRecords")

	brRW := models.GetBRReaderWriter()
	records, total, err := brRW.QueryBackupRecords(ctx, request.ClusterID, request.BackupID, "", request.StartTime, request.EndTime, request.Page, request.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query cluster backup records %+v failed %s", request, err.Error())
		return resp, page, errors.WrapError(errors.TIEM_BACKUP_RECORD_QUERY_FAILED, fmt.Sprintf("query cluster backup records %+v failed %s", request, err.Error()), err)
	}

	response := cluster.QueryBackupRecordsResp{
		BackupRecords: make([]*structs.BackupRecord, len(records)),
	}
	for index, record := range records {
		response.BackupRecords[index] = &structs.BackupRecord{
			ID:           record.ID,
			ClusterID:    record.ClusterID,
			BackupType:   record.BackupType,
			BackupMethod: record.BackupMethod,
			BackupMode:   record.BackupMode,
			FilePath:     record.FilePath,
			Size:         float32(record.Size) / bytes.MB, //Byte to MByte,
			BackupTSO:    strconv.FormatUint(record.BackupTso, 10),
			Status:       record.Status,
			StartTime:    record.StartTime,
			EndTime:      record.EndTime,
			CreateTime:   record.CreatedAt,
			UpdateTime:   record.UpdatedAt,
			DeleteTime:   record.DeletedAt.Time,
		}
	}

	return response, structs.Page{Page: request.Page, PageSize: request.PageSize, Total: int(total)}, nil
}

func (mgr *BRManager) DeleteBackupRecords(ctx context.Context, request cluster.DeleteBackupDataReq) (resp cluster.DeleteBackupDataResp, err error) {
	framework.LogWithContext(ctx).Infof("Begin DeleteBackupRecords, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End DeleteBackupRecords")

	if request.ClusterID == "" && request.BackupID == "" {
		framework.LogWithContext(ctx).Errorf("invalid param clusterId and backupId empty")
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid param clusterId and backupId empty")
	}

	deleteRecordMap := make(map[string]*backuprestore.BackupRecord)
	excludeBackupIdMap := make(map[string]string)
	if len(request.ExcludeBackupIDs) > 0 {
		for _, excludeId := range request.ExcludeBackupIDs {
			excludeBackupIdMap[excludeId] = excludeId
		}
	}

	brRW := models.GetBRReaderWriter()
	for page, pageSize := 1, defaultPageSize; ; page++ {
		records, _, err := brRW.QueryBackupRecords(ctx, request.ClusterID, request.BackupID, request.BackupMode, 0, 0, page, pageSize)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("query backup records of request %+v, failed, %s", request, err.Error())
			return resp, errors.WrapError(errors.TIEM_BACKUP_RECORD_QUERY_FAILED, fmt.Sprintf("query cluster %s backup records failed %s", request.ClusterID, err.Error()), err)
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			if _, ok := excludeBackupIdMap[record.ID]; ok {
				framework.LogWithContext(ctx).Infof("current backupId %s in excludeBackupIds, skip delete!", record.ID)
				continue
			}
			deleteRecordMap[record.ID] = record
		}
	}

	for recordId, record := range deleteRecordMap {
		framework.LogWithContext(ctx).Infof("begin delete backup record %+v", record)
		err = mgr.removeBackupFiles(ctx, record)
		if err != nil {
			framework.LogWithContext(ctx).Warnf("remove backup files of recordId %s failed, %s", recordId, err.Error())
		}
		err = brRW.DeleteBackupRecord(ctx, recordId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("delete backup record %s failed, %s", recordId, err.Error())
			return resp, errors.WrapError(errors.TIEM_BACKUP_RECORD_DELETE_FAILED, fmt.Sprintf("delete backup record %s failed, %s", recordId, err.Error()), err)
		}
		framework.LogWithContext(ctx).Infof("success delete backup record %+v", record)
	}

	return resp, nil
}

func (mgr *BRManager) GetBackupStrategy(ctx context.Context, request cluster.GetBackupStrategyReq) (resp cluster.GetBackupStrategyResp, err error) {
	framework.LogWithContext(ctx).Infof("Begin QueryBackupStrategy, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End QueryBackupStrategy")

	brRW := models.GetBRReaderWriter()
	strategy, err := brRW.GetBackupStrategy(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get backup strategy of cluster %s failed %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_STRATEGY_QUERY_FAILED, fmt.Sprintf("get backup strategy of cluster %s failed %s", request.ClusterID, err.Error()), err)
	}

	resp.Strategy = structs.BackupStrategy{
		ClusterID:  request.ClusterID,
		BackupDate: strategy.BackupDate,
		Period:     fmt.Sprintf("%d:00-%d:00", strategy.StartHour, strategy.EndHour),
	}
	return resp, nil
}

func (mgr *BRManager) SaveBackupStrategy(ctx context.Context, request cluster.SaveBackupStrategyReq) (resp cluster.SaveBackupStrategyResp, err error) {
	framework.LogWithContext(ctx).Infof("Begin SaveBackupStrategy, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End SaveBackupStrategy")

	if err = mgr.saveBackupStrategyPreCheck(ctx, request); err != nil {
		framework.LogWithContext(ctx).Errorf("save backup strategy precheck failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("save backup strategy precheck failed, %s", err.Error()), err)
	}

	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s meta failed, %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_NOT_FOUND, fmt.Sprintf("load cluster meta %s failed, %s", request.ClusterID, err.Error()), err)
	}

	period := strings.Split(request.Strategy.Period, "-")
	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, _ := strconv.Atoi(starts[0])
	endHour, _ := strconv.Atoi(ends[0])

	brRW := models.GetBRReaderWriter()
	strategy, err := brRW.SaveBackupStrategy(ctx, &backuprestore.BackupStrategy{
		Entity: dbModel.Entity{
			TenantId: meta.Cluster.TenantId,
		},
		ClusterID:  request.ClusterID,
		BackupDate: request.Strategy.BackupDate,
		StartHour:  uint32(startHour),
		EndHour:    uint32(endHour),
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("save backup strategy %+v failed %s", strategy, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_STRATEGY_SAVE_FAILED, fmt.Sprintf("save backup strategy %+v failed %s", strategy, err.Error()), err)
	}

	return resp, nil
}

func (mgr *BRManager) DeleteBackupStrategy(ctx context.Context, request cluster.DeleteBackupStrategyReq) (resp cluster.DeleteBackupStrategyResp, err error) {
	framework.LogWithContext(ctx).Infof("Begin DeleteBackupStrategy, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End DeleteBackupStrategy")

	brRW := models.GetBRReaderWriter()
	err = brRW.DeleteBackupStrategy(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete cluster %s backup strategy failed %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_BACKUP_STRATEGY_DELETE_FAILED, fmt.Sprintf("delete cluster %s backup strategy failed %s", request.ClusterID, err.Error()), err)
	}

	return resp, nil
}

func (mgr *BRManager) backupClusterPreCheck(ctx context.Context, request cluster.BackupClusterDataReq) error {
	configRW := models.GetConfigReaderWriter()
	storageTypeCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStorageType)
	if err != nil || storageTypeCfg.ConfigValue == "" {
		return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupStorageType)
	}
	storagePathCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStoragePath)
	if err != nil || storagePathCfg.ConfigValue == "" {
		return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupStoragePath)
	}
	switch storageTypeCfg.ConfigValue {
	case string(constants.StorageTypeS3):
		cfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3Endpoint)
		if err != nil || cfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3Endpoint)
		}
		cfg, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupS3AccessKey)
		if err != nil || cfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3AccessKey)
		}
		cfg, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupS3SecretAccessKey)
		if err != nil || cfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3SecretAccessKey)
		}
	case string(constants.StorageTypeNFS):
		absPath, err := filepath.Abs(storagePathCfg.ConfigValue)
		if err != nil {
			return fmt.Errorf("backup nfs path %s is not vaild", storagePathCfg.ConfigValue)
		}
		if !mgr.checkFilePathExists(absPath) {
			if err = os.MkdirAll(absPath, os.ModePerm); err != nil {
				return fmt.Errorf("make backup nfs path %s failed, %s", absPath, err.Error())
			}
		}
	default:
		return fmt.Errorf("conifg %s value %s is unknow", constants.ConfigKeyBackupStorageType, storageTypeCfg.ConfigValue)
	}

	if request.ClusterID == "" {
		return fmt.Errorf("empty param clusterID")
	}
	if request.BackupMode != string(constants.BackupModeManual) &&
		request.BackupMode != string(constants.BackupModeAuto) {
		return fmt.Errorf("invalid param backupMode %s", request.BackupMode)
	}

	return nil
}

func (mgr *BRManager) saveBackupStrategyPreCheck(ctx context.Context, request cluster.SaveBackupStrategyReq) error {
	period := strings.Split(request.Strategy.Period, "-")
	if len(period) != 2 {
		return fmt.Errorf("invalid param period, %s", request.Strategy.Period)
	}

	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, err := strconv.Atoi(starts[0])
	if err != nil {
		return fmt.Errorf("invalid param start hour, %s", err.Error())
	}
	endHour, err := strconv.Atoi(ends[0])
	if err != nil {
		return fmt.Errorf("invalid param end hour, %s", err.Error())
	}
	if startHour > 23 || startHour < 0 || endHour > 23 || endHour < 0 || startHour >= endHour {
		return fmt.Errorf("invalid param period, %s", request.Strategy.Period)
	}

	if request.Strategy.BackupDate != "" {
		backupDates := strings.Split(request.Strategy.BackupDate, ",")
		for _, day := range backupDates {
			if !checkWeekDayValid(day) {
				return fmt.Errorf("backupDate contains invalid weekday, %s", day)
			}
		}
	}

	return nil
}

func (mgr *BRManager) getBackupPath(backupPath, clusterId string, time time.Time, backupType string) string {
	return fmt.Sprintf("%s/%s/%s-%s", backupPath, clusterId, time.Format("2006-01-02-15-04-05"), backupType)
}

func (mgr *BRManager) checkFilePathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (mgr *BRManager) removeBackupFiles(ctx context.Context, record *backuprestore.BackupRecord) error {
	if string(constants.StorageTypeS3) == record.StorageType {
		configRW := models.GetConfigReaderWriter()
		endpointCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3Endpoint)
		if err != nil || endpointCfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3Endpoint)
		}
		akCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3AccessKey)
		if err != nil || akCfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3AccessKey)
		}
		skCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3SecretAccessKey)
		if err != nil || skCfg.ConfigValue == "" {
			return fmt.Errorf("get and check conifg %s failed", constants.ConfigKeyBackupS3SecretAccessKey)
		}
		go func() {
			endpoint := strings.TrimPrefix(endpointCfg.ConfigValue, "http://")
			s3Client, err := minio.New(endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(akCfg.ConfigValue, skCfg.ConfigValue, ""),
				Secure: false,
			})
			if err != nil {
				framework.LogWithContext(ctx).Warnf("create s3 client failed: %s", err.Error())
				return
			}
			s3Addr := strings.SplitN(record.FilePath, "/", 2)
			framework.LogWithContext(ctx).Infof("begin remove bucket %s object %s", s3Addr[0], s3Addr[1])
			if err = s3Client.RemoveObject(ctx, s3Addr[0], s3Addr[1], minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
				framework.LogWithContext(ctx).Warnf("remove bucket %s failed: %v", record.FilePath, err)
				return
			}
		}()
	} else {
		go func() {
			if err := os.RemoveAll(record.FilePath); err != nil {
				framework.LogWithContext(ctx).Warnf("remove backup filePath %s, result %v", record.FilePath, err)
				return
			}
		}()
	}
	return nil
}
