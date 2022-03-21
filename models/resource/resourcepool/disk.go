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
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	em_errors "github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/gorm"
)

type Disk struct {
	ID        string         `json:"diskId" gorm:"primaryKey"`
	HostID    string         `json:"omitempty" gorm:"not null"`
	Name      string         `json:"name" gorm:"not null;size:255"`         // [sda/sdb/nvmep0...]
	Capacity  int32          `json:"capacity"`                              // Disk size, Unit: GB
	Path      string         `json:"path" gorm:"size:255;not null"`         // Disk mount path: [/data1]
	Type      string         `json:"type" gorm:"not null"`                  // Disk type: [NVMeSSD/SSD/SATA]
	Status    string         `json:"status" gorm:"index;default:Available"` // Disk Status
	CreatedAt time.Time      `json:"-" gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt time.Time      `json:"-" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (d Disk) IsInused() bool {
	return d.Status == string(constants.DiskExhaust) || d.Status == string(constants.DiskInUsed)
}

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("host_id = ? and name = ? and path = ?", d.HostID, d.Name, d.Path).First(&Disk{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			d.ID = uuidutil.GenerateID()
			return nil
		}
		return err
	}

	return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_DISK_ALREADY_EXIST, "disk %s(%s) is existed on host %s", d.Name, d.Path, d.HostID)
}

func (d *Disk) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("ID = ?", d.ID).First(d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_DELETE_DISK_ERROR, "disk %s is not found", d.ID)
		}
		return err
	}
	if d.IsInused() {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_DISK_STILL_INUSED, "disk %s is still in used", d.ID)
	}

	return err
}

func (d *Disk) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("Path") {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update path for disk %s is not allowed", d.ID)
	}
	if tx.Statement.Changed("Type") {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update type for disk %s is not allowed", d.ID)
	}
	if tx.Statement.Changed("HostID") {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update host id for disk %s is not allowed", d.ID)
	}
	return
}

func (d *Disk) PrepareForUpdate(newDisk *Disk) (err error) {
	if newDisk.HostID != "" && newDisk.HostID != d.HostID {
		d.HostID = newDisk.HostID
	}
	if newDisk.Path != "" && newDisk.Path != d.Path {
		d.Path = newDisk.Path
	}
	if newDisk.Type != "" && newDisk.Type != d.Type {
		d.Type = newDisk.Type
	}
	if newDisk.Name != "" && newDisk.Name != d.Name {
		d.Name = newDisk.Name
	}
	if newDisk.Capacity != 0 && newDisk.Capacity != d.Capacity {
		d.Capacity = newDisk.Capacity
	}
	if newDisk.Status != "" && newDisk.Status != d.Status {
		d.Status = newDisk.Status
	}
	return nil
}

func (d *Disk) ValidateDisk(hostId string, hostDiskType string) (err error) {
	if d.Name == "" || d.Path == "" || d.Capacity <= 0 {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
			hostId, d.Name, d.Path, d.Capacity)
	}

	if d.HostID != "" {
		if hostId != "" && d.HostID != hostId {
			return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s failed, host id conflict %s vs %s",
				d.Name, d.Path, d.HostID, hostId)
		}
	}

	if d.Status != "" && !constants.DiskStatus(d.Status).IsValidStatus() {
		return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s specified a invalid status %s, [Available|Reserved]",
			d.Name, d.Path, hostId, d.Status)
	}

	if d.Type != "" {
		if err = constants.ValidDiskType(d.Type); err != nil {
			return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, %v", d.Name, d.Path, hostId, err)
		}
		if hostDiskType != "" && d.Type != hostDiskType {
			return em_errors.NewErrorf(em_errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, disk type conflict %s vs %s",
				d.Name, d.Path, hostId, d.Type, hostDiskType)
		}
	}

	return nil
}

func (d *Disk) ConstructFromDiskInfo(src *structs.DiskInfo) {
	d.HostID = src.HostId
	d.Name = src.Name
	d.Path = src.Path
	d.Capacity = src.Capacity
	d.Status = src.Status
	d.Type = src.Type
}

func (d *Disk) ToDiskInfo(dst *structs.DiskInfo) {
	dst.ID = d.ID
	dst.Name = d.Name
	dst.Path = d.Path
	dst.Capacity = d.Capacity
	dst.Status = d.Status
	dst.Type = d.Type
}
