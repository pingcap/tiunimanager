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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message/cluster"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

var brService BRService

func GetBRService() BRService {
	if brService == nil {
		brService = NewBRManager()
	}
	return brService
}

type BRManager struct {
	autoBackupMgr *autoBackupManager
}

func NewBRManager() *BRManager {
	mgr := &BRManager{
		autoBackupMgr: NewAutoBackupManager(),
	}

	flowManager := workflow.GetWorkFlowService()
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
	node.Success("success")
	return true
}

func updateBackupRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success("success")
	return true
}

func restoreFromSrcCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success("success")
	return true
}

func clusterEnd(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func clusterFail(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}
