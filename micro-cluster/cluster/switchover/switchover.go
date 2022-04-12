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

/*******************************************************************************
 * @File: switchover.go
 * @Description: switchover implementation
 * @Author: hansen@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/15 11:30
*******************************************************************************/

package switchover

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	emerr "github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	tsoLib "github.com/pingcap-inc/tiem/library/util/tso"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	changefeedMgr "github.com/pingcap-inc/tiem/micro-cluster/cluster/changefeed"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	changefeedModel "github.com/pingcap-inc/tiem/models/cluster/changefeed"
	clusterMgr "github.com/pingcap-inc/tiem/models/cluster/management"
	crypto "github.com/pingcap-inc/tiem/util/encrypt"
	"github.com/pingcap-inc/tiem/workflow"
)

type Manager struct {
	changefeedMgr CDCManagerAPI
}

func newManager() *Manager {
	return &Manager{}
}

var mgr = newManager()
var mgrOnceRegisterWorkFlow sync.Once

// do not include rollback workflow
var switchoverWorkflowNameMap = map[string]bool{
	constants.FlowMasterSlaveSwitchoverNormal:                     true,
	constants.FlowMasterSlaveSwitchoverForce:                      true,
	constants.FlowMasterSlaveSwitchoverForceWithMasterUnavailable: true,
}

func GetManager() *Manager {
	mgrOnceRegisterWorkFlow.Do(func() {
		mgr.changefeedMgr = changefeedMgr.GetManager()
		flowManager := workflow.GetWorkFlowService()
		flowManager.RegisterWorkFlow(context.TODO(), constants.FlowMasterSlaveSwitchoverNormal, &workflow.WorkFlowDefine{
			FlowName: constants.FlowMasterSlaveSwitchoverNormal,
			TaskNodes: map[string]*workflow.NodeDefine{
				"start": {
					"marshalSwitchoverMasterSlavesState", "checkHealthStatus", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepMarshalSwitchoverMasterSlavesState, wfStepFail)},
				"checkHealthStatus": {
					"checkHealthStatus", "checkSyncChangeFeedTaskMaxLagTime", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCheckOldSyncChangeFeedTaskHealth, wfStepFail)},
				"checkSyncChangeFeedTaskMaxLagTime": {
					"checkSyncChangeFeedTaskMaxLagTime", "setOldMasterReadOnly", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCheckSyncChangeFeedTaskMaxLagTime, wfStepFail)},
				"setOldMasterReadOnly": {
					"setOldMasterReadOnly", "waitOldMasterCDCsCaughtUp", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetOldMasterReadOnly, wfStepFail)},
				"waitOldMasterCDCsCaughtUp": {
					"waitOldMasterCDCsCaughtUp", "pauseOldSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepWaitOldMasterCDCsCaughtUp, wfStepFail)},
				"pauseOldSyncChangeFeedTask": {
					"pauseOldSyncChangeFeedTask", "createReverseSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepPauseOldSyncChangeFeedTask, wfStepFail)},
				"createReverseSyncChangeFeedTask": {
					"createReverseSyncChangeFeedTask", "checkNewSyncChangeFeedTaskHealth", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCreateReverseSyncChangeFeedTask, wfStepFail)},
				"checkNewSyncChangeFeedTaskHealth": {
					"checkNewSyncChangeFeedTaskHealth", "migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCheckNewSyncChangeFeedTaskHealth, wfStepFail)},
				"migrateAllDownStreamSyncChangeFeedTasksToNewMaster": {
					"migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster, wfStepFail)},
				"setNewMasterReadWrite": {
					"setNewMasterReadWrite", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetNewMasterReadWrite, wfStepFail)},
				"swapMasterSlaveRelationInDB": {
					"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSwapMasterSlaveRelationInDB, wfStepFail)},
				"end": {
					"end", "", "", workflow.SyncFuncNode, wfStepFinish},
				"fail": {
					"fail", "", "", workflow.SyncFuncNode, wfStepNOP},
			},
		})
		flowManager.RegisterWorkFlow(context.TODO(), constants.FlowMasterSlaveSwitchoverForce, &workflow.WorkFlowDefine{
			FlowName: constants.FlowMasterSlaveSwitchoverForce,
			TaskNodes: map[string]*workflow.NodeDefine{
				"start": {
					"marshalSwitchoverMasterSlavesState", "setOldMasterReadOnly", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepMarshalSwitchoverMasterSlavesState, wfStepFail)},
				"setOldMasterReadOnly": {
					"setOldMasterReadOnly", "pauseOldSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetOldMasterReadOnly, wfStepFail)},
				"pauseOldSyncChangeFeedTask": {
					"pauseOldSyncChangeFeedTask", "createReverseSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepPauseOldSyncChangeFeedTask, wfStepFail)},
				"createReverseSyncChangeFeedTask": {
					"createReverseSyncChangeFeedTask", "checkNewSyncChangeFeedTaskHealth", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCreateReverseSyncChangeFeedTask, wfStepFail)},
				"checkNewSyncChangeFeedTaskHealth": {
					"checkNewSyncChangeFeedTaskHealth", "migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepCheckNewSyncChangeFeedTaskHealth, wfStepFail)},
				"migrateAllDownStreamSyncChangeFeedTasksToNewMaster": {
					"migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster, wfStepFail)},
				"setNewMasterReadWrite": {
					"setNewMasterReadWrite", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetNewMasterReadWrite, wfStepFail)},
				"swapMasterSlaveRelationInDB": {
					"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSwapMasterSlaveRelationInDB, wfStepFail)},
				"end": {
					"end", "", "", workflow.SyncFuncNode, wfStepFinish},
				"fail": {
					"fail", "", "", workflow.SyncFuncNode, wfStepNOP},
			},
		})
		flowManager.RegisterWorkFlow(context.TODO(), constants.FlowMasterSlaveSwitchoverForceWithMasterUnavailable,
			&workflow.WorkFlowDefine{
				FlowName: constants.FlowMasterSlaveSwitchoverForceWithMasterUnavailable,
				TaskNodes: map[string]*workflow.NodeDefine{
					"start": {
						"marshalSwitchoverMasterSlavesState", "setOldMasterReadOnly", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepMarshalSwitchoverMasterSlavesState, wfStepFail)},
					"setOldMasterReadOnly": {
						"setOldMasterReadOnly", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetOldMasterReadOnly, wfStepFail)},
					"setNewMasterReadWrite": {
						"setNewMasterReadWrite", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSetNewMasterReadWrite, wfStepFail)},
					"swapMasterSlaveRelationInDB": {
						"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfGenStepWithRollbackCB(wfStepSwapMasterSlaveRelationInDB, wfStepFail)},
					"end": {
						"end", "", "", workflow.SyncFuncNode, wfStepFinish},
					"fail": {
						"fail", "", "", workflow.SyncFuncNode, wfStepNOP},
				},
			})
		flowManager.RegisterWorkFlow(context.TODO(), constants.FlowMasterSlaveSwitchoverRollback,
			&workflow.WorkFlowDefine{
				FlowName: constants.FlowMasterSlaveSwitchoverRollback,
				TaskNodes: map[string]*workflow.NodeDefine{
					"start": {
						"rollback", "end", "fail", workflow.SyncFuncNode, wfStepRollback},
					"end": {
						"end", "", "", workflow.SyncFuncNode, wfStepNOP},
					"fail": {
						"fail", "", "", workflow.SyncFuncNode, wfStepNOP},
				},
			})
	})
	return mgr
}

