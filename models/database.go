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
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/pingcap-inc/tiem/models/platform/product"

	"github.com/pingcap-inc/tiem/models/tiup"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	mm "github.com/pingcap-inc/tiem/models/resource/management"
	resourcePool "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/identification"
	"github.com/pingcap-inc/tiem/models/user/tenant"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/models/parametergroup"
	"github.com/pingcap-inc/tiem/models/platform/config"
	"github.com/pingcap-inc/tiem/models/resource"
	resource_rw "github.com/pingcap-inc/tiem/models/resource/gormreadwrite"
	"github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

var defaultDb *database

type database struct {
	base                             *gorm.DB
	workFlowReaderWriter             workflow.ReaderWriter
	importExportReaderWriter         importexport.ReaderWriter
	brReaderWriter                   backuprestore.ReaderWriter
	changeFeedReaderWriter           changefeed.ReaderWriter
	clusterReaderWriter              management.ReaderWriter
	parameterGroupReaderWriter       parametergroup.ReaderWriter
	clusterParameterReaderWriter     parameter.ReaderWriter
	configReaderWriter               config.ReaderWriter
	secondPartyOperationReaderWriter secondparty.ReaderWriter
	resourceReaderWriter             resource.ReaderWriter
	tenantReaderWriter               tenant.ReaderWriter
	accountReaderWriter              account.ReaderWriter
	tokenReaderWriter                identification.ReaderWriter
	productReaderWriter              product.ProductReadWriterInterface
	tiUPConfigReaderWriter           tiup.ReaderWriter
}

func Open(fw *framework.BaseFramework, reentry bool) error {
	dbFile := fw.GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName + "?_busy_timeout=60000"
	logins := framework.LogForkFile(constants.LogFileSystem)
	// todo tidb?
	//newLogger := framework.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags),
	//	framework.Config{
	//		SlowThreshold:             time.Second,
	//		LogLevel:                  framework.Info,
	//		IgnoreRecordNotFoundError: true,
	//	},
	//)
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		//Logger: newLogger,
	})

	if err != nil || db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
		return err
	} else {
		logins.Infof("open database succeed, filepath: %s", dbFile)
	}

	defaultDb = &database{
		base: db,
	}

	defaultDb.initReaderWriters()

	if !reentry {
		err := defaultDb.initTables()
		if err != nil {
			logins.Fatalf("init tables failed, %v", err)
			return err
		}
		defaultDb.initSystemData()
	}

	return nil
}

func (p *database) migrateStream(models ...interface{}) (err error) {
	for _, model := range models {
		err = p.base.AutoMigrate(model)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("init table failed, model = %v, err = %s", models, err.Error())
			return err
		}
	}
	return nil
}

func (p *database) initTables() (err error) {
	return p.migrateStream(
		new(changefeed.ChangeFeedTask),
		new(workflow.WorkFlow),
		new(workflow.WorkFlowNode),
		new(management.Cluster),
		new(management.ClusterInstance),
		new(management.ClusterRelation),
		new(management.ClusterTopologySnapshot),
		new(importexport.DataTransportRecord),
		new(backuprestore.BackupRecord),
		new(backuprestore.BackupStrategy),
		new(config.SystemConfig),
		new(secondparty.SecondPartyOperation),
		new(parametergroup.Parameter),
		new(parametergroup.ParameterGroup),
		new(parametergroup.ParameterGroupMapping),
		new(parameter.ClusterParameterMapping),
		new(account.Account),
		new(tenant.Tenant),
		new(identification.Token),
		new(tiup.TiupConfig),
		new(resourcePool.Host),
		new(resourcePool.Disk),
		new(resourcePool.Label),
		new(mm.UsedCompute),
		new(mm.UsedPort),
		new(mm.UsedDisk),
		new(product.Zone),
		new(product.Spec),
		new(product.Product),
		new(product.ProductComponent),
	)
}

