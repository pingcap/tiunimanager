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

/*******************************************************************************
 * @File: workflow.go
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
	"time"

	"github.com/pingcap/tiunimanager/common/constants"
	emerr "github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	changefeedModel "github.com/pingcap/tiunimanager/models/cluster/changefeed"
	workflowModel "github.com/pingcap/tiunimanager/models/workflow"
	workflow "github.com/pingcap/tiunimanager/workflow2"
)

const (
	wfContextReqKey                            string = "wfContextReqKey"
	wfContextOldSyncChangeFeedTaskIDKey        string = "wfContextOldSyncChangeFeedTaskIDKey"
	wfContextOtherSlavesMapToOldSyncCDCTaskKey string = "wfContextOtherSlavesMapToOldSyncCDCTaskKey"
	wfContextOtherSlavesMapToNewSyncCDCTaskKey string = "wfContextOtherSlavesMapToNewSyncCDCTaskKey"

	wfContextNewSyncChangeFeedTaskIDKey   string = "wfContextNewSyncChangeFeedTaskIDKey"
	wfContextCancelDeferStackKey          string = "wfContextCancelDeferStackKey"
	wfContextIsExecutingDeferStackFlagKey string = "wfContextIsExecutingDeferStackFlagKey"

	wfContextOldMasterPreviousMaintenanceStatusKey string = "wfContextOldMasterPreviousMaintenanceStatusKey"
	wfContextOldSlavePreviousMaintenanceStatusKey  string = "wfContextOldSlavePreviousMaintenanceStatusKey"
)

// wfGetReq workflow get request
func wfGetReq(ctx *workflow.FlowContext) *cluster.MasterSlaveClusterSwitchoverReq {
	var req cluster.MasterSlaveClusterSwitchoverReq
	err := ctx.GetData(wfContextReqKey, &req)
	if err != nil {
		return nil
	}
	return &req
}

func wfGetReqJson(ctx *workflow.FlowContext) string {
	bs, err := json.Marshal(wfGetReq(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("wfGetReqJson failed err:%s", err)
		return "-"
	}
	return string(bs)
}

// wfGetOtherSlavesMapToOldSyncCDCTask workflow get otherSlavesMapToOldSyncCDCTask
func wfGetOtherSlavesMapToOldSyncCDCTask(ctx *workflow.FlowContext) map[string]string {
	var m map[string]string
	var s string
	ctx.GetData(wfContextOtherSlavesMapToOldSyncCDCTaskKey, &s)
	if s != "" {
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			framework.Log().Panicf("wfGetOtherSlavesMapToOldSyncCDCTask: unmarshal failed, s:%s", s)
		}
	}
	return m
}

// wfSetOtherSlavesMapToOldSyncCDCTask workflow set otherSlavesMapToOldSyncCDCTask
/*
func wfSetOtherSlavesMapToOldSyncCDCTask(ctx *workflow.FlowContext, m map[string]string) {
	bs, err := json.Marshal(&m)
	if err != nil {
		framework.Log().Panicf("wfSetOtherSlavesMapToOldSyncCDCTask: marshal failed, err:%s", err)
	}
	ctx.SetData(wfContextOtherSlavesMapToOldSyncCDCTaskKey, string(bs))
}
*/
// wfGetOtherSlavesMapToNewSyncCDCTask workflow get otherSlavesMapToNewSyncCDCTask
func wfGetOtherSlavesMapToNewSyncCDCTask(ctx *workflow.FlowContext) map[string]string {
	var m map[string]string
	var s string
	ctx.GetData(wfContextOtherSlavesMapToNewSyncCDCTaskKey, &s)
	if s != "" {
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			framework.Log().Panicf("wfGetOtherSlavesMapToNewSyncCDCTask: unmarshal failed, s:%s", s)
		}
	}
	return m
}

// wfSetOtherSlavesMapToNewSyncCDCTask workflow set otherSlavesMapToNewSyncCDCTask
func wfSetOtherSlavesMapToNewSyncCDCTask(ctx *workflow.FlowContext, m map[string]string) {
	bs, err := json.Marshal(&m)
	if err != nil {
		framework.Log().Panicf("wfSetOtherSlavesMapToNewSyncCDCTask: marshal failed, err:%s", err)
	}
	ctx.SetData(wfContextOtherSlavesMapToNewSyncCDCTaskKey, string(bs))
}

