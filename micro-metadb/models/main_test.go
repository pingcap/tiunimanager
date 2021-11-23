
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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/gorm"
	"os"
	"testing"
)

var MetaDB *gorm.DB
var Dao *DAOManager

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.MetaDBService,
		func(d *framework.BaseFramework) error {
			Dao = NewDAOManager(d)
			Dao.InitDB(testFilePath)
			Dao.InitTables()
			Dao.AddTable("test_entitys", new(TestEntity))
			Dao.AddTable("test_entity2_s", new(TestEntity2))
			Dao.AddTable("test_records", new(TestRecord))
			Dao.AddTable("test_datas", new(TestData))

			MetaDB = Dao.Db()

			Dao.SetAccountManager(NewDAOAccountManager(Dao.Db()))
			Dao.SetClusterManager(NewDAOClusterManager(Dao.Db()))
			Dao.SetResourceManager(NewDAOResourceManager(Dao.Db()))
			Dao.SetChangeFeedTaskManager(NewDAOChangeFeedManager(Dao.Db()))
			return nil
		},
	)
	os.Exit(m.Run())
}