// Switchover switchover between master and slave cluster
func (p *Manager) Switchover(ctx context.Context, req *cluster.MasterSlaveClusterSwitchoverReq) (resp *cluster.MasterSlaveClusterSwitchoverResp, err error) {
	framework.LogWithContext(ctx).Info("Manager.Switchover")

	if len(req.RollbackWorkFlowID) > 0 {
		resp, err = p.switchoverRollback(ctx, req)
		return
	}

	if req.RollbackClearPreviousMaintenanceFlag {
		resp, err = p.switchoverEndMaintenance(ctx, req)
		return
	}

	flowName := constants.FlowMasterSlaveSwitchoverNormal
	reqJsonBs, err := json.Marshal(req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("req marshal to json failed err:%s", err)
		return resp, err
	}
	reqJson := string(reqJsonBs)
	funcName := "Switchover"
	oldMasterId := req.SourceClusterID
	oldSlaveId := req.TargetClusterID
	if req.CheckStandaloneClusterFlag {
		err := mgr.clusterCheckNoRelation(ctx, oldMasterId)
		framework.LogWithContext(ctx).Infof("req:%s clusterCheckNoRelation err:%v", reqJson, err)
		if err != nil {
			return resp, err
		}
		err = mgr.clusterCheckNoCDCs(ctx, oldMasterId)
		framework.LogWithContext(ctx).Infof("req:%s clusterCheckNoCDCs err:%v", reqJson, err)
		if err != nil {
			return resp, err
		}
		return resp, err
	}
	// pre check
	//   1. cluster relation is valid?
	//   2. cdc sync task is valid?
	//   3. slaveToBeNewMasterCluster has CDC component?
	oldSyncChangeFeedTaskId, err := mgr.getOldSyncChangeFeedTaskId(ctx, reqJson, funcName, oldMasterId, oldSlaveId)
	if err != nil {
		return resp, err
	}
	if len(oldSyncChangeFeedTaskId) <= 0 {
		return resp, emerr.Error(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_CDC_SYNC_TASK_NOT_FOUND)
	}
	otherSlavesMapToOldSyncCDCTask, err := mgr.clusterGetOtherSlavesMapToOldSyncCDCTask(ctx, oldMasterId, oldSlaveId)
	if err != nil {
		return resp, err
	}
	otherSlavesMapToOldSyncCDCTaskBs, err := json.Marshal(&otherSlavesMapToOldSyncCDCTask)
	if err != nil {
		return resp, fmt.Errorf("marshal otherSlavesMapToOldSyncCDCTask failed, err: %s", err)
	}
	otherSlavesMapToOldSyncCDCTaskStr := string(otherSlavesMapToOldSyncCDCTaskBs)
	err = mgr.clusterCheckHasCDCComponent(ctx, oldSlaveId, emerr.Error(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_SLAVE_NO_CDC_COMPONENT))
	if err != nil {
		return resp, err
	}
	if req.CheckSlaveReadOnlyFlag {
		readonlyFlag, err := mgr.clusterGetReadWriteMode(ctx, oldSlaveId)
		framework.LogWithContext(ctx).Infof("mgr.clusterGetReadWriteMode on oldSlaveId %s, ret readonlyFlag:%v err:%v",
			oldSlaveId, readonlyFlag, err)
		if err != nil {
			return resp, fmt.Errorf("mgr.clusterGetReadWriteMode failed, err: %s", err)
		}
		if readonlyFlag {
		} else {
			return resp, fmt.Errorf("oldSlave readonlyFlag:%v but expect readonly", readonlyFlag)
		}
	}
	if req.CheckMasterWritableFlag {
		readonlyFlag, err := mgr.clusterGetReadWriteMode(ctx, oldMasterId)
		framework.LogWithContext(ctx).Infof("mgr.clusterGetReadWriteMode on oldMasterId %s, ret readonlyFlag:%v err:%v",
			oldMasterId, readonlyFlag, err)
		if err != nil {
			return resp, fmt.Errorf("mgr.clusterGetReadWriteMode failed, err: %s", err)
		}
		if readonlyFlag {
			return resp, fmt.Errorf("oldMaster readonlyFlag:%v but expect writable", readonlyFlag)
		}
	}
	if req.OnlyCheck {
		return &cluster.MasterSlaveClusterSwitchoverResp{}, nil
	}
	// workflow switch
	if req.Force { // A -> B
		err = mgr.checkClusterReadWriteHealth(ctx, oldMasterId)
		if err == nil { // A rw-able
			framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldMasterId %s success", oldMasterId)
			err = mgr.checkSyncChangeFeedTaskHealth(ctx, reqJson, funcName, oldSyncChangeFeedTaskId)
			if err == nil {
				// A&B rw-able
				framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldSlaveId %s success", oldSlaveId)
				flowName = constants.FlowMasterSlaveSwitchoverForce
			} else {
				// A rw-able & B unavailable
				framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldSlaveId %s failed err:%s", oldSlaveId, err)
				flowName = "" // end
				return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED, "master/slave switchover failed: %s", "slave is unavailable")
			}
		} else {
			framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldMasterId %s failed err:%s", oldMasterId, err)
			err = mgr.checkClusterReadWriteHealth(ctx, oldSlaveId)
			if err == nil {
				// A unavailable & B rw-able
				framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldSlaveId %s success", oldSlaveId)
				flowName = constants.FlowMasterSlaveSwitchoverForceWithMasterUnavailable
			} else {
				// A unavailable & B unavailable
				framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldSlaveId %s failed err:%s", oldSlaveId, err)
				flowName = "" // end
				return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED, "master/slave switchover failed: %s", "master and slave are both unavailable")
			}
		}
	}
	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, oldMasterId, workflow.BizTypeCluster, flowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", flowName, err.Error())
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"create %s workflow failed: %s", flowName, err.Error())
	}

	flowManager.AddContext(flow, wfContextReqKey, req)
	flowManager.AddContext(flow, wfContextOldSyncChangeFeedTaskIDKey, oldSyncChangeFeedTaskId)
	flowManager.AddContext(flow, wfContextOtherSlavesMapToOldSyncCDCTaskKey, otherSlavesMapToOldSyncCDCTaskStr)

	var cancelFps []func()
	cancelFps = append(cancelFps, func() {
		flowManager.Destroy(ctx, flow, "start maintenance failed")
	})
	cancelFlag := true
	cancel := func() {
		if cancelFlag {
			for _, fp := range cancelFps {
				fp()
			}
		}
	}
	defer cancel()
	metaOfSource, err := meta.Get(ctx, req.SourceClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", req.SourceClusterID, err.Error())
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"get meta of %s failed, %s", req.SourceClusterID, err.Error())
	}

	if err := metaOfSource.StartMaintenance(ctx, constants.ClusterMaintenanceSwitching); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed:%s", err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
	}
	flowManager.AddContext(flow, wfContextOldMasterPreviousMaintenanceStatusKey, string(metaOfSource.Cluster.MaintenanceStatus))
	cancelFps = append(cancelFps, func() {
		err := metaOfSource.EndMaintenance(ctx, metaOfSource.Cluster.MaintenanceStatus)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", req.SourceClusterID, err.Error())
		}
	})

	metaOfTarget, err := meta.Get(ctx, req.TargetClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", req.TargetClusterID, err.Error())
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"get meta of cluster %s failed, %s", req.TargetClusterID, err.Error())
	}

	if err := metaOfTarget.StartMaintenance(ctx, constants.ClusterMaintenanceSwitching); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed:%s", err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
	}
	flowManager.AddContext(flow, wfContextOldSlavePreviousMaintenanceStatusKey, string(metaOfTarget.Cluster.MaintenanceStatus))
	cancelFps = append(cancelFps, func() {
		err := metaOfTarget.EndMaintenance(ctx, metaOfTarget.Cluster.MaintenanceStatus)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", req.TargetClusterID, err.Error())
		}
	})

	for otherSlaveClusterID := range otherSlavesMapToOldSyncCDCTask {
		metaOfOtherSlave, err := meta.Get(ctx, otherSlaveClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", otherSlaveClusterID, err.Error())
			return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
				"get meta of cluster %s failed, %s", otherSlaveClusterID, err.Error())
		}

		if err := metaOfOtherSlave.StartMaintenance(ctx, constants.ClusterMaintenanceSwitching); err != nil {
			framework.LogWithContext(ctx).Errorf("start maintenance failed:%s", err.Error())
			return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
		}
		//flowManager.AddContext(flow, wfContextOldSlavePreviousMaintenanceStatusKey, string(metaOfOtherSlave.Cluster.MaintenanceStatus))
		thisSlaveID := otherSlaveClusterID
		cancelFps = append(cancelFps, func() {
			err := metaOfOtherSlave.EndMaintenance(ctx, constants.ClusterMaintenanceSwitching)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", thisSlaveID, err.Error())
			}
		})
	}

	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", flowName, err.Error())
		return nil, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"async start %s workflow failed, %s", flowName, err.Error())
	}

	cancelFlag = false

	return &cluster.MasterSlaveClusterSwitchoverResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
	}, nil
}

