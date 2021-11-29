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

package secondparty

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/sirupsen/logrus"
	"sync"
)

var SecondParty MicroSrv

var logger *logrus.Entry

type MicroSrv interface {
	MicroInit()
	MicroSrvTiupDeploy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupScaleOut(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, configStrYaml string, timeoutS int, flags []string, bizID uint64)(taskID uint64, err error)
	MicroSrvTiupScaleIn(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, nodeId string, timeoutS int, flags []string, bizID uint64) (taskId uint64, err error)
	MicroSrvTiupStart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupRestart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupStop(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupList(ctx context.Context, tiupComponent TiUPComponentTypeStr, timeoutS int, flags []string) (resp *CmdListResp, err error)
	MicroSrvTiupDestroy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupDisplay(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error)
	MicroSrvTiupTransfer(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, collectorYaml string, remotePath string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupUpgrade(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, version string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupShowConfig(ctx context.Context, req *CmdShowConfigReq) (resp *CmdShowConfigResp, err error)
	MicroSrvTiupEditGlobalConfig(ctx context.Context, cmdEditGlobalConfigReq CmdEditGlobalConfigReq, bizID uint64) (uint64, error)
	MicroSrvTiupEditInstanceConfig(ctx context.Context, cmdEditInstanceConfigReq CmdEditInstanceConfigReq, bizID uint64) (uint64, error)
	MicroSrvTiupReload(ctx context.Context, cmdReloadConfigReq CmdReloadConfigReq, bizID uint64) (taskID uint64, err error)
	MicroSrvDumpling(ctx context.Context, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvLightning(ctx context.Context, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvBackUp(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID uint64) (taskID uint64, err error)
	MicroSrvShowBackUpInfo(ctx context.Context, cluster ClusterFacade) CmdShowBackUpInfoResp
	MicroSrvRestore(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID uint64) (taskID uint64, err error)
	MicroSrvShowRestoreInfo(ctx context.Context, cluster ClusterFacade) CmdShowRestoreInfoResp
	MicroSrvGetTaskStatus(ctx context.Context, taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error)
	MicroSrvGetTaskStatusByBizID(ctx context.Context, bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error)
}

type SecondMicro struct {
	TiupBinPath         string
	taskStatusCh        chan TaskStatusMember
	taskStatusMap       map[uint64]TaskStatusMapValue
	syncedTaskStatusMap map[uint64]TaskStatusMapValue
	taskStatusMapMutex  sync.Mutex
}

func (secondMicro *SecondMicro) MicroInit() {
	logger = framework.LogForkFile(common.LogFileSecondParty)
	framework.Log().Infof("microinit secondmicro: %+v", secondMicro)
	secondMicro.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondMicro.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondMicro.taskStatusMap = make(map[uint64]TaskStatusMapValue)
	go secondMicro.taskStatusMapSyncer()
}

func (secondMicro *SecondMicro) MicroSrvGetTaskStatus(ctx context.Context, taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error) {
	framework.LogWithContext(ctx).WithField("taskid", taskID).Infof("microsrvgettaskstatus")
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = taskID
	rsp, err := client.DBClient.FindTiupTaskByID(context.Background(), &req)
	if err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("err:%s, rsp.ErrCode:%d, rsp.ErrStr:%s", err, rsp.ErrCode, rsp.ErrStr)
		return stat, "", err
	} else {
		assert(rsp.TiupTask != nil && rsp.TiupTask.ID == taskID)
		stat = rsp.TiupTask.Status
		errStr = rsp.TiupTask.ErrorStr
		return stat, errStr, nil
	}
}

func (secondMicro *SecondMicro) MicroSrvGetTaskStatusByBizID(ctx context.Context, bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvgettaskstatusbybizid")
	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = bizID
	rsp, err := client.DBClient.GetTiupTaskStatusByBizID(context.Background(), &req)
	if err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("err:%s, rsp.ErrCode:%d, rsp.ErrStr:%s", err, rsp.ErrCode, rsp.ErrStr)
		return stat, "", err
	} else {
		return rsp.Stat, rsp.StatErrStr, nil
	}
}
