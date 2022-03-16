/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package structs

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"time"
)

type CheckString struct {
	Valid         bool   `json:"valid"`
	RealValue     string `json:"realValue"`
	ExpectedValue string `json:"expectedValue"`
}

type CheckInt32 struct {
	Valid         bool  `json:"valid"`
	RealValue     int32 `json:"realValue"`
	ExpectedValue int32 `json:"expectedValue"`
}

type CheckBool struct {
	Valid         bool `json:"valid"`
	RealValue     bool `json:"realValue"`
	ExpectedValue bool `json:"expectedValue"`
}

type CheckAny struct {
	Valid         bool        `json:"valid"`
	RealValue     interface{} `json:"realValue"`
	ExpectedValue interface{} `json:"expectedValue"`
}

type CheckStatus struct {
	Health  bool   `json:"health"`
	Message string `json:"message"`
}

type TenantCheck struct {
	ClusterCount CheckRangeInt32 `json:"clusterCount"`
	CPURatio     float32         `json:"cpuRatio"`
	MemoryRatio  float32         `json:"memoryRatio"`
	StorageRatio float32         `json:"storageRatio"`
	Clusters     []ClusterCheck  `json:"clusters"`
}

type ClusterCheck struct {
	ID                string                             `json:"clusterID"`
	MaintenanceStatus constants.ClusterMaintenanceStatus `json:"maintenanceStatus"`
	Status            string                             `json:"runningStatus"`
	CPU               int32                              `json:"cpu"`
	Memory            int32                              `json:"memory"`
	Storage           int32                              `json:"storage"`
	Copies            CheckInt32                         `json:"copies"`
	AccountStatus     CheckStatus                        `json:"accountStatus"`
	Topology          CheckString                        `json:"topology"`
	RegionStatus      CheckStatus                        `json:"regionStatus"`
	Instances         []InstanceCheck                    `json:"instances"`
	HealthStatus      CheckStatus                        `json:"healthStatus"`
	BackupStrategy    CheckString                        `json:"backupStrategy"`
	BackupRecordValid map[string]bool                    `json:"backupRecordValid"`
}

type InstanceCheck struct {
	ID         string                 `json:"instanceID"`
	Address    string                 `json:"address"`
	Parameters map[string]CheckAny    `json:"parameters"`
	Versions   map[string]CheckString `json:"versions"`
}

type CheckRangeInt32 struct {
	Valid         bool    `json:"valid"`
	RealValue     int32   `json:"realValue"`
	ExpectedRange []int32 `json:"expectedRange"` // Left closed right closed interval
}
type HostsCheck struct {
	NTP                 map[string]CheckRangeInt32 `json:"ntp"`
	TimeZoneConsistency map[string]bool            `json:"timeZoneConsistency"`
	Ping                map[string]bool            `json:"ping"`
	Delay               map[string]CheckRangeInt32 `json:"delay"`
	Hosts               map[string]HostCheck       `json:"hosts"`
}

type CheckSwitch struct {
	Enable bool `json:"enable"`
}

type CheckError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type HostCheck struct {
	Address         string                 `json:"address"`
	SELinux         CheckSwitch            `json:"selinux"`
	Firewall        CheckSwitch            `json:"firewall"`
	Swap            CheckSwitch            `json:"swap"`
	MemoryAllocated CheckInt32             `json:"memoryAllocated"`
	CPUAllocated    CheckInt32             `json:"cpuAllocated"`
	DiskAllocated   map[string]CheckString `json:"diskAllocated"`
	StorageRatio    float32                `json:"storageRatio"`
	Errors          []CheckError           `json:"errors"`
}

type CheckPlatformReportInfo struct {
	Tenants map[string]TenantCheck `json:"tenants"`
	Hosts   HostsCheck             `json:"hosts"`
}

type CheckReportMeta struct {
	ID        string    `json:"checkID"`
	Creator   string    `json:"creator"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createAt"`
}

type CheckClusterReportInfo struct {
	ClusterCheck
}
