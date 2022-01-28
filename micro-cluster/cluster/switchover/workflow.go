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

package switchover

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	emerr "github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	changefeedModel "github.com/pingcap-inc/tiem/models/cluster/changefeed"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	wfContextReqKey                       string = "wfContextReqKey"
	wfContextOldSyncChangeFeedTaskIDKey   string = "wfContextOldSyncChangeFeedTaskIDKey"
	wfContextNewSyncChangeFeedTaskIDKey   string = "wfContextNewSyncChangeFeedTaskIDKey"
	wfContextCancelDeferStackKey          string = "wfContextCancelDeferStackKey"
	wfContextIsExecutingDeferStackFlagKey string = "wfContextIsExecutingDeferStackFlagKey"

	wfContextOldMasterPreviousMaintenanceStatusKey string = "wfContextOldMasterPreviousMaintenanceStatusKey"
	wfContextOldSlavePreviousMaintenanceStatusKey  string = "wfContextOldSlavePreviousMaintenanceStatusKey"
)

// wfGetReq workflow get request
func wfGetReq(ctx *workflow.FlowContext) *cluster.MasterSlaveClusterSwitchoverReq {
	return ctx.GetData(wfContextReqKey).(*cluster.MasterSlaveClusterSwitchoverReq)
}

func wfGetReqJson(ctx *workflow.FlowContext) string {
	bs, err := json.Marshal(wfGetReq(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("wfGetReqJson failed err:%s", err)
		return "-"
	}
	return string(bs)
}

func wfGetIsExecutingDeferStackFlag(ctx *workflow.FlowContext) bool {
	s, ok := ctx.GetData(wfContextIsExecutingDeferStackFlagKey).(string)
	if ok {
		if len(s) > 0 {
			return true
		}
	} else {
	}
	return false
}

func wfSetExecutingDeferStackFlag(ctx *workflow.FlowContext) {
	ctx.SetData(wfContextIsExecutingDeferStackFlagKey, "true")
}

func myAssert(ctx context.Context, b bool) {
	if b {
	} else {
		framework.LogWithContext(ctx).Panicf("switchover: assert failed")
	}
}

type deferStack []string

func wfGetDeferStack(ctx *workflow.FlowContext) []string {
	var stk deferStack
	s, ok := ctx.GetData(wfContextCancelDeferStackKey).(string)
	if ok {
		err := json.Unmarshal([]byte(s), &stk)
		if err != nil {
			framework.Log().Panicf("wfGetDeferStack: unmarshal deferStack failed, s:%s", s)
		}
	} else {
	}
	return stk
}

func wfSetDeferStack(ctx *workflow.FlowContext, stk []string) {
	bs, err := json.Marshal(&stk)
	if err != nil {
		framework.Log().Panicf("wfSetDeferStack: marshal deferStack failed, err:%s", err)
	}
	ctx.SetData(wfContextCancelDeferStackKey, string(bs))
}

func wfGetDeferStackToConsume(ctx *workflow.FlowContext) []string {
	stk := wfGetDeferStack(ctx)
	length := len(stk)
	reversedStk := make(deferStack, length)
	for i := range stk {
		reversedStk[length-1-i] = stk[i]
	}
	return reversedStk
}

func wfPushDeferStack(ctx *workflow.FlowContext, taskName string) {
	if wfGetIsExecutingDeferStackFlag(ctx) {
		return
	}
	stk := wfGetDeferStack(ctx)
	stk = append(stk, taskName)
	wfSetDeferStack(ctx, stk)
}

func wfGetOldMasterClusterId(ctx *workflow.FlowContext) string {
	return wfGetReq(ctx).SourceClusterID
}

func wfGetNewMasterClusterId(ctx *workflow.FlowContext) string {
	return wfGetReq(ctx).TargetClusterID
}

func wfGetOldSlaveClusterId(ctx *workflow.FlowContext) string {
	return wfGetReq(ctx).TargetClusterID
}

func wfGetNewSlaveClusterId(ctx *workflow.FlowContext) string {
	return wfGetReq(ctx).SourceClusterID
}

func wfGetOldSyncChangeFeedTaskId(ctx *workflow.FlowContext) string {
	return ctx.GetData(wfContextOldSyncChangeFeedTaskIDKey).(string)
}

func wfGetNewSyncChangeFeedTaskId(ctx *workflow.FlowContext) string {
	s, ok := ctx.GetData(wfContextNewSyncChangeFeedTaskIDKey).(string)
	if ok {
		return s
	} else {
		return ""
	}
}

func wfSetNewSyncChangeFeedTaskId(ctx *workflow.FlowContext, newTaskID string) {
	ctx.SetData(wfContextNewSyncChangeFeedTaskIDKey, newTaskID)
}

func wfGetOldMasterPreviousMaintenanceStatus(ctx *workflow.FlowContext) string {
	return ctx.GetData(wfContextOldMasterPreviousMaintenanceStatusKey).(string)
}

func wfGetOldSlavePreviousMaintenanceStatus(ctx *workflow.FlowContext) string {
	return ctx.GetData(wfContextOldSlavePreviousMaintenanceStatusKey).(string)
}

// wfnSetNewMasterReadWrite workflow node set new master read-write mode
func wfnSetNewMasterReadWrite(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("start wfnSetNewMasterReadWrite")
	defer framework.LogWithContext(ctx).Info("exit wfnSetNewMasterReadWrite")
	err := mgr.clusterSetReadWrite(ctx, wfGetNewMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"wfnSetNewMasterReadWrite clusterId:%s failed err:%s", wfGetNewMasterClusterId(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"wfnSetNewMasterReadWrite clusterId:%s success", wfGetNewMasterClusterId(ctx))
	}
	return err
}

// wfnSetOldMasterReadWrite workflow node set new master read-write mode
func wfnSetOldMasterReadWrite(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("start wfnSetOldMasterReadWrite")
	defer framework.LogWithContext(ctx).Info("exit wfnSetOldMasterReadWrite")
	err := mgr.clusterSetReadWrite(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"wfnSetOldMasterReadWrite clusterId:%s failed err:%s", wfGetOldMasterClusterId(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"wfnSetOldMasterReadWrite clusterId:%s success", wfGetOldMasterClusterId(ctx))
	}
	return err
}

func wfnSetOldMasterReadOnly(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("start wfnSetOldMasterReadOnly")
	defer framework.LogWithContext(ctx).Info("exit wfnSetOldMasterReadOnly")
	err := mgr.clusterSetReadonly(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"wfnSetOldMasterReadOnly clusterId:%s failed err:%s", wfGetOldMasterClusterId(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"wfnSetOldMasterReadOnly clusterId:%s success", wfGetOldMasterClusterId(ctx))
	}
	return err
}

func wfnSetOldSlaveReadOnly(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("start wfnSetOldSlaveReadOnly")
	defer framework.LogWithContext(ctx).Info("exit wfnSetOldSlaveReadOnly")
	err := mgr.clusterSetReadonly(ctx, wfGetOldSlaveClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"wfnSetOldSlaveReadOnly clusterId:%s failed err:%s", wfGetOldSlaveClusterId(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"wfnSetOldSlaveReadOnly clusterId:%s success", wfGetOldSlaveClusterId(ctx))
	}
	return err
}

func wfnSwapMasterSlaveRelationInDB(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, newSyncChangeFeedTaskId string) error {
	funcName := "wfnSwapMasterSlaveRelationInDB"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	if len(newSyncChangeFeedTaskId) == 0 {
		framework.LogWithContext(ctx).Warnf("%s newTaskID is not exist", funcName)
	} else {
		framework.LogWithContext(ctx).Warnf("%s newTaskID is %s", funcName, newSyncChangeFeedTaskId)
	}
	err := mgr.swapClusterRelationInDB(ctx, wfGetOldMasterClusterId(ctx), wfGetOldSlaveClusterId(ctx), newSyncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

func wfnPauseSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnPauseSyncChangeFeedTask"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	err := mgr.pauseChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

func wfnResumeSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnResumeSyncChangeFeedTask"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	err := mgr.resumeChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

func wfnRemoveSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnRemoveSyncChangeFeedTask"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	err := mgr.removeChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

func wfnCreateReverseSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, oldSyncChangeFeedTask *cluster.ChangeFeedTask) (newSyncChangeFeedTaskId string, err error) {
	funcName := "wfnCreateReverseSyncChangeFeedTask"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)

	newTask, err := mgr.dupChangeFeedTaskStructWithTiDBDownStream(ctx, oldSyncChangeFeedTask)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s dupChangeFeedTaskStructWithTiDBDownStream req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return "", err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s dupChangeFeedTaskStructWithTiDBDownStream req:%s success", funcName, wfGetReqJson(ctx))
	}

	var zeroT time.Time
	newTask.ID = ""
	newTask.Name = ""
	newTask.ClusterID = wfGetNewMasterClusterId(ctx)
	newTask.StartTS = "0"
	newTask.Status = constants.ChangeFeedStatusNormal.ToString()
	newTask.CreateTime = zeroT
	newTask.UpdateTime = zeroT
	tidbDownStream := newTask.Downstream.(*changefeedModel.TiDBDownstream)

	var userName, password, ip string
	var port int
	var tls bool
	userName, password, err = mgr.clusterGetCDCUserNameAndPwd(ctx, wfGetNewSlaveClusterId(ctx))
	var err2 error
	ip, port, err2 = mgr.clusterGetOneConnectIPPort(ctx, wfGetNewSlaveClusterId(ctx))
	var err3 error
	tls, err3 = mgr.clusterGetTLSMode(ctx, wfGetNewSlaveClusterId(ctx))
	if err != nil || err2 != nil || err3 != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s get cluster connect addr & mysql user name and password, req:%s err:%s err2:%s err3:%s",
			funcName, wfGetReqJson(ctx), err, err2, err3)
		return "", err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s get cluster connect addr & mysql user name and password, req:%s success", funcName, wfGetReqJson(ctx))
	}
	tidbDownStream.TargetClusterId = wfGetNewSlaveClusterId(ctx)
	tidbDownStream.Username = userName
	tidbDownStream.Password = password
	tidbDownStream.Ip = ip
	tidbDownStream.Port = port
	tidbDownStream.Tls = tls

	newTaskId, err := mgr.createChangeFeedTask(ctx, newTask)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s createChangeFeedTask req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return "", err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s createChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
	}

	return newTaskId, err
}

func wfnCheckNewMasterReadWriteHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfnCheckClusterReadWriteHealth"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	err := mgr.checkClusterReadWriteHealth(ctx, wfGetNewMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

func wfnGetSyncChangeFeedTaskStatus(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) (constants.ChangeFeedStatus, error) {
	funcName := "wfnGetSyncChangeFeedTaskStatus"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	task, err := mgr.getChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return "", err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	return constants.ChangeFeedStatus(task.Status), err
}

func wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnMigrateAllDownStreamSyncChangeFeedTasks(node, ctx,
		wfGetOldMasterClusterId(ctx), wfGetNewMasterClusterId(ctx), wfGetOldSyncChangeFeedTaskId(ctx))
}

func wfnMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnMigrateAllDownStreamSyncChangeFeedTasks(node, ctx,
		wfGetNewMasterClusterId(ctx), wfGetOldMasterClusterId(ctx), wfGetNewSyncChangeFeedTaskId(ctx))
}

func wfnMigrateAllDownStreamSyncChangeFeedTasks(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext,
	fromCluster, toCluster, exceptSyncChangeFeedTaskId string) error {
	funcName := "wfnMigrateAllDownStreamSyncChangeFeedTasks"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)

	if len(fromCluster) == 0 || len(toCluster) == 0 {
		return fmt.Errorf("%s: unexpected clusterID len, fromCluster:%s toCluster:%s", funcName, fromCluster, toCluster)
	}

	tasks, err := mgr.getAllChangeFeedTasksOnCluster(ctx, fromCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success", funcName, wfGetReqJson(ctx))
	}
	var finalTasks []*cluster.ChangeFeedTask
	for _, v := range tasks {
		if v != nil && v.ID != exceptSyncChangeFeedTaskId {
			finalTasks = append(finalTasks, v)
		}
	}
	for _, v := range finalTasks {
		err := mgr.pauseChangeFeedTask(ctx, v.ID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s pauseChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), v.ID, err)
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s pauseChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
		}
	}
	for _, v := range finalTasks {
		var zeroT time.Time
		v.ID = ""
		v.Name = ""
		v.ClusterID = toCluster
		v.StartTS = "0"
		v.Status = constants.ChangeFeedStatusNormal.ToString()
		v.CreateTime = zeroT
		v.UpdateTime = zeroT
		id, err := mgr.createChangeFeedTask(ctx, v)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s pauseChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), v.ID, err)
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s pauseChangeFeedTask req:%s success taskId:%s", funcName, wfGetReqJson(ctx), id)
		}
	}
	return err
}

