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
 * @File: initdata.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package models

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/platform/config"
	"github.com/pingcap-inc/tiem/models/platform/system"
	resourcePool "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/pingcap-inc/tiem/models/user/account"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

var inTestingVersion = "InTesting"

var allVersionInitializers = []system.VersionInitializer{
	{"", fullDataBeforeVersions},
	{"v1.0.0-beta.10", func() error {
		defaultDb.base.Create(&system.VersionInfo{
			ID:          "v1.0.0-beta.10",
			Desc:        "beta 10",
			ReleaseNote: "release note",
		})
		upgradeSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/upgrades.sql"
		return initBySql(upgradeSqlFile, "tiup upgrade")
	}},
	{"v1.0.0-beta.11", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			_, err := defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{
				ConfigKey:   constants.ConfigKeyDefaultSSHPort,
				ConfigValue: "22",
			})
			if err != nil {
				return err
			}
			parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters_v1.0.0-beta.11.sql"
			err = tx.Create(&system.VersionInfo{
				ID:          "v1.0.0-beta.11",
				Desc:        "beta 11",
				ReleaseNote: "release note",
			}).Error
			if err != nil {
				return err
			}
			return initBySql(parameterSqlFile, "parameters")
		})
	}},

	{inTestingVersion, func() error {
		return defaultDb.base.Create(&system.VersionInfo{
			ID:          "InTesting",
			Desc:        "test version",
			ReleaseNote: "test",
		}).Error
	}},
}

func fullDataBeforeVersions() error {
	var defaultTenant *structs.TenantInfo
	return errors.OfNullable(nil).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init system info")
		defaultDb.base.Create(&system.SystemInfo{
			SystemName:       "EM",
			SystemLogo:       "",
			CurrentVersionID: "",
			LastVersionID:    "",
			State:            constants.SystemInitialing,
		})
		return nil
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default tenant")
		tenant, err := defaultDb.accountReaderWriter.CreateTenant(context.TODO(),
			&account.Tenant{
				ID:               "admin",
				Name:             "EM system administration",
				Creator:          "System",
				Status:           string(constants.TenantStatusNormal),
				OnBoardingStatus: string(constants.TenantOnBoarding)})
		defaultTenant = tenant
		return err
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default account")
		user := &account.User{
			DefaultTenantID: defaultTenant.ID,
			Name:            "admin",
			Creator:         "System",
		}
		user.GenSaltAndHash("admin")
		_, _, _, err := defaultDb.accountReaderWriter.CreateUser(context.TODO(), user, "admin")
		return err
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default labels")
		for _, v := range structs.DefaultLabelTypes {
			labelRecord := new(resourcePool.Label)
			labelRecord.ConstructLabelRecord(&v)
			if err := defaultDb.base.Create(labelRecord).Error; err != nil {
				return err
			}
		}
		return nil
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default system config")
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStorageType, ConfigValue: string(constants.StorageTypeS3)})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStoragePath, ConfigValue: constants.DefaultBackupStoragePath})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3AccessKey, ConfigValue: constants.DefaultBackupS3AccessKey})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3SecretAccessKey, ConfigValue: constants.DefaultBackupS3SecretAccessKey})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3Endpoint, ConfigValue: constants.DefaultBackupS3Endpoint})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupRateLimit, ConfigValue: constants.DefaultBackupRateLimit})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRestoreRateLimit, ConfigValue: constants.DefaultRestoreRateLimit})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupConcurrency, ConfigValue: constants.DefaultBackupConcurrency})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRestoreConcurrency, ConfigValue: constants.DefaultRestoreConcurrency})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyExportShareStoragePath, ConfigValue: constants.DefaultExportPath})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyImportShareStoragePath, ConfigValue: constants.DefaultImportPath})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyDumplingThreadNum, ConfigValue: constants.DefaultDumplingThreadNum})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRetainedPortRange, ConfigValue: constants.DefaultRetainedPortRange})
		return nil
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default parameters")
		parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters.sql"
		return initBySql(parameterSqlFile, "parameters")
	}).BreakIf(func() error {
		tiUPSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/tiup_configs.sql"
		return initBySql(tiUPSqlFile, "tiup config")
	}).If(func(err error) {
		framework.LogForkFile(constants.LogFileSystem).Errorf("init data failed, err = %s", err.Error())
	}).Else(func() {
		framework.LogForkFile(constants.LogFileSystem).Infof("init default data succeed")
	}).Present()
}

func initBySql(file string, module string) error {
	err := syscall.Access(file, syscall.F_OK)
	if !os.IsNotExist(err) {
		sqls, err := ioutil.ReadFile(file)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("import %s failed, err = %s", module, err.Error())
			return err
		}
		sqlArr := strings.Split(string(sqls), ";")
		for _, sql := range sqlArr {
			if strings.TrimSpace(sql) == "" {
				continue
			}
			// exec import sql
			defaultDb.base.Exec(sql)
		}
	}
	return nil
}
