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

package hostresource

import (
	"strings"

	"github.com/pingcap-inc/tiem/library/common/resource-type"
)

type HostInfo struct {
	ID           string     `json:"hostId"`
	IP           string     `json:"ip"`
	UserName     string     `json:"userName,omitempty"`
	Passwd       string     `json:"passwd,omitempty"`
	HostName     string     `json:"hostName"`
	Status       int32      `json:"status"`   // Host Status, 0 for Online, 1 for offline
	Stat         int32      `json:"loadStat"` // Host Resource Stat, 0 for loadless, 1 for inused, 2 for exhaust
	Arch         string     `json:"arch"`     // x86 or arm64
	OS           string     `json:"os"`
	Kernel       string     `json:"kernel"`
	Spec         string     `json:"spec"`         // Host Spec, init while importing
	CpuCores     int32      `json:"cpuCores"`     // Host cpu cores spec, init while importing
	Memory       int32      `json:"memory"`       // Host memroy, init while importing
	FreeCpuCores int32      `json:"freeCpuCores"` // Unused CpuCore, used for allocation
	FreeMemory   int32      `json:"freeMemory"`   // Unused memory size, Unit:GB, used for allocation
	Nic          string     `json:"nic"`          // Host network type: 1GE or 10GE
	Region       string     `json:"region"`
	AZ           string     `json:"az"`
	Rack         string     `json:"rack"`
	ClusterType  string     `json:"clusterType"` // What cluster is the host used for? [database/datamigration]
	Purpose      string     `json:"purpose"`     // What Purpose is the host used for? [compute/storage/general]
	DiskType     string     `json:"diskType"`    // Disk type of this host [sata/ssd/nvme_ssd]
	Reserved     bool       `json:"reserved"`    // Whether this host is reserved - will not be allocated
	Traits       int64      `json:"traits"`      // Traits of labels
	SysLabels    []string   `json:"sysLabels"`
	CreatedAt    int64      `json:"createTime"`
	UpdatedAt    int64      `json:"updateTime"`
	Disks        []DiskInfo `json:"disks"`
}

func (h *HostInfo) getPurposes() []string {
	return strings.Split(h.Purpose, ",")
}

func (h *HostInfo) addTraits(p string) (err error) {
	if trait, err := resource.GetTraitByName(p); err == nil {
		h.Traits = h.Traits | trait
	} else {
		return err
	}
	return nil
}

/*
func (h *HostInfo) buildDefaultTraits() (err error) {
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
*/
type DiskInfo struct {
	ID       string `json:"diskId"`
	HostId   string `json:"hostId,omitempty"`
	Name     string `json:"name"`     // [sda/sdb/nvmep0...]
	Capacity int32  `json:"capacity"` // Disk size, Unit: GB
	Path     string `json:"path"`     // Disk mount path: [/data1]
	Type     string `json:"type"`     // Disk type: [nvme-ssd/ssd/sata]
	Status   int32  `json:"status"`   // Disk Status, 0 for available, 1 for inused
	UsedBy   string `json:"usedBy,omitempty"`
}
