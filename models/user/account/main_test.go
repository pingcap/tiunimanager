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

package account

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

var testRW *AccountReadWrite

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
					SlowThreshold: time.Second,
					LogLevel:      framework.Info,
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
			db.Migrator().CreateTable(User{})
			db.Migrator().CreateTable(Tenant{})
			db.Migrator().CreateTable(UserLogin{})
			db.Migrator().CreateTable(UserTenantRelation{})

			testRW = NewAccountReadWrite(db)
			return nil
		},
	)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
