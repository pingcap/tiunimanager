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
 * @File: ReadWriterImpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/5
*******************************************************************************/

package upgrade

import (
	"context"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormProductUpgradePathReadWrite struct {
	dbCommon.GormDB
}

func NewGormProductUpgradePath(db *gorm.DB) *GormProductUpgradePathReadWrite {
	m := &GormProductUpgradePathReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *GormProductUpgradePathReadWrite) Create(ctx context.Context, path *ProductUpgradePath) (*ProductUpgradePath, error) {
	return path, m.DB(ctx).Create(path).Error
}

func (m *GormProductUpgradePathReadWrite) Delete(ctx context.Context, pathId string) (err error) {
	if "" == pathId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	path := &ProductUpgradePath{}

	return m.DB(ctx).First(path, "id = ?", pathId).Delete(path).Error
}

func (m *GormProductUpgradePathReadWrite) UpdateConfig(ctx context.Context, updateTemplate *ProductUpgradePath) error {
	if "" == updateTemplate.ID {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	return m.DB(ctx).Omit("type", "product_id", "src_version", "dst_version").
		Save(updateTemplate).Error
}

func (m *GormProductUpgradePathReadWrite) Get(ctx context.Context, pathId string) (*ProductUpgradePath, error) {
	if "" == pathId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ProductUpgradePath{}
	err := m.DB(ctx).First(task, "id = ?", pathId).Error

	if err != nil {
		return nil, framework.SimpleError(common.TIEM_CHANGE_FEED_NOT_FOUND)
	} else {
		return task, nil
	}
}

func (m *GormProductUpgradePathReadWrite) QueryBySrcVersion(ctx context.Context, srcVersion string) (paths []*ProductUpgradePath, err error) {
	if "" == srcVersion {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	return paths, m.DB(ctx).Model(&ProductUpgradePath{}).
		Where("src_version = ?", srcVersion).
		Order("created_at").Find(&paths).Error
}
