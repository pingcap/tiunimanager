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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
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

func (s *SystemReadWrite) GetHistoryVersions(ctx context.Context) ([]*VersionInfo, error) {
	panic("implement me")
}

func (s *SystemReadWrite) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{}
	err := s.DB(ctx).First(info).Error
	if err != nil {
		return nil, errors.WrapError(errors.TIEM_SYSTEM_MISSING_DATA, "", err)
	}
	return info, err
}

func (s *SystemReadWrite) UpdateStatus(ctx context.Context, original, target constants.SystemStatus) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}

	if info.Status != original {
		return errors.NewErrorf(errors.TIEM_SYSTEM_STATUS_CONFLICT, "current status = %s, expected original status = %s", info.Status, original)
	}
	err = s.DB(ctx).First(info).Update("status", target).Error
	if err != nil {
		return dbCommon.WrapDBError(err)
	}
	return nil
}

func (s *SystemReadWrite) UpdateVersion(ctx context.Context, target constants.SystemStatus) error {
	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		return err
	}
	err = s.DB(ctx).First(info).Update("last_version_id", info.CurrentVersionID).Update("current_version_id", target).Error

	if err != nil {
		return dbCommon.WrapDBError(err)
	}
	return nil
}





