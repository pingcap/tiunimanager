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

	"github.com/pingcap-inc/tiem/common/resource"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/resource/management"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"gorm.io/gorm"
)

type GormResourceReadWrite struct {
	db *gorm.DB
}

func NewGormChangeFeedReadWrite() *GormResourceReadWrite {
	m := new(GormResourceReadWrite)
	return m
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
		if err = rw.initSystemDefaultLabels(ctx); err != nil {
			log.Errorf("init table Label failed, error: %v", err)
			return err
		}
	}
	return nil
}

func (rw *GormResourceReadWrite) initSystemDefaultLabels(ctx context.Context) (err error) {
	for _, v := range management.DefaultLabelTypes {
		err = rw.db.Create(&v).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INIT_LABELS_ERROR, "init default label table failed, error: %v", err)
		}
	}
	return nil
}

func (rw *GormResourceReadWrite) Create(ctx context.Context, hosts []rp.Host) (hostIds []string, err error) {
	tx := rw.DB().Begin()
	for _, host := range hosts {
		err = tx.Create(host).Error
		if err != nil {
			tx.Rollback()
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_CREATE_HOST_ERROR, "create %s(%s) error, %v", host.HostName, host.IP, err)
		}
		hostIds = append(hostIds, host.ID)
	}
	err = tx.Commit().Error
	return
}

func (rw *GormResourceReadWrite) Delete(ctx context.Context, hostIds []string) (err error) {
	tx := rw.DB().Begin()
	for _, hostId := range hostIds {
		var host rp.Host
		if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&host, "ID = ?", hostId).Error; err != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_LOCK_TABLE_ERROR, "lock host %s(%s) error, %v", hostId, host.IP, err)
		}
		err = tx.Delete(&host).Error
		if err != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_DELETE_HOST_ERROR, "delete host %s(%s) error, %v", hostId, host.IP, err)
		}
	}
	err = tx.Commit().Error
	return
}

func (rw *GormResourceReadWrite) Get(ctx context.Context, hostId string) (rp.Host, error)
func (rw *GormResourceReadWrite) Query(ctx context.Context, cond QueryCond) (hosts []rp.Host, total int64, err error)

func (rw *GormResourceReadWrite) UpdateHostStatus(ctx context.Context, status string) (err error)
func (rw *GormResourceReadWrite) ReserveHost(ctx context.Context, reserved bool) (err error)
func (rw *GormResourceReadWrite) GetHierarchy(ctx context.Context, filter resource.HostFilter, level int32, depth int32) (root resource.HierarchyTreeNode, err error)
func (rw *GormResourceReadWrite) GetStocks(ctx context.Context, filter StockFilter) (stock Stock, err error)
