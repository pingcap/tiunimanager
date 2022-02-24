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
	ClusterID string `json:"clusterId"`
}

// DeleteClusterReq Message for delete a new cluster
type DeleteClusterReq struct {
	ClusterID                string `json:"clusterID" swaggerignore:"true" validate:"required,min=4,max=64"`
	AutoBackup               bool   `json:"autoBackup" form:"autoBackup"`
	KeepHistoryBackupRecords bool   `json:"keepHistoryBackupRecords" form:"keepHistoryBackupRecords"`
	Force                    bool   `json:"force" form:"force"`
}

// DeleteClusterResp Reply message for delete a new cluster
type DeleteClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterID"`
}

// StopClusterReq Message for stop a new cluster
type StopClusterReq struct {
	ClusterID string `json:"clusterId" validate:"required,min=4,max=64"`
}

// StopClusterResp Reply message for stop a new cluster
type StopClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// RestartClusterReq Message for restart a new cluster
type RestartClusterReq struct {
	ClusterID string `json:"clusterId" validate:"required,min=4,max=64"`
}

// RestartClusterResp Reply message for restart a new cluster
type RestartClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// ScaleInClusterReq Message for delete an instance in the cluster
type ScaleInClusterReq struct {
	ClusterID  string `json:"clusterId" form:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`
	InstanceID string `json:"instanceId"  form:"instanceId"`
}

// ScaleInClusterResp Reply message for delete an instance in the cluster
type ScaleInClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// PreviewScaleOutClusterReq Message for cluster expansion operation
type PreviewScaleOutClusterReq struct {
	ClusterID string `json:"clusterId" form:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`
	structs.ClusterResourceInfo
}

// ScaleOutClusterReq Message for cluster expansion operation
type ScaleOutClusterReq struct {
	ClusterID string `json:"clusterId" form:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`

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
	BackupID          string                      `json:"backupId" validate:"required,min=8,max=64"`
	ResourceParameter structs.ClusterResourceInfo `json:"resourceParameters"`
}

//RestoreNewClusterResp Restore to a new cluster using the backup file Reply Message
type RestoreNewClusterResp struct {
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
	ClusterID                     string `json:"clusterID"`
}

//RestoreExistClusterReq Restore to exist cluster message using the backup file
type RestoreExistClusterReq struct {
	ClusterID string `json:"clusterID" validate:"required,min=4,max=64"`
	BackupID  string `json:"backupID" validate:"required,min=8,max=64"`
}

//RestoreExistClusterResp Restore to exist cluster using the backup file Reply Message
type RestoreExistClusterResp struct {
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
}

// CloneClusterReq Message for clone a new cluster
type CloneClusterReq struct {
	structs.CreateClusterParameter
	ResourceParameter structs.ClusterResourceInfo `json:"resourceParameters" form:"resourceParameters"`
	CloneStrategy     string                      `json:"cloneStrategy" validate:"required"`                // specify clone strategy, include empty, snapshot and sync, default empty(option)
	SourceClusterID   string                      `json:"sourceClusterId" validate:"required,min=4,max=64"` // specify source cluster id(require)
}

// CloneClusterResp Reply message for clone a new cluster
type CloneClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
}

// MasterSlaveClusterSwitchoverReq Master and slave cluster switchover messages
type MasterSlaveClusterSwitchoverReq struct {
	// old master/new slave
	SourceClusterID string `json:"sourceClusterID" validate:"required,min=4,max=64"`
	// new master/old slave
	TargetClusterID string `json:"targetClusterID" validate:"required,min=4,max=64"`
	Force           bool   `json:"force"`
}

// MasterSlaveClusterSwitchoverResp Master and slave cluster switchover reply message
type MasterSlaveClusterSwitchoverResp struct {
	structs.AsyncTaskWorkFlowInfo
}

// TakeoverClusterReq Requests to take over an existing TiDB cluster, requiring TiDB version >= 4.0 when taking over
type TakeoverClusterReq struct {
	TiUPIp           string `json:"TiUPIp" example:"172.16.4.147" form:"TiUPIp" validate:"required,ip"`
	TiUPPort         int    `json:"TiUPPort" example:"22" form:"TiUPPort" validate:"required"`
	TiUPUserName     string `json:"TiUPUserName" example:"root" form:"TiUPUserName" validate:"required"`
	TiUPUserPassword string `json:"TiUPUserPassword" example:"password" form:"TiUPUserPassword" validate:"required"`
	TiUPPath         string `json:"TiUPPath" example:".tiup/" form:"TiUPPath" validate:"required"`
	ClusterName      string `json:"clusterName" example:"myClusterName" form:"clusterName" validate:"required,min=4,max=64"`
	DBPassword       string `json:"dbPassword" example:"myPassword" form:"dbPassword" validate:"required"`
}

