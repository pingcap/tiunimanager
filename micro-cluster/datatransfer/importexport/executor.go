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
	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"strconv"
	"time"
)

func buildDataImportConfig(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin buildDataImportConfig")
	defer framework.LogWithContext(ctx).Info("end buildDataImportConfig")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	config := NewDataImportConfig(meta, info)
	if config == nil {
		framework.LogWithContext(ctx).Errorf("convert toml config failed, cluster: %s", meta.Cluster.ID)
		node.Fail(fmt.Errorf("convert toml config failed, cluster: %s", meta.Cluster.ID))
		return false
	}

	filePath := fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create import toml config failed, %s", err.Error())
		node.Fail(fmt.Errorf("create import toml config failed, %s", err.Error()))
		return false
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		framework.LogWithContext(ctx).Errorf("encode data import toml config failed, %s", err.Error())
		node.Fail(fmt.Errorf("encode data import toml config failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Infof("build lightning toml config sucess, %v", config)
	node.Success()
	return true
}

func importDataToCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin importDataToCluster")
	defer framework.LogWithContext(ctx).Info("end importDataToCluster")

	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	importTaskId, err := secondparty.Manager.Lightning(ctx, 0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)},
		node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup lightning api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup lightning api failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr tidb-lightning api success, importTaskId %d", importTaskId)
	node.Success()
	return true
}

func updateDataImportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin updateDataImportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataImportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Success()
	return true
}

func exportDataFromCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin exportDataFromCluster")
	defer framework.LogWithContext(ctx).Info("end exportDataFromCluster")

	//meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)

	//todo: get from meta
	tidbHost := ""
	tidbPort := 4000
	if tidbPort == 0 {
		tidbPort = constants.DefaultTiDBPort
	}

	if common.NfsStorageType == info.StorageType {
		if err := cleanDataTransportDir(ctx, info.FilePath); err != nil {
			framework.LogWithContext(ctx).Errorf("clean export directory failed, %s", err.Error())
			node.Fail(fmt.Errorf("clean export directory failed, %s", err.Error()))
			return false
		}
	}

	//tiup dumpling -u root -P 4000 --host 127.0.0.1 --filetype sql -t 8 -o /tmp/test -r 200000 -F 256MiB --filter "user*"
	//todo: replace root password
	cmd := []string{"-u", info.UserName,
		"-p", info.Password,
		"-P", strconv.Itoa(tidbPort),
		"--host", tidbHost,
		"--filetype", info.FileType,
		"-t", "8",
		"-o", info.FilePath,
		"-r", "200000",
		"-F", "256MiB"}
	if info.Filter != "" {
		cmd = append(cmd, "--filter", info.Filter)
	}
	if FileTypeCSV == info.FileType && info.Filter == "" && info.Sql != "" {
		cmd = append(cmd, "--sql", info.Sql)
	}
	if common.S3StorageType == info.StorageType && info.BucketRegion != "" {
		cmd = append(cmd, "--s3.region", fmt.Sprintf("\"%s\"", info.BucketRegion))
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr dumpling api, cmd: %v", cmd)
	exportTaskId, err := secondparty.Manager.Dumpling(ctx, 0, cmd, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup dumpling api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup dumpling api failed, %s", err.Error()))
		return false
	}

	framework.LogWithContext(ctx).Infof("call tiupmgr succee, exportTaskId: %d", exportTaskId)
	node.Success()
	return true
}

func updateDataExportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin updateDataExportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataExportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Success()
	return true
}

func importDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin importDataFailed")
	defer framework.LogWithContext(ctx).Info("end importDataFailed")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)
	if err := updateTransportRecordFailed(ctx, info.RecordId, meta.Cluster.ID); err != nil {
		node.Fail(err)
		return false
	}

	return clusterFail(node, ctx)
}

func exportDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	return clusterFail(node, ctx)
}

func clusterEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	node.Success()
	return true
}

func clusterFail(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	node.Success()
	return true
}

func cleanDataTransportDir(ctx context.Context, filepath string) error {
	framework.LogWithContext(ctx).Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath, os.ModeDir); err != nil {
		return err
	}
	return nil
}

func updateTransportRecordFailed(ctx context.Context, recordId string, clusterId string) error {
	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, recordId, string(constants.DataImportExportFailed), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return err
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	return nil
}