func wfGetIsExecutingDeferStackFlag(ctx *workflow.FlowContext) bool {
	var s string
	ctx.GetData(wfContextIsExecutingDeferStackFlagKey, &s)
	if s != "" {
		if len(s) > 0 {
			return true
		}
	}
	return false
}

func wfSetExecutingDeferStackFlag(ctx *workflow.FlowContext) {
	ctx.SetData(wfContextIsExecutingDeferStackFlagKey, "true")
}

func myAssert(ctx context.Context, b bool) {
	if !b {
		framework.LogWithContext(ctx).Panicf("switchover: assert failed")
	}
}

type deferStack []string

func wfGetDeferStack(ctx *workflow.FlowContext) []string {
	var stk deferStack
	var s string
	ctx.GetData(wfContextCancelDeferStackKey, &s)
	if s != "" {
		err := json.Unmarshal([]byte(s), &stk)
		if err != nil {
			framework.Log().Panicf("wfGetDeferStack: unmarshal deferStack failed, s:%s", s)
		}
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
	var s string
	ctx.GetData(wfContextOldSyncChangeFeedTaskIDKey, &s)
	return s
}

func wfGetNewSyncChangeFeedTaskId(ctx *workflow.FlowContext) string {
	var s string
	ctx.GetData(wfContextNewSyncChangeFeedTaskIDKey, &s)
	if s != "" {
		return s
	} else {
		return ""
	}
}

func wfSetNewSyncChangeFeedTaskId(ctx *workflow.FlowContext, newTaskID string) {
	ctx.SetData(wfContextNewSyncChangeFeedTaskIDKey, newTaskID)
}

func wfGetOldMasterPreviousMaintenanceStatus(ctx *workflow.FlowContext) string {
	var s string
	ctx.GetData(wfContextOldMasterPreviousMaintenanceStatusKey, &s)
	return s
}

func wfGetOldSlavePreviousMaintenanceStatus(ctx *workflow.FlowContext) string {
	var s string
	ctx.GetData(wfContextOldSlavePreviousMaintenanceStatusKey, &s)
	return s
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
	otherSlavesMapToNewSyncCDCTask := wfGetOtherSlavesMapToNewSyncCDCTask(ctx)
	oldMasterClusterId := wfGetOldMasterClusterId(ctx)
	otherSlavesMapToNewSyncCDCTask[oldMasterClusterId] = newSyncChangeFeedTaskId
	err := mgr.swapClusterRelationsInDB(ctx, wfGetOldMasterClusterId(ctx), wfGetOldSlaveClusterId(ctx), otherSlavesMapToNewSyncCDCTask)
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
	framework.LogWithContext(ctx).Infof(
		"%s newMasterClusterId:%s, newSlaveClusterId:%s", funcName, wfGetNewMasterClusterId(ctx), wfGetNewSlaveClusterId(ctx))
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
	framework.LogWithContext(ctx).Infof(
		"%s clusterGetOneConnectIPPort clusterID:%s, ip:%s port:%d err:%v", funcName, wfGetNewMasterClusterId(ctx), ip, port, err2)
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
			"%s createChangeFeedTask req:%s success, taskID:%s", funcName, wfGetReqJson(ctx), newTaskId)
	}

	return newTaskId, err
}

