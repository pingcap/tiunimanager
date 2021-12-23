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

/*******************************************************************************
 * @File: test_main.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10
*******************************************************************************/

package specs

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	logins := framework.LogForkFile(constants.LogFileSystem)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + constants.DBDirPrefix + constants.DatabaseFileName
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}
			db.Migrator().CreateTable(Zone{})
			db.Migrator().CreateTable(Region{})
			db.Migrator().CreateTable(ResourceSpec{})
			db.Migrator().CreateTable(ProductComponent{})
			db.Migrator().CreateTable(Product{})
			db.Migrator().CreateTable(Vendor{})
			//init test data
			prw = NewProductReadWriter(db)
			return nil
		},
	)

	os.Exit(m.Run())
}
