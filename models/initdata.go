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
 * @File: initdata.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/pingcap/tiunimanager/models/platform/product"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models/platform/config"
	"github.com/pingcap/tiunimanager/models/platform/system"
	resourcePool "github.com/pingcap/tiunimanager/models/resource/resourcepool"
	"github.com/pingcap/tiunimanager/models/user/account"
	"gorm.io/gorm"
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
		return initBySql(nil, upgradeSqlFile, "tiup upgrade")
	}},
	{"v1.0.0-beta.11", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			err := tx.Create(&config.SystemConfig{
				ConfigKey:   constants.ConfigKeyDefaultSSHPort,
				ConfigValue: "22",
			}).Error
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
			return initBySql(tx, parameterSqlFile, "parameters")
		})
	}},
	{"v1.0.0-beta.12", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return errors.OfNullable(nil).BreakIf(func() error {
				return tx.Create(&system.VersionInfo{
					ID:          "v1.0.0-beta.12",
					Desc:        "beta 12",
					ReleaseNote: "release note",
				}).Error
			}).BreakIf(func() error {
				return nil
			}).BreakIf(func() error {
				existed := false
				if existed {
					// todo migrate product infos
				} else {
					return initDefaultProductsAndVendors(tx)
				}
				return nil
			}).Present()
		})
	}},
	{"v1.0.0-beta.13", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return errors.OfNullable(nil).BreakIf(func() error {
				return tx.Create(&config.SystemConfig{
					ConfigKey:   constants.ConfigKeyExtraVMFacturer,
					ConfigValue: "",
				}).Error
			}).BreakIf(func() error {
				return tx.Create(&system.VersionInfo{
					ID:          "v1.0.0-beta.13",
					Desc:        "beta 13",
					ReleaseNote: "release note",
				}).Error
			}).BreakIf(func() error {
				parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters_v1.0.0-beta.13.sql"
				return initBySql(tx, parameterSqlFile, "parameters")
			}).Present()
		})
	}},
	{"v1.0.0", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return tx.Create(&system.VersionInfo{
				ID:          "v1.0.0",
				Desc:        "",
				ReleaseNote: "",
			}).Error
		})
	}},
	{"v1.0.1", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return errors.OfNullable(nil).BreakIf(func() error {
				return tx.Create(&system.VersionInfo{
					ID:          "v1.0.1",
					Desc:        "",
					ReleaseNote: "",
				}).Error
			}).BreakIf(func() error {
				parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters_v1.0.1.sql"
				return initBySql(tx, parameterSqlFile, "parameters")
			}).BreakIf(func() error {
				framework.LogForkFile(constants.LogFileSystem).Info("reset admin password")
				adminUser, _ := defaultDb.accountReaderWriter.GetUserByName(context.TODO(), "admin")
				if adminUser != nil && adminUser.ID != "" {
					framework.LogForkFile(constants.LogFileSystem).Infof("reset admin password for userId %s", adminUser.ID)
					adminUser.GenSaltAndHash("admin")
					b, err := json.Marshal(adminUser.FinalHash)
					if err != nil {
						return err
					}
					framework.LogForkFile(constants.LogFileSystem).Infof("reset admin password for userId %s, salt: %s, finalHash: %v", adminUser.ID, adminUser.Salt, adminUser.FinalHash)
					tx.Exec(fmt.Sprintf("update users set salt = '%s', final_hash = '%s'  where name = 'admin'", adminUser.Salt, string(b)))
					return nil
				}
				framework.LogForkFile(constants.LogFileSystem).Errorf("get empty adminUser %+v", adminUser)
				return fmt.Errorf("get empty adminUser %+v", adminUser)
			}).Present()
		})
	}},
	{"v1.0.2", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return errors.OfNullable(nil).BreakIf(func() error {
				return tx.Create(&system.VersionInfo{
					ID:          "v1.0.2",
					Desc:        "",
					ReleaseNote: "",
				}).Error
			}).BreakIf(func() error {
				parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters_v1.0.2.sql"
				return initBySql(tx, parameterSqlFile, "parameters")
			}).ContinueIf(func() error {
				tiUPSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/tiup_configs.sql"
				return initBySql(tx, tiUPSqlFile, "tiup config")
			}).If(func(err error) {
				framework.LogForkFile(constants.LogFileSystem).Errorf("init v1.0.2 data failed, err = %s", err.Error())
			}).Present()
		})
	}},
	{"v1.0.3", func() error {
		return defaultDb.base.WithContext(context.TODO()).Transaction(func(tx *gorm.DB) error {
			return errors.OfNullable(nil).BreakIf(func() error {
				return tx.Create(&system.VersionInfo{
					ID:          "v1.0.3",
					Desc:        "",
					ReleaseNote: "",
				}).Error
			}).BreakIf(func() error {
				parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters_v1.0.3.sql"
				return initBySql(tx, parameterSqlFile, "parameters")
			}).Present()
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
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyDefaultTiUPHome, ConfigValue: constants.DefaultTiUPHome})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyDefaultEMHome, ConfigValue: constants.DefaultEMHome})
		return nil
	}).BreakIf(func() error {
		framework.LogForkFile(constants.LogFileSystem).Info("init default parameters")
		parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters.sql"
		return initBySql(nil, parameterSqlFile, "parameters")
	}).If(func(err error) {
		framework.LogForkFile(constants.LogFileSystem).Errorf("init data failed, err = %s", err.Error())
	}).Else(func() {
		framework.LogForkFile(constants.LogFileSystem).Infof("init default data succeed")
	}).Present()
}

