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

package structs

import (
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
)

//ClusterResourceParameterComputeResource Single component resource parameters when creating a cluster
type ClusterResourceParameterComputeResource struct {
	Zone         string `json:"zoneCode"` //
	DiskType     string `json:"diskType"` //NVMeSSD/SSD/SATA
	DiskCapacity int    `json:"diskCapacity"`
	Spec         string `json:"specCode"` //4C8G/8C16G ?
	Count        int    `json:"count"`
}

func (p *ClusterResourceParameterComputeResource) Equal(zone, spec, diskType string, diskCapacity int) bool {
	return p.Zone == zone &&
		p.DiskType == diskType &&
		p.DiskCapacity == diskCapacity &&
		p.Spec == spec
}

//ClusterResourceParameterCompute Component resource parameters when creating a cluster, including: compute resources, storage resources
type ClusterResourceParameterCompute struct {
	Type     string                                    `json:"componentType"` //TiDB/TiKV/PD/TiFlash/CDC/DM-Master/DM-Worker
	Count    int                                       `json:"totalNodeCount"`
	Resource []ClusterResourceParameterComputeResource `json:"resource"`
}

//ClusterResourceInfo Resource information for creating database cluster input
type ClusterResourceInfo struct {
	InstanceResource []ClusterResourceParameterCompute `json:"instanceResource"`
}

func (p ClusterResourceInfo) GetComponentCount(idType constants.EMProductComponentIDType) int32 {
	for _, i := range p.InstanceResource {
		if i.Type == string(idType) {
			return int32(i.Count)
		}
	}
	return 0
}

//CreateClusterParameter User input parameters when creating a cluster
type CreateClusterParameter struct {
	Name string `json:"clusterName" validate:"required,min=8,max=64"`
	// todo delete?
	DBUser           string   `json:"dbUser" validate:"max=32"` //The username and password for the newly created database cluster, default is the root user, which is not valid for Data Migration clusters
	DBPassword       string   `json:"dbPassword" validate:"required,min=8,max=32"`
	Type             string   `json:"clusterType" validate:"required,oneof=TiDB DM TiKV"`
	Version          string   `json:"clusterVersion" validate:"required,startswith=v"`
	Tags             []string `json:"tags"`
	TLS              bool     `json:"tls"`
	Copies           int      `json:"copies"`                     //The number of copies of the newly created cluster data, consistent with the number of copies set in PD
	Exclusive        bool     `json:"exclusive" form:"exclusive"` //Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization
	Vendor           string   `json:"vendor" form:"vendor"`
	Region           string   `json:"region" form:"region" validate:"required,max=32"`                                       //The Region where the cluster is located
	CpuArchitecture  string   `json:"cpuArchitecture" form:"cpuArchitecture" validate:"required,oneof=X86 X86_64 ARM ARM64"` //X86/X86_64/ARM
	ParameterGroupID string   `json:"parameterGroupID" form:"parameterGroupID"`
}

// ClusterInfo Cluster details information
type ClusterInfo struct {
	ID      string `json:"clusterId"`
	UserID  string `json:"userId"`
	Name    string `json:"clusterName"`
	Type    string `json:"clusterType"`
	Version string `json:"clusterVersion"`
	//DBUser                   string    `json:"dbUser"` //The username and password for the newly created database cluster, default is the root user, which is not valid for Data Migration clusters
	Vendor                   string    `json:"vendor" form:"vendor"`
	Tags                     []string  `json:"tags"`
	TLS                      bool      `json:"tls"`
	Region                   string    `json:"region"`
	Status                   string    `json:"status"`
	Role                     string    `json:"role"`
	Copies                   int       `json:"copies"`                                 //The number of copies of the newly created cluster data, consistent with the number of copies set in PD
	Exclusive                bool      `json:"exclusive" form:"exclusive"`             //Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization
	CpuArchitecture          string    `json:"cpuArchitecture" form:"cpuArchitecture"` //X86/X86_64/ARM
	AlertUrl                 string    `json:"alertUrl" example:"http://127.0.0.1:9093"`
	GrafanaUrl               string    `json:"grafanaUrl" example:"http://127.0.0.1:3000"`
	MaintainStatus           string    `json:"maintainStatus"`
	MaintainWindow           string    `json:"maintainWindow"`
	IntranetConnectAddresses []string  `json:"intranetConnectAddresses"`
	ExtranetConnectAddresses []string  `json:"extranetConnectAddresses"`
	Whitelist                []string  `json:"whitelist"`
	CpuUsage                 Usage     `json:"cpuUsage"`
	MemoryUsage              Usage     `json:"memoryUsage"`
	StorageUsage             Usage     `json:"storageUsage"`
	BackupSpaceUsage         Usage     `json:"backupFileUsage"`
	CreateTime               time.Time `json:"createTime"`
	UpdateTime               time.Time `json:"updateTime"`
	DeleteTime               time.Time `json:"deleteTime"`
}

// ClusterInstanceInfo Details of the instances in the cluster
type ClusterInstanceInfo struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Version      string          `json:"version"`
	Status       string          `json:"status"`
	HostID       string          `json:"hostID"`
	Addresses    []string        `json:"addresses"`
	Ports        []int32         `json:"ports"`
	CpuUsage     Usage           `json:"cpuUsage"`
	MemoryUsage  Usage           `json:"memoryUsage"`
	StorageUsage Usage           `json:"storageUsage"`
	IOUtil       float32         `json:"ioUtil"`
	IOPS         []float32       `json:"iops"`
	Spec         ProductSpecInfo `json:"spec"` //??
	Zone         ZoneInfo        `json:"zone"` //??
}

