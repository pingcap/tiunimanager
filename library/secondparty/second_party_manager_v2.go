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

	"github.com/pingcap-inc/tiem/library/spec"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/library/framework"
)

var Manager SecondPartyService

type SecondPartyService interface {
	Init()
	ClusterDeploy(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, version string,
		configStrYaml string, timeoutS int, flags []string, workFlowNodeID string, password string) (operationID string, err error)
	ClusterScaleOut(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, configStrYaml string,
		timeoutS int, flags []string, workFlowNodeID string, password string) (operationID string, err error)
	ClusterScaleIn(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, nodeId string,
		timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterStart(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, timeoutS int,
		flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterRestart(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, timeoutS int,
		flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterStop(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, timeoutS int,
		flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterList(ctx context.Context, tiUPComponent TiUPComponentTypeStr, timeoutS int, flags []string) (
		resp *CmdListResp, err error)
	ClusterDestroy(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, timeoutS int,
		flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterDisplay(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, timeoutS int,
		flags []string) (resp *CmdDisplayResp, err error)
	ClusterUpgrade(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, version string,
		timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error)
	ClusterShowConfig(ctx context.Context, req *CmdShowConfigReq) (resp *CmdShowConfigResp, err error)
	ClusterEditGlobalConfig(ctx context.Context, cmdEditGlobalConfigReq CmdEditGlobalConfigReq, workFlowNodeID string) (
		string, error)
	ClusterEditInstanceConfig(ctx context.Context, cmdEditInstanceConfigReq CmdEditInstanceConfigReq,
		workFlowNodeID string) (string, error)
	ClusterReload(ctx context.Context, cmdReloadConfigReq CmdReloadConfigReq, workFlowNodeID string) (
		operationID string, err error)
	ClusterExec(ctx context.Context, cmdClusterExecReq CmdClusterExecReq, workFlowNodeID string) (
		operationID string, err error)
	// Check
	// @Description: tiup cluster check
	// @param ctx
	// @param componentType
	// @param checkObject, it could be cluster-name or topology-file
	// @param flags[]
	// @param timeoutS
	// @return resp json string
	// @return err
	Check(ctx context.Context, tiUPComponent TiUPComponentTypeStr, checkObject string, timeoutS int,
		flags []string) (result string, err error)
	Dumpling(ctx context.Context, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error)
	Lightning(ctx context.Context, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error)
	Transfer(ctx context.Context, tiUPComponent TiUPComponentTypeStr, instanceName string, collectorYaml string,
		remotePath string, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error)
	BackUp(ctx context.Context, cluster ClusterFacade, storage BrStorage, workFlowNodeID string) (operationID string,
		err error)
	ShowBackUpInfo(ctx context.Context, cluster ClusterFacade) CmdShowBackUpInfoResp
	ShowBackUpInfoThruMetaDB(ctx context.Context, operationID string) (resp CmdBrResp, err error)
	Restore(ctx context.Context, cluster ClusterFacade, storage BrStorage, workFlowNodeID string) (operationID string,
		err error)
	ShowRestoreInfo(ctx context.Context, cluster ClusterFacade) CmdShowRestoreInfoResp
	GetOperationStatus(ctx context.Context, operationID string) (resp GetOperationStatusResp, err error)
	GetOperationStatusByWorkFlowNodeID(ctx context.Context, workFlowNodeID string) (resp GetOperationStatusResp, err error)
	ApiEditConfig(ctx context.Context, apiEditConfigReq ApiEditConfigReq) (bool, error)
	EditClusterConfig(ctx context.Context, req ClusterEditConfigReq, workFlowNodeID string) error
	SetClusterDbPassword(ctx context.Context, req ClusterSetDbPswReq, workFlowNodeID string) error
	ClusterComponentCtl(ctx context.Context, str TiUPComponentTypeStr, clusterVersion string,
		component spec.TiDBClusterComponent, flags []string, timeoutS int) (string, error)
	CreateChangeFeedTask(ctx context.Context, req ChangeFeedCreateReq) (resp ChangeFeedCmdAcceptResp, err error)
	UpdateChangeFeedTask(ctx context.Context, req ChangeFeedUpdateReq) (resp ChangeFeedCmdAcceptResp, err error)
	PauseChangeFeedTask(ctx context.Context, req ChangeFeedPauseReq) (resp ChangeFeedCmdAcceptResp, err error)
	ResumeChangeFeedTask(ctx context.Context, req ChangeFeedResumeReq) (resp ChangeFeedCmdAcceptResp, err error)
	DeleteChangeFeedTask(ctx context.Context, req ChangeFeedDeleteReq) (resp ChangeFeedCmdAcceptResp, err error)
	QueryChangeFeedTasks(ctx context.Context, req ChangeFeedQueryReq) (resp ChangeFeedQueryResp, err error)
	DetailChangeFeedTask(ctx context.Context, req ChangeFeedDetailReq) (ChangeFeedDetailResp, error)
}

type SecondPartyManager struct {
	TiUPBinPath              string
	operationStatusCh        chan OperationStatusMember
	operationStatusMap       map[string]OperationStatusMapValue // key: operationID
	syncedOperationStatusMap map[string]OperationStatusMapValue // key: operationID
	operationStatusMapMutex  sync.Mutex
}

func (manager *SecondPartyManager) Init() {
	framework.Log().Infof("init secondpartymanager: %+v", manager)
	manager.syncedOperationStatusMap = make(map[string]OperationStatusMapValue)
	manager.operationStatusCh = make(chan OperationStatusMember, 1024)
	manager.operationStatusMap = make(map[string]OperationStatusMapValue)
	go manager.operationStatusMapSyncer()
}

func (manager *SecondPartyManager) GetOperationStatus(ctx context.Context, operationID string) (
	resp GetOperationStatusResp, err error) {
	framework.LogWithContext(ctx).WithField("operationid", operationID).Infof("getoperationstatus")
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Get(ctx, operationID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondPartyOperation:%v, err:%v", secondPartyOperation, err)
		return
	} else {
		resp.Status = secondPartyOperation.Status
		resp.Result = secondPartyOperation.Result
		resp.ErrorStr = secondPartyOperation.ErrorStr
		return
	}
}

func (manager *SecondPartyManager) GetOperationStatusByWorkFlowNodeID(ctx context.Context, workFlowNodeID string) (
	resp GetOperationStatusResp, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("getoperationstatusbybizid")
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().QueryByWorkFlowNodeID(ctx, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartypperation:%v, err:%v", secondPartyOperation, err)
		return
	} else {
		resp.Status = secondPartyOperation.Status
		resp.Result = secondPartyOperation.Result
		resp.ErrorStr = secondPartyOperation.ErrorStr
		return
	}
}
