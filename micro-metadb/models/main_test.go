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
	testFile := "tmp/" + GenerateID()
	os.MkdirAll(testFile, 0755)

	defer func() {
		os.Remove(testFile)
	}()
	framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			Dao = new(DAOManager)
			Dao.InitDB(testFile)
			MetaDB = Dao.Db()
			Dao.SetAccountManager(NewDAOAccountManager(Dao.Db()))
			Dao.SetClusterManager(NewDAOClusterManager(Dao.Db()))
			return nil
		},
	)
	m.Run()
}