func (p *Manager) switchoverEndMaintenance(ctx context.Context, req *cluster.MasterSlaveClusterSwitchoverReq) (resp *cluster.MasterSlaveClusterSwitchoverResp, err error) {
	framework.LogWithContext(ctx).Info("Manager.switchoverEndMaintenance")

	if err != nil {
		framework.LogWithContext(ctx).Errorf("req marshal to json failed err:%s", err)
		return resp, err
	}
	funcName := "switchoverEndMaintenance"
	oldMasterId := req.SourceClusterID
	framework.LogWithContext(ctx).Infof("%s OldMasterClusterID %s", funcName, oldMasterId)

	slaveRels, err := mgr.clusterGetRelationsByMasterClusterId(ctx, oldMasterId)
	if err != nil {
		return resp, err
	}
	var slaveIDs []string
	for _, rel := range slaveRels {
		slaveIDs = append(slaveIDs, rel.ObjectClusterID)
	}

	metaOfOldMaster, err := meta.Get(ctx, oldMasterId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", oldMasterId, err.Error())
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
			"get meta of %s failed, %s", oldMasterId, err.Error())
	}
	err = metaOfOldMaster.EndMaintenance(ctx, constants.ClusterMaintenanceSwitching)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("rollbackClearPreviousMaintenance end maintenance of cluster %s failed:%s", oldMasterId, err.Error())
		return resp, err
	}

	for _, slaveClusterID := range slaveIDs {
		metaOfSlave, err := meta.Get(ctx, slaveClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", slaveClusterID, err.Error())
			return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
				"get meta of cluster %s failed, %s", slaveClusterID, err.Error())
		}
		err = metaOfSlave.EndMaintenance(ctx, constants.ClusterMaintenanceSwitching)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("rollbackClearPreviousMaintenance end maintenance of cluster %s failed:%s", slaveClusterID, err.Error())
			return resp, err
		}
	}

	return &cluster.MasterSlaveClusterSwitchoverResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: "",
		},
	}, nil
}

