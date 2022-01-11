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
 * @File: zone.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package product

import (
	"gorm.io/gorm"
	"time"
)

// Zone information provided by Enterprise Manager
type Zone struct {
	VendorID   string         `gorm:"primaryKey;size:32"`
	VendorName string         `gorm:"primaryKey;size:32"`
	RegionID   string         `gorm:"primaryKey;size:32"`
	RegionName string         `gorm:"primaryKey;size:32"`
	ZoneID     string         `gorm:"primaryKey;size:32"`
	ZoneName   string         `gorm:"primaryKey;size:32"`
	Comment    string         `gorm:"size:1024;"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:""`
}
