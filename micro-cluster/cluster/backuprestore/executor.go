/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"strconv"
	"time"
)

func backupCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin backupCluster")
	defer framework.LogWithContext(ctx).Info("end backupCluster")

	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)
	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)

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

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: tidbUserInfo.Password,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
		DbName:      "", //todo: support db table backup
		TableName:   "",
		ClusterId:   meta.Cluster.ID,
		ClusterName: meta.Cluster.Name,
	}
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        getBRStoragePath(ctx, record.StorageType, record.FilePath),
		//Root:        fmt.Sprintf("%s/%s", record.FilePath, "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"),
	}

	framework.LogWithContext(ctx).Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)

	backupTaskId, err := secondparty.Manager.BackUp(ctx, clusterFacade, storage, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		return err
	}
	ctx.SetData(contextBackupTiupTaskIDKey, backupTaskId)
	node.Record(fmt.Sprintf("backup cluster %s ", meta.Cluster.ID))
	return nil
}

func updateBackupRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin updateBackupRecord")
	defer framework.LogWithContext(ctx).Info("end updateBackupRecord")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)
	backupTaskId := ctx.GetData(contextBackupTiupTaskIDKey).(string)

	resp, err := secondparty.Manager.ShowBackUpInfoThruMetaDB(ctx, backupTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("show backup info of backupId %s failed", backupTaskId)
		return err
	}

	brRW := models.GetBRReaderWriter()
	err = brRW.UpdateBackupRecord(ctx, record.ID, string(constants.ClusterBackupFinished), resp.Size, resp.BackupTS, time.Now())
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

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)

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

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: tidbUserInfo.Password,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
		DbName:      "", //todo: support db table restore
		TableName:   "",
		ClusterId:   meta.Cluster.ID,
		ClusterName: meta.Cluster.Name,
	}
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        getBRStoragePath(ctx, record.StorageType, record.FilePath),
		//Root:        fmt.Sprintf("%s/%s", record.FilePath, "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"),
	}
	framework.LogWithContext(ctx).Infof("begin call brmgr restore api, clusterFacade %v, storage %v", clusterFacade, storage)
	_, err = secondparty.Manager.Restore(ctx, clusterFacade, storage, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call restore api failed, %s", err.Error())
		return err
	}
	node.Record(fmt.Sprintf("update backup record %s of cluster %s ", record.ID, meta.Cluster.ID))
	return nil
}

func defaultEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin defaultEnd")
	defer framework.LogWithContext(ctx).Info("end defaultEnd")

	clusterMeta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	maintenanceStatusChange := ctx.GetData(contextMaintenanceStatusChangeKey).(bool)
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

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)

	brRW := models.GetBRReaderWriter()
	err := brRW.UpdateBackupRecord(ctx, record.ID, string(constants.ClusterBackupFailed), 0, 0, time.Now())
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

func convertBrStorageType(storageType string) (secondparty.StorageType, error) {
	if string(constants.StorageTypeS3) == storageType {
		return secondparty.StorageTypeS3, nil
	} else if string(constants.StorageTypeNFS) == storageType {
		return secondparty.StorageTypeLocal, nil
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
