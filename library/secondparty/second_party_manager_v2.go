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
* @File: second_party_manager_v2.go
* @Description:
* @Author: shenhaibo@pingcap.com
* @Version: 1.0.0
* @Date: 2021/12/7
*******************************************************************************/

package secondparty

import (
	"context"
	"fmt"
	"sync"

	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
)

var Manager SecondPartyService

//var logger *logrus.Entry

type SecondPartyService interface {
	Init()
	ClusterDeploy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterScaleOut(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, configStrYaml string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterScaleIn(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, nodeId string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterStart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterRestart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterStop(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterList(ctx context.Context, tiupComponent TiUPComponentTypeStr, timeoutS int, flags []string) (resp *CmdListResp, err error)
	ClusterDestroy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterDisplay(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error)
	ClusterUpgrade(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, version string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	ClusterShowConfig(ctx context.Context, req *CmdShowConfigReq) (resp *CmdShowConfigResp, err error)
	ClusterEditGlobalConfig(ctx context.Context, cmdEditGlobalConfigReq CmdEditGlobalConfigReq, bizID string) (uint64, error)
	ClusterEditInstanceConfig(ctx context.Context, cmdEditInstanceConfigReq CmdEditInstanceConfigReq, bizID string) (uint64, error)
	ClusterReload(ctx context.Context, cmdReloadConfigReq CmdReloadConfigReq, bizID string) (taskID uint64, err error)
	Transfer(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, collectorYaml string, remotePath string, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	Dumpling(ctx context.Context, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	Lightning(ctx context.Context, timeoutS int, flags []string, bizID string) (taskID uint64, err error)
	BackUp(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID string) (taskID uint64, err error)
	ShowBackUpInfo(ctx context.Context, cluster ClusterFacade) CmdShowBackUpInfoResp
	Restore(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID string) (taskID uint64, err error)
	ShowRestoreInfo(ctx context.Context, cluster ClusterFacade) CmdShowRestoreInfoResp
	GetTaskStatus(ctx context.Context, taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error)
	GetTaskStatusByBizID(ctx context.Context, bizID string) (stat dbPb.TiupTaskStatus, statErrStr string, err error)
}

type SecondPartyManager struct {
	TiupBinPath         string
	taskStatusCh        chan TaskStatusMember
	taskStatusMap       map[uint64]TaskStatusMapValue
	syncedTaskStatusMap map[uint64]TaskStatusMapValue
	taskStatusMapMutex  sync.Mutex
}

func (manager *SecondPartyManager) Init() {
	logger = framework.LogForkFile(common.LogFileSecondParty)
	framework.Log().Infof("microinit secondpartymanager: %+v", manager)
	manager.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	manager.taskStatusCh = make(chan TaskStatusMember, 1024)
	manager.taskStatusMap = make(map[uint64]TaskStatusMapValue)
	go manager.taskStatusMapSyncer()
}

func (manager *SecondPartyManager) GetTaskStatus(ctx context.Context, taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error) {
	framework.LogWithContext(ctx).WithField("taskid", taskID).Infof("microsrvgettaskstatus")
	var req dbPb.FindTiupOperatorRecordByIDRequest
	req.Id = taskID
	rsp, err := client.DBClient.FindTiupOperatorRecordByID(context.Background(), &req)
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

func (manager *SecondPartyManager) GetTaskStatusByBizID(ctx context.Context, bizID string) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvgettaskstatusbybizid")
	var req dbPb.GetTiupOperatorRecordStatusByBizIDRequest
	req.BizID = bizID
	rsp, err := client.DBClient.GetTiupOperatorRecordStatusByBizID(context.Background(), &req)
	if err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("err:%s, rsp.ErrCode:%d, rsp.ErrStr:%s", err, rsp.ErrCode, rsp.ErrStr)
		return stat, "", err
	} else {
		return rsp.Stat, rsp.StatErrStr, nil
	}
}
