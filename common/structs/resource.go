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
 * @File: resource.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import (
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/common/constants"
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

//CPUInfo Information describing the CPU, which will currently be used for Telemetry
type CPUInfo struct {
	Num     int     `json:"num"`     //go's reported runtime.NUMCPU()
	Sockets int     `json:"sockets"` //number of cpus reported
	Cores   int32   `json:"cores"`   //reported cores for first cpu
	Model   string  `json:"model"`   //reported model name e.g. `Intel(R) Core(TM) i7-7920HQ CPU @ 3.10GHz`
	HZ      float64 `json:"hz"`      //speed of first cpu e.g. 3100
}

//MemoryInfo Information describing the Memory, which will currently be used for Telemetry
type MemoryInfo struct {
	Total     uint64 `json:"total"`
	Available uint64 `json:"available"`
}

//OSInfo Information describing the OS, which will currently be used for Telemetry
type OSInfo struct {
	Family   string `json:"family"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

type DiskInfo struct {
	ID       string `json:"diskId"`
	HostId   string `json:"hostId,omitempty"`
	Name     string `json:"name"`     // [sda/sdb/nvmep0...]
	Capacity int32  `json:"capacity"` // Disk size, Unit: GB
	Path     string `json:"path"`     // Disk mount path: [/data1]
	Type     string `json:"type"`     // Disk type: [nvme-ssd/ssd/sata]
	Status   string `json:"status"`   // Disk Status, 0 for available, 1 for inused
}

type HostInfo struct {
	ID           string     `json:"hostId"`
	IP           string     `json:"ip"`
	UserName     string     `json:"userName,omitempty"`
	Passwd       string     `json:"passwd,omitempty"`
	HostName     string     `json:"hostName"`
	Status       string     `json:"status"`   // Host Status, 0 for Online, 1 for offline
	Stat         string     `json:"loadStat"` // Host Resource Stat, 0 for loadless, 1 for inused, 2 for exhaust
	Arch         string     `json:"arch"`     // x86 or arm64
	OS           string     `json:"os"`
	Kernel       string     `json:"kernel"`
	Spec         string     `json:"spec"`         // Host Spec, init while importing
	CpuCores     int32      `json:"cpuCores"`     // Host cpu cores spec, init while importing
	Memory       int32      `json:"memory"`       // Host memory, init while importing
	FreeCpuCores int32      `json:"freeCpuCores"` // Unused CpuCore, used for allocation
	FreeMemory   int32      `json:"freeMemory"`   // Unused memory size, Unit:GB, used for allocation
	Nic          string     `json:"nic"`          // Host network type: 1GE or 10GE
	Region       string     `json:"region"`
	AZ           string     `json:"az"`
	Rack         string     `json:"rack"`
	ClusterType  string     `json:"clusterType"` // What cluster is the host used for? [database/data migration]
	Purpose      string     `json:"purpose"`     // What Purpose is the host used for? [compute/storage/general]
	DiskType     string     `json:"diskType"`    // Disk type of this host [sata/ssd/nvme_ssd]
	Reserved     bool       `json:"reserved"`    // Whether this host is reserved - will not be allocated
	Traits       int64      `json:"traits"`      // Traits of labels
	SysLabels    []string   `json:"sysLabels"`
	CreatedAt    int64      `json:"createTime"`
	UpdatedAt    int64      `json:"updateTime"`
	Disks        []DiskInfo `json:"disks"`
}

func (h *HostInfo) GetPurposes() []string {
	return strings.Split(h.Purpose, ",")
}

func (h *HostInfo) GetSpecString() string {
	return fmt.Sprintf("%dC%dG", h.CpuCores, h.Memory)
}

func (h *HostInfo) AddTraits(p string) (err error) {
	if trait, err := GetTraitByName(p); err == nil {
		h.Traits = h.Traits | trait
	} else {
		return err
	}
	return nil
}

func (h HostInfo) IsExhaust() (stat constants.HostLoadStatus, isExhaust bool) {
	diskExaust := true
	for _, disk := range h.Disks {
		if disk.Status == string(constants.DiskAvailable) {
			diskExaust = false
			break
		}
	}
	computeExaust := (h.FreeCpuCores == 0 || h.FreeMemory == 0)
	if diskExaust && computeExaust {
		return constants.HostLoadExhaust, true
	} else if computeExaust {
		return constants.HostLoadComputeExhaust, true
	} else if diskExaust {
		return constants.HostLoadDiskExhaust, true
	} else {
		return constants.HostLoadStatusWhatever, false
	}
}

func (h HostInfo) IsLoadless() bool {
	diskLoadless := true
	for _, disk := range h.Disks {
		if disk.Status == string(constants.DiskExhaust) || disk.Status == string(constants.DiskInUsed) {
			diskLoadless = false
			break
		}
	}
	return diskLoadless && h.FreeCpuCores == h.CpuCores && h.FreeMemory == h.Memory
}

type Location struct {
	Region string `json:"Region"`
	Zone   string `json:"Zone"`
	Rack   string `json:"Rack"`
	HostIp string `json:"HostIp"`
}

type HostFilter struct {
	HostID  string `json:"hostId" form:"hostId"`
	Purpose string `json:"purpose" form:"purpose"`
	Status  string `json:"status" form:"status"`
	Stat    string `json:"loadStat" form:"loadStat"`
	Arch    string `json:"arch" form:"arch"`
}

type DiskFilter struct {
	DiskType   string `json:"DiskType"`
	DiskStatus string `json:"DiskStatus"`
	Capacity   int32  `json:"Capacity"`
}

type HierarchyTreeNode struct {
	Code     string               `json:"Code"`
	Name     string               `json:"Name"`
	Prefix   string               `json:"Prefix"`
	SubNodes []*HierarchyTreeNode `json:"SubNodes"`
}

type Stocks struct {
	Zone             string `json:"zone"`
	FreeHostCount    int32  `json:"freeHostCount"`
	FreeCpuCores     int32  `json:"freeCpuCores"`
	FreeMemory       int32  `json:"freeMemory"`
	FreeDiskCount    int32  `json:"freeDiskCount"`
	FreeDiskCapacity int32  `json:"freeDiskCapacity"`
}
