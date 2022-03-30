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

package importexport

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func buildDataImportConfig(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin buildDataImportConfig")
	defer framework.LogWithContext(ctx).Info("end buildDataImportConfig")

	meta := ctx.GetData(contextClusterMetaKey).(*meta.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*importInfo)

	config := NewDataImportConfig(ctx, meta, info)
	if config == nil {
		framework.LogWithContext(ctx).Errorf("convert toml config failed, cluster: %s", meta.Cluster.ID)
		return fmt.Errorf("convert toml config failed, cluster: %s", meta.Cluster.ID)
	}

	filePath := fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create import toml config failed, %s", err.Error())
		return fmt.Errorf("create import toml config failed, %s", err.Error())
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		framework.LogWithContext(ctx).Errorf("encode data import toml config failed, %s", err.Error())
		return fmt.Errorf("encode data import toml config failed, %s", err.Error())
	}
	framework.LogWithContext(ctx).Infof("build lightning toml config sucess, %v", config)
	node.Record("build lightning toml config ")
	return nil
}

func importDataToCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin importDataToCluster")
	defer framework.LogWithContext(ctx).Info("end importDataToCluster")

	info := ctx.GetData(contextDataTransportRecordKey).(*importInfo)

	framework.LogWithContext(ctx).Infof("begin do tiup tidb-lightning -config %s/tidb-lightning.toml, timeout: %d", info.ConfigPath, lightningTimeout)
	//tiup tidb-lightning -config tidb-lightning.toml
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	importTaskId, err := deployment.M.Lightning(ctx, tiupHomeForTidb, node.ParentID, []string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)}, lightningTimeout)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup lightning api failed, %s", err.Error())
		return fmt.Errorf("call tiup lightning api failed, %s", err.Error())
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr tidb-lightning api success, importTaskId %s", importTaskId)
	node.Record("import data to cluster ")
	node.OperationID = importTaskId
	return nil
}

func updateDataImportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin updateDataImportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataImportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*importInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return fmt.Errorf("update data transport record failed, %s", err.Error())
	}
	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Record("update data transport record ")
	return nil
}

func exportDataFromCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin exportDataFromCluster")
	defer framework.LogWithContext(ctx).Info("end exportDataFromCluster")

	meta := ctx.GetData(contextClusterMetaKey).(*meta.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*exportInfo)

	tidbServers := meta.GetClusterConnectAddresses()
	if len(tidbServers) == 0 {
		framework.LogWithContext(ctx).Error("get tidb servers from meta result empty")
		return fmt.Errorf("get tidb servers from meta result empty")
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb address from meta, %+v", meta.Cluster.ID, tidbServers)
	tidbHost := tidbServers[0].IP
	tidbPort := tidbServers[0].Port

	if string(constants.StorageTypeNFS) == info.StorageType {
		if err := cleanDataTransportDir(ctx, info.FilePath); err != nil {
			framework.LogWithContext(ctx).Errorf("clean export directory failed, %s", err.Error())
			return fmt.Errorf("clean export directory failed, %s", err.Error())
		}
	}

	configRW := models.GetConfigReaderWriter()
	dumplingThreadNumConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyDumplingThreadNum)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get conifg %s failed: %s", constants.ConfigKeyDumplingThreadNum, err.Error())
	}

	//tiup dumpling -u root -P 4000 --host 127.0.0.1 --filetype sql -t 8 -o /tmp/test -r 200000 -F 256MiB --filter "db.tb"
	cmd := []string{"-u", info.UserName,
		"-p", info.Password,
		"-P", strconv.Itoa(tidbPort),
		"--host", tidbHost,
		"--filetype", info.FileType,
		"-o", info.FilePath,
		"-r", "200000",
		"-F", "256MiB",
		"--loglevel", "debug",
		"--logfile", fmt.Sprintf("%s/dumpling.log", info.ConfigPath)}
	if dumplingThreadNumConfig != nil && dumplingThreadNumConfig.ConfigValue != "" {
		cmd = append(cmd, "-t", dumplingThreadNumConfig.ConfigValue)
	} else {
		cmd = append(cmd, "-t", "8")
	}
	if info.Filter != "" {
		cmd = append(cmd, "--filter", info.Filter)
	}
	if fileTypeCSV == info.FileType && info.Filter == "" && info.Sql != "" {
		cmd = append(cmd, "--sql", info.Sql)
	}
	//framework.LogWithContext(ctx).Infof("call tiupmgr dumpling api, timeout: %d", dumplingTimeout)
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	exportTaskId, err := deployment.M.Dumpling(ctx, tiupHomeForTidb, node.ParentID, cmd, dumplingTimeout)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup dumpling api failed, %s", err.Error())
		return fmt.Errorf("call tiup dumpling api failed, %s", err.Error())
	}

	framework.LogWithContext(ctx).Infof("call tiupmgr succee, exportTaskId: %s", exportTaskId)
	node.Record(fmt.Sprintf("export data from cluster %s ", meta.Cluster.ID),
		fmt.Sprintf("host: %s, port: %d ", tidbHost, tidbPort))
	node.OperationID = exportTaskId
	return nil
}

func updateDataExportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin updateDataExportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataExportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*exportInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return fmt.Errorf("update data transport record failed, %s", err.Error())
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Record("update data transport record ")
	return nil
}

func importDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin importDataFailed")
	defer framework.LogWithContext(ctx).Info("end importDataFailed")

	info := ctx.GetData(contextDataTransportRecordKey).(*importInfo)
	if err := updateTransportRecordFailed(ctx, info.RecordId); err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return fmt.Errorf("update data transport record failed, %s", err.Error())
	}

	return nil
}

func exportDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin exportDataFailed")
	defer framework.LogWithContext(ctx).Info("end exportDataFailed")

	info := ctx.GetData(contextDataTransportRecordKey).(*exportInfo)
	if err := updateTransportRecordFailed(ctx, info.RecordId); err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return fmt.Errorf("update data transport record failed, %s", err.Error())
	}

	return nil
}

func defaultEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return nil
}

func cleanDataTransportDir(ctx context.Context, filepath string) error {
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

func updateTransportRecordFailed(ctx context.Context, recordId string) error {
	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, recordId, string(constants.DataImportExportFailed), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return err
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	return nil
}
