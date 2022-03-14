/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 * @File: readwriteimpl.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package system

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type SystemReadWrite struct {
	dbCommon.GormDB
}

func NewSystemReadWrite(db *gorm.DB) ReaderWriter {
	return &SystemReadWrite{
		dbCommon.WrapDB(db),
	}
}
func (s *SystemReadWrite) QueryVersions(ctx context.Context) ([]*VersionInfo, error) {
	versions := make([]*VersionInfo, 0)
	return versions, s.DB(ctx).Find(&versions).Error
}

func (s *SystemReadWrite) GetVersion(ctx context.Context, versionID string) (*VersionInfo, error) {
	if "" == versionID {
		errInfo := "get version info failed : empty versionID"
		framework.LogWithContext(ctx).Error(errInfo)
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, errInfo)
	}

	version := &VersionInfo{}
	err := s.DB(ctx).First(version, "id = ?", versionID).Error

	if err != nil {
		errInfo := fmt.Sprintf("get version info failed : versionID = %s, err = %s", versionID, err.Error())
		framework.LogWithContext(ctx).Error(errInfo)

		return nil, errors.WrapError(errors.TIEM_SYSTEM_INVALID_VERSION, errInfo, err)
	} else {
		return version, nil
	}
}

func (s *SystemReadWrite) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{}
	err := s.DB(ctx).First(info).Error
	if err != nil {
		return nil, errors.WrapError(errors.TIEM_SYSTEM_MISSING_DATA, "", err)
	}
	return info, err
}

func (s *SystemReadWrite) UpdateState(ctx context.Context, original, target constants.SystemState) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}

	if info.State != original {
		return errors.NewErrorf(errors.TIEM_SYSTEM_STATE_CONFLICT, "current state = %s, expected original state = %s", info.State, original)
	}
	err = s.DB(ctx).Model(info).Where("system_name is not null").Update("state", target).Error
	return dbCommon.WrapDBError(err)
}

func (s *SystemReadWrite) UpdateVersion(ctx context.Context, target string) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}

	if info.CurrentVersionID == target {
		framework.LogWithContext(ctx).Infof("current version %s is equal to target version", info.CurrentVersionID)
		return nil
	}
	err = s.DB(ctx).Model(info).Where("system_name is not null").
		Update("last_version_id", info.CurrentVersionID).
		Update("current_version_id", target).
		Error
	return dbCommon.WrapDBError(err)
}

func (s *SystemReadWrite) VendorInitialized(ctx context.Context) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}
	err = s.DB(ctx).Model(info).Where("system_name is not null").
		Update("vendor_zones_initialized", true).
		Update("vendor_specs_initialized", true).
		Error
	return dbCommon.WrapDBError(err)
}

func (s *SystemReadWrite) ProductInitialized(ctx context.Context) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}
	err = s.DB(ctx).Model(info).Where("system_name is not null").
		Update("product_components_initialized", true).
		Update("product_versions_initialized", true).
		Error
	return dbCommon.WrapDBError(err)
}
