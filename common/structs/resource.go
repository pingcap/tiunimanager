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
	"strconv"
	"strings"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
)

// generate code by names, for example:
// region name "Region1" and zone name "Zone1" will get zoneID "Region1,Zone1"
func GenDomainCodeByName(pre string, name string) string {
	if pre == "" {
		return name
	}
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
	ID                 string              `json:"hostId"`
	IP                 string              `json:"ip"`
	SSHPort            int32               `json:"sshPort,omitempty"`
	UserName           string              `json:"userName,omitempty"`
	Passwd             string              `json:"passwd,omitempty"`
	HostName           string              `json:"hostName"`
	Status             string              `json:"status"`   // Host status, Online, Offline, Failed, Deleted, etc
	Stat               string              `json:"loadStat"` // Host load stat, Loadless, Inused, Exhaust, etc
	Arch               string              `json:"arch"`     // x86 or arm64
	OS                 string              `json:"os"`
	Kernel             string              `json:"kernel"`
	Spec               string              `json:"spec"`         // Host Spec, init while importing
	CpuCores           int32               `json:"cpuCores"`     // Host cpu cores spec, init while importing
	Memory             int32               `json:"memory"`       // Host memory, init while importing
	UsedCpuCores       int32               `json:"usedCpuCores"` // Unused CpuCore, used for allocation
	UsedMemory         int32               `json:"usedMemory"`   // Unused memory size, Unit:GiB, used for allocation
	Nic                string              `json:"nic"`          // Host network type: 1GE or 10GE
	Vendor             string              `json:"vendor"`
	Region             string              `json:"region"`
	AZ                 string              `json:"az"`
	Rack               string              `json:"rack"`
	ClusterType        string              `json:"clusterType"` // What cluster is the host used for? [database/data migration]
	Purpose            string              `json:"purpose"`     // What Purpose is the host used for? [compute/storage/schedule]
	DiskType           string              `json:"diskType"`    // Disk type of this host [SATA/SSD/NVMeSSD]
	Reserved           bool                `json:"reserved"`    // Whether this host is reserved - will not be allocated
	Traits             int64               `json:"traits"`      // Traits of labels
	SysLabels          []string            `json:"sysLabels"`
	Instances          map[string][]string `json:"instances"`
	CreatedAt          int64               `json:"createTime"`
	UpdatedAt          int64               `json:"updateTime"`
	AvailableDiskCount int32               `json:"availableDiskCount"` // available disk count which could be used for allocation
	Disks              []DiskInfo          `json:"disks"`
}

func ParseCpu(specCode string) int {
	cpu, err := strconv.Atoi(strings.Split(specCode, "C")[0])
	if err != nil {
		framework.Log().Errorf("ParseCpu error, specCode = %s", specCode)
	}
	return cpu
}

func ParseMemory(specCode string) int {
	memory, err := strconv.Atoi(strings.Split(strings.Split(specCode, "C")[1], "G")[0])
	if err != nil {
		framework.Log().Errorf("ParseMemory error, specCode = %s", specCode)
	}
	return memory
}

func GenSpecCode(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}

// validate the disk info before create disk
func (d *DiskInfo) ValidateDisk(hostId string, hostDiskType string) (err error) {
	// disk name, disk path, disk capacity is required
	if d.Name == "" || d.Path == "" || d.Capacity <= 0 {
		return errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
			hostId, d.Name, d.Path, d.Capacity)
	}
	// disk's host id is optional, if specified, should be equal to the existed host id
	if d.HostId != "" {
		if hostId != "" && d.HostId != hostId {
			return errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s failed, host id conflict %s vs %s",
				d.Name, d.Path, d.HostId, hostId)
		}
	}

	if d.Status != "" && !constants.DiskStatus(d.Status).IsValidStatus() {
		return errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s specified a invalid status %s, [Available|Reserved]",
			d.Name, d.Path, hostId, d.Status)
	}

	if d.Type != "" {
		if err = constants.ValidDiskType(d.Type); err != nil {
			return errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, %v", d.Name, d.Path, hostId, err)
		}
		if hostDiskType != "" && d.Type != hostDiskType {
			return errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, disk type conflict %s vs %s",
				d.Name, d.Path, hostId, d.Type, hostDiskType)
		}
	}

	return nil
}

func (h *HostInfo) GetPurposes() []string {
	if h.Purpose == "" {
		return nil
	}
	purposes := strings.Split(h.Purpose, ",")
	for i := range purposes {
		purposes[i] = strings.TrimSpace(purposes[i])
	}
	return purposes
}

func (h *HostInfo) FormatPurpose() {
	purposes := h.GetPurposes()
	h.Purpose = strings.Join(purposes, ",")
}

func (h *HostInfo) GetSpecString() string {
	return fmt.Sprintf("%dC%dG", h.CpuCores, h.Memory)
}

func (h *HostInfo) AddTraits(p string) (err error) {
	if p == "" {
		return nil
	}
	trait, err := GetTraitByName(p)
	if err != nil {
		return err
	}
	h.Traits = h.Traits | trait
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
	computeExaust := (h.UsedCpuCores == h.CpuCores || h.UsedMemory == h.Memory)
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
	return diskLoadless && h.UsedCpuCores == 0 && h.UsedMemory == 0
}

type Location struct {
	Region string `json:"region" form:"region"`
	Zone   string `json:"zone" form:"zone"`
	Rack   string `json:"rack" form:"rack"`
	HostIp string `json:"hostIp" form:"hostIp"`
}

type ImportCondition struct {
	ReserveHost   bool `json:"reserved" form:"reserved"`
	SkipHostInit  bool `json:"skipHostInit" form:"skipHostInit"`
	IgnoreWarings bool `json:"ignoreWarnings" form:"ignoreWarnings"`
}

type HostFilter struct {
	HostID  string `json:"hostId" form:"hostId"`
	Purpose string `json:"purpose" form:"purpose"`
	Status  string `json:"status" form:"status"`
	Stat    string `json:"loadStat" form:"loadStat"`
	Arch    string `json:"arch" form:"arch"`
}

type DiskFilter struct {
	DiskType   string `json:"diskType" form:"diskType"`
	DiskStatus string `json:"diskStatus" form:"diskStatus"`
	Capacity   int32  `json:"capacity" form:"capacity"`
}

type HierarchyTreeNode struct {
	Code     string               `json:"code"`
	Name     string               `json:"name"`
	Prefix   string               `json:"prefix"`
	SubNodes []*HierarchyTreeNode `json:"subNodes"`
}

type Stocks struct {
	Zone             string `json:"zone"`
	FreeHostCount    int32  `json:"freeHostCount"`
	FreeCpuCores     int32  `json:"freeCpuCores"`
	FreeMemory       int32  `json:"freeMemory"`
	FreeDiskCount    int32  `json:"freeDiskCount"`
	FreeDiskCapacity int32  `json:"freeDiskCapacity"`
}