func wfnCheckSyncChangeFeedTaskMaxLagTime(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnCheckSyncChangeFeedTaskLagTime"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	slaveT, err := mgr.queryChangeFeedTaskCheckpointedTime(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s queryChangeFeedTaskCheckpointedTime req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s queryChangeFeedTaskCheckpointedTime req:%s success", funcName, wfGetReqJson(ctx))
	}
	nowT := time.Now()
	if nowT.After(slaveT) {
		minSyncT := nowT.Add(-constants.SwitchoverCheckMasterSlaveMaxLagTime)
		if slaveT.After(minSyncT) {
			return nil
		} else {
			return fmt.Errorf("%s slave checkpoint lag time is bigger than %v", funcName, constants.SwitchoverCheckMasterSlaveMaxLagTime)
		}
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s suspicious checkpoint t, req:%s salveCheckpointT:%v nowT:%v", funcName, wfGetReqJson(ctx), slaveT, nowT)
		return nil
	}
}

func wfnCheckSyncChangeFeedTaskHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnCheckSyncChangeFeedTaskHealth"
	return mgr.checkSyncChangeFeedTaskHealth(ctx, wfGetReqJson(ctx), funcName, syncChangeFeedTaskId)
}

func wfnCheckSyncCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, syncChangeFeedTaskId string) error {
	funcName := "wfnCheckSyncCaughtUp"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	var slaveT time.Time
	var firstInfo *cluster.ChangeFeedTaskInfo
	var err error
	firstInfo, err = mgr.queryChangeFeedTask(ctx, syncChangeFeedTaskId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s queryChangeFeedTask req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s queryChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
	}
	if firstInfo.Status != constants.ChangeFeedStatusNormal.ToString() {
		return fmt.Errorf("%s syncChangeFeedTaskInfo status is %s instead of Normal", funcName, firstInfo.Status)
	}
	slaveT, err = mgr.calcCheckpointedTimeFromChangeFeedTaskInfo(ctx, firstInfo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s calcCheckpointedTimeFromChangeFeedTaskInfo req:%s success", funcName, wfGetReqJson(ctx))
	}
	nowT := time.Now()
	if nowT.After(slaveT) {
		minSyncT := nowT.Add(-constants.SwitchoverCheckSyncChangeFeedTaskCaughtUpMaxLagTime)
		if slaveT.After(minSyncT) {
			return nil
		} else {
			return fmt.Errorf("%s slave checkpoint lag time is bigger than %v", funcName, constants.SwitchoverCheckSyncChangeFeedTaskCaughtUpMaxLagTime)
		}
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s suspicious checkpoint t, req:%s slaveCheckpointT:%v nowT:%v", funcName, wfGetReqJson(ctx), slaveT, nowT)
		return nil
	}
}

