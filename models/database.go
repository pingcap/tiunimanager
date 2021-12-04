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

package models

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var defaultDb *database

type database struct {
	base                   *gorm.DB
	changeFeedReaderWriter changefeed.ReaderWriter
}

func Open(fw *framework.BaseFramework, reentry bool) error {
	fw.GetDataDir()
	dbFile := fw.GetDataDir() + common.DBDirPrefix + common.SqliteFileName
	logins := framework.LogForkFile(common.LogFileSystem)
	// todo how to
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil || db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
		return err
	} else {
		logins.Infof("open database succeed, filepath: %s", dbFile)
	}

	defaultDb = &database{
		base: db,
	}

	if !reentry {
		defaultDb.initTables()
		defaultDb.initSystemData()
	}

	defaultDb.initReaderWriters()

	return nil
}

func (p *database) initTables() {
	p.addTable(new(changefeed.ChangeFeedTask))

	// other tables
}

func (p *database) initReaderWriters() {
	defaultDb.changeFeedReaderWriter = changefeed.NewGormChangeFeedReadWrite(defaultDb.base)
}

func (p *database) initSystemData() {
	// todo
}

func (p *database) addTable(gormModel interface{}) error {
	log := framework.LogForkFile(common.LogFileSystem)
	if !p.base.Migrator().HasTable(gormModel) {
		err := p.base.Migrator().CreateTable(gormModel)
		if err != nil {
			log.Errorf("create table failed, error : %v.", err)
			return err
		}
	}

	return nil
}

func GetChangeFeedReaderWriter() changefeed.ReaderWriter {
	return defaultDb.changeFeedReaderWriter
}