// ClusterTopologyInfo Topology of the cluster
type ClusterTopologyInfo struct {
	Topology []ClusterInstanceInfo `json:"topology"`
}

// BackupStrategy Timed or scheduled data backup strategy
type BackupStrategy struct {
	ClusterID  string `json:"clusterId"`
	BackupDate string `json:"backupDate"`
	Period     string `json:"period"`
}

// BackupRecord Single backup file details
type BackupRecord struct {
	ID           string    `json:"id"`
	ClusterID    string    `json:"clusterId"`
	BackupType   string    `json:"backupType"`
	BackupMethod string    `json:"backupMethod"`
	BackupMode   string    `json:"backupMode"`
	FilePath     string    `json:"filePath"`
	Size         float32   `json:"size"`
	BackupTSO    string    `json:"backupTso"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
	DeleteTime   time.Time `json:"deleteTime"`
}

type ClusterLogItem struct {
	Index      string                 `json:"index" example:"tiem-tidb-cluster-2021.09.23"`
	Id         string                 `json:"id" example:"zvadfwf"`
	Level      string                 `json:"level" example:"warn"`
	SourceLine string                 `json:"sourceLine" example:"main.go:210"`
	Message    string                 `json:"message"  example:"tidb log"`
	Ip         string                 `json:"ip" example:"127.0.0.1"`
	ClusterId  string                 `json:"clusterId" example:"abc"`
	Module     string                 `json:"module" example:"tidb"`
	Ext        map[string]interface{} `json:"ext"`
	Timestamp  string                 `json:"timestamp" example:"2021-09-23 14:23:10"`
}

type ProductUpgradePathItem struct {
	UpgradeType string   `json:"upgradeType"  validate:"required" enums:"in-place,migration"`
	UpgradeWay  string   `json:"upgradeWay,omitempty"  enums:"offline,online"`
	Versions    []string `json:"versions" validate:"required" example:"v5.0.0,v5.3.0"`
}
type ProductUpgradeVersionConfigDiffItem struct {
	ParamId      string   `json:"paramId" validate:"required" example:"1"`
	Category     string   `json:"category" validate:"required" example:"basic"`
	Name         string   `json:"name" validate:"required" example:"max-merge-region-size"`
	InstanceType string   `json:"instanceType" validate:"required" example:"pd-server"`
	CurrentValue string   `json:"currentValue" validate:"required" example:"20"`
	SuggestValue string   `json:"suggestValue" validate:"required" example:"30"`
	Type         int      `json:"type" validate:"required" example:"0" enums:"0,1,2,3,4"`
	Unit         string   `json:"unit" validate:"required" example:"MB"`
	UnitOptions  []string `json:"unitOptions" validate:"required" example:"KB,MB,GB"`
	Range        []string `json:"range" validate:"required" example:"1, 1000"`
	RangeType    int      `json:"rangeType" validate:"required" example:"1" enums:"0,1,2"`
	Description  string   `json:"description" example:"desc for max-merge-region-size"`
}

type ClusterUpgradeVersionConfigItem struct {
	ParamId      string `json:"paramId" validate:"required" example:"1"`
	Name         string `json:"name" validate:"required" example:"max-merge-region-size"`
	InstanceType string `json:"instanceType" validate:"required" example:"pd-server"`
	Value        string `json:"value" validate:"required" example:"20"`
}

type ClusterInstanceParameterValue struct {
	ID    string `json:"instanceId"`
	Value string `json:"value"`
}

type ParameterRealValue struct {
	ClusterValue  string                           `json:"clusterValue"`
	InstanceValue []*ClusterInstanceParameterValue `json:"instanceValue"`
}

type ClusterParameterSampleInfo struct {
	ParamId   string             `json:"paramId" example:"1" validate:"required,min=1,max=64"`
	RealValue ParameterRealValue `json:"realValue" validate:"required"`
}

type ClusterParameterInfo struct {
	ParamId        string             `json:"paramId" example:"1"`
	Category       string             `json:"category" example:"basic"`
	Name           string             `json:"name" example:"binlog_size"`
	InstanceType   string             `json:"instanceType" example:"tidb"`
	SystemVariable string             `json:"systemVariable" example:"log.log_level"`
	Type           int                `json:"type" example:"0" enums:"0,1,2,3,4"`
	Unit           string             `json:"unit" example:"MB"`
	UnitOptions    []string           `json:"unitOptions" example:"KB,MB,GB"`
	Range          []string           `json:"range" example:"1, 1000"`
	RangeType      int                `json:"rangeType" example:"1" enums:"0,1,2"`
	HasReboot      int                `json:"hasReboot" example:"0" enums:"0,1"`
	HasApply       int                `json:"hasApply" example:"1" enums:"0,1"`
	UpdateSource   int                `json:"updateSource" example:"0" enums:"0,1,2,3"`
	ReadOnly       int                `json:"readOnly" example:"0" enums:"0,1"`
	DefaultValue   string             `json:"defaultValue" example:"1"`
	RealValue      ParameterRealValue `json:"realValue"`
	Description    string             `json:"description" example:"binlog cache size"`
	Note           string             `json:"note" example:"binlog cache size"`
	CreatedAt      int64              `json:"createTime" example:"1636698675"`
	UpdatedAt      int64              `json:"updateTime" example:"1636698675"`
}

type ResourceStockCheckResult struct {
	Type string `json:"componentType"`
	Name string `json:"componentName"`
	ClusterResourceParameterComputeResource
	Enough bool `json:"enough"`
}
