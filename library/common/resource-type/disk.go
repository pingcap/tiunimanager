
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package resource

import (
	"errors"
	"time"

	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/gorm"
)

type DiskType string

const (
	NvmeSSD DiskType = "nvme_ssd"
	SSD     DiskType = "ssd"
	Sata    DiskType = "sata"
)

func ValidDiskType(diskType string) error {
	if diskType == string(NvmeSSD) || diskType == string(SSD) || diskType == string(Sata) {
		return nil
	}
	return errors.New("valid disk type: [nvme_ssd | ssd | sata]")
}

type DiskStatus int32

const (
	DISK_AVAILABLE DiskStatus = iota
	DISK_RESERVED
	DISK_INUSED
	DISK_EXHAUST
	DISK_ERROR
)

func (s DiskStatus) IsInused() bool {
	return s == DISK_INUSED
}

func (s DiskStatus) IsExhaust() bool {
	return s == DISK_EXHAUST
}

func (s DiskStatus) IsAvailable() bool {
	return s == DISK_AVAILABLE
}

type Disk struct {
	ID        string         `json:"diskId" gorm:"PrimaryKey"`
	HostId    string         `json:"omitempty"`
	Name      string         `json:"name" gorm:"size:255"`          // [sda/sdb/nvmep0...]
	Capacity  int32          `json:"capacity"`                      // Disk size, Unit: GB
	Path      string         `json:"path" gorm:"size:255;not null"` // Disk mount path: [/data1]
	Type      string         `json:"type"`                          // Disk type: [nvme-ssd/ssd/sata]
	Status    int32          `json:"status" gorm:"index"`           // Disk Status, 0 for available, 1 for reserved
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuidutil.GenerateID()
	return nil
}
