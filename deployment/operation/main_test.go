/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: main_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package operation

import (
	"fmt"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRW *GormOperationReadWrite

func TestMain(m *testing.M) {
	testTiUPHome := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testTiUPHome, 0755)

	logins := framework.LogForkFile(constants.LogFileSystem)

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := fmt.Sprintf("%s/operation.db", testTiUPHome)
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %v, meta database error: %v", dbFile, err, db.Error)
				return err
			} else {
				logins.Infof("open database succeed, filepath: %s", dbFile)
			}
			db.Migrator().CreateTable(Operation{})

			testRW = NewGormOperationReadWrite(db)
			return nil
		},
	)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
