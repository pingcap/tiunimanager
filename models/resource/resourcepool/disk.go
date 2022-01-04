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
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"time"

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

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuidutil.GenerateID()
	return nil
}