func wfnExecuteDeferStack(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	wfSetExecutingDeferStackFlag(ctx)

	funcName := "wfnExecuteDeferStack"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)

	fpMap := map[string]func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error{
		//"wfStepCheckNewMasterReadWriteHealth":                          wfStepCheckNewMasterReadWriteHealth,
		//"wfStepCheckNewSyncChangeFeedTaskHealth":                       wfStepCheckNewSyncChangeFeedTaskHealth,
		//"wfStepCheckOldSyncChangeFeedTaskHealth":                       wfStepCheckOldSyncChangeFeedTaskHealth,
		//"wfStepCheckSyncCaughtUp":                                      wfStepCheckSyncCaughtUp,
		//"wfStepCheckSyncChangeFeedTaskMaxLagTime":                      wfStepCheckSyncChangeFeedTaskMaxLagTime,
		//"wfStepCreateReverseSyncChangeFeedTask": wfStepCreateReverseSyncChangeFeedTask,
		//"wfStepFail":                                                   wfStepFail,
		//"wfStepFinish":                                                 wfStepFinish,
		"wfStepMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster": wfStepMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster,
		//"wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster":     wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster,
		//"wfStepPauseOldSyncChangeFeedTask":  wfStepPauseOldSyncChangeFeedTask,
		"wfStepRemoveNewSyncChangeFeedTask": wfStepRemoveNewSyncChangeFeedTask,
		"wfStepResumeOldSyncChangeFeedTask": wfStepResumeOldSyncChangeFeedTask,
		//"wfStepSetNewMasterReadWrite":       wfStepSetNewMasterReadWrite,
		//"wfStepSetOldMasterReadOnly":        wfStepSetOldMasterReadOnly,
		"wfStepSetOldMasterReadWrite": wfStepSetOldMasterReadWrite,
		"wfStepSetOldSlaveReadonly":   wfStepSetOldSlaveReadonly,
		//"wfStepSwapMasterSlaveRelationInDB": wfStepSwapMasterSlaveRelationInDB,
	}
	list := wfGetDeferStackToConsume(ctx)
	wfSetDeferStack(ctx, nil)
	defer func() {
		myAssert(ctx, len(wfGetDeferStackToConsume(ctx)) == 0)
	}()
	var errAfterAllRetryies error
	framework.LogWithContext(ctx).Debugf("%s start consume defer stack, len: %d", funcName, len(list))
	for i, stepName := range list {
		var err error
		framework.LogWithContext(ctx).Debugf("%s consume defer stack idx:%d", funcName, i)
		fp, ok := fpMap[stepName]
		if ok == false {
			err = fmt.Errorf("%s didn't found step function which named %s", funcName, stepName)
			framework.LogWithContext(ctx).Panicf("%s err:%s", funcName, err)
			return err
		}
		for retryCt := range make([]struct{}, constants.SwitchoverCancelOpRetriesCount+1) {
			if err != nil {
				time.Sleep(constants.SwitchoverCancelOpRetryWait)
			}
			err = fp(node, ctx)
			if err == nil {
				framework.LogWithContext(ctx).Debugf("%s stepName:%s success, retry count:%d", funcName, stepName, retryCt)
				break
			} else {
				framework.LogWithContext(ctx).Debugf("%s stepName:%s failed, err:%s, retry count:%d", funcName, stepName, err, retryCt)
			}
		}
		if err != nil {
			errAfterAllRetryies = err
			framework.LogWithContext(ctx).Warnf("%s stepName:%s errAfterAllRetryies:%s", funcName, stepName, err)
			if constants.SwitchoverCancelOpRunAllStepsEvenOnFail == false {
				framework.LogWithContext(ctx).Errorf("%s break cancel op iteration, err:%s", funcName, err)
				break
			}
		} else {
			if i < len(list)-1 {
				framework.LogWithContext(ctx).Warnf("%s continue to run other steps even when step %s failed", funcName, stepName)
			}
		}
	}
	return errAfterAllRetryies
}

