
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
	"github.com/pingcap-inc/tiem/micro-api/controller/resource/warehouse"
)

type ClusterBaseInfo struct {
	ClusterName    string      `json:"clusterName"`
	DbPassword     string      `json:"dbPassword"`
	ClusterType    string      `json:"clusterType"`
	ClusterVersion string      `json:"clusterVersion"`
	Tags           []string    `json:"tags"`
	Tls            bool        `json:"tls"`
	RecoverInfo    RecoverInfo `json:"recoverInfo"`
}

type ClusterInstanceInfo struct {
	IntranetConnectAddresses []string         `json:"intranetConnectAddresses"`
	ExtranetConnectAddresses []string         `json:"extranetConnectAddresses"`
	Whitelist                []string         `json:"whitelist"`
	PortList                 []int            `json:"portList"`
	DiskUsage                controller.Usage `json:"diskUsage"`
	CpuUsage                 controller.Usage `json:"cpuUsage"`
	MemoryUsage              controller.Usage `json:"memoryUsage"`
	StorageUsage             controller.Usage `json:"storageUsage"`
	BackupFileUsage          controller.Usage `json:"backupFileUsage"`
}

type RecoverInfo struct {
	SourceClusterId string `json:"sourceClusterId"`
	BackupRecordId  int64  `json:"backupRecordId"`
}

type ClusterMaintenanceInfo struct {
	MaintainTaskCron string `json:"maintainTaskCron"`
}

type ClusterDisplayInfo struct {
	ClusterId string `json:"clusterId"`
	ClusterBaseInfo
	controller.StatusInfo
	ClusterInstanceInfo
}

type StockCheckItem struct {
	Region          string `json:"region"`
	CpuArchitecture string `json:"cpuArchitecture"`

	DistributionItem
	ComponentType string `json:"componentType"`
	Enough        bool   `json:"enough"`
}

type ClusterNodeDemand struct {
	ComponentType     string             `json:"componentType"`
	TotalNodeCount    int                `json:"totalNodeCount"`
	DistributionItems []DistributionItem `json:"distributionItems"`
}

type DistributionItem struct {
	ZoneCode string `json:"zoneCode"`
	SpecCode string `json:"specCode"`
	Count    int    `json:"count"`
}

type ComponentInstance struct {
	ComponentBaseInfo
	Nodes []ComponentNodeDisplayInfo `json:"nodes"`
}

type ComponentNodeDisplayInfo struct {
	NodeId  string `json:"nodeId"`
	Version string `json:"version"`
	Status  string `json:"status"`
	ComponentNodeInstanceInfo
	ComponentNodeUsageInfo
}

type ComponentNodeInstanceInfo struct {
	HostId string                 `json:"hostId"`
	Port   int                    `json:"port"`
	Role   ComponentNodeRole      `json:"role"`
	Spec   warehouse.SpecBaseInfo `json:"spec"`
	Zone   warehouse.ZoneBaseInfo `json:"zone"`
}

type ComponentNodeUsageInfo struct {
	IoUtil       float32          `json:"ioUtil"`
	Iops         []float32        `json:"iops"`
	CpuUsage     controller.Usage `json:"cpuUsage"`
	MemoryUsage  controller.Usage `json:"memoryUsage"`
	StorageUsage controller.Usage `json:"storageUsage"`
}

type ComponentNodeRole struct {
	RoleCode string `json:"roleCode"`
	RoleName string `json:"roleName"`
}

type ComponentBaseInfo struct {
	ComponentType string `json:"componentType"`
	ComponentName string `json:"componentName"`
}

type ClusterCommonDemand struct {
	Exclusive       bool   `json:"exclusive" form:"exclusive"`
	Region          string `json:"region" form:"region"`
	CpuArchitecture string `json:"cpuArchitecture" form:"cpuArchitecture"`
}