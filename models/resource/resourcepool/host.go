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
	"errors"
	"strings"
	"time"

	"github.com/pingcap/tiunimanager/models/common"
	"github.com/pingcap/tiunimanager/util/uuidutil"

	"github.com/pingcap/tiunimanager/common/constants"
	em_errors "github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"gorm.io/gorm"
)

type Host struct {
	ID           string          `json:"hostId" gorm:"primaryKey"`
	IP           string          `json:"ip" gorm:"not null"`
	UserName     string          `json:"userName,omitempty" gorm:"size:32"`
	Passwd       common.Password `json:"passwd,omitempty" gorm:"size:256"`
	HostName     string          `json:"hostName" gorm:"size:255"`
	Status       string          `json:"status" gorm:"index;default:Online"` // Host Status
	Stat         string          `json:"stat" gorm:"index;default:LoadLess"` // Host Resource Stat
	Arch         string          `json:"arch" gorm:"index"`                  // x86 or arm64
	OS           string          `json:"os" gorm:"size:32"`
	Kernel       string          `json:"kernel" gorm:"size:32"`
	Spec         string          `json:"spec"`               // Host Spec, init while importing
	CpuCores     int32           `json:"cpuCores"`           // Host cpu cores spec, init while importing
	Memory       int32           `json:"memory"`             // Host memory, init while importing
	FreeCpuCores int32           `json:"freeCpuCores"`       // Unused CpuCore, used for allocation
	FreeMemory   int32           `json:"freeMemory"`         // Unused memory size, Unit:GB, used for allocation
	Nic          string          `json:"nic" gorm:"size:32"` // Host network type: 1GE or 10GE
	Vendor       string          `json:"vendor" gorm:"size:32"`
	Region       string          `json:"region" gorm:"size:32"`
	AZ           string          `json:"az" gorm:"index"`
	Rack         string          `json:"rack" gorm:"index"`
	ClusterType  string          `json:"clusterType" gorm:"index"` // What Cluster is the host used for? [database/datamigration]
	Purpose      string          `json:"purpose" gorm:"index"`     // What Purpose is the host used for? [compute/storage/schedule]
	DiskType     string          `json:"diskType" gorm:"index"`    // Disk type of this host [sata/ssd/nvme_ssd]
	Reserved     bool            `json:"reserved" gorm:"index"`    // Whether this host is reserved - will not be allocated
	Traits       int64           `json:"traits" gorm:"index"`      // Traits of labels
	Disks        []Disk          `json:"disks" gorm:"-"`
	//UsedDisks    []UsedDisk     `json:"-" gorm:"-"`
	//UsedComputes []UsedCompute  `json:"-" gorm:"-"`
	//UsedPorts    []UsedPort     `json:"-" gorm:"-"`
	CreatedAt time.Time      `json:"createTime" gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt time.Time      `json:"-" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (h Host) IsInused() bool {
	return h.Stat == string(constants.HostLoadInUsed)
}

func (h Host) IsLoadless() bool {
	diskLoadless := true
	for _, disk := range h.Disks {
		if disk.Status == string(constants.DiskExhaust) || disk.Status == string(constants.DiskInUsed) {
			diskLoadless = false
			break
		}
	}
	return diskLoadless && h.FreeCpuCores == h.CpuCores && h.FreeMemory == h.Memory
}

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("IP = ? and HOST_NAME = ?", h.IP, h.HostName).First(&Host{}).Error
	if err == nil {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_HOST_ALREADY_EXIST, "host %s(%s) is existed", h.HostName, h.IP)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.ID = uuidutil.GenerateID()
		return nil
	} else {
		return err
	}
}

func (h *Host) AfterCreate(tx *gorm.DB) (err error) {
	for i := range h.Disks {
		h.Disks[i].HostID = h.ID
		err = tx.Create(&(h.Disks[i])).Error
		if err != nil {
			return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_CREATE_DISK_ERROR, "create disk %s for host %s(%s) failed, %v", h.Disks[i].Name, h.HostName, h.IP, err)
		}
	}
	return nil
}

func (h *Host) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("ID = ?", h.ID).First(h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_HOST_NOT_FOUND, "host %s is not found", h.ID)
		}
	} else {
		if h.IsInused() {
			return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_HOST_STILL_INUSED, "host %s is still in used", h.ID)
		}
	}

	return err
}

func (h *Host) AfterDelete(tx *gorm.DB) (err error) {
	var disks []Disk
	err = tx.Find(&disks, "host_id = ?", h.ID).Error
	if err != nil {
		return
	}
	for i := range disks {
		err = tx.Delete(&disks[i]).Error
		if err != nil {
			return
		}
	}
	h.Status = string(constants.HostDeleted)
	err = tx.Model(&h).Update("Status", h.Status).Error
	return
}

