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
 ******************************************************************************/

package models

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/library/framework"
	"github.com/pingcap-inc/tiunimanager/models/platform/system"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGetReaderWriter(t *testing.T) {
	err := Open(framework.Current.(*framework.BaseFramework))
	assert.NoError(t, err)
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()
	assert.NotEmpty(t, GetBRReaderWriter())
	SetBRReaderWriter(nil)
	assert.Empty(t, GetBRReaderWriter())

	assert.NotEmpty(t, GetChangeFeedReaderWriter())
	SetChangeFeedReaderWriter(nil)
	assert.Empty(t, GetChangeFeedReaderWriter())

	assert.NotEmpty(t, GetWorkFlowReaderWriter())
	SetWorkFlowReaderWriter(nil)
	assert.Empty(t, GetWorkFlowReaderWriter())

	assert.NotEmpty(t, GetImportExportReaderWriter())
	SetImportExportReaderWriter(nil)
	assert.Empty(t, GetImportExportReaderWriter())

	assert.NotEmpty(t, GetResourceReaderWriter())
	SetResourceReaderWriter(nil)
	assert.Empty(t, GetResourceReaderWriter())

	assert.NotEmpty(t, GetClusterReaderWriter())
	SetClusterReaderWriter(nil)
	assert.Empty(t, GetClusterReaderWriter())

	assert.NotEmpty(t, GetConfigReaderWriter())
	SetConfigReaderWriter(nil)
	assert.Empty(t, GetConfigReaderWriter())

	assert.NotEmpty(t, GetSecondPartyOperationReaderWriter())
	SetSecondPartyOperationReaderWriter(nil)
	assert.Empty(t, GetSecondPartyOperationReaderWriter())

	assert.NotEmpty(t, GetParameterGroupReaderWriter())
	SetParameterGroupReaderWriter(nil)
	assert.Empty(t, GetParameterGroupReaderWriter())

	assert.NotEmpty(t, GetClusterParameterReaderWriter())
	SetClusterParameterReaderWriter(nil)
	assert.Empty(t, GetClusterParameterReaderWriter())

	assert.NotEmpty(t, GetAccountReaderWriter())
	SetAccountReaderWriter(nil)
	assert.Empty(t, GetAccountReaderWriter())

	assert.NotEmpty(t, GetTokenReaderWriter())
	SetTokenReaderWriter(nil)
	assert.Empty(t, GetTokenReaderWriter())

	assert.NotEmpty(t, GetTiUPConfigReaderWriter())
	SetTiUPConfigReaderWriter(nil)
	assert.Empty(t, GetTiUPConfigReaderWriter())

	assert.NotEmpty(t, GetProductReaderWriter())
	SetProductReaderWriter(nil)
	assert.Empty(t, GetProductReaderWriter())

	assert.NotEmpty(t, GetUpgradeReaderWriter())
	SetUpgradeReaderWriter(nil)
	assert.Empty(t, GetUpgradeReaderWriter())

	assert.NotEmpty(t, GetSystemReaderWriter())
	SetSystemReaderWriter(nil)
	assert.Empty(t, GetSystemReaderWriter())
}

func Test_Open(t *testing.T) {
	// open
	err := Open(framework.Current.(*framework.BaseFramework))
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()
	assert.NoError(t, err)

	// reopen
	err = Open(framework.Current.(*framework.BaseFramework))
	assert.NoError(t, err)
}

var mockVersionInitializers =  []system.VersionInitializer {
	{"", fullDataBeforeVersions},
	{"v1", func() error {
		return defaultDb.base.Create(&system.VersionInfo {
			ID: "v1",
			Desc: "v1",
			ReleaseNote: "v1",
		}).Error
	}},
	{"v2", func() error {
		return defaultDb.base.Create(&system.VersionInfo {
			ID: "v2",
			Desc: "v2",
			ReleaseNote: "v2",
		}).Error
	}},
	{"v3", func() error {
		return defaultDb.base.Create(&system.VersionInfo {
			ID: "v3",
			Desc: "v3",
			ReleaseNote: "v3",
		}).Error
	}},
	{"v4", func() error {
		return defaultDb.base.Create(&system.VersionInfo {
			ID: "v4",
			Desc: "v4",
			ReleaseNote: "v4",
		}).Error
	}},
}

func Test_MockVersionData(t *testing.T) {
	temp := allVersionInitializers
	defer func() {
		allVersionInitializers = temp
	}()
	allVersionInitializers = mockVersionInitializers

	// open empty
	err := Open(framework.Current.(*framework.BaseFramework))
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()
	assert.NoError(t, err)

	// return error if targetVersion is empty
	err = IncrementVersionData("", "")
	assert.Error(t, err)

	// reopen and upgrade from empty version to v2
	err = IncrementVersionData("", "v2")
	assert.NoError(t, err)
	v := &system.VersionInfo{}

	// v1 existed
	v.ID = "v1"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v2 existed
	v.ID = "v2"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v3 is not existed
	v.ID = "v3"
	err = defaultDb.base.First(v).Error
	assert.Error(t, err)

	// reopen
	err = IncrementVersionData("v2", "v2")
	assert.NoError(t, err)

	// reopen and upgrade from v2 to v3
	err = IncrementVersionData("v2", "v3")
	assert.NoError(t, err)
	// v1 existed
	v.ID = "v1"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v2 existed
	v.ID = "v2"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v3 existed
	v.ID = "v3"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// reopen and upgrade from v3 to v1
	err = IncrementVersionData("v3", "v1")
	assert.Error(t, err)

	// reopen and upgrade to empty version
	err = IncrementVersionData("v3", "")
	assert.Error(t, err)
}