func (p *database) initReaderWriters() {
	defaultDb.changeFeedReaderWriter = changefeed.NewGormChangeFeedReadWrite(defaultDb.base)
	defaultDb.workFlowReaderWriter = workflow.NewFlowReadWrite(defaultDb.base)
	defaultDb.importExportReaderWriter = importexport.NewImportExportReadWrite(defaultDb.base)
	defaultDb.brReaderWriter = backuprestore.NewBRReadWrite(defaultDb.base)
	defaultDb.resourceReaderWriter = resource_rw.NewGormResourceReadWrite(defaultDb.base)
	defaultDb.parameterGroupReaderWriter = parametergroup.NewParameterGroupReadWrite(defaultDb.base)
	defaultDb.clusterParameterReaderWriter = parameter.NewClusterParameterReadWrite(defaultDb.base)
	defaultDb.configReaderWriter = config.NewConfigReadWrite(defaultDb.base)
	defaultDb.secondPartyOperationReaderWriter = secondparty.NewGormSecondPartyOperationReadWrite(defaultDb.base)
	defaultDb.clusterReaderWriter = management.NewClusterReadWrite(defaultDb.base)
	defaultDb.tenantReaderWriter = tenant.NewTenantReadWrite(defaultDb.base)
	defaultDb.accountReaderWriter = account.NewAccountReadWrite(defaultDb.base)
	defaultDb.tokenReaderWriter = identification.NewTokenReadWrite(defaultDb.base)
	defaultDb.productReaderWriter = product.NewProductReadWriter(defaultDb.base)
	defaultDb.tiUPConfigReaderWriter = tiup.NewGormTiupConfigReadWrite(defaultDb.base)
}