func wfStepFinish(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepFinish"
	framework.LogWithContext(ctx).Infof("%s", funcName)
	err := mgr.removeChangeFeedTask(ctx, wfGetOldSyncChangeFeedTaskId(ctx))
	if err == nil {
		framework.LogWithContext(ctx).Infof("%s remove old change feed task success", funcName)
	} else {
		framework.LogWithContext(ctx).Warnf("%s remove old change feed task failed, id:%s, err:%s",
			funcName, wfGetOldSyncChangeFeedTaskId(ctx), err)
	}

	errToRet := err
	metaOfSource, err := meta.Get(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		errToRet = err
		framework.LogWithContext(ctx).Warnf("%s get meta of cluster %s failed:%s",
			funcName, wfGetOldMasterClusterId(ctx), err)
		errToRet = emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"get meta of %s failed, %s", wfGetOldMasterClusterId(ctx), err.Error())
	} else {
		previousStatus := constants.ClusterMaintenanceStatus(wfGetOldMasterPreviousMaintenanceStatus(ctx))
		err := metaOfSource.EndMaintenance(ctx, previousStatus)
		if err != nil {
			errToRet = err
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", wfGetOldMasterClusterId(ctx), err.Error())
		}
	}

	metaOfTarget, err := meta.Get(ctx, wfGetOldSlaveClusterId(ctx))
	if err != nil {
		errToRet = err
		framework.LogWithContext(ctx).Warnf("%s get meta of cluster %s failed:%s",
			funcName, wfGetOldSlaveClusterId(ctx), err)
		errToRet = emerr.NewErrorf(emerr.TIEM_MASTER_SLAVE_SWITCHOVER_FAILED,
			"get meta of %s failed, %s", wfGetOldSlaveClusterId(ctx), err.Error())
	} else {
		previousStatus := constants.ClusterMaintenanceStatus(wfGetOldSlavePreviousMaintenanceStatus(ctx))
		err := metaOfTarget.EndMaintenance(ctx, previousStatus)
		if err != nil {
			errToRet = err
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", wfGetOldSlaveClusterId(ctx), err.Error())
		}
	}

	return errToRet
}

func wfStepFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepFail"
	framework.LogWithContext(ctx).Errorf("%s", funcName)
	return wfnExecuteDeferStack(node, ctx)
}

func ifNoErrThenPushDefer(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext,
	fp func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error, deferTaskName string) error {

	err := fp(node, ctx)
	if err == nil {
		wfPushDeferStack(ctx, deferTaskName)
		return nil
	} else {
		return err
	}
}

func wfStepSetNewMasterReadWrite(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return ifNoErrThenPushDefer(node, ctx, wfnSetNewMasterReadWrite, "wfStepSetOldSlaveReadonly")
}

func wfStepSetOldSlaveReadonly(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	myAssert(ctx, wfGetIsExecutingDeferStackFlag(ctx))
	return wfnSetOldSlaveReadOnly(node, ctx)
}

func wfStepSetOldMasterReadOnly(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return ifNoErrThenPushDefer(node, ctx, wfnSetOldMasterReadOnly, "wfStepSetOldMasterReadWrite")
}

func wfStepSetOldMasterReadWrite(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	myAssert(ctx, wfGetIsExecutingDeferStackFlag(ctx))
	return wfnSetOldMasterReadWrite(node, ctx)
}

func wfStepSwapMasterSlaveRelationInDB(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	// no defer task needed because it is the last step
	return wfnSwapMasterSlaveRelationInDB(node, ctx, wfGetNewSyncChangeFeedTaskId(ctx))
}

func wfStepResumeOldSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	myAssert(ctx, wfGetIsExecutingDeferStackFlag(ctx))
	return wfnResumeSyncChangeFeedTask(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
}

func wfStepPauseOldSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return ifNoErrThenPushDefer(node, ctx,
		func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
			return wfnPauseSyncChangeFeedTask(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		},
		"wfStepResumeOldSyncChangeFeedTask",
	)
}

func wfStepRemoveNewSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	myAssert(ctx, wfGetIsExecutingDeferStackFlag(ctx))
	return wfnRemoveSyncChangeFeedTask(node, ctx, wfGetNewSyncChangeFeedTaskId(ctx))
}

func wfStepCreateReverseSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	fp := func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
		funcName := "wfStepCreateReverseSyncChangeFeedTask"
		framework.LogWithContext(ctx).Infof("start %s", funcName)
		defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
		task, err := mgr.getChangeFeedTask(ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s getChangeFeedTask req:%s err:%s", funcName, wfGetReqJson(ctx), err)
			return err
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s getChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
		}
		var newTaskID string
		newTaskID, err = wfnCreateReverseSyncChangeFeedTask(node, ctx, task)
		if err == nil {
			wfSetNewSyncChangeFeedTaskId(ctx, newTaskID)
		}
		return err
	}
	return ifNoErrThenPushDefer(node, ctx, fp, "wfStepRemoveNewSyncChangeFeedTask")
}

func wfStepCheckNewMasterReadWriteHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	for range make([]struct{}, constants.SwitchoverCheckClusterReadWriteHealthRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckClusterReadWriteHealthRetryWait)
		}
		err = wfnCheckNewMasterReadWriteHealth(node, ctx)
		if err == nil {
			return nil
		}
	}
	return err
}

func wfStepCheckNewSyncChangeFeedTaskHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepCheckNewSyncChangeFeedTaskHealth"
	for i := range make([]struct{}, constants.SwitchoverCheckSyncChangeFeedTaskHealthRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckSyncChangeFeedTaskHealthRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
		}
		err = wfnCheckSyncChangeFeedTaskHealth(node, ctx, wfGetNewSyncChangeFeedTaskId(ctx))
		if err == nil {
			return nil
		}
	}
	return err
}

func wfStepMigrateAllDownStreamSyncChangeFeedTasksToNewMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return ifNoErrThenPushDefer(node, ctx, wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster,
		"wfStepMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster")
}

func wfStepMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	myAssert(ctx, wfGetIsExecutingDeferStackFlag(ctx))
	return wfnMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster(node, ctx)
}

func wfStepCheckSyncChangeFeedTaskMaxLagTime(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfnStepCheckSyncChangeFeedTaskLagTime"
	for i := range make([]struct{}, constants.SwitchoverCheckMasterSlaveMaxLagTimeRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckMasterSlaveMaxLagTimeRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
		}
		err = wfnCheckSyncChangeFeedTaskMaxLagTime(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		if err == nil {
			return nil
		}
	}
	return err
}

func wfStepCheckOldSyncChangeFeedTaskHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepCheckOldSyncChangeFeedTaskHealth"
	for i := range make([]struct{}, constants.SwitchoverCheckSyncChangeFeedTaskHealthRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckSyncChangeFeedTaskHealthRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
		}
		err = wfnCheckSyncChangeFeedTaskHealth(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		if err == nil {
			return nil
		}
	}
	return err
}

func wfStepCheckSyncCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepCheckSyncCaughtUp"
	for i := range make([]struct{}, constants.SwitchoverCheckSyncChangeFeedTaskCaughtUpRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckSyncChangeFeedTaskCaughtUpRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
		}
		err = wfnCheckSyncCaughtUp(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		if err == nil {
			return nil
		}
	}
	return err
}
