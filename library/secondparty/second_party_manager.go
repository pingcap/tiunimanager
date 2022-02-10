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
* @File: second_party_manager.go
* @Description:
* @Author: shenhaibo@pingcap.com
* @Version: 1.0.0
* @Date: 2021/12/7
*******************************************************************************/

package secondparty

import (
	"context"
	"sync"

	"github.com/pingcap-inc/tiem/library/framework"
)

var Manager SecondPartyService

type SecondPartyService interface {
	Init()

	// TODO: @istudies move to util/api
	ApiEditConfig(ctx context.Context, apiEditConfigReq ApiEditConfigReq) (bool, error)
	// TODO: @istudies, @haiboumich move to util/api/tidb/sql
	EditClusterConfig(ctx context.Context, req ClusterEditConfigReq, workFlowNodeID string) error

	// TODO: @zhangpeijin-milo, move to util/api/cdc
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
