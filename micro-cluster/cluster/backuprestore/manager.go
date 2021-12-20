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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
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
	mgr := &BRManager{
		autoBackupMgr: NewAutoBackupManager(),
	}

	flowManager := workflow.GetWorkFlowService()
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowBackupCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowBackupCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"backup", "backupDone", "fail", workflow.PollingNode, backupCluster},
			"backupDone":       {"updateBackupRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateBackupRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, backupFail},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestoreExistCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowRestoreExistCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":       {"restoreFromSrcCluster", "restoreDone", "fail", workflow.PollingNode, restoreFromSrcCluster},
			"restoreDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":        {"fail", "", "", workflow.SyncFuncNode, restoreFail},
		},
	})

	return mgr
}

func (mgr *BRManager) BackupCluster(ctx context.Context, request cluster.BackupClusterDataReq) (resp cluster.BackupClusterDataResp, backupErr error) {
	framework.LogWithContext(ctx).Infof("Begin BackupCluster, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End BackupCluster")

	if err := mgr.backupClusterPreCheck(ctx, request); err != nil {
		framework.LogWithContext(ctx).Errorf("backup cluster precheck failed: %s", err.Error())
		return resp, framework.WrapError(common.TIEM_PARAMETER_INVALID, fmt.Sprintf("backup cluster precheck failed: %s", err.Error()), err)
	}
	configRW := models.GetConfigReaderWriter()
	storageTypeConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStorageType, err.Error())
		return resp, framework.WrapError(common.TIEM_BACKUP_SYSTEM_CONFIG_INVAILD, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyBackupStorageType, err.Error()), err)
	}
	storagePathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStoragePath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStoragePath, err.Error())
		return resp, framework.WrapError(common.TIEM_BACKUP_SYSTEM_CONFIG_INVAILD, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyBackupStoragePath, err.Error()), err)
	}

	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, fmt.Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
	}

	if err := meta.StartMaintenance(ctx, constants.ClusterMaintenanceBackUp); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed, %s", err.Error())
		return resp, fmt.Errorf("start maintenance failed, %s", err.Error())
	}
	defer func() {
		if backupErr != nil {
			if endErr := meta.EndMaintenance(ctx, meta.Cluster.MaintenanceStatus); endErr != nil {
				framework.LogWithContext(ctx).Warnf("end maintenance failed, %s", err.Error())
			}
		}
	}()

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
		return resp, err
	}
	defer func() {
		if backupErr != nil {
			if delErr := brRW.DeleteBackupRecord(ctx, recordCreate.ID); delErr != nil {
				framework.LogWithContext(ctx).Warnf("delete backup record %+v failed, %s", recordCreate, err.Error())
			}
		}
	}()

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, constants.FlowBackupCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
		return resp, fmt.Errorf("create %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
	}

	flowManager.AddContext(flow, contextBackupRecordKey, recordCreate)
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
		return resp, fmt.Errorf("async start %s workflow failed, %s", constants.FlowBackupCluster, err.Error())
	}

	resp.WorkFlowID = flow.Flow.ID
	resp.BackupID = recordCreate.ID

	return resp, nil
}

func (mgr *BRManager) RestoreExistCluster(ctx context.Context, request cluster.RestoreExistClusterReq) (resp cluster.RestoreExistClusterResp, restoreErr error) {
	framework.LogWithContext(ctx).Infof("Begin RestoreExistCluster, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End RestoreExistCluster")

	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, fmt.Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
	}

	brRW := models.GetBRReaderWriter()
	record, err := brRW.GetBackupRecord(ctx, request.BackupID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get backup record %s failed, %s", request.BackupID, err.Error())
		return resp, fmt.Errorf("get backup record %s failed, %s", request.BackupID, err.Error())
	}

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, constants.FlowRestoreExistCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
		return resp, fmt.Errorf("create %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
	}

	flowManager.AddContext(flow, contextBackupRecordKey, record)
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
		return resp, fmt.Errorf("async start %s workflow failed, %s", constants.FlowRestoreExistCluster, err.Error())
	}

	resp.WorkFlowID = flow.Flow.ID
	return resp, nil
}

