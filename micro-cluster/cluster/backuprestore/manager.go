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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-metadb/service"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"strconv"
	"time"
)

type BRManager struct {
	autoBackupMgr *autoBackupManager
}

var manager *BRManager

func GetBRManager() *BRManager {
	if manager == nil {
		manager = NewBRManager()
	}
	return manager
}

func NewBRManager() *BRManager {
	mgr := &BRManager{
		autoBackupMgr: NewAutoBackupManager(),
	}

	flowManager := workflow.GetWorkFlowManager()
	flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowBackupCluster, &workflow.WorkFlowDefine{
		FlowName: constants.WorkFlowBackupCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"backup", "backupDone", "fail", workflow.PollingNode, backupCluster},
			"backupDone":       {"updateBackupRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateBackupRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, clusterEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, clusterFail},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), "ExportData", &workflow.WorkFlowDefine{
		FlowName: constants.WorkFlowRestoreExistCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":       {"restoreFromSrcCluster", "restoreDone", "fail", workflow.PollingNode, restoreFromSrcCluster},
			"restoreDone": {"end", "", "", workflow.SyncFuncNode, clusterEnd},
			"fail":        {"fail", "", "", workflow.SyncFuncNode, clusterFail},
		},
	})

	return mgr
}

func (mgr *BRManager) BackupCluster(ctx context.Context, request *cluster.BackupClusterDataReq) (*cluster.BackupClusterDataResp, error) {
	return nil, nil
}

func (mgr *BRManager) RestoreExistCluster(ctx context.Context, request *cluster.RestoreExistClusterReq) (*cluster.RestoreExistClusterResp, error) {
	return nil, nil
}

func (mgr *BRManager) QueryClusterBackupRecords(ctx context.Context, request *cluster.QueryBackupRecordsReq) (*cluster.QueryBackupRecordsResp, error) {
	return nil, nil
}

func (mgr *BRManager) DeleteBackupRecords(ctx context.Context, request *cluster.DeleteBackupDataReq) (*cluster.DeleteBackupDataResp, error) {
	return nil, nil
}

func (mgr *BRManager) QueryBackupStrategy(ctx context.Context, request *cluster.QueryBackupStrategyReq) (*cluster.QueryBackupStrategyResp, error) {
	return nil, nil
}

func (mgr *BRManager) SaveBackupStrategy(ctx context.Context, request *cluster.UpdateBackupStrategyReq) (*cluster.UpdateBackupStrategyResp, error) {
	return nil, nil
}

func (mgr *BRManager) DeleteBackupStrategy(ctx context.Context, request *cluster.DeleteBackupStrategyReq) (*cluster.DeleteBackupStrategyResp, error) {
	return nil, nil
}

func backupCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin backupCluster")
	defer framework.LogWithContext(ctx).Info("end backupCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	record := clusterAggregation.LastBackupRecord
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]
	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = common.DefaultTidbPort
	}

	storageType, err := convertBrStorageType(string(record.StorageType))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert storage type failed, %s", err.Error())
		node.Fail(err)
		return false
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			IP:       tidbServer.Host,
			Port:     strconv.Itoa(tidbServerPort),
		},
		DbName:      "", //todo: support db table backup
		TableName:   "",
		ClusterId:   cluster.Id,
		ClusterName: cluster.ClusterName,
		TaskID:      uint64(task.Id),
	}
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.FilePath, "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}

	framework.LogWithContext(ctx).Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)

	backupTaskId, err := secondparty.SecondParty.MicroSrvBackUp(ctx, clusterFacade, storage, uint64(node.ID))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		node.Fail(err)
		return false
	}
	flowContext.SetData("backupTaskId", backupTaskId)
	node.Success("success")
	return true
}

func updateBackupRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin updateBackupRecord")
	defer framework.LogWithContext(ctx).Info("end updateBackupRecord")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	record := clusterAggregation.LastBackupRecord

	var req dbpb.FindTiupTaskByIDRequest
	var resp *dbpb.FindTiupTaskByIDResponse
	var err error
	req.Id = flowContext.GetData("backupTaskId").(uint64)

	resp, err = client.DBClient.FindTiupTaskByID(flowContext, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get backup task err = %s", err.Error())
		node.Fail(err)
		return false
	}
	if resp.TiupTask.Status == dbpb.TiupTaskStatus_Error {
		framework.LogWithContext(ctx).Errorf("backup cluster error, %s", resp.TiupTask.ErrorStr)
		node.Fail(errors.New(resp.TiupTask.ErrorStr))
		return false
	}

	var backupInfo secondparty.CmdBrResp
	err = json.Unmarshal([]byte(resp.GetTiupTask().GetErrorStr()), &backupInfo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("json unmarshal backup info resp: %+v, failed, %s", resp, err.Error())
	} else {
		record.Size = backupInfo.Size
		record.BackupTso = backupInfo.BackupTS
	}

	updateResp, err := client.DBClient.UpdateBackupRecord(flowContext, &dbpb.DBUpdateBackupRecordRequest{
		BackupRecord: &dbpb.DBBackupRecordDTO{
			Id:        record.Id,
			Size:      record.Size,
			BackupTso: record.BackupTso,
			EndTime:   time.Now().Unix(),
		},
	})
	if err != nil {
		msg := fmt.Sprintf("update backup record for cluster %s failed", clusterAggregation.Cluster.Id)
		tiemError := framework.WrapError(common.TIEM_METADB_SERVER_CALL_ERROR, msg, err)
		framework.LogWithContext(ctx).Error(tiemError)
		node.Fail(tiemError)
		return false
	}
	if updateResp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		msg := fmt.Sprintf("update backup record for cluster %s failed, %s", clusterAggregation.Cluster.Id, updateResp.Status.Message)
		tiemError := framework.NewTiEMError(common.TIEM_ERROR_CODE(updateResp.GetStatus().GetCode()), msg)

		framework.LogWithContext(ctx).Error(tiemError)
		node.Fail(tiemError)
		return false
	}
	node.Success("success")
	return true
}

func restoreFromSrcCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin recoverFromSrcCluster")
	defer framework.LogWithContext(ctx).Info("end recoverFromSrcCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	recoverInfo := cluster.RecoverInfo
	if recoverInfo.SourceClusterId == "" || recoverInfo.BackupRecordId <= 0 {
		framework.LogWithContext(ctx).Infof("cluster %s no need recover", cluster.Id)
		node.Success("success, cluster no need recover")
		return true
	}

	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	record, err := client.DBClient.QueryBackupRecords(flowContext, &dbpb.DBQueryBackupRecordRequest{ClusterId: recoverInfo.SourceClusterId, RecordId: recoverInfo.BackupRecordId})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query backup record failed, %s", err.Error())
		node.Fail(fmt.Errorf("query backup record failed, %s", err.Error()))
		return false
	}
	if record.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		framework.LogWithContext(ctx).Errorf("query backup record failed, %s", record.GetStatus().GetMessage())
		node.Fail(fmt.Errorf("query backup record failed, %s", record.GetStatus().GetMessage()))
		return false
	}

	storageType, err := convertBrStorageType(record.GetBackupRecords().GetBackupRecord().GetStorageType())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("convert br storage type failed, %s", err.Error())
		node.Fail(fmt.Errorf("convert br storage type failed, %s", err.Error()))
		return false
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			IP:       tidbServer.Host,
			Port:     strconv.Itoa(tidbServer.Port),
		},
		DbName:      "", //todo: support db table restore
		TableName:   "",
		ClusterId:   cluster.Id,
		ClusterName: cluster.ClusterName,
	}
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.GetBackupRecords().GetBackupRecord().GetFilePath(), "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}
	framework.LogWithContext(ctx).Infof("begin call brmgr restore api, clusterFacade %v, storage %v", clusterFacade, storage)

	_, err = secondparty.SecondParty.MicroSrvRestore(ctx, clusterFacade, storage, uint64(task.Id))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call restore api failed, %s", err.Error())
		node.Fail(err)
		return false
	}
	node.Success("success")
	return true
}

func clusterEnd(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true

	return true
}

func clusterFail(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Status = constants.WorkFlowStatusError
	node.Result = "fail"
	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true
	return true
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
