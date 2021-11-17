
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
 *                                                                            *
 ******************************************************************************/

package backuprestore

import (
	"time"

	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
)

type BackupRecordQueryReq struct {
	controller.PageRequest
	ClusterId string `json:"clusterId" form:"clusterId"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
}

type BackupDeleteReq struct {
	ClusterId string `json:"clusterId"`
}

type BackupStrategy struct {
	ClusterId      string    `json:"clusterId"`
	BackupDate     string    `json:"backupDate"`
	Period         string    `json:"period"`
	NextBackupTime time.Time `json:"nextBackupTime"`
}

type BackupStrategyUpdateReq struct {
	Strategy BackupStrategy `json:"strategy"`
}

type BackupReq struct {
	ClusterId    string `json:"clusterId"`
	BackupType   string `json:"backupType"`
	BackupMethod string `json:"backupMethod"`
	FilePath     string `json:"filePath"`
}
type BackupRecoverReq struct {
	ClusterId string `json:"clusterId"`
}

type RestoreReq struct {
	management.ClusterBaseInfo
	management.ClusterCommonDemand
	NodeDemandList []management.ClusterNodeDemand `json:"nodeDemandList"`
}
