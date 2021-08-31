package models

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/gorm"
	"os"
	"testing"
)

var MetaDB *gorm.DB
var Dao *DAOManager

func TestMain(m *testing.M) {
	testFilePath := "tmp/" + GenerateID()
	os.MkdirAll(testFilePath, 0755)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			Dao = new(DAOManager)
			Dao.InitDB(testFilePath)
			Dao.InitTables()
			Dao.AddTable("test_entitys", new(TestEntity))
			Dao.AddTable("test_entity2_s", new(TestEntity2))
			Dao.AddTable("test_records", new(TestRecord))
			Dao.AddTable("test_datas", new(TestData))

			MetaDB = Dao.Db()
			Dao.SetAccountManager(NewDAOAccountManager(Dao.Db()))
			Dao.SetClusterManager(NewDAOClusterManager(Dao.Db()))
			return nil
		},
	)
	m.Run()
}
