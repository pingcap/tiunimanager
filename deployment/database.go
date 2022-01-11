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
 * @File: database
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package deployment

import (
	"fmt"

	"github.com/pingcap-inc/tiem/deployment/operation"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var defaultDb *database

type database struct {
	base                  *gorm.DB
	operationReaderWriter operation.ReaderWriter
}

const DatabaseFileName string = "operation.db"

func Open(tiUPHome string, reentry bool) error {
	dbFile := fmt.Sprintf("%s/storage/%s?_busy_timeout=60000", tiUPHome, DatabaseFileName)
	logins := framework.LogForkFile(constants.LogFileSystem)
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil || db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %v, meta database error: %v", dbFile, err, db.Error)
		return err
	} else {
		logins.Infof("open database succeed, filepath: %s", dbFile)
	}

	defaultDb = &database{
		base: db,
	}

	defaultDb.initReaderWriters()

	if !reentry {
		err := defaultDb.initTables()
		if err != nil {
			logins.Fatalf("init tables failed, %v", err)
			return err
		}
		defaultDb.initSystemData()
	}

	return nil
}

func (p *database) migrateStream(models ...interface{}) (err error) {
	for _, model := range models {
		err = p.base.AutoMigrate(model)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("init table failed, model = %v, err = %s", models, err.Error())
			return err
		}
	}
	return nil
}

func (p *database) initTables() (err error) {
	return p.migrateStream(
		new(operation.Operation),
	)
}

func (p *database) initReaderWriters() {
	defaultDb.operationReaderWriter = operation.NewGormOperationReadWrite(defaultDb.base)
}

func (p *database) initSystemData() {

}

func GetOperationReaderWriter() operation.ReaderWriter {
	return defaultDb.operationReaderWriter
}

func SetOperationReaderWriter(rw operation.ReaderWriter) {
	defaultDb.operationReaderWriter = rw
}

func MockDB() {
	defaultDb = &database{}
}