func (mgr *BRManager) QueryClusterBackupRecords(ctx context.Context, request cluster.QueryBackupRecordsReq) (resp cluster.QueryBackupRecordsResp, page structs.Page, err error) {
	framework.LogWithContext(ctx).Infof("Begin QueryClusterBackupRecords, request: %+v", request)
	defer framework.LogWithContext(ctx).Infof("End QueryClusterBackupRecords")

	brRW := models.GetBRReaderWriter()
	records, total, err := brRW.QueryBackupRecords(ctx, request.ClusterID, request.BackupID, "", request.StartTime.Unix(), request.EndTime.Unix(), request.Page, request.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query cluster backup records %+v failed %s", request, err.Error())
		return resp, page, err
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
			BackupTSO:    record.BackupTso,
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

	brRW := models.GetBRReaderWriter()
	for page, pageSize := 1, 100; ; page++ {
		records, _, err := brRW.QueryBackupRecords(ctx, request.ClusterID, request.BackupID, request.BackupMode, 0, 0, page, pageSize)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("query backup records of request %+v, failed, %s", request, err.Error())
			return resp, fmt.Errorf("query backup records of request %+v, failed, %s", request, err.Error())
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			if string(constants.StorageTypeS3) != record.StorageType {
				filePath := record.FilePath
				err = os.RemoveAll(filePath)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("remove backup filePath %s failed, %s", filePath, err.Error())
					return resp, fmt.Errorf("remove backup filePath %s failed, %s", filePath, err.Error())
				}
			}

			err = brRW.DeleteBackupRecord(ctx, record.ID)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("delete backup record %s failed, %s", record.ID, err.Error())
				return resp, fmt.Errorf("delete backup record %s failed, %s", record.ID, err.Error())
			}
		}
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
		return resp, err
	}

	resp.Strategy = structs.BackupStrategy{
		ClusterID:  strategy.ClusterID,
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
		return resp, err
	}

	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s meta failed, %s", request.ClusterID, err.Error())
		return resp, fmt.Errorf("get cluster %s meta failed, %s", request.ClusterID, err.Error())
	}

	period := strings.Split(request.Strategy.Period, "-")
	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, _ := strconv.Atoi(starts[0])
	endHour, _ := strconv.Atoi(ends[0])

	brRW := models.GetBRReaderWriter()
	strategy, err := brRW.CreateBackupStrategy(ctx, &backuprestore.BackupStrategy{
		Entity: dbModel.Entity{
			TenantId: meta.Cluster.TenantId,
		},
		ClusterID:  request.ClusterID,
		BackupDate: request.Strategy.BackupDate,
		StartHour:  uint32(startHour),
		EndHour:    uint32(endHour),
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create backup strategy %+v failed %s", strategy, err.Error())
		return resp, err
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
		return resp, err
	}

	return resp, nil
}

func (mgr *BRManager) backupClusterPreCheck(ctx context.Context, request cluster.BackupClusterDataReq) error {
	configRW := models.GetConfigReaderWriter()
	storageTypeCfg, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupStorageType)
	if err != nil {
		return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStorageType, err.Error())
	}
	_, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupStoragePath)
	if err != nil {
		return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupStoragePath, err.Error())
	}
	if string(constants.StorageTypeS3) == storageTypeCfg.ConfigValue {
		_, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupS3Endpoint)
		if err != nil {
			return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3Endpoint, err.Error())
		}
		_, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupS3AccessKey)
		if err != nil {
			return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3AccessKey, err.Error())
		}
		_, err = configRW.GetConfig(ctx, constants.ConfigKeyBackupS3SecretAccessKey)
		if err != nil {
			return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3SecretAccessKey, err.Error())
		}
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
	return fmt.Sprintf("%s/%s/%s_%s", backupPath, clusterId, time.Format("2006-01-02_15:04:05"), backupType)
}
