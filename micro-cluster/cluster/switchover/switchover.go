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

package switchover

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	tsoLib "github.com/pingcap-inc/tiem/library/util/tso"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	changefeedModel "github.com/pingcap-inc/tiem/models/cluster/changefeed"
	clusterMgr "github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/workflow"
)

type Manager struct{}

func newManager() *Manager {
	return &Manager{}
}

var mgr = newManager()
var mgrOnceRegisterWorkFlow sync.Once

func GetManager() *Manager {
	mgrOnceRegisterWorkFlow.Do(func() {
		flowManager := workflow.GetWorkFlowService()
		flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowMasterSlaveSwitchoverNormal, &workflow.WorkFlowDefine{
			FlowName: constants.WorkFlowMasterSlaveSwitchoverNormal,
			TaskNodes: map[string]*workflow.NodeDefine{
				"start": {
					"checkHealthStatus", "checkSyncChangeFeedTaskMaxLagTime", "fail", workflow.SyncFuncNode, wfStepCheckOldSyncChangeFeedTaskHealth},
				"checkSyncChangeFeedTaskMaxLagTime": {
					"checkSyncChangeFeedTaskMaxLagTime", "setOldMasterReadOnly", "fail", workflow.SyncFuncNode, wfStepCheckSyncChangeFeedTaskMaxLagTime},
				"setOldMasterReadOnly": {
					"setOldMasterReadOnly", "checkSyncCaughtUp", "fail", workflow.SyncFuncNode, wfStepSetOldMasterReadOnly},
				"checkSyncCaughtUp": {
					"checkSyncCaughtUp", "pauseOldSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfStepCheckSyncCaughtUp},
				"pauseOldSyncChangeFeedTask": {
					"pauseOldSyncChangeFeedTask", "createReverseSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfStepPauseOldSyncChangeFeedTask},
				"createReverseSyncChangeFeedTask": {
					"createReverseSyncChangeFeedTask", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfStepCreateReverseSyncChangeFeedTask},
				"setNewMasterReadWrite": {
					"setNewMasterReadWrite", "checkNewMasterReadWriteHealth", "fail", workflow.SyncFuncNode, wfStepSetNewMasterReadWrite},
				"checkNewMasterReadWriteHealth": {
					"checkNewMasterReadWriteHealth", "checkNewSyncChangeFeedTaskHealth", "fail", workflow.SyncFuncNode, wfStepCheckNewMasterReadWriteHealth},
				"checkNewSyncChangeFeedTaskHealth": {
					"checkNewSyncChangeFeedTaskHealth", "migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "fail", workflow.SyncFuncNode, wfStepCheckNewSyncChangeFeedTaskHealth},
				"migrateAllDownStreamSyncChangeFeedTasksToNewMaster": {
					"migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster},
				"swapMasterSlaveRelationInDB": {
					"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfStepSwapMasterSlaveRelationInDB},
				"end": {
					"finish", "", "", workflow.SyncFuncNode, wfStepFinish},
				"fail": {
					"fail", "", "", workflow.SyncFuncNode, wfStepFail},
			},
		})
		flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowMasterSlaveSwitchoverForce, &workflow.WorkFlowDefine{
			FlowName: constants.WorkFlowMasterSlaveSwitchoverForce,
			TaskNodes: map[string]*workflow.NodeDefine{
				"start": {
					"setOldMasterReadOnly", "pauseOldSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfStepSetOldMasterReadOnly},
				"pauseOldSyncChangeFeedTask": {
					"pauseOldSyncChangeFeedTask", "createReverseSyncChangeFeedTask", "fail", workflow.SyncFuncNode, wfStepPauseOldSyncChangeFeedTask},
				"createReverseSyncChangeFeedTask": {
					"createReverseSyncChangeFeedTask", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfStepCreateReverseSyncChangeFeedTask},
				"setNewMasterReadWrite": {
					"setNewMasterReadWrite", "checkNewMasterReadWriteHealth", "fail", workflow.SyncFuncNode, wfStepSetNewMasterReadWrite},
				"checkNewMasterReadWriteHealth": {
					"checkNewMasterReadWriteHealth", "checkNewSyncChangeFeedTaskHealth", "fail", workflow.SyncFuncNode, wfStepCheckNewMasterReadWriteHealth},
				"checkNewSyncChangeFeedTaskHealth": {
					"checkNewSyncChangeFeedTaskHealth", "migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "fail", workflow.SyncFuncNode, wfStepCheckNewSyncChangeFeedTaskHealth},
				"migrateAllDownStreamSyncChangeFeedTasksToNewMaster": {
					"migrateAllDownStreamSyncChangeFeedTasksToNewMaster", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster},
				"swapMasterSlaveRelationInDB": {
					"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfStepSwapMasterSlaveRelationInDB},
				"end": {
					"finish", "", "", workflow.SyncFuncNode, wfStepFinish},
				"fail": {
					"fail", "", "", workflow.SyncFuncNode, wfStepFail},
			},
		})
		flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable,
			&workflow.WorkFlowDefine{
				FlowName: constants.WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable,
				TaskNodes: map[string]*workflow.NodeDefine{
					"start": {
						"setOldMasterReadOnly", "setNewMasterReadWrite", "fail", workflow.SyncFuncNode, wfStepSetOldMasterReadOnly},
					"setNewMasterReadWrite": {
						"setNewMasterReadWrite", "swapMasterSlaveRelationInDB", "fail", workflow.SyncFuncNode, wfStepSetNewMasterReadWrite},
					"swapMasterSlaveRelationInDB": {
						"swapMasterSlaveRelationInDB", "end", "fail", workflow.SyncFuncNode, wfStepSwapMasterSlaveRelationInDB},
					"end": {
						"finish", "", "", workflow.SyncFuncNode, wfStepFinish},
					"fail": {
						"fail", "", "", workflow.SyncFuncNode, wfStepFail},
				},
			})
	})
	return mgr
}

func (p *Manager) Switchover(ctx context.Context, req *cluster.MasterSlaveClusterSwitchoverReq) (resp *cluster.MasterSlaveClusterSwitchoverResp, err error) {
	framework.LogWithContext(ctx).Info("Manager.Switchover")

	flowName := constants.WorkFlowMasterSlaveSwitchoverNormal
	reqJsonBs, err := json.Marshal(req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("req marshal to json failed err:%s", err)
		return resp, err
	}
	reqJson := string(reqJsonBs)
	funcName := "Switchover"
	oldMasterId := req.SourceClusterID
	oldSlaveId := req.TargetClusterID
	oldSyncChangeFeedTaskId, err := mgr.getOldSyncChangeFeedTaskId(ctx, reqJson, funcName, oldMasterId, oldSlaveId)
	if err != nil {
		return resp, err
	}
	if req.Force {
		err = mgr.checkClusterReadWriteHealth(ctx, oldMasterId)
		if err == nil { // A rw-able
			framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldMasterId %s success", oldMasterId)
			err = mgr.checkSyncChangeFeedTaskHealth(ctx, reqJson, funcName, oldSyncChangeFeedTaskId)
			if err == nil {
				// A&B rw-able
				framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldSlaveId %s success", oldSlaveId)
				flowName = constants.WorkFlowMasterSlaveSwitchoverForce
			} else {
				// A rw-able & B unavailable
				framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldSlaveId %s failed err:%s", oldSlaveId, err)
				flowName = "" // end
				return resp, framework.NewTiEMErrorf(common.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED, "master/slave switchover failed: %s", "slave is unavailable")
			}
		} else {
			framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldMasterId %s failed err:%s", oldMasterId, err)
			err = mgr.checkClusterReadWriteHealth(ctx, oldSlaveId)
			if err == nil {
				// A unavailable & B rw-able
				framework.LogWithContext(ctx).Infof("checkClusterReadWriteHealth on oldSlaveId %s success", oldSlaveId)
				flowName = constants.WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable
			} else {
				// A unavailable & B unavailable
				framework.LogWithContext(ctx).Errorf("checkClusterReadWriteHealth on oldSlaveId %s failed err:%s", oldSlaveId, err)
				flowName = "" // end
				return resp, framework.NewTiEMErrorf(common.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED, "master/slave switchover failed: %s", "slave is unavailable")
			}
		}
	}
	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, oldMasterId, flowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", flowName, err.Error())
		return resp, framework.NewTiEMErrorf(common.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"create %s workflow failed: %s", flowName, err.Error())
	}

	flowManager.AddContext(flow, wfContextReqKey, req)
	flowManager.AddContext(flow, wfContextOldSyncChangeFeedTaskIDKey, oldSyncChangeFeedTaskId)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", flowName, err.Error())
		return nil, framework.NewTiEMErrorf(common.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"async start %s workflow failed, %s", flowName, err.Error())
	}

	return &cluster.MasterSlaveClusterSwitchoverResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
	}, nil
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
	userName, password, err := p.clusterGetMysqlUserNameAndPwd(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster's mysql userName and password, err:%s", err)
	}
	var addr string
	addr, err = p.clusterGetOneConnectAddresses(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster's mysql access addr, err:%s", err)
	}
	return p.checkClusterWritable(ctx, userName, password, addr)
}

func (p *Manager) clusterGetMysqlUserNameAndPwd(ctx context.Context, clusterID string) (userName, password string, err error) {
	framework.LogWithContext(ctx).Info("clusterGetMysqlUserNameAndPwd clusterID:", clusterID)
	db := models.GetClusterReaderWriter()
	cluster, err := db.Get(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Error("clusterGetMysqlUserNameAndPwd get cluster record err:", err)
		return userName, password, err
	} else {
		userName = cluster.DBUser
		password = cluster.DBPassword
		framework.LogWithContext(ctx).Infof(
			"clusterGetMysqlUserNameAndPwd get cluster record userName:%s password:%s err:%s", userName, password, err)
		return userName, password, err
	}
}

// addr: ip:port
func (p *Manager) clusterGetOneConnectAddresses(ctx context.Context, clusterID string) (string, error) {
	m, err := handler.Get(ctx, clusterID)
	s := m.GetConnectAddresses()
	if err != nil {
		return "", err
	}
	if len(s) == 0 {
		return "", fmt.Errorf("no connect address available")
	}
	return s[0], nil
}

// addr: ip:port
func (p *Manager) clusterGetOneConnectIPPort(ctx context.Context, clusterID string) (ip string, port int, err error) {
	panic("NIY")
	return
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
	db := models.GetClusterReaderWriter()
	cluster, err := db.Get(ctx, clusterID)
	if err != nil {
		return readOnlyFlag, err
	} else {
		readOnlyFlag = cluster.ReadOnlyFlag
		return readOnlyFlag, err
	}
}

func (p *Manager) clusterSetReadonly(ctx context.Context, clusterID string) error {
	db := models.GetClusterReaderWriter()
	return db.UpdateReadOnlyFlag(ctx, clusterID, true)
}

func (p *Manager) clusterSetReadWrite(ctx context.Context, clusterID string) error {
	db := models.GetClusterReaderWriter()
	return db.UpdateReadOnlyFlag(ctx, clusterID, false)
}

func (p *Manager) convertTSOToPhysicalTime(ctx context.Context, tso uint64) time.Time {
	t, lt := tsoLib.ParseTS(tso)
	_ = lt
	return t
}

func (p *Manager) convertPhysicalTimeToTSO(ctx context.Context, t time.Time) (tso uint64) {
	return tsoLib.GenerateTSO(t, 0)
}

func (m *Manager) clusterGetRelationByMasterSlaveClusterId(ctx context.Context, masterClusterId, slaveClusterId string) (relation *clusterMgr.ClusterRelation, err error) {
	panic("NIY")
	return nil, nil
}

func (m *Manager) swapClusterRelationInDB(ctx context.Context, oldMasterClusterId, oldSlaveClusterId, newSyncChangeFeedTaskId string) error {
	panic("NIY")
	return nil
}

func (m *Manager) getAllChangeFeedTasksOnCluster(ctx context.Context, clusterId string) ([]*cluster.ChangeFeedTask, error) {
	panic("NIY")
	return nil, nil
}

func (m *Manager) getChangeFeedTask(ctx context.Context, changeFeedTaskId string) (*cluster.ChangeFeedTask, error) {
	panic("NIY")
	return nil, nil
}

func (m *Manager) createChangeFeedTask(ctx context.Context, task *cluster.ChangeFeedTask) (changeFeedTaskId string, err error) {
	panic("NIY")
	return "", nil
}

func (m *Manager) pauseChangeFeedTask(ctx context.Context, changeFeedTaskId string) error {
	panic("NIY")
	return nil
}

func (m *Manager) queryChangeFeedTask(ctx context.Context, changeFeedTaskId string) (*cluster.ChangeFeedTaskInfo, error) {
	panic("NIY")
	return nil, nil
}

func (m *Manager) calcCheckpointedTimeFromChangeFeedTaskInfo(ctx context.Context, info *cluster.ChangeFeedTaskInfo) (time.Time, error) {
	var t time.Time
	panic("NIY")
	return t, nil
}

func (m *Manager) queryChangeFeedTaskCheckpointedTime(ctx context.Context, changeFeedTaskId string) (time.Time, error) {
	var t time.Time
	panic("NIY")
	return t, nil
}

func (m *Manager) deleteChangeFeedTask(ctx context.Context, changeFeedTaskId string) error {
	panic("NIY")
	return nil
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
		return "", framework.NewTiEMErrorf(common.TIEM_MASTER_SLAVE_SWITCHOVER_NOT_FOUND, "master/slave relation not found: %s", err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s getRelation req:%s success", funcName, reqJson)
	}
	return relation.SyncChangeFeedTaskId, nil
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
	if firstInfo.Status != constants.Normal.ToString() {
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
	if firstInfo.Status != constants.Normal.ToString() {
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