func (p *Manager) switchoverRollback(ctx context.Context, req *cluster.MasterSlaveClusterSwitchoverReq) (resp *cluster.MasterSlaveClusterSwitchoverResp, err error) {
	framework.LogWithContext(ctx).Info("Manager.switchoverRollback")

	flowName := constants.FlowMasterSlaveSwitchoverRollback
	reqJsonBs, err := json.Marshal(req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("req marshal to json failed err:%s", err)
		return resp, err
	}
	reqJson := string(reqJsonBs)
	funcName := "switchoverRollback"
	if req.CheckStandaloneClusterFlag || req.OnlyCheck || req.CheckSlaveReadOnlyFlag || req.CheckMasterWritableFlag {
		return resp, fmt.Errorf("checkFlags should not be set if rollbackWorkFlowID is valid")
	}
	s, _, err := mgr.getSwitchoverMasterSlavesStateFromASwitchoverWorkflow(ctx, req.RollbackWorkFlowID)
	if err != nil {
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
			"getSwitchoverMasterSlavesStateFromAFailedWorkflow failed: %s", err.Error())
	}
	oldMasterId := s.OldMasterClusterID
	framework.LogWithContext(ctx).Infof("%s OldMasterClusterID %s", funcName, oldMasterId)
	for slaveID := range s.OldSlavesClusterIDMapToSyncTaskID {
		framework.LogWithContext(ctx).Infof("%s iterate OldSlavesClusterID %s", funcName, slaveID)
	}
	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, oldMasterId, workflow.BizTypeCluster, flowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("%s create %s workflow failed, err:%s, req:%s", funcName, flowName, err.Error(), reqJson)
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
			"create %s workflow failed: %s", flowName, err.Error())
	}

	flowManager.AddContext(flow, wfContextReqKey, req)

	var cancelFps []func()
	cancelFps = append(cancelFps, func() {
		flowManager.Destroy(ctx, flow, "start maintenance failed")
	})
	cancelFlag := true
	cancel := func() {
		if cancelFlag {
			for _, fp := range cancelFps {
				fp()
			}
		}
	}
	defer cancel()
	metaOfOldMaster, err := meta.Get(ctx, oldMasterId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", oldMasterId, err.Error())
		return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
			"get meta of %s failed, %s", oldMasterId, err.Error())
	}
	if req.RollbackClearPreviousMaintenanceFlag {
		err := metaOfOldMaster.EndMaintenance(ctx, constants.ClusterMaintenanceSwitching)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("rollbackClearPreviousMaintenance end maintenance of cluster %s failed:%s", oldMasterId, err.Error())
			return resp, err
		}
	}
	if err := metaOfOldMaster.StartMaintenance(ctx, constants.ClusterMaintenanceSwitchoverRollback); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed:%s", err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
	}
	flowManager.AddContext(flow, wfContextOldMasterPreviousMaintenanceStatusKey, string(metaOfOldMaster.Cluster.MaintenanceStatus))
	cancelFps = append(cancelFps, func() {
		err := metaOfOldMaster.EndMaintenance(ctx, metaOfOldMaster.Cluster.MaintenanceStatus)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", oldMasterId, err.Error())
		}
	})

	for slaveClusterID := range s.OldSlavesClusterIDMapToSyncTaskID {
		metaOfSlave, err := meta.Get(ctx, slaveClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", slaveClusterID, err.Error())
			return resp, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
				"get meta of cluster %s failed, %s", slaveClusterID, err.Error())
		}
		if req.RollbackClearPreviousMaintenanceFlag {
			err := metaOfSlave.EndMaintenance(ctx, constants.ClusterMaintenanceSwitching)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("rollbackClearPreviousMaintenance end maintenance of cluster %s failed:%s", slaveClusterID, err.Error())
				return resp, err
			}
		}
		if err := metaOfSlave.StartMaintenance(ctx, constants.ClusterMaintenanceSwitchoverRollback); err != nil {
			framework.LogWithContext(ctx).Errorf("start maintenance failed:%s", err.Error())
			return resp, errors.WrapError(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT, fmt.Sprintf("start maintenance failed, %s", err.Error()), err)
		}
		flowManager.AddContext(flow, wfContextOldSlavePreviousMaintenanceStatusKey, string(metaOfSlave.Cluster.MaintenanceStatus))
		thisSlaveID := slaveClusterID
		cancelFps = append(cancelFps, func() {
			err := metaOfSlave.EndMaintenance(ctx, constants.ClusterMaintenanceSwitchoverRollback)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", thisSlaveID, err.Error())
			}
		})
	}

	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", flowName, err.Error())
		return nil, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_ROLLBACK_FAILED,
			"async start %s workflow failed, %s", flowName, err.Error())
	}

	cancelFlag = false

	return &cluster.MasterSlaveClusterSwitchoverResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
	}, nil
}

