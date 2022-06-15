/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package backuprestore

import (
	"context"
	"fmt"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/backuprestore"
	wfModel "github.com/pingcap/tiunimanager/models/workflow"
	"github.com/pingcap/tiunimanager/util/api/tidb/sql"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	"os"
	"strconv"
	"time"
)

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
)

func backupCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin backupCluster")
	defer framework.LogWithContext(ctx).Info("end backupCluster")

	var record backuprestore.BackupRecord
	var meta meta.ClusterMeta
	err := ctx.GetData(contextBackupRecordKey, &record)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextClusterMetaKey, &meta)
	if err != nil {
		return err
	}

	if string(constants.StorageTypeNFS) == record.StorageType {
		if err := cleanBackupNfsPath(ctx, record.FilePath); err != nil {
			framework.LogWithContext(ctx).Errorf("clean backup nfs path failed, %s", err.Error())
			return fmt.Errorf("clean backup nfs path failed, %s", err.Error())
		}
	}

	tidbAddress := meta.GetClusterConnectAddresses()
	if len(tidbAddress) == 0 {
		framework.LogWithContext(ctx).Errorf("get tidb address from meta failed, empty address")
		return fmt.Errorf("get tidb address from meta failed, empty address")
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb address from meta, %+v", meta.Cluster.ID, tidbAddress)
	tidbServerHost := tidbAddress[0].IP
	tidbServerPort := tidbAddress[0].Port
	node.Record(fmt.Sprintf("get cluster %s tidb address: %s:%d ", meta.Cluster.ID, tidbServerHost, tidbServerPort))

	tidbUserInfo, err := meta.GetDBUserNamePassword(ctx, constants.DBUserBackupRestore)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s user info from meta falied, %s ", meta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(ctx).Infof("get cluster %s user info from meta", meta.Cluster.ID)

	storageType, err := convertBrStorageType(record.StorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert storage type failed, %s", err.Error())
		return err
	}
	node.Record(fmt.Sprintf("convert storage type: %s ", storageType))

	configRW := models.GetConfigReaderWriter()
	rateLimitConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupRateLimit)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get conifg %s failed: %s", constants.ConfigKeyBackupRateLimit, err.Error())
	}
	concurrencyConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupConcurrency)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get conifg %s failed: %s", constants.ConfigKeyBackupConcurrency, err.Error())
	}

	backupSQLReq := sql.BackupSQLReq{
		NodeID:         node.ID,
		DbName:         "", //todo: support db table backup
		TableName:      "",
		StorageAddress: fmt.Sprintf("%s://%s", storageType, getBRStoragePath(ctx, record.StorageType, record.FilePath)),
		DbConnParameter: sql.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: tidbUserInfo.Password.Val,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
	}
	if rateLimitConfig != nil && rateLimitConfig.ConfigValue != "" {
		backupSQLReq.RateLimitM = rateLimitConfig.ConfigValue
	}
	if concurrencyConfig != nil && concurrencyConfig.ConfigValue != "" {
		backupSQLReq.Concurrency = concurrencyConfig.ConfigValue
	}
	framework.LogWithContext(ctx).Infof("begin do backup sql, request[%+v]", backupSQLReq)
	resp, err := sql.ExecBackupSQL(ctx, backupSQLReq, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		return err
	}

	err = ctx.SetData(contextBRInfoKey, &resp)
	if err != nil {
		return err
	}
	node.Record(fmt.Sprintf("backup cluster %s ", meta.Cluster.ID))
	return nil
}

func updateBackupRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin updateBackupRecord")
	defer framework.LogWithContext(ctx).Info("end updateBackupRecord")

	var record backuprestore.BackupRecord
	var meta meta.ClusterMeta
	var brInfo sql.BRSQLResp
	err := ctx.GetData(contextBackupRecordKey, &record)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextClusterMetaKey, &meta)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextBRInfoKey, &brInfo)
	if err != nil {
		return err
	}

	brRW := models.GetBRReaderWriter()
	err = brRW.UpdateBackupRecord(ctx, record.ID, string(constants.ClusterBackupFinished), brInfo.Size, brInfo.BackupTS, time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update backup reocrd %s of cluster %s failed", record.ID, meta.Cluster.ID)
		return err
	}

	node.Record(fmt.Sprintf("update backup record %s of cluster %s ", record.ID, meta.Cluster.ID))
	return nil
}

func restoreFromSrcCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin recoverFromSrcCluster")
	defer framework.LogWithContext(ctx).Info("end recoverFromSrcCluster")

	var record backuprestore.BackupRecord
	var meta meta.ClusterMeta
	err := ctx.GetData(contextBackupRecordKey, &record)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextClusterMetaKey, &meta)
	if err != nil {
		return err
	}

	tidbServers := meta.GetClusterConnectAddresses()
	if len(tidbServers) == 0 {
		framework.LogWithContext(ctx).Error("get tidb servers from meta result empty")
		return fmt.Errorf("get tidb servers from meta result empty")
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb address from meta, %+v", meta.Cluster.ID, tidbServers)
	tidbServerHost := tidbServers[0].IP
	tidbServerPort := tidbServers[0].Port
	node.Record(fmt.Sprintf("get cluster %s tidb address: %s:%d ", meta.Cluster.ID, tidbServerHost, tidbServerPort))

	tidbUserInfo, err := meta.GetDBUserNamePassword(ctx, constants.DBUserBackupRestore)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s user info from meta falied, %s ", meta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(ctx).Infof("get cluster %s user info from meta", meta.Cluster.ID)

	storageType, err := convertBrStorageType(record.StorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert br storage type failed, %s", err.Error())
		return fmt.Errorf("convert br storage type failed, %s", err.Error())
	}
	node.Record(fmt.Sprintf("convert br storage type: %s ", storageType))

	configRW := models.GetConfigReaderWriter()
	rateLimitConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyRestoreRateLimit)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get conifg %s failed: %s", constants.ConfigKeyRestoreRateLimit, err.Error())
	}
	concurrencyConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyRestoreConcurrency)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get conifg %s failed: %s", constants.ConfigKeyRestoreConcurrency, err.Error())
	}

	restoreSQLReq := sql.RestoreSQLReq{
		NodeID:         node.ID,
		DbName:         "", //todo: support db table backup
		TableName:      "",
		StorageAddress: fmt.Sprintf("%s://%s", storageType, getBRStoragePath(ctx, record.StorageType, record.FilePath)),
		DbConnParameter: sql.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: tidbUserInfo.Password.Val,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
	}

	if rateLimitConfig != nil && rateLimitConfig.ConfigValue != "" {
		restoreSQLReq.RateLimitM = rateLimitConfig.ConfigValue
	}
	if concurrencyConfig != nil && concurrencyConfig.ConfigValue != "" {
		restoreSQLReq.Concurrency = concurrencyConfig.ConfigValue
	}
	framework.LogWithContext(ctx).Infof("begin do backup sql, request[%+v]", restoreSQLReq)
	_, err = sql.ExecRestoreSQL(ctx, restoreSQLReq, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		return err
	}

	node.Record(fmt.Sprintf("update backup record %s of cluster %s ", record.ID, meta.Cluster.ID))
	return nil
}

func defaultEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin defaultEnd")
	defer framework.LogWithContext(ctx).Info("end defaultEnd")

	var maintenanceStatusChange bool
	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextMaintenanceStatusChangeKey, &maintenanceStatusChange)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextClusterMetaKey, &clusterMeta)
	if err != nil {
		return err
	}

	if maintenanceStatusChange {
		if err := clusterMeta.EndMaintenance(ctx, clusterMeta.Cluster.MaintenanceStatus); err != nil {
			framework.LogWithContext(ctx).Errorf("end cluster %s maintenance status failed, %s", clusterMeta.Cluster.ID, err.Error())
			return err
		}
	}

	return nil
}

func backupFail(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin backupFail")
	defer framework.LogWithContext(ctx).Info("end backupFail")

	var record backuprestore.BackupRecord
	var meta meta.ClusterMeta
	err := ctx.GetData(contextBackupRecordKey, &record)
	if err != nil {
		return err
	}
	err = ctx.GetData(contextClusterMetaKey, &meta)
	if err != nil {
		return err
	}
	brRW := models.GetBRReaderWriter()
	err = brRW.UpdateBackupRecord(ctx, record.ID, string(constants.ClusterBackupFailed), 0, 0, time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update backup reocrd %s of cluster %s failed", record.ID, meta.Cluster.ID)
		return err
	}

	return defaultEnd(node, ctx)
}

func restoreFail(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin restoreFail")
	defer framework.LogWithContext(ctx).Info("end restoreFail")

	return defaultEnd(node, ctx)
}

func getBRStoragePath(ctx context.Context, storageType string, filePath string) string {
	configRW := models.GetConfigReaderWriter()
	if string(constants.StorageTypeS3) == storageType {
		endpointConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3Endpoint)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3Endpoint, err.Error())
			return ""
		}
		accessKeyConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3AccessKey)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3AccessKey, err.Error())
			return ""
		}
		secretAccessKeyConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyBackupS3SecretAccessKey)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyBackupS3SecretAccessKey, err.Error())
			return ""
		}
		return fmt.Sprintf("%s/?access-key=%s\\&secret-access-key=%s\\&endpoint=%s\\&force-path-style=true",
			filePath, accessKeyConfig.ConfigValue, secretAccessKeyConfig.ConfigValue, endpointConfig.ConfigValue)
	} else {
		return filePath
	}
}

func convertBrStorageType(storageType string) (StorageType, error) {
	if string(constants.StorageTypeS3) == storageType {
		return StorageTypeS3, nil
	} else if string(constants.StorageTypeNFS) == storageType {
		return StorageTypeLocal, nil
	} else {
		return "", fmt.Errorf("invalid storage type, %s", storageType)
	}
}

func cleanBackupNfsPath(ctx context.Context, filepath string) error {
	framework.LogWithContext(ctx).Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		framework.LogWithContext(ctx).Errorf("remove data dir: %s failed %s", filepath, err.Error())
		return err
	}

	if err := os.MkdirAll(filepath, os.ModePerm); err != nil {
		framework.LogWithContext(ctx).Errorf("re-mkdir data dir: %s failed %s", filepath, err.Error())
		return err
	}
	return nil
}
