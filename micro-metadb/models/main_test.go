package models

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testFile := GenerateID() + ".db"

	defer func() {
		os.Remove(testFile)
	}()
	framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			MetaDB, _ = gorm.Open(sqlite.Open(testFile), &gorm.Config{})

			err := MetaDB.Migrator().CreateTable(
				&AccountDO{},
				&RoleBindingDO{},
				&RoleDO{},
				&PermissionBindingDO{},
				&PermissionDO{},
				&TestEntity{},
				&TestEntity2{},
				&TestRecord{},
				&TestData{},
				&DemandRecordDO{},
				&Host{},
				&Disk{},
				&ClusterDO{},
				&TiUPConfigDO{},
				&Tenant{},
				&FlowDO{},
				&TaskDO{},
				&Token{},
				&BackupRecordDO{},
				&RecoverRecordDO{},
				&ParametersRecordDO{},
			)
			return err
		},
	)
	m.Run()
}

