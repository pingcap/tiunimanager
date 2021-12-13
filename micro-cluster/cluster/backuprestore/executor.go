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
	"strconv"
	"time"
)

func backupCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin backupCluster")
	defer framework.LogWithContext(ctx).Info("end backupCluster")

	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)
	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)

	//todo: get from meta
	tidbServerHost := ""
	tidbServerPort := 0
	if tidbServerPort == 0 {
		tidbServerPort = constants.DefaultTiDBPort
	}

	storageType, err := convertBrStorageType(record.StorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert storage type failed, %s", err.Error())
		return err
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
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
	ctx.SetData("backupTaskId", backupTaskId)
	return nil
}

func updateBackupRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin updateBackupRecord")
	defer framework.LogWithContext(ctx).Info("end updateBackupRecord")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)

	//todo: show backup info and update backupTSO,size,endTime
	size := uint64(0)
	backupTSO := uint64(0)
	brRW := models.GetBRReaderWriter()
	err := brRW.UpdateBackupRecord(ctx, record.ID, string(constants.ClusterBackupFinished), size, backupTSO, time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update backup reocrd %s of cluster %s failed", record.ID, meta.Cluster.ID)
		return err
	}

	return nil
}

func restoreFromSrcCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin recoverFromSrcCluster")
	defer framework.LogWithContext(ctx).Info("end recoverFromSrcCluster")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	record := ctx.GetData(contextBackupRecordKey).(*backuprestore.BackupRecord)

	//todo: get from meta
	tidbServerHost := ""
	tidbServerPort := 0
	storageType, err := convertBrStorageType(record.StorageType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert br storage type failed, %s", err.Error())
		return fmt.Errorf("convert br storage type failed, %s", err.Error())
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
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
	return nil
}

func defaultEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin defaultEnd")
	defer framework.LogWithContext(ctx).Info("end defaultEnd")

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

	//todo: update cluster status

	return nil
}

func restoreFail(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin restoreFail")
	defer framework.LogWithContext(ctx).Info("end restoreFail")

	return nil
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
	} else if string(constants.StorageTypeLocal) == storageType {
		return secondparty.StorageTypeLocal, nil
	} else {
		return "", fmt.Errorf("invalid storage type, %s", storageType)
	}
}
