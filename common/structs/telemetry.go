/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
 * limitations under the License                                              *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: telemetry.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/21
*******************************************************************************/

package structs

import "time"

// NodeInfo Identifying a machine For use with Telemetry only
type NodeInfo struct {
	ID        string     `json:"Id"`
	Cpus      CPUInfo    `json:"cpus"`
	Memory    MemoryInfo `json:"memory"`
	Os        OSInfo     `json:"os"`
	Disks     []DiskInfo `json:"disks"`
	Loadavg15 float64    `json:"loadavg15"`
}

// InstanceInfoForTelemetry Identifying an instance For use with Telemetry only
type InstanceInfoForTelemetry struct {
	ID         string                        `json:"Id"`
	Type       string                        `json:"type"` //Type of instance, e.g., tidb,pd,etc
	Version    string                        `json:"version"`
	Status     string                        `json:"status"`
	CreateTime time.Time                     `json:"createTime"`
	DeleteTime time.Time                     `json:"deleteTime"`
	Spec       ComponentInstanceResourceSpec `json:"spec"`
	Regions    RegionInfo                    `json:"region"`
}

// ClusterInfoForTelemetry Identifying a cluster For use with Telemetry only
type ClusterInfoForTelemetry struct {
	ID              string                     `json:"Id"`
	Type            string                     `json:"type"` //Type of cluster, e.g., TiDB or DM
	Version         Version                    `json:"version"`
	Status          string                     `json:"status"`
	Exclusive       bool                       `json:"exclusive" form:"exclusive"`             //Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization
	CpuArchitecture string                     `json:"cpuArchitecture" form:"cpuArchitecture"` //X86/X86_64/ARM
	CreateTime      time.Time                  `json:"createTime"`
	DeleteTime      time.Time                  `json:"deleteTime"`
	Instances       []InstanceInfoForTelemetry `json:"instances"`
}

// EnterpriseManagerFeatureUsage Statistics of Enterprise Manager function usage, For use with Telemetry only
type EnterpriseManagerFeatureUsage struct {
	APIName string `json:"apiName"`
	Count   uint32 `json:"Count"`
	//AvgLatency int64  `json:"AvgLatency"`
}

// EnterpriseManagerInfo Identifying an enterprise manager For use with Telemetry only
type EnterpriseManagerInfo struct {
	ID         string                          `json:"Id"`
	Version    Version                         `json:"version"`
	CreateTime time.Time                       `json:"createTime"`
	LicenseID  string                          `json:"licenseId"` //License expiration time for Business Certificates
	EmAPICount []EnterpriseManagerFeatureUsage `json:"emAPICount"`
	EmNodes    []NodeInfo                      `json:"emNodes"`
	Nodes      []NodeInfo                      `json:"nodes"`
	Clusters   []ClusterInfoForTelemetry       `json:"clusters"`
}