func (p *database) initSystemData() {
	tenant, err := defaultDb.tenantReaderWriter.AddTenant(context.TODO(), "EM system administration", 1, 0)

	// todo determine if default data needed
	if err == nil {
		// system admin account
		account := &account.Account{
			Entity: dbCommon.Entity{
				TenantId: tenant.ID,
			},
			Name: "admin",
		}
		account.GenSaltAndHash("admin")
		defaultDb.accountReaderWriter.AddAccount(context.TODO(), tenant.ID, account.Name, account.Salt, account.FinalHash, 0)

		// label
		for _, v := range structs.DefaultLabelTypes {
			labelRecord := new(resourcePool.Label)
			labelRecord.ConstructLabelRecord(&v)
			if err = defaultDb.base.Create(labelRecord).Error; err != nil {
				framework.LogForkFile(constants.LogFileSystem).Errorf("create label error: %s", err.Error())
				return
			}
		}

		// system config
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStorageType, ConfigValue: string(constants.StorageTypeS3)})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStoragePath, ConfigValue: constants.DefaultBackupStoragePath})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3AccessKey, ConfigValue: constants.DefaultBackupS3AccessKey})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3SecretAccessKey, ConfigValue: constants.DefaultBackupS3SecretAccessKey})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3Endpoint, ConfigValue: constants.DefaultBackupS3Endpoint})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyExportShareStoragePath, ConfigValue: constants.DefaultExportPath})
		defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyImportShareStoragePath, ConfigValue: constants.DefaultImportPath})

		// batch import parameters & default parameter group sql
		parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters.sql"
		err := syscall.Access(parameterSqlFile, syscall.F_OK)
		if !os.IsNotExist(err) {
			sqls, err := ioutil.ReadFile(parameterSqlFile)
			if err != nil {
				framework.LogForkFile(constants.LogFileSystem).Errorf("batch import parameters failed, err = %s", err.Error())
				return
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

		// import TiUP configs
		tiUPSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/tiup_configs.sql"
		err = syscall.Access(tiUPSqlFile, syscall.F_OK)
		if !os.IsNotExist(err) {
			sqls, err := ioutil.ReadFile(tiUPSqlFile)
			if err != nil {
				framework.LogForkFile(constants.LogFileSystem).Errorf("import tiupconfigs failed, err = %s", err.Error())
				return
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
	}
}

func GetChangeFeedReaderWriter() changefeed.ReaderWriter {
	return defaultDb.changeFeedReaderWriter
}

func SetChangeFeedReaderWriter(rw changefeed.ReaderWriter) {
	defaultDb.changeFeedReaderWriter = rw
}

func GetWorkFlowReaderWriter() workflow.ReaderWriter {
	return defaultDb.workFlowReaderWriter
}

func SetWorkFlowReaderWriter(rw workflow.ReaderWriter) {
	defaultDb.workFlowReaderWriter = rw
}

func GetImportExportReaderWriter() importexport.ReaderWriter {
	return defaultDb.importExportReaderWriter
}

func SetImportExportReaderWriter(rw importexport.ReaderWriter) {
	defaultDb.importExportReaderWriter = rw
}

func GetBRReaderWriter() backuprestore.ReaderWriter {
	return defaultDb.brReaderWriter
}

func GetResourceReaderWriter() resource.ReaderWriter {
	return defaultDb.resourceReaderWriter
}

func SetBRReaderWriter(rw backuprestore.ReaderWriter) {
	defaultDb.brReaderWriter = rw
}

func GetClusterReaderWriter() management.ReaderWriter {
	return defaultDb.clusterReaderWriter
}

func SetClusterReaderWriter(rw management.ReaderWriter) {
	defaultDb.clusterReaderWriter = rw
}

func GetConfigReaderWriter() config.ReaderWriter {
	return defaultDb.configReaderWriter
}

func SetConfigReaderWriter(rw config.ReaderWriter) {
	defaultDb.configReaderWriter = rw
}

func GetSecondPartyOperationReaderWriter() secondparty.ReaderWriter {
	return defaultDb.secondPartyOperationReaderWriter
}

func SetSecondPartyOperationReaderWriter(rw secondparty.ReaderWriter) {
	defaultDb.secondPartyOperationReaderWriter = rw
}

func GetParameterGroupReaderWriter() parametergroup.ReaderWriter {
	return defaultDb.parameterGroupReaderWriter
}

func SetParameterGroupReaderWriter(rw parametergroup.ReaderWriter) {
	defaultDb.parameterGroupReaderWriter = rw
}

func GetClusterParameterReaderWriter() parameter.ReaderWriter {
	return defaultDb.clusterParameterReaderWriter
}

func SetClusterParameterReaderWriter(rw parameter.ReaderWriter) {
	defaultDb.clusterParameterReaderWriter = rw
}

func GetAccountReaderWriter() account.ReaderWriter {
	return defaultDb.accountReaderWriter
}

func SetAccountReaderWriter(rw account.ReaderWriter) {
	defaultDb.accountReaderWriter = rw
}

func GetTenantReaderWriter() tenant.ReaderWriter {
	return defaultDb.tenantReaderWriter
}

func SetTenantReaderWriter(rw tenant.ReaderWriter) {
	defaultDb.tenantReaderWriter = rw
}
func GetTokenReaderWriter() identification.ReaderWriter {
	return defaultDb.tokenReaderWriter
}

func SetTokenReaderWriter(rw identification.ReaderWriter) {
	defaultDb.tokenReaderWriter = rw
}

func GetProductReaderWriter() product.ProductReadWriterInterface {
	return defaultDb.productReaderWriter
}
func SetProductReaderWriter(rw product.ProductReadWriterInterface) {
	defaultDb.productReaderWriter = rw
}

func GetTiUPConfigReaderWriter() tiup.ReaderWriter {
	return defaultDb.tiUPConfigReaderWriter
}

func SetTiUPConfigReaderWriter(rw tiup.ReaderWriter) {
	defaultDb.tiUPConfigReaderWriter = rw
}

func MockDB() {
	defaultDb = &database{}
}
