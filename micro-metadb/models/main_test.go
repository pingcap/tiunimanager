package models

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

var MetaDB *gorm.DB
var Dao *DAOManager

func TestMain(m *testing.M) {
	testFile := GenerateID() + ".db"

	defer func() {
		os.Remove(testFile)
	}()
	f := framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			MetaDB, _ = gorm.Open(sqlite.Open(testFile), &gorm.Config{})

			err := MetaDB.Migrator().CreateTable(
				&Account{},
				&RoleBinding{},
				&Role{},
				&PermissionBinding{},
				&Permission{},
				&TestEntity{},
				&TestEntity2{},
				&TestRecord{},
				&TestData{},
				&DemandRecord{},
				&Host{},
				&Disk{},
				&Cluster{},
				&TiUPConfig{},
				&Tenant{},
				&FlowDO{},
				&TaskDO{},
				&Token{},
				&BackupRecord{},
				&RecoverRecord{},
				&ParametersRecord{},
			)
			return err
		},
	)
	Dao = new(DAOManager)
	Dao.SetDb(MetaDB)
	Dao.SetFramework(f)
	Dao.SetClusterManager(new(DAOClusterManager))
	Dao.ClusterManager().SetDb(MetaDB)
	m.Run()
}
