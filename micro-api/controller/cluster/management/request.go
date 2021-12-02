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

package management

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type CreateReq struct {
	ClusterBaseInfo
	ClusterCommonDemand
	NodeDemandList []ClusterNodeDemand `json:"nodeDemandList"`
}

type DeleteReq struct {
	AutoBackup      bool `json:"autoBackup" form:"autoBackup"`
	ClearBackupData bool `json:"clearBackupData" form:"clearBackupData"`
}

type QueryReq struct {
	controller.PageRequest
	ClusterId     string `json:"clusterId" form:"clusterId"`
	ClusterName   string `json:"clusterName" form:"clusterName"`
	ClusterType   string `json:"clusterType" form:"clusterType"`
	ClusterStatus string `json:"clusterStatus" form:"clusterStatus"`
	ClusterTag    string `json:"clusterTag" form:"clusterTag"`
}

type TakeoverReq struct {
	TiupIp           string   `json:"tiupIp" example:"172.16.4.147" form:"tiupIp"`
	TiupPort         int      `json:"tiupPort" example:"22" form:"tiupPort"`
	TiupUserName     string   `json:"tiupUserName" example:"root" form:"tiupUserName"`
	TiupUserPassword string   `json:"tiupUserPassword" example:"password" form:"tiupUserPassword"`
	TiupPath         string   `json:"tiupPath" example:".tiup/" form:"tiupPath"`
	ClusterNames     []string `json:"clusterNames" form:"clusterNames"`
}

type ScaleOutReq struct {
	NodeDemandList []ClusterNodeDemand `json:"nodeDemandList"`
}

type ScaleInReq struct {
	ComponentType string `json:"componentType"`
	NodeId        string `json:"nodeId"`
}