func Test_RealVersionData(t *testing.T) {
	temp := allVersionInitializers
	defer func() {
		allVersionInitializers = temp
	}()
	allVersionInitializers = mockVersionInitializers

	// open empty
	err := Open(framework.Current.(*framework.BaseFramework))
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()
	assert.NoError(t, err)

	// return error if targetVersion is empty
	err = IncrementVersionData("", "")
	assert.Error(t, err)

	// reopen and upgrade from empty version to v2
	err = IncrementVersionData("", "v2")
	assert.NoError(t, err)
	v := &system.VersionInfo{}

	// v1 existed
	v.ID = "v1"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v2 existed
	v.ID = "v2"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v3 is not existed
	v.ID = "v3"
	err = defaultDb.base.First(v).Error
	assert.Error(t, err)

	// reopen and upgrade from v2 to v3
	err = IncrementVersionData("v2", "v3")
	assert.NoError(t, err)
	// v1 existed
	v.ID = "v1"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v2 existed
	v.ID = "v2"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// v3 existed
	v.ID = "v3"
	err = defaultDb.base.First(v).Error
	assert.NoError(t, err)

	// reopen and upgrade from v3 to v1
	err = IncrementVersionData("v3", "v1")
	assert.Error(t, err)

	// reopen and upgrade to empty version
	err = IncrementVersionData("v3", "")
	assert.Error(t, err)
}

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			return os.MkdirAll(testFilePath, 0755)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func TestTransaction(t *testing.T) {
	Open(framework.Current.(*framework.BaseFramework))
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()

	t.Run("normal", func(t *testing.T) {
		defer GetSystemReaderWriter().UpdateState(context.TODO(), "Running", "Initialing")
		err := Transaction(context.TODO(), func(transactionCtx context.Context) error {
			GetSystemReaderWriter().UpdateVersion(transactionCtx, "999")
			return GetSystemReaderWriter().UpdateState(transactionCtx, "Initialing", "Running")
		})
		assert.NoError(t, err)
		info, err := GetSystemReaderWriter().GetSystemInfo(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, "999", info.CurrentVersionID)
		assert.Equal(t, "Running", string(info.State))
		// revert
	})

	t.Run("rollback", func(t *testing.T) {
		GetSystemReaderWriter().UpdateVersion(context.TODO(), "444")
		err := Transaction(context.TODO(), func(transactionCtx context.Context) error {
			GetSystemReaderWriter().UpdateVersion(transactionCtx, "333")
			return GetSystemReaderWriter().UpdateState(transactionCtx, "Upgrading", "Running")
		})
		assert.Error(t, err)
		info, err := GetSystemReaderWriter().GetSystemInfo(context.TODO())
		assert.NoError(t, err)
		assert.NotEqual(t, "333", info.CurrentVersionID)
		assert.Equal(t, "444", info.CurrentVersionID)
		assert.NotEqual(t, "Running", string(info.State))
		assert.Equal(t, "Initialing", string(info.State))
	})

	t.Run("not in a transaction", func(t *testing.T) {
		GetSystemReaderWriter().UpdateVersion(context.TODO(), "555")

		err := Transaction(context.TODO(), func(transactionCtx context.Context) error {
			// timeout
			te := GetSystemReaderWriter().UpdateVersion(context.TODO(), "666")
			assert.Error(t, te)
			return GetSystemReaderWriter().UpdateState(transactionCtx, "Upgrading", "Running")
		})
		assert.Error(t, err)
		info, err := GetSystemReaderWriter().GetSystemInfo(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, "555", info.CurrentVersionID)
		assert.NotEqual(t, "Running", string(info.State))
		assert.Equal(t, "Initialing", string(info.State))
	})

	t.Run("mock db", func(t *testing.T) {
		db := defaultDb
		defer func() {
			defaultDb = db
		}()
		MockDB()
		err := Transaction(context.TODO(), func(transactionCtx context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("concurrency", func(t *testing.T) {
		go func() {
			Transaction(context.TODO(), func(transactionCtx context.Context) error {
				fmt.Println("111 begin")
				e := GetSystemReaderWriter().UpdateVersion(transactionCtx, "111")
				assert.NoError(t, e)
				fmt.Println("111 update")

				time.Sleep(time.Second * 2)
				fmt.Println("111 commit")
				return e
			})
		}()
		go func() {
			time.Sleep(time.Millisecond * 5)
			Transaction(context.TODO(), func(transactionCtx context.Context) error {
				fmt.Println("222 begin")
				e := GetSystemReaderWriter().UpdateVersion(transactionCtx, "222")
				assert.NoError(t, e)
				fmt.Println("222 update")

				fmt.Println("222 commit")
				return e
			})
		}()
		time.Sleep(time.Second * 5)
	})
}