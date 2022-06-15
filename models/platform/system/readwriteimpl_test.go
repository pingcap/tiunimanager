/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: readwriteimpl_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package system

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemReadWrite_GetSystemInfo(t *testing.T) {
	info, err := testRW.GetSystemInfo(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_MISSING_DATA, err.(errors.EMError).GetCode())

	systemInfo := &SystemInfo{
		SystemName:       "EM",
		SystemLogo:       "sss",
		CurrentVersionID: "v1",
		LastVersionID:    "",
		State:            constants.SystemRunning,
	}
	err = testRW.DB(context.TODO()).Create(systemInfo).Error
	defer testRW.DB(context.TODO()).Delete(&SystemInfo{}, "system_name = 'EM'")

	info, err = testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "EM", info.SystemName)
}

func TestSystemReadWrite_VersionInfo(t *testing.T) {
	testRW.DB(context.TODO()).Create(&VersionInfo{
		"v1", "test", "",
	})
	testRW.DB(context.TODO()).Create(&VersionInfo{
		"v2", "test", "",
	})
	testRW.DB(context.TODO()).Create(&VersionInfo{
		"v3", "test", "",
	})
	defer testRW.DB(context.TODO()).Delete(&SystemInfo{}, "desc = 'test'")

	t.Run("query", func(t *testing.T) {
		versions, err := testRW.QueryVersions(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 3, len(versions))
		assert.Equal(t, "v2", versions[1].ID)
	})

	t.Run("get v2", func(t *testing.T) {
		version, err := testRW.GetVersion(context.TODO(), "v2")
		assert.NoError(t, err)
		assert.Equal(t, "v2", version.ID)
	})
	t.Run("get empty", func(t *testing.T) {
		_, err := testRW.GetVersion(context.TODO(), "")
		assert.Error(t, err)
		assert.Equal(t, errors.TIUNIMANAGER_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})
	t.Run("get not found", func(t *testing.T) {
		_, err := testRW.GetVersion(context.TODO(), "v5")
		assert.Error(t, err)
		assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_INVALID_VERSION, err.(errors.EMError).GetCode())
	})
}

func TestSystemReadWrite_UpdateState(t *testing.T) {
	err := testRW.UpdateState(context.TODO(), constants.SystemUnserviceable, constants.SystemRunning)
	assert.Error(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_MISSING_DATA, err.(errors.EMError).GetCode())

	systemInfo := &SystemInfo{
		SystemName:       "EM",
		SystemLogo:       "sss",
		CurrentVersionID: "",
		LastVersionID:    "",
		State:            constants.SystemRunning,
	}
	err = testRW.DB(context.TODO()).Create(systemInfo).Error
	defer testRW.DB(context.TODO()).Delete(&SystemInfo{}, "system_name = 'EM'")
	assert.NoError(t, err)

	t.Run("conflict", func(t *testing.T) {
		err = testRW.UpdateState(context.TODO(), constants.SystemUpgrading, constants.SystemServiceReady)
		assert.Error(t, err)
		newInfo, _ := testRW.GetSystemInfo(context.TODO())
		assert.Equal(t, constants.SystemRunning, newInfo.State)
	})

	t.Run("normal", func(t *testing.T) {
		err = testRW.UpdateState(context.TODO(), constants.SystemRunning, constants.SystemServiceReady)
		assert.NoError(t, err)
		newInfo, _ := testRW.GetSystemInfo(context.TODO())
		assert.Equal(t, constants.SystemServiceReady, newInfo.State)

		err = testRW.UpdateState(context.TODO(), constants.SystemRunning, constants.SystemServiceReady)
		assert.Error(t, err)

		err = testRW.UpdateState(context.TODO(), constants.SystemServiceReady, constants.SystemRunning)
		assert.NoError(t, err)
	})
}

func TestSystemReadWrite_UpdateVersion(t *testing.T) {
	err := testRW.UpdateVersion(context.TODO(), "v1")
	assert.Error(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_MISSING_DATA, err.(errors.EMError).GetCode())

	systemInfo := &SystemInfo{
		SystemName:       "EM",
		SystemLogo:       "sss",
		CurrentVersionID: "",
		LastVersionID:    "",
		State:            constants.SystemRunning,
	}
	err = testRW.DB(context.TODO()).Create(systemInfo).Error
	defer testRW.DB(context.TODO()).Delete(&SystemInfo{}, "system_name = 'EM'")
	assert.NoError(t, err)

	err = testRW.UpdateVersion(context.TODO(), "v1")
	assert.NoError(t, err)

	newInfo, err := testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "", newInfo.LastVersionID)
	assert.Equal(t, "v1", newInfo.CurrentVersionID)

	err = testRW.UpdateVersion(context.TODO(), "v2")
	assert.NoError(t, err)

	newInfo, err = testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "v1", newInfo.LastVersionID)
	assert.Equal(t, "v2", newInfo.CurrentVersionID)

	err = testRW.UpdateVersion(context.TODO(), "v2")
	assert.NoError(t, err)

	newInfo, err = testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "v1", newInfo.LastVersionID)
	assert.Equal(t, "v2", newInfo.CurrentVersionID)
}

func TestSystemReadWrite_Initialized(t *testing.T) {
	err := testRW.VendorInitialized(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_MISSING_DATA, err.(errors.EMError).GetCode())

	err = testRW.ProductInitialized(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_SYSTEM_MISSING_DATA, err.(errors.EMError).GetCode())

	systemInfo := &SystemInfo{
		SystemName:       "EM",
		SystemLogo:       "sss",
		CurrentVersionID: "",
		LastVersionID:    "",
		State:            constants.SystemRunning,
	}
	err = testRW.DB(context.TODO()).Create(systemInfo).Error
	defer testRW.DB(context.TODO()).Delete(&SystemInfo{}, "system_name = 'EM'")
	assert.NoError(t, err)

	info, err := testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, false, info.VendorSpecsInitialized)
	assert.Equal(t, false, info.VendorZonesInitialized)
	assert.Equal(t, false, info.ProductVersionsInitialized)
	assert.Equal(t, false, info.ProductComponentsInitialized)

	err = testRW.VendorInitialized(context.TODO())
	assert.NoError(t, err)

	err = testRW.ProductInitialized(context.TODO())
	assert.NoError(t, err)

	newInfo, err := testRW.GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, true, newInfo.VendorSpecsInitialized)
	assert.Equal(t, true, newInfo.VendorZonesInitialized)
	assert.Equal(t, true, newInfo.ProductVersionsInitialized)
	assert.Equal(t, true, newInfo.ProductComponentsInitialized)

}
