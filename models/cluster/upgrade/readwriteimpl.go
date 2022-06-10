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
 * @File: ReadWriterImpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/5
*******************************************************************************/

package upgrade

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiunimanager/common/errors"

	"github.com/pingcap-inc/tiunimanager/common/constants"

	dbCommon "github.com/pingcap-inc/tiunimanager/models/common"
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

func (m *GormProductUpgradePathReadWrite) Create(ctx context.Context, upgradeType constants.UpgradeType,
	emProductIDType constants.EMProductIDType, srcVersion string, dstVersion string) (*ProductUpgradePath, error) {
	if "" == upgradeType || "" == emProductIDType || "" == srcVersion || "" == dstVersion {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "at least one of upgradetype(actual: "+
			"%s), emproductidtype(actual: %s), srcversion(actual: %s), dstversion(actual: %s) is nil", upgradeType,
			emProductIDType, srcVersion, dstVersion)
	}

	exist, err := m.queryByPathParam(ctx, upgradeType, emProductIDType, srcVersion, dstVersion)
	if err == nil && exist.ID != "" {
		return exist, nil
	}

	path := &ProductUpgradePath{
		Type:       string(upgradeType),
		ProductID:  string(emProductIDType),
		SrcVersion: srcVersion,
		DstVersion: dstVersion,
	}

	return path, m.DB(ctx).Create(path).Error
}

// Check if the path has been created and not soft deleted
func (m *GormProductUpgradePathReadWrite) queryByPathParam(ctx context.Context, upgradeType constants.UpgradeType,
	emProductIDType constants.EMProductIDType, srcVersion string, dstVersion string) (*ProductUpgradePath, error) {
	if "" == upgradeType || "" == emProductIDType || "" == srcVersion || "" == dstVersion {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "at least one of upgradetype(actual: "+
			"%s), emproductidtype(actual: %s), srcversion(actual: %s), dstversion(actual: %s) is nil", upgradeType,
			emProductIDType, srcVersion, dstVersion)
	}

	path := &ProductUpgradePath{}

	return path, m.DB(ctx).Where(fmt.Sprintf("%s = ? and %s = ? and %s = ? and %s = ?", ColumnType, ColumnProductID,
		ColumnSrcVersion, ColumnDstVersion), string(upgradeType), string(emProductIDType), srcVersion, dstVersion).Find(path).Error
}

func (m *GormProductUpgradePathReadWrite) Get(ctx context.Context, id string) (*ProductUpgradePath, error) {
	if "" == id {
		return nil, errors.Error(errors.TIUNIMANAGER_PARAMETER_INVALID)
	}

	path := &ProductUpgradePath{}
	err := m.DB(ctx).First(path, "id = ?", id).Error

	if err != nil {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_UPGRADE_QUERY_PATH_FAILED, err.Error())
	} else {
		return path, nil
	}
}

func (m *GormProductUpgradePathReadWrite) QueryBySrcVersion(ctx context.Context, srcVersion string) (paths []*ProductUpgradePath, err error) {
	if "" == srcVersion {
		return nil, errors.Error(errors.TIUNIMANAGER_PARAMETER_INVALID)
	}

	return paths, m.DB(ctx).Model(&ProductUpgradePath{}).
		Where(fmt.Sprintf("%s = ?", ColumnSrcVersion), srcVersion).
		Order("created_at").Find(&paths).Error
}

func (m *GormProductUpgradePathReadWrite) Delete(ctx context.Context, id string) (err error) {
	if "" == id {
		return errors.Error(errors.TIUNIMANAGER_PARAMETER_INVALID)
	}
	path := &ProductUpgradePath{}

	return m.DB(ctx).First(path, "id = ?", id).Delete(path).Error
}
