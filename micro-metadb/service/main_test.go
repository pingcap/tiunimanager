package service

import (
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"gorm.io/gorm"
)

var MetaDB *gorm.DB
var Dao *models.DAOManager
var handler *DBServiceHandler

func TestMain(m *testing.M) {
	testFilePath := "tmp/" + uuidutil.GenerateID()
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
			Dao.SetResourceManager(models.NewDAOResourceManager(Dao.Db()))
			return nil
		},
	)
	m.Run()
}
