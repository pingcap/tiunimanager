package account

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

var testRW *AccountReadWrite

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	logins := framework.LogForkFile(common.LogFileSystem)

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + common.DBDirPrefix + common.DatabaseFileName
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}
			db.Migrator().CreateTable(Account{})

			testRW = NewAccountReadWrite(db)
			return nil
		},
	)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