/*
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
*/
func wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfnMigrateAllDownStreamSyncChangeFeedTasksToNewMaster"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	originCDCTaskIDMapToNewCDCID, err := wfnMigrateAllDownStreamSyncChangeFeedTasks(node, ctx,
		wfGetOldMasterClusterId(ctx), wfGetNewMasterClusterId(ctx), wfGetOldSyncChangeFeedTaskId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	}
	otherSlavesMapToOldSyncCDCTaskID := wfGetOtherSlavesMapToOldSyncCDCTask(ctx)
	otherSlavesMapToNewSyncCDCTaskID := make(map[string]string)
	for otherSlaveID, oldSyncCDCID := range otherSlavesMapToOldSyncCDCTaskID {
		newSyncCDCID, ok := originCDCTaskIDMapToNewCDCID[oldSyncCDCID]
		if !ok {
			err := fmt.Errorf("oldSyncCDCID %s map to newSyncCDCID not found", oldSyncCDCID)
			framework.LogWithContext(ctx).Errorf("%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
			return err
		} else {
			otherSlavesMapToNewSyncCDCTaskID[otherSlaveID] = newSyncCDCID
		}
	}
	if len(otherSlavesMapToOldSyncCDCTaskID) != len(otherSlavesMapToNewSyncCDCTaskID) {
		err := fmt.Errorf("len(otherSlavesMapToOldSyncCDCTaskID) != len(otherSlavesMapToNewSyncCDCTaskID): %d != %d",
			len(otherSlavesMapToOldSyncCDCTaskID), len(otherSlavesMapToNewSyncCDCTaskID))
		framework.LogWithContext(ctx).Errorf("%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	}
	for otherSlaveID := range otherSlavesMapToOldSyncCDCTaskID {
		newSyncCDCID, ok := otherSlavesMapToNewSyncCDCTaskID[otherSlaveID]
		if !ok {
			err := fmt.Errorf("newSyncCDCID of otherSlaveID %s not found in otherSlavesMapToNewSyncCDCTaskID", otherSlaveID)
			framework.LogWithContext(ctx).Errorf("%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
			return err
		} else {
			if len(newSyncCDCID) <= 0 {
				err := fmt.Errorf("length of newSyncCDCID <= 0: %d", len(newSyncCDCID))
				framework.LogWithContext(ctx).Errorf("%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
				return err
			}
		}
	}
	wfSetOtherSlavesMapToNewSyncCDCTask(ctx, otherSlavesMapToNewSyncCDCTaskID)
	return nil
}

func wfnMigrateAllDownStreamSyncChangeFeedTasksBackToOldMaster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	_, err := wfnMigrateAllDownStreamSyncChangeFeedTasks(node, ctx,
		wfGetNewMasterClusterId(ctx), wfGetOldMasterClusterId(ctx), wfGetNewSyncChangeFeedTaskId(ctx))
	return err
}

// ret map[oldCDCTaskID]NewCDCTaskID,err
// the map is the superset of otherSlavesMapToOldSyncCDCTask and doesn't contain the old sync cdc task between old master and old slave
func wfnMigrateAllDownStreamSyncChangeFeedTasks(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext,
	fromCluster, toCluster, exceptSyncChangeFeedTaskId string) (map[string]string, error) {
	funcName := "wfnMigrateAllDownStreamSyncChangeFeedTasks"
	framework.LogWithContext(ctx).Infof("start %s: fromCluster %s, toCluster: %s, exceptSyncChangeFeedTaskId: %s",
		funcName, fromCluster, toCluster, exceptSyncChangeFeedTaskId)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)

	if len(fromCluster) == 0 || len(toCluster) == 0 {
		return nil, fmt.Errorf("%s: unexpected clusterID len, fromCluster:%s toCluster:%s", funcName, fromCluster, toCluster)
	}

	tasks, err := mgr.getAllChangeFeedTasksOnCluster(ctx, fromCluster)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return nil, err
	} else {
		var ids []string
		for _, v := range tasks {
			ids = append(ids, v.ID)
		}
		framework.LogWithContext(ctx).Infof(
			"%s req:%s success taskIDs:%s", funcName, wfGetReqJson(ctx), ids)
	}
	var finalTasks []*cluster.ChangeFeedTask
	for _, v := range tasks {
		if v != nil && v.ID != exceptSyncChangeFeedTaskId {
			finalTasks = append(finalTasks, v)
		}
	}
	{
		var ids []string
		for _, v := range finalTasks {
			ids = append(ids, v.ID)
		}
		framework.LogWithContext(ctx).Infof(
			"%s req:%s finalTaskIDs:%s", funcName, wfGetReqJson(ctx), ids)
	}
	var cancelFps []func()
	cancelFlag := true
	cancel := func() {
		if cancelFlag {
			for _, fp := range cancelFps {
				fp()
			}
		}
	}
	defer cancel()
	for _, v := range finalTasks {
		err := mgr.pauseChangeFeedTask(ctx, v.ID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s pauseChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), v.ID, err)
			return nil, err
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s pauseChangeFeedTask req:%s taskId:%s success", funcName, wfGetReqJson(ctx), v.ID)
		}
		thisCDCID := v.ID
		cancelFps = append(cancelFps, func() {
			err := mgr.resumeChangeFeedTask(ctx, thisCDCID)
			if err != nil {
				framework.LogWithContext(ctx).Errorf(
					"%s resumeChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), thisCDCID, err)
			} else {
				framework.LogWithContext(ctx).Infof(
					"%s resumeChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
			}
		})
	}
	retMap := make(map[string]string)
	for _, v := range finalTasks {
		originID := v.ID
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
				"%s createChangeFeedTask req:%s originTaskId:%s err:%s", funcName, wfGetReqJson(ctx), originID, err)
			return nil, err
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s createChangeFeedTask req:%s success, originTaskId:%s, ret new taskId:%s", funcName, wfGetReqJson(ctx), originID, id)
			retMap[originID] = id
		}
		cancelFps = append(cancelFps, func() {
			err := mgr.removeChangeFeedTask(ctx, id)
			if err != nil {
				framework.LogWithContext(ctx).Errorf(
					"%s removeChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), id, err)
			} else {
				framework.LogWithContext(ctx).Infof(
					"%s removeChangeFeedTask req:%s success", funcName, wfGetReqJson(ctx))
			}
		})
	}
	cancelFlag = false
	for originCDCID := range retMap {
		err := mgr.removeChangeFeedTask(ctx, originCDCID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s removeChangeFeedTask req:%s taskId:%s err:%s", funcName, wfGetReqJson(ctx), originCDCID, err)
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s removeChangeFeedTask req:%s taskId:%s, success", funcName, wfGetReqJson(ctx), originCDCID)
		}
	}
	return retMap, err
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
		framework.LogWithContext(ctx).Infof(
			"%s req:%s nowT.Sub(slaveT):%s", funcName, wfGetReqJson(ctx), nowT.Sub(slaveT))
		minSyncT := nowT.Add(-constants.SwitchoverCheckMasterSlaveMaxLagTime)
		if slaveT.After(minSyncT) {
			return nil
		} else {
			return fmt.Errorf("%s slave checkpoint lag time is bigger than %v, nowT:%s, slaveT:%s", funcName,
				constants.SwitchoverCheckMasterSlaveMaxLagTime, nowT, slaveT)
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

func wfnCheckCDCsCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, changeFeedTaskIDs []string) error {
	var err error
	funcName := "wfnCheckCDCsCaughtUp"
	for _, changeFeedTaskID := range changeFeedTaskIDs {
		err = nil
		for i := range make([]struct{}, constants.SwitchoverCheckChangeFeedTaskCaughtUpRetriesCount+1) {
			if err != nil {
				time.Sleep(constants.SwitchoverCheckChangeFeedTaskCaughtUpRetryWait)
				framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
			}
			err = wfnCheckCDCCaughtUp(node, ctx, changeFeedTaskID)
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
	}
	return err
}

func wfnCheckCDCCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, changeFeedTaskId string) error {
	funcName := "wfnCheckCDCCaughtUp"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	for retryCt := range make([]struct{}, constants.SwitchoverCheckChangeFeedTaskCaughtUpMakeSureRetriesCount+1) {
		time.Sleep(constants.SwitchoverCheckChangeFeedTaskCaughtUpMaxLagTime)
		err := wfnCheckCDCCaughtUpHelper(node, ctx, changeFeedTaskId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"%s call wfnCheckCDCCaughtUpHelper retryCt %d failed, req:%s err:%s", funcName, retryCt, wfGetReqJson(ctx), err)
			return err
		} else {
			framework.LogWithContext(ctx).Infof(
				"%s call wfnCheckCDCCaughtUpHelper retryCt %d success, req:%s err:%s", funcName, retryCt, wfGetReqJson(ctx), err)
		}
	}
	return nil
}