func initBySql(tx *gorm.DB, file string, module string) error {
	err := syscall.Access(file, syscall.F_OK)
	if !os.IsNotExist(err) {
		sqls, err := ioutil.ReadFile(file)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("import %s failed, err = %s", module, err.Error())
			return err
		}
		sqlArr := strings.Split(string(sqls), ";")
		if tx == nil {
			tx = defaultDb.base
		}
		for _, sql := range sqlArr {
			if strings.TrimSpace(sql) == "" {
				continue
			}
			// exec import sql
			tx.Exec(sql)
		}
	}
	return nil
}

// initDefaultProductsAndVendors
// @Description: init default products and vendors data for TiUniManager v1.0.0-beta.12
// @Parameter tx
// @return error
func initDefaultProductsAndVendors(tx *gorm.DB) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return tx.Create(&product.ProductInfo{ProductID: "TiDB", ProductName: "TiDB"}).Error
	}).BreakIf(func() error {
		return tx.CreateInBatches([]*product.ProductComponentInfo{
			{ProductID: "TiDB", ComponentID: "TiDB", ComponentName: "TiDB", PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 128, SuggestedInstancesCount: []int32{}},
			{ProductID: "TiDB", ComponentID: "TiKV", ComponentName: "TiKV", PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 128, SuggestedInstancesCount: []int32{}},
			{ProductID: "TiDB", ComponentID: "PD", ComponentName: "PD", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 128, SuggestedInstancesCount: []int32{1, 3, 5, 7}},
			{ProductID: "TiDB", ComponentID: "TiFlash", ComponentName: "TiFlash", PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 0, MaxInstance: 128, SuggestedInstancesCount: []int32{}},
			{ProductID: "TiDB", ComponentID: "CDC", ComponentName: "CDC", PurposeType: "Compute", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 0, MaxInstance: 128, SuggestedInstancesCount: []int32{}},
		}, 5).Error
	}).BreakIf(func() error {
		return tx.CreateInBatches([]*product.ProductVersion{
			{ProductID: "TiDB", Arch: string(constants.ArchX8664), Version: "v5.2.2"},
			{ProductID: "TiDB", Arch: string(constants.ArchArm64), Version: "v5.2.2"},
		}, 1).Error
	}).BreakIf(func() error {
		return tx.Create(&product.Vendor{VendorID: "Local", VendorName: "local datacenter"}).Error
	}).BreakIf(func() error {
		return tx.CreateInBatches([]product.VendorSpec{
			{VendorID: "Local", SpecID: "c.large", SpecName: "4C8G", CPU: 4, Memory: 8, DiskType: "SATA", PurposeType: "Compute"},
			{VendorID: "Local", SpecID: "c.xlarge", SpecName: "8C16G", CPU: 8, Memory: 16, DiskType: "SATA", PurposeType: "Compute"},
			{VendorID: "Local", SpecID: "st.large", SpecName: "4C8G", CPU: 4, Memory: 8, DiskType: "SATA", PurposeType: "Storage"},
			{VendorID: "Local", SpecID: "st.xlarge", SpecName: "8C16G", CPU: 8, Memory: 16, DiskType: "SATA", PurposeType: "Storage"},
			{VendorID: "Local", SpecID: "sc.large", SpecName: "4C8G", CPU: 4, Memory: 8, DiskType: "SATA", PurposeType: "Schedule"},
			{VendorID: "Local", SpecID: "sc.xlarge", SpecName: "8C16G", CPU: 8, Memory: 16, DiskType: "SATA", PurposeType: "Schedule"},
		}, 8).Error
	}).BreakIf(func() error {
		return tx.CreateInBatches([]product.VendorZone{
			{VendorID: "Local", RegionID: "Region1", RegionName: "Region1", ZoneID: "Zone1_1", ZoneName: "Zone1_1"},
			{VendorID: "Local", RegionID: "Region1", RegionName: "Region1", ZoneID: "Zone1_2", ZoneName: "Zone1_2"},
			{VendorID: "Local", RegionID: "Region2", RegionName: "Region2", ZoneID: "Zone2_1", ZoneName: "Zone2_1"},
			{VendorID: "Local", RegionID: "Region2", RegionName: "Region2", ZoneID: "Zone2_2", ZoneName: "Zone2_2"},
		}, 4).Error
	}).Present()
}