//CheckSwitchover check whether the master cluster and slave cluster are in normal state
func (p *Manager) CheckSwitchover(ctx context.Context, masterClusterID, slaveClusterID string) (resp *structs.ClusterRelationsCheck, err error) {
	funcName := "CheckSwitchover"
	framework.LogWithContext(ctx).Infof("enter %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	oldMasterId := masterClusterID
	oldSlaveId := slaveClusterID
	// check
	//   1. cluster relation is valid?
	//   2. cdc sync task is valid?
	//   3. get cdc task status
	oldSyncChangeFeedTaskId, err := mgr.getOldSyncChangeFeedTaskId(ctx, "-", funcName, oldMasterId, oldSlaveId)
	framework.LogWithContext(ctx).Infof("%s getOldSyncChangeFeedTaskId, oldMasterId:%s oldSlaveId:%s oldSyncChangeFeedTaskId:%s err:%v",
		funcName, oldMasterId, oldSlaveId, oldSyncChangeFeedTaskId, err)
	if err != nil {
		return resp, err
	}
	if len(oldSyncChangeFeedTaskId) <= 0 {
		return resp, emerr.Error(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_CDC_SYNC_TASK_NOT_FOUND)
	}
	cdcTaskInfo, err := mgr.queryChangeFeedTask(ctx, oldSyncChangeFeedTaskId)
	framework.LogWithContext(ctx).Infof("%s queryChangeFeedTask, oldSyncChangeFeedTaskId:%s err:%v",
		funcName, oldSyncChangeFeedTaskId, err)
	if err != nil {
		return resp, err
	}
	retV := structs.ClusterRelationsCheck{
		SourceClusterID: oldMasterId,
		TargetClusterID: oldSlaveId,
		Relation:        string(constants.ClusterRelationStandBy),
		TaskID:          oldSyncChangeFeedTaskId,
		TaskStatus:      cdcTaskInfo.Status,
	}
	return &retV, err
}

func (p *Manager) checkMasterSalveRelation(ctx context.Context, masterClusterID, slaveClusterID string) error {
	_, err := p.clusterGetRelationByMasterSlaveClusterId(ctx, masterClusterID, slaveClusterID)
	return err
}

func (p *Manager) checkClusterDetailedHealthStatus(ctx context.Context, clusterID string) error {
	// TODO: use info below to compute a more precise result
	//    pd ctl region / tiup cluster display / tikv store status / tikv region status
	return p.checkClusterReadWriteHealth(ctx, clusterID)
}

func (p *Manager) checkClusterReadWriteHealth(ctx context.Context, clusterID string) error {
	userName, password, err := p.clusterGetCDCUserNameAndPwd(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster's mysql userName and password, err:%s", err)
	}
	var addr string
	addr, err = p.clusterGetOneConnectAddress(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster's mysql access addr, err:%s", err)
	}
	return p.checkClusterWritable(ctx, clusterID, userName, password, addr)
}

func (p *Manager) clusterGetMysqlUserNameAndPwd(ctx context.Context, clusterID string) (userName, password string, err error) {
	panic("NIY")
	framework.LogWithContext(ctx).Info("clusterGetMysqlUserNameAndPwd clusterID:", clusterID)
	db := models.GetClusterReaderWriter()
	_, err = db.Get(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Error("clusterGetMysqlUserNameAndPwd get cluster record err:", err)
		return userName, password, err
	} else {
		userName = "" //cluster.DBUser
		password = "" //cluster.DBPassword
		framework.LogWithContext(ctx).Infof(
			"clusterGetMysqlUserNameAndPwd get cluster record userName:%s password:%s err:%s", userName, password, err)
		return userName, password, err
	}
}

// with special `RESTRICTED_REPLICA_WRITER_ADMIN` privilege already set
func (p *Manager) clusterGetCDCUserNameAndPwd(ctx context.Context, clusterID string) (userName, password string, err error) {
	framework.LogWithContext(ctx).Info("clusterGetCDCUserNameAndPwd clusterID:", clusterID)
	m, err := meta.Get(ctx, clusterID)
	if err != nil {
		return "", "", err
	}
	user, err := m.GetDBUserNamePassword(ctx, constants.DBUserCDCDataSync)
	if err != nil {
		return "", "", err
	}
	return user.Name, user.Password.Val, nil
}

// addr: ip:port
func (p *Manager) clusterGetOneConnectAddress(ctx context.Context, clusterID string) (string, error) {
	ip, port, err := p.clusterGetOneConnectIPPort(ctx, clusterID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", ip, port), nil
}

// addr: ip:port
func (p *Manager) clusterGetOneConnectIPPort(ctx context.Context, clusterID string) (ip string, port int, err error) {
	m, err := meta.Get(ctx, clusterID)
	if err != nil {
		return "", 0, err
	}
	s := m.GetClusterConnectAddresses()
	if len(s) == 0 {
		return "", 0, fmt.Errorf("no connect address available")
	}
	return s[0].IP, s[0].Port, nil
}

func (p *Manager) clusterGetTLSMode(ctx context.Context, clusterID string) (tls bool, err error) {
	db := models.GetClusterReaderWriter()
	cluster, err := db.Get(ctx, clusterID)
	if err != nil {
		return tls, err
	} else {
		tls = cluster.TLS
		return tls, err
	}
}

func (p *Manager) clusterGetReadWriteMode(ctx context.Context, clusterID string) (readOnlyFlag bool, err error) {
	readOnlyFlag, err = mgr.clusterRestrictedReadOnlyOp(ctx, clusterID, "get", false)
	return
}

// set cluster Readonly to normal user but still Read-Writeable to changeFeedTask's user
func (p *Manager) clusterSetReadonly(ctx context.Context, clusterID string) error {
	_, err := mgr.clusterRestrictedReadOnlyOp(ctx, clusterID, "set", true)
	if err != nil {
		return err
	}
	return nil
}

// set cluster Read-Writeable to normal user and changeFeedTask's user
func (p *Manager) clusterSetReadWrite(ctx context.Context, clusterID string) error {
	_, err := mgr.clusterRestrictedReadOnlyOp(ctx, clusterID, "set", false)
	if err != nil {
		return err
	}
	return nil
}

func (p *Manager) convertTSOToPhysicalTime(ctx context.Context, tso uint64) time.Time {
	t, lt := tsoLib.ParseTS(tso)
	_ = lt
	return t
}

func (p *Manager) convertPhysicalTimeToTSO(ctx context.Context, t time.Time) (tso uint64) {
	return tsoLib.GenerateTSO(t, 0)
}

func (m *Manager) clusterCheckHasCDCComponent(ctx context.Context, clusterId string, myNotFoundErr error) error {
	myMeta, err := meta.Get(ctx, clusterId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"load cluster %s meta from db error: %s", clusterId, err.Error())
		return err
	}
	toplogy, _ := myMeta.DisplayInstanceInfo(ctx)
	for _, t := range toplogy.Topology {
		if t.Type == string(constants.ComponentIDCDC) {
			return nil
		}
	}
	return myNotFoundErr
}

func (m *Manager) clusterCheckNoRelation(ctx context.Context, clusterId string) error {
	funcName := "clusterCheckNoRelation"
	mRels, err := models.GetClusterReaderWriter().GetMasters(ctx, clusterId)
	framework.LogWithContext(ctx).Infof("%s: GetMasters len:%d err:%v", funcName, len(mRels), err)
	if err != nil {
		return err
	}
	if len(mRels) > 0 {
		return fmt.Errorf("cluster should have no master, clusterID:%s, mastersCount:%d", clusterId, len(mRels))
	}
	sRels, err := models.GetClusterReaderWriter().GetSlaves(ctx, clusterId)
	framework.LogWithContext(ctx).Infof("%s: GetSlaves len:%d err:%v", funcName, len(sRels), err)
	if err != nil {
		return err
	}
	if len(sRels) > 0 {
		return fmt.Errorf("cluster should have no slave, clusterID:%s, slavesCount:%d", clusterId, len(sRels))
	}
	return nil
}

func (m *Manager) clusterCheckNoCDCs(ctx context.Context, clusterId string) error {
	funcName := "clusterCheckNoCDCs"
	tasks, err := mgr.getAllChangeFeedTasksOnCluster(ctx, clusterId)
	framework.LogWithContext(ctx).Infof("%s: getAllChangeFeedTasksOnCluster clusterID:%s len:%d err:%v", funcName, clusterId, len(tasks), err)
	if err != nil {
		return err
	}
	if len(tasks) > 0 {
		return fmt.Errorf("cluster should have no cdc, clusterID:%s, cdcCount:%d", clusterId, len(tasks))
	}
	return nil
}

func (m *Manager) getSwitchoverMasterSlavesStateFromASwitchoverWorkflow(ctx context.Context, oldWorkFlowID string) (retState *switchoverMasterSlavesState, workflowRollbackSucceedFlag bool, retErr error) {
	funcName := "getSwitchoverMasterSlavesStateFromASwitchoverWorkflow"
	framework.LogWithContext(ctx).Infof("enter %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	var req message.QueryWorkFlowDetailReq
	req.WorkFlowID = oldWorkFlowID
	result, err := workflow.GetWorkFlowService().DetailWorkFlow(ctx, req)
	if err != nil {
		return nil, false, err
	}
	if switchoverWorkflowNameMap[result.Info.Name] {
	} else {
		return nil, false, fmt.Errorf("workflow %v is not a switchover workflow, result.Info.Name:%s", oldWorkFlowID, result.Info.Name)
	}
	if result.Info.Status != constants.WorkFlowStatusError {
		return nil, false, fmt.Errorf("workflow %v is not under Error state, result.Info.Status:%s", oldWorkFlowID, result.Info.Status)
	}
	if len(result.NodeInfo) <= 0 {
		return nil, false, fmt.Errorf("workflow %v has no valid NodeInfo field", oldWorkFlowID)
	}
	str := result.NodeInfo[0].Result
	if len(str) <= 0 {
		return nil, false, fmt.Errorf("result.NodeInfo[0].Result is zero-length string, workflow %v", oldWorkFlowID)
	}
	previousRollbackSuccessFlag := false
	{
		toParseStr := result.NodeInfo[len(result.NodeInfo)-1].Result
		units := strings.Split(toParseStr, "\n")
		if len(units) > 0 {
			if len(units[len(units)-1]) == 0 {
				units = units[:len(units)-1]
			}
		}
		if len(units) > 0 {
			framework.LogWithContext(ctx).Infof("%s check previousRollbackSuccess, last unit:%q", funcName, units[len(units)-1])
			if units[len(units)-1] == constants.SwitchoverRollbackSuccessInfoString {
				previousRollbackSuccessFlag = true
			}
		}
		framework.LogWithContext(ctx).Infof("%s previousRollbackSuccessFlag:%s", funcName, previousRollbackSuccessFlag)
	}
	framework.LogWithContext(ctx).Infof("%s result.NodeInfo[0].Result:%s", funcName, str)
	units := strings.Split(str, "\n")
	if len(units) <= 0 {
		return nil, previousRollbackSuccessFlag, fmt.Errorf("result.NodeInfo[0].Result strings.Split(str, \"\\n\") got zero unit, workflow %v", oldWorkFlowID)
	}
	framework.LogWithContext(ctx).Infof("%s units after strings.Split:%q", funcName, units)
	encryptedStr := units[0]
	str, err = mgr.decryptMarshaledMasterSlavesStateStr(ctx, encryptedStr)
	if err != nil {
		return nil, previousRollbackSuccessFlag, err
	}
	var s switchoverMasterSlavesState
	err = json.Unmarshal([]byte(str), &s)
	if err != nil {
		return nil, previousRollbackSuccessFlag, err
	}
	return &s, previousRollbackSuccessFlag, nil
}

func (m *Manager) relationsResetSyncChangeFeedTaskIDs(ctx context.Context, masterClusterID string, slavesClusterIDMapToSyncTaskID map[string]string) error {
	return models.GetClusterReaderWriter().RelationsResetSyncChangeFeedTaskIDs(ctx, masterClusterID, slavesClusterIDMapToSyncTaskID)
}

func (m *Manager) clusterGetRelationByMasterSlaveClusterId(ctx context.Context, masterClusterId, slaveClusterId string) (relation *clusterMgr.ClusterRelation, err error) {
	relations, err := models.GetClusterReaderWriter().GetRelations(ctx, slaveClusterId)
	if err != nil {
		return nil, err
	}
	for _, v := range relations {
		if v.SubjectClusterID == masterClusterId && v.RelationType == constants.ClusterRelationStandBy {
			relation = v
			break
		}
	}
	if relation == nil {
		err = fmt.Errorf("clusterGetRelationByMasterSlaveClusterId: master/slave relation not found, masterClusterID:%s, slaveClusterID:%s",
			masterClusterId, slaveClusterId)
	}
	return relation, err
}

func (m *Manager) clusterGetRelationsByMasterClusterId(ctx context.Context, masterClusterId string) ([]*clusterMgr.ClusterRelation, error) {
	relations, err := models.GetClusterReaderWriter().GetSlaves(ctx, masterClusterId)
	if err != nil {
		return nil, err
	}
	var ret []*clusterMgr.ClusterRelation
	for _, v := range relations {
		if v.RelationType == constants.ClusterRelationStandBy {
			ret = append(ret, v)
		}
	}
	return ret, err
}

func (m *Manager) clusterGetOtherSlavesMapToOldSyncCDCTask(ctx context.Context, oldMasterClusterId, oldSlaveClusterId string) (map[string]string, error) {
	relations, err := m.clusterGetRelationsByMasterClusterId(ctx, oldMasterClusterId)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string)
	for _, v := range relations {
		if len(v.ObjectClusterID) <= 0 {
			return nil, fmt.Errorf("clusterGetOtherSlavesMapToOldSyncCDCTask: ObjectClusterID is invalid"+
				"relationID:%v, objectClusterID:%s, syncChangeFeedTaskID:%s",
				v.ID, v.ObjectClusterID, v.SyncChangeFeedTaskID,
			)
		}
		if len(v.SyncChangeFeedTaskID) <= 0 {
			return nil, emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_CDC_SYNC_TASK_NOT_FOUND,
				"clusterGetOtherSlavesMapToOldSyncCDCTask: SyncChangeFeedTaskID is invalid"+
					"relationID:%v, objectClusterID:%s, syncChangeFeedTaskID:%s",
				v.ID, v.ObjectClusterID, v.SyncChangeFeedTaskID,
			)
		}
		if v.ObjectClusterID != oldSlaveClusterId {
			ret[v.ObjectClusterID] = v.SyncChangeFeedTaskID
		}
	}
	return ret, err
}