// TakeoverClusterResp Reply message for takeover a cluster
type TakeoverClusterResp struct {
	structs.AsyncTaskWorkFlowInfo
	ClusterID string `json:"clusterId"`
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

// GetDashboardInfoReq Message to query the dashboard address information of a cluster
type GetDashboardInfoReq struct {
	ClusterID string `json:"clusterId" example:"abc" swaggerignore:"true"`
}

// GetDashboardInfoResp Reply message for querying the dashboard address information of the cluster
type GetDashboardInfoResp struct {
	ClusterID string `json:"clusterId" example:"abc"`
	Url       string `json:"url" example:"http://127.0.0.1:9093"`
	Token     string `json:"token"`
}

//QueryClusterLogReq Messages that query cluster log information can be filtered based on query criteria
type QueryClusterLogReq struct {
	ClusterID string `json:"clusterId" swaggerignore:"true"`
	Module    string `form:"module" example:"tidb"`
	Level     string `form:"level" example:"warn"`
	Ip        string `form:"ip" example:"127.0.0.1"`
	Message   string `form:"message" example:"tidb log"`
	StartTime int64  `form:"startTime" example:"1630468800"`
	EndTime   int64  `form:"endTime" example:"1638331200"`
	structs.PageRequest
}

//QueryClusterLogResp Reply message for querying cluster log information
type QueryClusterLogResp struct {
	Took    int                      `json:"took" example:"10"`
	Results []structs.ClusterLogItem `json:"results"`
}

type QueryClusterParametersReq struct {
	ClusterID    string `json:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`
	ParamName    string `json:"paramName" form:"paramName"`
	InstanceType string `json:"instanceType" form:"instanceType"`
	structs.PageRequest
}

type QueryClusterParametersResp struct {
	ParamGroupId string                         `json:"paramGroupId"`
	Params       []structs.ClusterParameterInfo `json:"params"`
}

type UpdateClusterParametersReq struct {
	ClusterID string                               `json:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`
	Params    []structs.ClusterParameterSampleInfo `json:"params" validate:"required"`
	Reboot    bool                                 `json:"reboot"`
	Nodes     []string                             `json:"nodes" swaggerignore:"true"`
}

type UpdateClusterParametersResp struct {
	ClusterID string `json:"clusterId" example:"1"`
	structs.AsyncTaskWorkFlowInfo
}

type InspectParametersReq struct {
	ClusterID  string `json:"clusterId" swaggerignore:"true" validate:"required,min=4,max=64"`
	InstanceID string `json:"instanceId" form:"instanceId"`
}

type InspectParametersResp struct {
	Params []InspectParameters `json:"params"`
}

type InspectParameters struct {
	InstanceID     string                 `json:"instanceId"`
	InstanceType   string                 `json:"instanceType"`
	ParameterInfos []InspectParameterInfo `json:"parameterInfos"`
}

type InspectParameterInfo struct {
	ParamID        string                     `json:"paramId" example:"1"`
	Category       string                     `json:"category" example:"log"`
	Name           string                     `json:"name" example:"binlog_cache"`
	SystemVariable string                     `json:"systemVariable" example:"log.log_level"`
	Type           int                        `json:"type" example:"0" enums:"0,1,2,3,4"`
	Unit           string                     `json:"unit" example:"MB"`
	UnitOptions    []string                   `json:"unitOptions" example:"KB,MB,GB"`
	Range          []string                   `json:"range" example:"1, 1000"`
	RangeType      int                        `json:"rangeType" example:"1" enums:"0,1,2"`
	HasReboot      int                        `json:"hasReboot" example:"0" enums:"0,1"`
	ReadOnly       int                        `json:"readOnly" example:"0" enums:"0,1"`
	Description    string                     `json:"description" example:"binlog cache size"`
	Note           string                     `json:"note" example:"binlog cache size"`
	RealValue      structs.ParameterRealValue `json:"realValue"`
	InspectValue   interface{}                `json:"inspectValue"`
}

type PreviewClusterResp struct {
	Region          string `json:"region" form:"region"`
	CpuArchitecture string `json:"cpuArchitecture" form:"cpuArchitecture"`
	ClusterType     string `json:"clusterType"`
	ClusterVersion  string `json:"clusterVersion"`

	ClusterName string `json:"clusterName"`

	StockCheckResult  []structs.ResourceStockCheckResult `json:"stockCheckResult"`
	CapabilityIndexes []structs.Index                    `json:"capabilityIndexes"`
}

type ScaleOutPreviewResp struct {
	StockCheckResult  []structs.ResourceStockCheckResult `json:"stockCheckResult"`
	CapabilityIndexes []structs.Index                    `json:"capabilityIndexes"`
}

type ApiEditConfigReq struct {
	InstanceHost string
	InstancePort uint
	Headers      map[string]string
	ConfigMap    map[string]interface{}
}

type ApiShowConfigReq struct {
	InstanceHost string
	InstancePort uint
	Params       map[string]string
	Headers      map[string]string
}
