
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
