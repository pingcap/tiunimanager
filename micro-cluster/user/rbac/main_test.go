/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package rbac

import (
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/user/rbac"
	"github.com/pingcap/tiunimanager/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
	"time"
)

var RW rbac.ReaderWriter

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

			RW = rbac.NewRBACReadWrite(db)
			models.MockDB()
			models.SetRBACReaderWriter(RW)
			return nil
		},
	)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