func wfnCheckCDCCaughtUpHelper(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, changeFeedTaskId string) error {
	funcName := "wfnCheckCDCCaughtUpHelper"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	var slaveT time.Time
	var firstInfo *cluster.ChangeFeedTaskInfo
	var err error
	firstInfo, err = mgr.queryChangeFeedTask(ctx, changeFeedTaskId)
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
		framework.LogWithContext(ctx).Infof(
			"%s req:%s nowT.Sub(slaveT):%s", funcName, wfGetReqJson(ctx), nowT.Sub(slaveT))
		minSyncT := nowT.Add(-constants.SwitchoverCheckChangeFeedTaskCaughtUpMaxLagTime)
		if slaveT.After(minSyncT) {
			return nil
		} else {
			return fmt.Errorf("%s slave checkpoint lag time is bigger than %v, nowT:%s, slaveT:%s", funcName,
				constants.SwitchoverCheckChangeFeedTaskCaughtUpMaxLagTime, nowT, slaveT)
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
		if !ok {
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

func wfnEndMaintenanceWithClusterIDs(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, clusterIDs []string) error {
	funcName := "wfnEndMaintenanceWithClusterIDs"
	var errToRet error

	for _, clusterID := range clusterIDs {
		framework.LogWithContext(ctx).Infof("%s clusterID:%s", funcName, clusterID)
		metaOfCluster, err := meta.Get(ctx, clusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("%s get meta of cluster %s failed:%s", funcName, clusterID, err.Error())
			framework.LogWithContext(ctx).Errorf("%s end maintenance of cluster %s failed:%s", funcName, clusterID, err.Error())
			errToRet = err
			continue
		}

		if err := metaOfCluster.EndMaintenance(ctx, constants.ClusterMaintenanceStatus(wfGetOldSlavePreviousMaintenanceStatus(ctx))); err != nil {
			errToRet = err
			framework.LogWithContext(ctx).Errorf("%s end maintenance of cluster %s failed:%s", funcName, clusterID, err.Error())
		}
	}

	return errToRet
}

func wfnEndMaintenance(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfnEndMaintenance"
	var errToRet error
	metaOfSource, err := meta.Get(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Warnf("%s get meta of cluster %s failed:%s",
			funcName, wfGetOldMasterClusterId(ctx), err)
		errToRet = emerr.NewErrorf(emerr.TIUNIMANAGER_MASTER_SLAVE_SWITCHOVER_FAILED,
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
		framework.LogWithContext(ctx).Warnf("%s get meta of cluster %s failed:%s",
			funcName, wfGetOldSlaveClusterId(ctx), err)
		errToRet = emerr.NewErrorf(emerr.TIUNIMANAGER_MASTER_SLAVE_SWITCHOVER_FAILED,
			"get meta of %s failed, %s", wfGetOldSlaveClusterId(ctx), err.Error())
	} else {
		previousStatus := constants.ClusterMaintenanceStatus(wfGetOldSlavePreviousMaintenanceStatus(ctx))
		err := metaOfTarget.EndMaintenance(ctx, previousStatus)
		if err != nil {
			errToRet = err
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", wfGetOldSlaveClusterId(ctx), err.Error())
		}
	}

	for otherSlaveClusterID := range wfGetOtherSlavesMapToOldSyncCDCTask(ctx) {
		metaOfOtherSlave, err := meta.Get(ctx, otherSlaveClusterID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get meta of cluster %s failed:%s", otherSlaveClusterID, err.Error())
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", otherSlaveClusterID, err.Error())
			errToRet = err
			continue
		}

		if err := metaOfOtherSlave.EndMaintenance(ctx, constants.ClusterMaintenanceStatus(wfGetOldSlavePreviousMaintenanceStatus(ctx))); err != nil {
			errToRet = err
			framework.LogWithContext(ctx).Errorf("end maintenance of cluster %s failed:%s", otherSlaveClusterID, err.Error())
		}
	}

	return errToRet
}

func wfnGetSwitchoverMasterSlavesStateFromASwitchoverWorkflow(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext, oldWorkFlowID string) (*switchoverMasterSlavesState, bool, error) {
	funcName := "wfnGetSwitchoverMasterSlavesStateFromASwitchoverWorkflow"
	framework.LogWithContext(ctx).Infof("enter %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	s, flag, err := mgr.getSwitchoverMasterSlavesStateFromASwitchoverWorkflow(ctx, oldWorkFlowID)
	return s, flag, err
}

func wfStepRollback(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepRollback"
	framework.LogWithContext(ctx).Infof("enter %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)

	req := wfGetReq(ctx)
	oldWorkflowID := req.RollbackWorkFlowID
	state, workflowRollbackSucceedFlag, err := wfnGetSwitchoverMasterSlavesStateFromASwitchoverWorkflow(node, ctx, oldWorkflowID)
	if err != nil {
		return err
	}
	{
		var allClusterIDs []string
		allClusterIDs = append(allClusterIDs, state.OldMasterClusterID)
		for slaveID := range state.OldSlavesClusterIDMapToSyncTaskID {
			allClusterIDs = append(allClusterIDs, slaveID)
		}
		defer wfnEndMaintenanceWithClusterIDs(node, ctx, allClusterIDs)
	}
	if workflowRollbackSucceedFlag {
		node.Record("The previous workflow has already been rollbacked successfully.")
		return nil
	}
	// previous failed switchover: A->B on clusters(A, B, C)
	// set all cluster readonly
	err = mgr.clusterSetReadonly(ctx, state.OldMasterClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("%s master clusterSetReadonly failed, clusterID:%s", funcName, state.OldMasterClusterID)
		return err
	}
	for slaveID := range state.OldSlavesClusterIDMapToSyncTaskID {
		err = mgr.clusterSetReadonly(ctx, slaveID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("%s slave clusterSetReadonly failed, clusterID:%s", funcName, slaveID)
			return err
		}
	}
	// delete all cdc on B C
	for slaveID := range state.OldSlavesClusterIDMapToSyncTaskID {
		ids, err := mgr.getAllChangeFeedTaskIDsOnCluster(ctx, slaveID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			skipFlag := false
			for _, previousCDCID := range state.OldSlavesClusterIDMapToCDCIDs[slaveID] {
				if id == previousCDCID {
					skipFlag = true
					break
				}
			}
			if skipFlag {
				continue
			}
			err := mgr.removeChangeFeedTask(ctx, id)
			if err != nil {
				return err
			}
		}
	}
	// filter
	previousCDCsM := make(map[string]*cluster.ChangeFeedTask)
	for _, v := range state.CDCsOnMaster {
		previousCDCsM[v.ID] = v
	}
	alreadyExistCDCs := make(map[string]bool)
	{
		cdcs, err := mgr.getAllChangeFeedTasksOnCluster(ctx, state.OldMasterClusterID)
		if err != nil {
			return err
		}
		for _, task := range cdcs {
			if previousCDCsM[task.ID] != nil {
				// preserve old cdc
				alreadyExistCDCs[task.ID] = true
				err := mgr.resumeChangeFeedTask(ctx, task.ID)
				if err != nil {
					return err
				}
			} else {
				// delete unrecognized cdc
				err := mgr.removeChangeFeedTask(ctx, task.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	oldCDCMapToNewlyCreatedCDC := make(map[string]string)
	for k, v := range previousCDCsM {
		if alreadyExistCDCs[k] {
			continue
		}
		v.ID = ""
		v.StartTS = "0"
		var zeroT time.Time
		v.Status = constants.ChangeFeedStatusNormal.ToString()
		v.CreateTime = zeroT
		v.UpdateTime = zeroT

		newTaskID, err := mgr.createChangeFeedTask(ctx, v)
		if err != nil {
			return err
		}
		oldCDCMapToNewlyCreatedCDC[k] = newTaskID
	}
	oldSyncTaskIDMapToNewID := make(map[string]string)
	for _, v := range state.OldSlavesClusterIDMapToSyncTaskID {
		oldCDCTaskID := v
		if alreadyExistCDCs[oldCDCTaskID] {
			oldSyncTaskIDMapToNewID[oldCDCTaskID] = oldCDCTaskID
			continue
		}
		newID, ok := oldCDCMapToNewlyCreatedCDC[oldCDCTaskID]
		if !ok {
			return fmt.Errorf("oldCDC %s map to new newCDC not found", oldCDCTaskID)
		}
		if len(newID) <= 0 {
			return fmt.Errorf("newCDCID is invalid, oldCDC: %s", oldCDCTaskID)
		}
		oldSyncTaskIDMapToNewID[oldCDCTaskID] = newID
	}
	fixSlaveIDMapToCDCID := make(map[string]string)
	for slaveID, oldCDCID := range state.OldSlavesClusterIDMapToSyncTaskID {
		fixSlaveIDMapToCDCID[slaveID] = oldSyncTaskIDMapToNewID[oldCDCID]
		delete(oldSyncTaskIDMapToNewID, oldCDCID)
	}
	err = mgr.clusterSetReadWrite(ctx, state.OldMasterClusterID)
	if err != nil {
		return err
	}
	return mgr.relationsResetSyncChangeFeedTaskIDs(ctx, state.OldMasterClusterID, fixSlaveIDMapToCDCID)
}

func wfStepMarshalSwitchoverMasterSlavesState(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	str, err := mgr.marshalSwitchoverMasterSlavesState(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		return err
	}
	node.Record(fmt.Sprintf("%s\n", str))
	return nil
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
	err = wfnEndMaintenance(node, ctx)
	if err != nil {
		errToRet = err
	}
	return errToRet
}

func wfStepFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepFail"
	framework.LogWithContext(ctx).Errorf("%s", funcName)
	err := wfnExecuteDeferStack(node, ctx)
	err2 := wfnEndMaintenance(node, ctx)
	if err == nil && err2 == nil {
		//node.Record("Rollback Successfully.")
		return nil
	}
	finalErr := fmt.Errorf("wfStepFail: err:%s err2:%s", err, err2)
	//node.Record("Rollback Failed:", finalErr)
	return finalErr
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

/*
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
*/
func wfStepCheckNewSyncChangeFeedTaskHealth(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepCheckNewSyncChangeFeedTaskHealth"
	for i := range make([]struct{}, constants.SwitchoverCheckSyncChangeFeedTaskHealthRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckSyncChangeFeedTaskHealthRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d, previous err:%v", funcName, i, err)
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

/*
func wfStepCheckSyncCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepCheckSyncCaughtUp"
	for i := range make([]struct{}, constants.SwitchoverCheckChangeFeedTaskCaughtUpRetriesCount+1) {
		if err != nil {
			time.Sleep(constants.SwitchoverCheckChangeFeedTaskCaughtUpRetryWait)
			framework.LogWithContext(ctx).Infof("%s retry %d", funcName, i)
		}
		err = wfnCheckCDCCaughtUp(node, ctx, wfGetOldSyncChangeFeedTaskId(ctx))
		if err == nil {
			return nil
		}
	}
	return err
}
*/
func wfStepWaitOldMasterCDCsCaughtUp(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var err error
	funcName := "wfStepWaitOldMasterCDCsCaughtUp"
	ids, err := mgr.getAllChangeFeedTaskIDsOnCluster(ctx, wfGetOldMasterClusterId(ctx))
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s getAllChangeFeedTaskIDsOnCluster req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s getAllChangeFeedTaskIDsOnCluster req:%s success", funcName, wfGetReqJson(ctx))
	}
	err = wfnCheckCDCsCaughtUp(node, ctx, ids)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"%s wfnCheckCDCsCaughtUp req:%s err:%s", funcName, wfGetReqJson(ctx), err)
		return err
	} else {
		framework.LogWithContext(ctx).Infof(
			"%s wfnCheckCDCsCaughtUp req:%s success", funcName, wfGetReqJson(ctx))
	}
	return err
}

type wfStepCallback func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error

func wfGenStepWithRollbackCB(fp, rollbackFp wfStepCallback) func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	return func(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
		funcName := "wfStepWithRollbackCB"
		framework.LogWithContext(ctx).Infof("enter %s", funcName)
		defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
		err0 := fp(node, ctx)
		if err0 != nil {
			framework.LogWithContext(ctx).Errorf("%s start rollback, previous err:%v", funcName, err0)
			err1 := rollbackFp(node, ctx)
			framework.LogWithContext(ctx).Infof("%s rollback err:%v", funcName, err1)
			if err1 == nil {
				err := fmt.Errorf("err:%v \n%s", err0, constants.SwitchoverRollbackSuccessInfoString)
				return err
			} else {
				err := fmt.Errorf("err:%v \nRollback Failed:%v", err0, err1)
				return err
			}
		}
		return nil
	}
}

func wfStepNOP(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	funcName := "wfStepNOP"
	framework.LogWithContext(ctx).Infof("enter %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	return nil
}