func (h *Host) AfterFind(tx *gorm.DB) (err error) {
	err = tx.Find(&(h.Disks), "HOST_ID = ?", h.ID).Error
	return
}

func (h *Host) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("IP") {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update ip on host %s is not allowed", h.ID)
	}
	if tx.Statement.Changed("DiskType", "Arch", "ClusterType", "Stat") {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update disk type or arch type or cluster type or load stat on host %s is not allowed", h.ID)
	}
	if tx.Statement.Changed("Vendor", "Region", "AZ", "Rack") {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update vendor/region/zone/rack info on host %s is not allowed", h.ID)
	}
	return
}

func (h *Host) PrepareForUpdate(newHost *Host) (updates map[string]interface{}, err error) {
	updates = make(map[string]interface{})
	h.prepareForUpdateName(newHost.HostName, newHost.IP, updates)
	h.prepareForUpdateLoginInfo(newHost.UserName, newHost.Passwd, updates)
	h.prepareForUpdateLocation(newHost.Vendor, newHost.Region, newHost.AZ, newHost.Rack, updates)
	h.prepareForUpdateKernel(newHost.OS, newHost.Kernel, updates)
	h.prepareForUpdateNic(newHost.Nic, updates)
	h.prepareForUpdateSpec(newHost.CpuCores, newHost.Memory, updates)
	err = h.prepareForUpdateType(newHost.Arch, newHost.DiskType, newHost.ClusterType, updates)
	if err != nil {
		return nil, err
	}
	err = h.prepareForUpdatePurpose(newHost.Purpose, updates)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (h *Host) ConstructFromHostInfo(src *structs.HostInfo) error {
	h.HostName = src.HostName
	h.IP = src.IP
	h.UserName = src.UserName
	h.Passwd = common.Password(src.Passwd)
	h.Arch = src.Arch
	h.OS = src.OS
	h.Kernel = src.Kernel
	h.FreeCpuCores = src.CpuCores - src.UsedCpuCores
	h.FreeMemory = src.Memory - src.UsedMemory
	h.Spec = src.GetSpecString()
	h.CpuCores = src.CpuCores
	h.Memory = src.Memory
	h.Nic = src.Nic
	h.Vendor = src.Vendor
	h.Region = src.Region
	h.AZ = structs.GenDomainCodeByName(h.Region, src.AZ)
	h.Rack = structs.GenDomainCodeByName(h.AZ, src.Rack)
	h.Status = src.Status
	h.Stat = src.Stat
	h.ClusterType = src.ClusterType
	// make sure purpose store in db is a normal string
	src.FormatPurpose()
	h.Purpose = src.Purpose
	h.DiskType = src.DiskType
	h.Reserved = src.Reserved
	h.Traits = src.Traits
	for _, disk := range src.Disks {
		var dbDisk Disk
		dbDisk.ConstructFromDiskInfo(&disk)
		h.Disks = append(h.Disks, dbDisk)
	}
	return nil
}

func (h *Host) ToHostInfo(dst *structs.HostInfo) {
	dst.ID = h.ID
	dst.HostName = h.HostName
	dst.IP = h.IP
	dst.Arch = h.Arch
	dst.OS = h.OS
	dst.Kernel = h.Kernel
	dst.UsedCpuCores = (h.CpuCores - h.FreeCpuCores)
	dst.UsedMemory = (h.Memory - h.FreeMemory)
	dst.Spec = h.Spec
	dst.CpuCores = h.CpuCores
	dst.Memory = h.Memory
	dst.Nic = h.Nic
	dst.Vendor = h.Vendor
	dst.Region = h.Region
	dst.AZ = structs.GetDomainNameFromCode(h.AZ)
	dst.Rack = structs.GetDomainNameFromCode(h.Rack)
	dst.Status = h.Status
	dst.ClusterType = h.ClusterType
	dst.Purpose = h.Purpose
	dst.DiskType = h.DiskType
	dst.CreatedAt = h.CreatedAt.Unix()
	dst.UpdatedAt = h.UpdatedAt.Unix()
	dst.Reserved = h.Reserved
	dst.Traits = h.Traits
	dst.SysLabels = structs.GetLabelNamesByTraits(dst.Traits)

	dst.AvailableDiskCount = 0
	for _, disk := range h.Disks {
		var diskInfo structs.DiskInfo
		disk.ToDiskInfo(&diskInfo)
		dst.Disks = append(dst.Disks, diskInfo)
		if disk.Status == string(constants.DiskAvailable) {
			dst.AvailableDiskCount++
		}
	}
	// Update Host's load stat after diskInfo is updated
	dst.Stat = h.Stat
	if dst.Stat == string(constants.HostLoadInUsed) {
		stat, isExhaust := dst.IsExhaust()
		if isExhaust {
			dst.Stat = string(stat)
		}
	}
}

func (h *Host) getPurposes() []string {
	if h.Purpose == "" {
		return nil
	}
	return strings.Split(h.Purpose, ",")
}

func (h *Host) addTraits(p string) (err error) {
	if p == "" {
		return nil
	}
	trait, err := structs.GetTraitByName(p)
	if err != nil {
		return err
	}
	h.Traits = h.Traits | trait
	return nil
}

func (h *Host) BuildDefaultTraits() (err error) {
	h.Traits = 0
	if err := h.addTraits(h.ClusterType); err != nil {
		return err
	}
	purposes := h.getPurposes()
	for _, p := range purposes {
		if err := h.addTraits(p); err != nil {
			return err
		}
	}
	if err := h.addTraits(h.DiskType); err != nil {
		return err
	}
	return nil
}

func (h *Host) prepareForUpdateName(hostName, ip string, updates map[string]interface{}) {
	if hostName != "" && hostName != h.HostName {
		updates["host_name"] = hostName
	}
	if ip != "" && ip != h.IP {
		updates["ip"] = ip
	}
}

func (h *Host) prepareForUpdateLoginInfo(userName string, password common.Password, updates map[string]interface{}) {
	if userName != "" && userName != h.UserName {
		updates["user_name"] = userName
	}
	if password != "" && password != h.Passwd {
		updates["passwd"] = password
	}
}

func (h *Host) prepareForUpdateSpec(cpuCores, memory int32, updates map[string]interface{}) {
	if (cpuCores == 0 && memory == 0) || (cpuCores == h.CpuCores && memory == h.Memory) {
		// no need to update
		return
	}
	updateCpu := false
	updateMem := false
	if cpuCores != 0 && cpuCores != h.CpuCores {
		alreadyUsedCpuCores := h.CpuCores - h.FreeCpuCores
		// update free cpu cores after set to the new cpucores
		freeCpuCores := cpuCores - alreadyUsedCpuCores
		updates["free_cpu_cores"] = freeCpuCores
		updates["cpu_cores"] = cpuCores
		updateCpu = true
	}
	if memory != 0 && memory != h.Memory {
		alreadyUsedMemory := h.Memory - h.FreeMemory
		// update free memory after set to the new memory
		freeMemory := memory - alreadyUsedMemory
		updates["free_memory"] = freeMemory
		updates["memory"] = memory
		updateMem = true
	}

	if updateCpu || updateMem {
		// either cpu cores or memory is updated, need to reset spec
		updatedCpuCores := h.CpuCores
		updatedMemory := h.Memory
		if updateCpu {
			updatedCpuCores = cpuCores
		}
		if updateMem {
			updatedMemory = memory
		}
		updates["spec"] = (&structs.HostInfo{CpuCores: updatedCpuCores, Memory: updatedMemory}).GetSpecString()
	}
}

func (h *Host) prepareForUpdateKernel(os, kernel string, updates map[string]interface{}) {
	if os != "" && os != h.OS {
		updates["os"] = os
	}
	if kernel != "" && kernel != h.Kernel {
		updates["kernel"] = kernel
	}
}

func (h *Host) prepareForUpdateNic(nic string, updates map[string]interface{}) {
	if nic != "" && nic != h.Nic {
		updates["nic"] = nic
	}
}

func (h *Host) prepareForUpdatePurpose(purpose string, updates map[string]interface{}) error {
	if purpose == "" || purpose == h.Purpose {
		// no need to update
		return nil
	}
	updates["purpose"] = purpose
	// Not change h in prepareForUpdateXXX function, so we copy to a patch to build traits for update
	patch := *h
	patch.Purpose = purpose
	// rebuild traits
	err := patch.BuildDefaultTraits()
	if err != nil {
		return err
	}
	updates["traits"] = patch.Traits
	return nil
}

// update region/zone/rack is not allowed by now, and it will be terminated in update hook
func (h *Host) prepareForUpdateLocation(vendor, region, zone, rack string, updates map[string]interface{}) {
	if vendor != "" && vendor != h.Vendor {
		updates["vendor"] = vendor
	}
	if region != "" && region != h.Region {
		updates["region"] = region
	}
	if zone != "" && zone != h.AZ {
		updates["az"] = zone
	}
	if rack != "" && rack != h.Rack {
		updates["rack"] = rack
	}
}

func (h *Host) prepareForUpdateType(arch, diskType, clusterType string, updates map[string]interface{}) (err error) {
	if arch != "" && arch != h.Arch {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update arch on host %s is not allowed", h.ID)
	}
	if diskType != "" && diskType != h.DiskType {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update disk type on host %s is not allowed", h.ID)
	}
	if clusterType != "" && clusterType != h.ClusterType {
		return em_errors.NewErrorf(em_errors.TIUNIMANAGER_RESOURCE_UPDATE_HOSTINFO_ERROR, "update cluster type on host %s is not allowed", h.ID)
	}
	return nil
}
