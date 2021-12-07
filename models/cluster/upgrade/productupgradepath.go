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

/*******************************************************************************
 * @File: productupgradepath
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/3
*******************************************************************************/

package upgrade

import (
	"gorm.io/gorm"
	"time"
)

type ProductUpgradePath struct {
	ID         string         `gorm:"primaryKey;"`
	Type       string         `gorm:"not null;comment:'in-place/migration'"`
	ProductID  string         `gorm:"not null;size:64;comment:'TiDB/DM/TiEM'"`
	SrcVersion string         `gorm:"index:path_index,unique;not null;size:64;comment:'original version of the cluster'"`
	DstVersion string         `gorm:"index:path_index,unique;not null;size:64;comment:'available upgrade version of the cluster'"`
	Status     string         `gorm:"not null;size:32;"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type DeploymentPackage struct {
	ID          string         `gorm:"primaryKey;"`
	ProductID   string         `gorm:"not null;size:64;comment:'TiDB/DM/TiEM'"`
	Version     string         `gorm:"not null;size:64;comment:'installed version, separated by comma, no space'"`
	Platforms   string         `gorm:"not null;size:64;comment:'supported platforms, separated by comma, no space'"`
	Url         string         `gorm:"not null"`
	Description string         `gorm:"size:64;comment:'description'"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