func (m *Manager) swapClusterRelationsInDB(ctx context.Context, oldMasterClusterId, slaveToBeMasterClusterId string, newSlaveClusterIdMapToSyncCDCTaskId map[string]string) error {
	return models.GetClusterReaderWriter().SwapMasterSlaveRelations(ctx, oldMasterClusterId, slaveToBeMasterClusterId, newSlaveClusterIdMapToSyncCDCTaskId)
}

func (m *Manager) getAllChangeFeedTasksOnCluster(ctx context.Context, clusterId string) ([]*cluster.ChangeFeedTask, error) {
	req := cluster.QueryChangeFeedTaskReq{
		ClusterId: clusterId,
		PageRequest: structs.PageRequest{
			Page:     0,
			PageSize: 0,
		},
	}
	tasks, _, err := mgr.changefeedMgr.Query(ctx, req)
	if err != nil {
		return nil, err
	}
	framework.LogWithContext(ctx).Infof("getAllChangeFeedTasksOnCluster tasks:%v err:%v", tasks, err)
	var myTasks []*cluster.ChangeFeedTask
	for _, v := range tasks {
		framework.LogWithContext(ctx).Infof("getAllChangeFeedTasksOnCluster task:%v id:%v", v, v.ID)
		dupV := v
		myTasks = append(myTasks, &dupV.ChangeFeedTask)
	}

	framework.LogWithContext(ctx).Infof("getAllChangeFeedTasksOnCluster ret myTasks:%v err:%v", myTasks, err)
	for _, v := range myTasks {
		framework.LogWithContext(ctx).Infof("getAllChangeFeedTasksOnCluster ret task:%v id:%v", v, v.ID)
	}

	return myTasks, err
}

