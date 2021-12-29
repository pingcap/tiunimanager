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
 * @Date: 2021/12/28
*******************************************************************************/

package mirror

import (
	"context"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormMirrorReadWrite struct {
	common.GormDB
}

func NewGormMirrorReadWrite(db *gorm.DB) *GormMirrorReadWrite {
	m := &GormMirrorReadWrite{
		common.WrapDB(db),
	}
	return m
}

func (m *GormMirrorReadWrite) Create(ctx context.Context, componentType string, mirrorAddr string) (*Mirror, error) {
	if "" == componentType || "" == mirrorAddr {
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "either componentType(actual: %s) or "+
			"mirrorAddr(actual: %s) is nil", componentType, mirrorAddr)
	}

	mirror := &Mirror{
		ComponentType: componentType,
		MirrorAddr:    mirrorAddr,
	}

	return mirror, m.DB(ctx).Create(mirror).Error
}

func (m *GormMirrorReadWrite) Update(ctx context.Context, updateTemplate *Mirror) error {
	if "" == updateTemplate.ID {
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "id is nil for %+v", updateTemplate)
	}

	return m.DB(ctx).Save(updateTemplate).Error
}

func (m *GormMirrorReadWrite) Get(ctx context.Context, id string) (*Mirror, error) {
	if "" == id {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "id required")
	}

	mirror := &Mirror{}
	err := m.DB(ctx).First(mirror, "id = ?", id).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return mirror, nil
	}
}

func (m *GormMirrorReadWrite) QueryByComponentType(ctx context.Context, componentType string) (*Mirror, error) {
	if "" == componentType {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "componenttype is required")
	}

	mirror := &Mirror{}
	err := m.DB(ctx).First(mirror, "component_type = ?", componentType).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return mirror, nil
	}
}
