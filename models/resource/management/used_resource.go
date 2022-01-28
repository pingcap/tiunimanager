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

package management

import (
	"errors"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"time"

	em_errors "github.com/pingcap-inc/tiem/common/errors"
	"gorm.io/gorm"
)

type Holder struct {
	HolderId  string `gorm:"index"` // who(clusterId) hold the resource
	RequestId string `gorm:"index"` // the resource is allocated in which request
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UsedCompute struct {
	Holder
	ID       string `gorm:"PrimaryKey"`
	HostId   string `gorm:"index;not null"`
	CpuCores int32
	Memory   int32
}

func (d *UsedCompute) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuidutil.GenerateID()
	return nil
}

type UsedPort struct {
	Holder
	ID     string `gorm:"PrimaryKey"`
	HostId string `gorm:"index;not null"`
	Port   int32
}

func (d *UsedPort) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("host_id = ? and port = ?", d.HostId, d.Port).First(&UsedPort{}).Error
	if err == nil {
		return em_errors.NewErrorf(em_errors.TIEM_SQL_ERROR, "port %d in host(%s) is already inused", d.Port, d.HostId)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		d.ID = uuidutil.GenerateID()
		return nil
	} else {
		return err
	}
}

type UsedDisk struct {
	Holder
	ID       string `gorm:"PrimaryKey"`
	HostId   string `gorm:"not null"`
	DiskId   string `gorm:"not null"`
	Capacity int32
}

func (d *UsedDisk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuidutil.GenerateID()
	return nil
}