type switchoverMasterSlavesState struct {
	OldMasterClusterID                string
	OldSlavesClusterIDMapToSyncTaskID map[string]string
	OldSlavesClusterIDMapToCDCIDs     map[string][]string
	CDCsOnMaster                      []*cluster.ChangeFeedTask
}

func (m *Manager) marshalSwitchoverMasterSlavesState(ctx context.Context, oldMasterClusterID string) (string, error) {
	rels, err := mgr.clusterGetRelationsByMasterClusterId(ctx, oldMasterClusterID)
	if err != nil {
		return "", err
	}
	var s switchoverMasterSlavesState
	s.OldSlavesClusterIDMapToSyncTaskID = make(map[string]string)
	s.OldSlavesClusterIDMapToCDCIDs = make(map[string][]string)
	s.OldMasterClusterID = oldMasterClusterID
	uniqM := make(map[string]bool)
	for _, v := range rels {
		slaveClusterID := v.ObjectClusterID
		if uniqM[slaveClusterID] {
			return "", fmt.Errorf("duplicate slaveClusterID found:%v", slaveClusterID)
		} else {
			uniqM[slaveClusterID] = true
		}
		s.OldSlavesClusterIDMapToSyncTaskID[slaveClusterID] = v.SyncChangeFeedTaskID
		if len(v.SyncChangeFeedTaskID) <= 0 {
			return "", fmt.Errorf("found zero-length v.SyncChangeFeedTaskID, slaveClusterID:%v", slaveClusterID)
		}
		ids, err := mgr.getAllChangeFeedTaskIDsOnCluster(ctx, slaveClusterID)
		if err != nil {
			return "", err
		}
		s.OldSlavesClusterIDMapToCDCIDs[slaveClusterID] = ids
	}
	tasks, err := mgr.getAllChangeFeedTasksOnCluster(ctx, s.OldMasterClusterID)
	if err != nil {
		return "", err
	}
	s.CDCsOnMaster = tasks
	{
		hitCt := 0
		for _, task := range s.CDCsOnMaster {
			for _, syncTaskID := range s.OldSlavesClusterIDMapToSyncTaskID {
				if syncTaskID == task.ID {
					hitCt++
					break
				}
			}
		}
		if hitCt != len(s.OldSlavesClusterIDMapToSyncTaskID) {
			return "", fmt.Errorf("hitCt != len(s.OldSlavesClusterIDMapToSyncTaskID), %d != %d", hitCt, len(s.CDCsOnMaster))
		}
	}
	bs, err := json.Marshal(&s)
	if err != nil {
		return "", err
	}
	encryptedStr, err := mgr.encryptMarshaledMasterSlavesStateStr(ctx, string(bs))
	return encryptedStr, err
}

func (m *Manager) getChangeFeedTask(ctx context.Context, changeFeedTaskId string) (*cluster.ChangeFeedTask, error) {
	req := cluster.DetailChangeFeedTaskReq{
		ID: changeFeedTaskId,
	}
	resp, err := mgr.changefeedMgr.Detail(ctx, req)
	if err != nil {
		return nil, err
	}
	return &resp.ChangeFeedTask, err
}

func (m *Manager) createChangeFeedTask(ctx context.Context, task *cluster.ChangeFeedTask) (changeFeedTaskId string, err error) {
	req := cluster.CreateChangeFeedTaskReq{
		Name:           task.Name,
		ClusterID:      task.ClusterID,
		StartTS:        task.StartTS,
		FilterRules:    task.FilterRules,
		DownstreamType: task.DownstreamType,
		Downstream:     task.Downstream,
	}
	{ //add default FilterRules and uniq
		defaultFilterRules := constants.DefaultFilterRules
		oldRulesM := make(map[string]bool)
		for _, v := range req.FilterRules {
			oldRulesM[v] = true
		}
		for _, v := range defaultFilterRules {
			if oldRulesM[v] {
			} else {
				req.FilterRules = append(req.FilterRules, v)
			}
		}
	}
	resp, err := mgr.changefeedMgr.Create(ctx, req)
	framework.LogWithContext(ctx).Infof("createChangeFeedTask clusterID:%s", task.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("createChangeFeedTask err:%s", err)
		return "", err
	}
	framework.LogWithContext(ctx).Infof("createChangeFeedTask success, ret taskID:%s", resp.ID)
	return resp.ID, nil
}

func (m *Manager) pauseChangeFeedTask(ctx context.Context, changeFeedTaskId string) error {
	req := cluster.PauseChangeFeedTaskReq{
		ID: changeFeedTaskId,
	}
	_, err := mgr.changefeedMgr.Pause(ctx, req)
	return err
}

func (m *Manager) resumeChangeFeedTask(ctx context.Context, changeFeedTaskId string) error {
	req := cluster.ResumeChangeFeedTaskReq{
		ID: changeFeedTaskId,
	}
	_, err := mgr.changefeedMgr.Resume(ctx, req)
	return err
}

