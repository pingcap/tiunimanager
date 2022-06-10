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

/*******************************************************************************
 * @File: readwriteimpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/30
*******************************************************************************/

package tiup

import (
	"context"

	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/models/common"
	"gorm.io/gorm"
)

type GormTiupConfigReadWrite struct {
	common.GormDB
}

func NewGormTiupConfigReadWrite(db *gorm.DB) *GormTiupConfigReadWrite {
	m := &GormTiupConfigReadWrite{
		common.WrapDB(db),
	}
	return m
}

func (m *GormTiupConfigReadWrite) Create(ctx context.Context, componentType string, tiUPHome string) (*TiupConfig, error) {
	if "" == componentType {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "componenttype nil")
	}

	tiUPConfig := &TiupConfig{
		ComponentType: componentType,
		TiupHome:      tiUPHome,
	}

	return tiUPConfig, m.DB(ctx).Create(tiUPConfig).Error
}

func (m *GormTiupConfigReadWrite) Update(ctx context.Context, updateTemplate *TiupConfig) error {
	if "" == updateTemplate.ID {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "id is nil for %+v", updateTemplate)
	}

	return m.DB(ctx).Save(updateTemplate).Error
}

func (m *GormTiupConfigReadWrite) Get(ctx context.Context, id string) (*TiupConfig, error) {
	if "" == id {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "id required")
	}

	tiUPConfig := &TiupConfig{}
	err := m.DB(ctx).First(tiUPConfig, "id = ?", id).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return tiUPConfig, nil
	}
}

func (m *GormTiupConfigReadWrite) QueryByComponentType(ctx context.Context, componentType string) (*TiupConfig, error) {
	if "" == componentType {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "componenttype is required")
	}

	tiUPConfig := &TiupConfig{}
	err := m.DB(ctx).First(tiUPConfig, "component_type = ?", componentType).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return tiUPConfig, nil
	}
}
