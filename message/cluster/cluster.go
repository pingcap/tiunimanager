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

/*******************************************************************************
 * @File: cluster.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package cluster

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
)

//CreateClusterReq Message for creating a new cluster
type CreateClusterReq struct {
	structs.CreateClusterParameter
	ResourceParameter structs.ClusterResourceInfo `json:"resourceParameters" form:"resourceParameters"`
}

// CreateClusterResp Reply message for creating a new cluster
type CreateClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID                     string `json:"clusterId"`
}

// DeleteClusterReq Message for delete a new cluster
type DeleteClusterReq struct {
	ClusterID       string `json:"clusterID"`
	AutoBackup      bool   `json:"autoBackup" form:"autoBackup"`
	ClearBackupData bool   `json:"clearBackupData" form:"clearBackupData"`
}

// DeleteClusterResp Reply message for delete a new cluster
type DeleteClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID                     string `json:"clusterID"`
}

// StopClusterReq Message for stop a new cluster
type StopClusterReq struct {
	ClusterID string `json:"clusterId"`
}

// StopClusterResp Reply message for stop a new cluster
type StopClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID                     string `json:"clusterId"`
}

// RestartClusterReq Message for restart a new cluster
type RestartClusterReq struct {
	ClusterID string `json:"clusterId"`
}

// RestartClusterResp Reply message for restart a new cluster
type RestartClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID                     string `json:"clusterId"`
}

// ScaleInClusterReq Message for delete an instance in the cluster
type ScaleInClusterReq struct {
	ClusterID  string `json:"clusterId" form:"clusterId"`
	InstanceID string `json:"instanceId"  form:"instanceId"`
}

// ScaleInClusterResp Reply message for delete an instance in the cluster
type ScaleInClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// ScaleOutClusterReq Message for cluster expansion operation
type ScaleOutClusterReq struct {
	ClusterID                        string `json:"clusterId" form:"clusterId"`
	structs.ClusterResourceInfo
}

// ScaleOutClusterResp Reply message for cluster expansion operation
type ScaleOutClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

//RestoreNewClusterReq Restore to a new cluster message using the backup file
type RestoreNewClusterReq struct {
	structs.CreateClusterParameter
	BackupID          string                      `json:"backupId"`
	ResourceParameter structs.ClusterResourceInfo `json:"resourceParameters"`
}

//RestoreNewClusterResp Restore to a new cluster using the backup file Reply Message
type RestoreNewClusterResp struct {
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
	ClusterID                     string `json:"clusterID"`
}

//RestoreExistClusterReq Restore to exist cluster message using the backup file
type RestoreExistClusterReq struct {
	ClusterID string `json:"clusterID"`
	BackupID  string `json:"backupID"`
}

//RestoreExistClusterResp Restore to exist cluster using the backup file Reply Message
type RestoreExistClusterResp struct {
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
}

// CloneClusterReq Message for clone a new cluster
type CloneClusterReq struct {
	structs.CreateClusterParameter
	ParamGroupID    int64                          `json:"paramGroupId"`    // specify cloned cluster parameter group id(option)
	CloneStrategy   constants.ClusterCloneStrategy `json:"cloneStrategy"`   // specify clone strategy, include empty, snapshot and sync, default empty(option)
	SourceClusterID string                         `json:"sourceClusterId"` // specify source cluster id(require)
}

// CloneClusterResp Reply message for clone a new cluster
type CloneClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// MasterSlaveClusterSwitchoverReq Master and slave cluster switchover messages
type MasterSlaveClusterSwitchoverReq struct {
	SourceClusterID string `json:"sourceClusterID"`
	TargetClusterID string `json:"targetClusterID"`
	Force           bool   `json:"force"`
}

// MasterSlaveClusterSwitchoverResp Master and slave cluster switchover reply message
type MasterSlaveClusterSwitchoverResp struct {
	structs.AsyncTaskWorkFlowInfo
}

type QueryUpgradeVersionDiffInfoReq struct {
	ClusterID string `json:"clusterId"`
	Version   string `json:"version"`
}

type QueryUpgradeVersionDiffInfoResp struct {
	ConfigDiffInfos []structs.ProductUpgradeVersionConfigDiffItem `json:"configDiffInfos"`
}

type ClusterUpgradeVersionConfigItem struct {
	Name         string `json:"name"`
	InstanceType string `json:"instanceType"`
	Value        string `json:"value"`
}

type ClusterUpgradeReq struct {
	ClusterID     string `json:"ClusterId"`
	TargetVersion string `json:"targetVersion"`
	Configs       []ClusterUpgradeVersionConfigItem
}

type ClusterUpgradeResp struct {
	structs.AsyncTaskWorkFlowInfo
}

// TakeoverClusterReq Requests to take over an existing TiDB cluster, requiring TiDB version >= 4.0 when taking over
type TakeoverClusterReq struct {
	TiUPIp           string   `json:"TiUPIp" example:"172.16.4.147" form:"TiUPIp"`
	TiUPPort         int      `json:"TiUPPort" example:"22" form:"TiUPPort"`
	TiUPUserName     string   `json:"TiUPUserName" example:"root" form:"TiUPUserName"`
	TiUPUserPassword string   `json:"TiUPUserPassword" example:"password" form:"TiUPUserPassword"`
	TiUPPath         string   `json:"TiUPPath" example:".tiup/" form:"TiUPPath"`
	ClusterNames     []string `json:"clusterNames" form:"clusterNames"`
}

// TakeoverClusterResp Reply message for takeover a cluster
type TakeoverClusterResp struct {

	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
	Clusters                      []structs.ClusterInfo `json:"clusters"`
}

// QueryClustersReq Query cluster list messages
type QueryClustersReq struct {
	structs.PageRequest
	ClusterID string `json:"clusterId" form:"clusterId"`
	Name      string `json:"clusterName" form:"clusterName"`
	Type      string `json:"clusterType" form:"clusterType"`
	Status    string `json:"clusterStatus" form:"clusterStatus"`
	Tag       string `json:"clusterTag" form:"clusterTag"`
}

// QueryClusterResp Query the cluster list to reply to messages
type QueryClusterResp struct {
	Clusters []structs.ClusterInfo `json:"clusters"`
}

// QueryClusterDetailReq Query cluster detail messages
type QueryClusterDetailReq struct {
	ClusterID string `json:"clusterId" form:"clusterId"`
}

// QueryClusterDetailResp Query the cluster detail to reply to messages
type QueryClusterDetailResp struct {
	Info structs.ClusterInfo `json:"info"`
	structs.ClusterTopologyInfo
	structs.ClusterResourceInfo
}

// QueryMonitorInfoReq Message to query the monitoring address information of a cluster
type QueryMonitorInfoReq struct {
	ClusterID string `json:"clusterId" example:"abc"`
}

// QueryMonitorInfoResp Reply message for querying the monitoring address information of the cluster
type QueryMonitorInfoResp struct {
	ClusterID  string `json:"clusterId" example:"abc"`
	AlertUrl   string `json:"alertUrl" example:"http://127.0.0.1:9093"`
	GrafanaUrl string `json:"grafanaUrl" example:"http://127.0.0.1:3000"`
}

// QueryDashboardInfoReq Message to query the dashboard address information of a cluster
type QueryDashboardInfoReq struct {
	ClusterID string `json:"clusterId" example:"abc"`
}

// QueryDashboardInfoResp Reply message for querying the dashboard address information of the cluster
type QueryDashboardInfoResp struct {
	ClusterID string `json:"clusterId" example:"abc"`
	Url       string `json:"url" example:"http://127.0.0.1:9093"`
	Token     string `json:"token"`
}

//QueryClusterLogReq Messages that query cluster log information can be filtered based on query criteria
type QueryClusterLogReq struct {
	ClusterID string `json:"clusterId"`
	Module    string `form:"module" example:"tidb"`
	Level     string `form:"level" example:"warn"`
	Ip        string `form:"ip" example:"127.0.0.1"`
	Message   string `form:"message" example:"tidb log"`
	StartTime string `form:"startTime" example:"2021-09-01 12:00:00"`
	EndTime   string `form:"endTime" example:"2021-12-01 12:00:00"`
	structs.PageRequest
}

//QueryClusterLogResp Reply message for querying cluster log information
type QueryClusterLogResp struct {
	Took    int                      `json:"took" example:"10"`
	Results []structs.ClusterLogItem `json:"results"`
}

type QueryClusterParametersReq struct {
	ClusterID string `json:"clusterId"`
	structs.PageRequest
}

type QueryClusterParametersResp struct {
	structs.ClusterParameterInfo
}

type UpdateClusterParametersReq struct {
	ClusterID string                               `json:"clusterId"`
	Params    []structs.ClusterParameterSampleInfo `json:"params"`
	Reboot    bool                                 `json:"reboot"`
}

type UpdateClusterParametersResp struct {
	ClusterID                     string `json:"clusterId" example:"1"`
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
}

type InspectClusterParametersReq struct {
	ClusterID string `json:"clusterId"`
}

type InspectClusterParametersResp struct {
	ParamId       int64                      `json:"paramId" example:"1"`
	Name          string                     `json:"name" example:"binlog_cache"`
	ComponentType string                     `json:"componentType" example:"tidb"`
	Instance      string                     `json:"instance" example:"172.16.5.23"`
	RealValue     structs.ParameterRealValue `json:"realValue"`
	InspectValue  string                     `json:"inspectValue" example:"1"`
}
