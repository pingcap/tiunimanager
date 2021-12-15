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
	"encoding/json"
	"fmt"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	changefeedModel "github.com/pingcap-inc/tiem/models/cluster/changefeed"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	wfContextReqKey                     string = "wfContextReqKey"
	wfContextOldSyncChangeFeedTaskIDKey string = "wfContextOldSyncChangeFeedTaskIDKey"
	wfContextNewSyncChangeFeedTaskIDKey string = "wfContextNewSyncChangeFeedTaskIDKey"
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
	newTask.Id = ""
	newTask.Name = ""
	newTask.ClusterId = wfGetNewMasterClusterId(ctx)
	newTask.StartTS = 0
	newTask.Status = string(constants.Normal)
	newTask.CreateTime = zeroT
	newTask.UpdateTime = zeroT
	tidbDownStream := newTask.Downstream.(*changefeedModel.TiDBDownstream)

	var userName, password, ip string
	var port int
	var tls bool
	userName, password, err = mgr.clusterGetMysqlUserNameAndPwd(ctx, wfGetNewSlaveClusterId(ctx))
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

func wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, exceptOldSyncChangeFeedTaskId string) error {
	funcName := "wfnCloneAllDownStreamSyncChangeFeedTasksToNewMaster"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	tasks, err := mgr.getAllChangeFeedTasksOnCluster(ctx, wfGetOldMasterClusterId(ctx))
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
		if v != nil && v.Id != exceptOldSyncChangeFeedTaskId {
			finalTasks = append(finalTasks, v)
		}
	}
	for _, v := range finalTasks {
		err := mgr.pauseChangeFeedTask(ctx, v.Id)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s pauseChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), v.Id, err)
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s pauseChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
		}
	}
	for _, v := range finalTasks {
		var zeroT time.Time
		v.ClusterId = wfGetNewMasterClusterId(ctx)
		v.Id = ""
		v.Name = ""
		v.ClusterId = wfGetNewMasterClusterId(ctx)
		v.StartTS = 0
		v.Status = string(constants.Normal)
		v.CreateTime = zeroT
		v.UpdateTime = zeroT
		id, err := mgr.createChangeFeedTask(ctx, v)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s pauseChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), v.Id, err)
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
	if firstInfo.Status != constants.Normal.ToString() {
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

func wfStepFinish(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepFinish"
	framework.LogWithContext(ctx).Infof("%s", funcName)
	return nil
}

func wfStepFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepFail"
	framework.LogWithContext(ctx).Errorf("%s", funcName)
	return nil
}

func wfStepSetNewMasterReadWrite(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnSetNewMasterReadWrite(node, ctx)
}

func wfStepSetOldMasterReadOnly(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnSetOldMasterReadOnly(node, ctx)
}

func wfStepSwapMasterSlaveRelationInDB(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnSwapMasterSlaveRelationInDB(node, ctx, wfGetNewSyncChangeFeedTaskId(ctx))
}

func wfStepPauseOldSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return wfnPauseSyncChangeFeedTask(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
}

func wfStepCreateReverseSyncChangeFeedTask(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
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
	return wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
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