func (m *Manager) queryChangeFeedTask(ctx context.Context, changeFeedTaskId string) (*cluster.ChangeFeedTaskInfo, error) {
	funcName := "queryChangeFeedTask"
	req := cluster.DetailChangeFeedTaskReq{
		ID: changeFeedTaskId,
	}
	resp, err := mgr.changefeedMgr.Detail(ctx, req)
	if err != nil {
		return nil, err
	}
	framework.LogWithContext(ctx).Infof(
		"%s queryChangeFeedTask changeFeedTaskId:%s success, status:%s", funcName, changeFeedTaskId, resp.Status)
	return &resp.ChangeFeedTaskInfo, err
}

func (m *Manager) getAllChangeFeedTaskIDsOnCluster(ctx context.Context, clusterID string) ([]string, error) {
	resp, err := mgr.getAllChangeFeedTasksOnCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, v := range resp {
		ids = append(ids, v.ID)
	}
	return ids, err
}

func (m *Manager) removeChangeFeedTask(ctx context.Context, changeFeedTaskId string) error {
	req := cluster.DeleteChangeFeedTaskReq{
		ID: changeFeedTaskId,
	}
	_, err := mgr.changefeedMgr.Delete(ctx, req)
	return err
}

func (m *Manager) calcCheckpointedTimeFromChangeFeedTaskInfo(ctx context.Context, info *cluster.ChangeFeedTaskInfo) (time.Time, error) {
	var t time.Time
	tso, err := strconv.ParseUint(info.DownstreamSyncTS, 10, 64)
	if err != nil {
		return t, err
	}
	t, _ = tsoLib.ParseTS(tso)
	return t, nil
}

func (m *Manager) queryChangeFeedTaskCheckpointedTime(ctx context.Context, changeFeedTaskId string) (time.Time, error) {
	var t time.Time
	info, err := m.queryChangeFeedTask(ctx, changeFeedTaskId)
	if err != nil {
		return t, err
	}
	t, err = m.calcCheckpointedTimeFromChangeFeedTaskInfo(ctx, info)
	return t, err
}

func (m *Manager) dupChangeFeedTaskStructWithTiDBDownStream(ctx context.Context, old *cluster.ChangeFeedTask) (*cluster.ChangeFeedTask, error) {
	new := cluster.ChangeFeedTask{}
	new = *old
	if old.DownstreamType == string(constants.DownstreamTypeTiDB) {
		var t changefeedModel.TiDBDownstream
		orig, ok := old.Downstream.(*changefeedModel.TiDBDownstream)
		if ok && orig != nil {
			t = *orig
			new.Downstream = &t
			return &new, nil
		} else {
			return nil, fmt.Errorf("dupChangeFeedTaskStructWithTiDBDownStream fail to convert old.Downstream")
		}
	}
	return nil, fmt.Errorf("dupChangeFeedTaskStructWithTiDBDownStream unsupported DownstreamType %s", old.DownstreamType)
}

func (m *Manager) getOldSyncChangeFeedTaskId(ctx context.Context, reqJson, logName, oldMasterId, oldSlaveId string) (string, error) {
	funcName := "wfGetSyncChangeFeedTaskId"
	relation, err := mgr.clusterGetRelationByMasterSlaveClusterId(ctx, oldMasterId, oldSlaveId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s getRelation req:%s err:%s", funcName, reqJson, err)
		return "", emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_NOT_FOUND, "master/slave relation not found: %s", err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s getRelation req:%s success", funcName, reqJson)
	}
	return relation.SyncChangeFeedTaskID, nil
}

func (m *Manager) checkSyncChangeFeedTaskHealth(ctx context.Context, reqJson, logName, syncChangeFeedTaskId string) error {
	funcName := logName
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	var firstT, secondT time.Time
	var firstInfo, secondInfo *cluster.ChangeFeedTaskInfo
	var err error
	firstInfo, err = mgr.queryChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s queryChangeFeedTask req:%s err:%s", funcName, reqJson, err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s queryChangeFeedTask req:%s success", funcName, reqJson)
	}
	if firstInfo.Status != constants.ChangeFeedStatusNormal.ToString() {
		return fmt.Errorf("%s syncChangeFeedTaskInfo status is %s instead of Normal", funcName, firstInfo.Status)
	}
	firstT, err = mgr.calcCheckpointedTimeFromChangeFeedTaskInfo(ctx, firstInfo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s err:%s", funcName, reqJson, err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s success", funcName, reqJson)
	}

	time.Sleep(constants.SwitchoverCheckSyncChangeFeedTaskHealthTimeInterval)

	secondInfo, err = mgr.queryChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s queryChangeFeedTask req:%s err:%s", funcName, reqJson, err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s queryChangeFeedTask req:%s success", funcName, reqJson)
	}
	if firstInfo.Status != constants.ChangeFeedStatusNormal.ToString() {
		return fmt.Errorf("%s syncChangeFeedTaskInfo status is %s instead of Normal", funcName, firstInfo.Status)
	}
	secondT, err = mgr.calcCheckpointedTimeFromChangeFeedTaskInfo(ctx, secondInfo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s err:%s", funcName, reqJson, err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s success", funcName, reqJson)
	}
	if secondT.After(firstT) {
		return nil
	} else {
		return fmt.Errorf("%s secondT is not bigger than firstT, firstT:%v secondT:%v", funcName, firstT, secondT)
	}
}

func (m *Manager) encryptStr(ctx context.Context, toEncrypt string) (string, error) {
	str, err := crypto.AesEncryptCFB(toEncrypt)
	return str, err
}

func (m *Manager) decryptStr(ctx context.Context, toDescryptStr string) (string, error) {
	str, err := crypto.AesDecryptCFB(toDescryptStr)
	return str, err
}

type marshalMasterSlavesState struct {
	MarshaledMasterSlavesState string
}

func (m *Manager) encryptMarshaledMasterSlavesStateStr(ctx context.Context, toEncrypt string) (string, error) {
	str, err := crypto.AesEncryptCFB(toEncrypt)
	if err != nil {
		return "", err
	}
	var ms marshalMasterSlavesState
	ms.MarshaledMasterSlavesState = str
	bs, err := json.Marshal(&ms)
	if err != nil {
		return "", err
	}
	return string(bs), err
}

func (m *Manager) decryptMarshaledMasterSlavesStateStr(ctx context.Context, toDescryptStr string) (string, error) {
	var ms marshalMasterSlavesState
	err := json.Unmarshal([]byte(toDescryptStr), &ms)
	if err != nil {
		return "", err
	}
	str, err := crypto.AesDecryptCFB(ms.MarshaledMasterSlavesState)
	return str, err
}
