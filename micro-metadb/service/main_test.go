package service

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"gorm.io/gorm"
	"os"
	"testing"
)

var MetaDB *gorm.DB
var Dao *models.DAOManager
var handler *DBServiceHandler

func TestMain(m *testing.M) {
	testFilePath := "tmp/" + models.GenerateID()
	os.MkdirAll(testFilePath, 0755)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			handler = NewDBServiceHandler(testFilePath, d)

			Dao = handler.Dao()
			Dao.InitDB(testFilePath)
			Dao.InitTables()

			MetaDB = Dao.Db()
			Dao.SetAccountManager(models.NewDAOAccountManager(Dao.Db()))
			Dao.SetClusterManager(models.NewDAOClusterManager(Dao.Db()))
			return nil
		},
	)
	m.Run()
}
