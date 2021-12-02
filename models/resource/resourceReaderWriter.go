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

package resource

import (
	"context"

	"github.com/pingcap-inc/tiem/library/common"
	rt "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/resource/management"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"gorm.io/gorm"
)

type GormResourceReadWrite struct {
	db *gorm.DB
}

func (rw *GormResourceReadWrite) SetDB(db *gorm.DB) {
	rw.db = db
}

func (rw *GormResourceReadWrite) DB() (db *gorm.DB) {
	return db
}

func (rw *GormResourceReadWrite) addTable(ctx context.Context, tableModel interface{}) (newTable bool, err error) {
	if !rw.db.Migrator().HasTable(tableModel) {
		err := rw.db.Migrator().CreateTable(tableModel)
		if nil != err {
			return true, framework.NewTiEMErrorf(common.TIEM_RESOURCE_ADD_TABLE_ERROR, "crete table %v failed, error: %v", tableModel, err)
		}
		return true, nil
	} else {
		return false, nil
	}
}

func (rw *GormResourceReadWrite) InitTables(ctx context.Context) error {
	log := framework.LogWithContext(ctx)
	_, err := rw.addTable(ctx, new(rp.Host))
	if err != nil {
		log.Errorf("create table Host failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(rp.Disk))
	if err != nil {
		log.Errorf("create table Disk failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(management.UsedCompute))
	if err != nil {
		log.Errorf("create table UsedCompute failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(management.UsedPort))
	if err != nil {
		log.Errorf("create table UsedPort failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(management.UsedDisk))
	if err != nil {
		log.Errorf("create table UsedDisk failed, error: %v", err)
		return err
	}
	newTable, err := rw.addTable(ctx, new(management.Label))
	if err != nil {
		log.Errorf("create table Label failed, error: %v", err)
		return err
	}
	if newTable {

	}
	return nil
}

func (rw *GormResourceReadWrite) initSystemDefaultLabels(ctx context.Context) (err error) {
	for _, v := range rt.DefaultLabelTypes {
		err = rw.db.Create(&v).Error
		if err != nil {
			return err
		}
	}
	return err
}

func NewGormChangeFeedReadWrite() *GormResourceReadWrite {
	m := new(GormResourceReadWrite)
	return m
}
