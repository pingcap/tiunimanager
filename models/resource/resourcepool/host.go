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

package resourcepool

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GenDomainCodeByName(pre string, name string) string {
	return fmt.Sprintf("%s,%s", pre, name)
}

func GetDomainNameFromCode(failureDomain string) string {
	pos := strings.LastIndex(failureDomain, ",")
	return failureDomain[pos+1:]
}

func GetDomainPrefixFromCode(failureDomain string) string {
	pos := strings.LastIndex(failureDomain, ",")
	if pos == -1 {
		// No found ","
		return failureDomain
	}
	return failureDomain[:pos]
}

type Host struct {
	ID           string `json:"hostId" gorm:"primaryKey"`
	IP           string `json:"ip" gorm:"not null"`
	UserName     string `json:"userName,omitempty" gorm:"size:32"`
	Passwd       string `json:"passwd,omitempty" gorm:"size:32"`
	HostName     string `json:"hostName" gorm:"size:255"`
	Status       string `json:"status" gorm:"index"`         // Host Status, 0 for Online, 1 for offline
	Stat         string `json:"stat" gorm:"index;default:0"` // Host Resource Stat, 0 for loadless, 1 for inused, 2 for exhaust
	Arch         string `json:"arch" gorm:"index"`           // x86 or arm64
	OS           string `json:"os" gorm:"size:32"`
	Kernel       string `json:"kernel" gorm:"size:32"`
	Spec         string `json:"spec"`               // Host Spec, init while importing
	CpuCores     int32  `json:"cpuCores"`           // Host cpu cores spec, init while importing
	Memory       int32  `json:"memory"`             // Host memory, init while importing
	FreeCpuCores int32  `json:"freeCpuCores"`       // Unused CpuCore, used for allocation
	FreeMemory   int32  `json:"freeMemory"`         // Unused memory size, Unit:GB, used for allocation
	Nic          string `json:"nic" gorm:"size:32"` // Host network type: 1GE or 10GE
	Region       string `json:"region" gorm:"size:32"`
	AZ           string `json:"az" gorm:"index"`
	Rack         string `json:"rack" gorm:"index"`
	ClusterType  string `json:"clusterType" gorm:"index"` // What Cluster is the host used for? [database/datamigration]
	Purpose      string `json:"purpose" gorm:"index"`     // What Purpose is the host used for? [compute/storage/schedule]
	DiskType     string `json:"diskType" gorm:"index"`    // Disk type of this host [sata/ssd/nvme_ssd]
	Reserved     bool   `json:"reserved" gorm:"index"`    // Whether this host is reserved - will not be allocated
	Traits       int64  `json:"traits" gorm:"index"`      // Traits of labels
	Disks        []Disk `json:"disks" gorm:"-"`
	//UsedDisks    []UsedDisk     `json:"-" gorm:"-"`
	//UsedComputes []UsedCompute  `json:"-" gorm:"-"`
	//UsedPorts    []UsedPort     `json:"-" gorm:"-"`
	CreatedAt time.Time      `json:"createTime" gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt time.Time      `json:"-" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
