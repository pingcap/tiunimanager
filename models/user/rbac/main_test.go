package rbac

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
	"time"
)

var RW ReaderWriter

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	logins := framework.LogForkFile(constants.LogFileSystem)

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + constants.DBDirPrefix + constants.DatabaseFileName
			newLogger := framework.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				framework.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  framework.Info,
					IgnoreRecordNotFoundError: true,
				},
			)
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
				Logger: newLogger,
				//Logger: d.GetRootLogger(),
			})
			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}

			RW = NewRBACReadWrite(db)
			return nil
		},
	)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
