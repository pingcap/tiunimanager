/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
	"time"

	"github.com/pingcap/tiunimanager/util/uuidutil"
	"gorm.io/gorm"
)

type ProductUpgradePath struct {
	ID         string         `gorm:"primaryKey;"`
	Type       string         `gorm:"not null;comment:'in-place/migration'"`
	ProductID  string         `gorm:"not null;size:64;comment:'TiDB/DM/TiUniManager'"`
	SrcVersion string         `gorm:"not null;size:64;comment:'original version of the cluster'"`
	DstVersion string         `gorm:"not null;size:64;comment:'available upgrade version of the cluster'"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (s *ProductUpgradePath) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuidutil.GenerateID()
	return nil
}

const (
	ColumnType       = "type"
	ColumnProductID  = "product_id"
	ColumnSrcVersion = "src_version"
	ColumnDstVersion = "dst_version"
	ColumnStatus     = "status"
)

const (
	StatusValid   = "valid"
	StatusInvalid = "invalid"
)
