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
 * @File: system_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package system

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/message"
	"github.com/pingcap-inc/tiunimanager/models"
	"github.com/pingcap-inc/tiunimanager/models/platform/system"
	"github.com/pingcap-inc/tiunimanager/test/mockmodels/mocksystem"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemManager_AcceptSystemEvent(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		err := GetSystemManager().AcceptSystemEvent(context.TODO(), "")
		assert.Error(t, err)
	})
	t.Run("unknown", func(t *testing.T) {
		err := GetSystemManager().AcceptSystemEvent(context.TODO(), "unknown")
		assert.Error(t, err)
	})
}

func TestSystemManager_GetSystemInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	systemRW := mocksystem.NewMockReaderWriter(ctrl)
	models.SetSystemReaderWriter(systemRW)

	systemRW.EXPECT().GetVersion(gomock.Any(), "v1").Return(&system.VersionInfo{
		ID: "v1",
		Desc: "v1",
		ReleaseNote: "this is v1",
	}, nil).AnyTimes()
	systemRW.EXPECT().GetVersion(gomock.Any(), "v2").Return(&system.VersionInfo{
		ID: "v2",
		Desc: "v2",
		ReleaseNote: "this is v2",
	}, nil).AnyTimes()
	systemRW.EXPECT().GetVersion(gomock.Any(), "v3").Return(nil, errors.Error(errors.TIUNIMANAGER_SYSTEM_INVALID_VERSION)).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "v2",
			LastVersionID: "v1",
			State: constants.SystemRunning,
		}, nil).Times(1)

		got, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})

		assert.NoError(t, err)
		assert.NotEmpty(t, got)
		assert.Equal(t, "EM", got.Info.SystemName)
		assert.Equal(t, "111", got.Info.SystemLogo)
		assert.Equal(t, "v2", got.Info.CurrentVersionID)
		assert.Equal(t, "v1", got.Info.LastVersionID)
		assert.Equal(t, string(constants.SystemRunning), got.Info.State)
		assert.Equal(t, "v2", got.CurrentVersion.VersionID)
		assert.Equal(t, "v1", got.LastVersion.VersionID)
	})
	t.Run("error", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(nil, errors.Error(errors.TIUNIMANAGER_SYSTEM_MISSING_CONFIG)).Times(1)
		_, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})
		assert.Error(t, err)
	})
	t.Run("without version", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "v2",
			LastVersionID: "v1",
			State: constants.SystemRunning,
		}, nil).Times(1)

		got, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: false})

		assert.NoError(t, err)
		assert.Equal(t, "", got.LastVersion.VersionID)
		assert.Equal(t, "", got.CurrentVersion.VersionID)
	})
	t.Run("empty current version", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "",
			LastVersionID: "",
			State: constants.SystemRunning,
		}, nil).Times(1)

		got, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})

		assert.NoError(t, err)
		assert.Equal(t, "", got.LastVersion.VersionID)
		assert.Equal(t, "", got.CurrentVersion.VersionID)
	})
	t.Run("empty last version", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "v2",
			LastVersionID: "",
			State: constants.SystemRunning,
		}, nil).Times(1)

		got, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})

		assert.NoError(t, err)
		assert.Equal(t, "v2", got.CurrentVersion.VersionID)
		assert.Equal(t, "", got.LastVersion.VersionID)
	})
	t.Run("version error", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "v3",
			LastVersionID: "v2",
			State: constants.SystemRunning,
		}, nil).Times(1)

		_, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})
		assert.Error(t, err)
	})
	t.Run("version error", func(t *testing.T) {
		systemRW.EXPECT().GetSystemInfo(gomock.Any()).Return(&system.SystemInfo{
			SystemName: "EM",
			SystemLogo: "111",
			CurrentVersionID: "v2",
			LastVersionID: "v3",
			State: constants.SystemRunning,
		}, nil).Times(1)

		_, err := GetSystemManager().GetSystemInfo(context.TODO(), message.GetSystemInfoReq{WithVersionDetail: true})
		assert.Error(t, err)
	})
}