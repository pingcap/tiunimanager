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

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
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
	//m.db, _ = gorm.Open(sqlite.Open("./t2.db"), &gorm.Config{})
	//m.InitTables(context.TODO())
	return m
}

func (rw *GormResourceReadWrite) SetDB(db *gorm.DB) {
	rw.db = db
}

func (rw *GormResourceReadWrite) DB() (db *gorm.DB) {
	return rw.db
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
		err = tx.Create(&host).Error
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

func (rw *GormResourceReadWrite) Query(ctx context.Context, filter *structs.HostFilter, offset int, limit int) (hosts []rp.Host, err error) {
	db := rw.DB()
	// Check Host Detail
	if filter.HostID != "" {
		err = db.Where("id = ?", filter.HostID).Find(&hosts).Error
		if err != nil {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_HOST_NOT_FOUND, "query host %s error, %v", filter.HostID, err)
		}
		return
	}

	if filter.Arch != "" {
		db = db.Where("arch = ?", filter.Arch)
	}

	if filter.Status != "" {
		if filter.Status != string(constants.HostDeleted) {
			db = db.Where("status = ?", filter.Status)
		} else {
			db = db.Unscoped().Where("status = ?", filter.Status)
		}
	}
	if filter.Stat != "" {
		db = db.Where("stat = ?", filter.Stat)
	}
	if filter.Purpose != "" {
		label, err := structs.GetTraitByName(filter.Purpose)
		if err != nil {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_LABEL_NAEM, "query host use a invalid purpose name %s, %v", filter.Purpose, err)
		}
		db = db.Where("traits & ? = ?", label, label)
	}
	err = db.Offset(offset).Limit(limit).Find(&hosts).Error
	return
}

func (rw *GormResourceReadWrite) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	tx := rw.DB().Begin()
	for _, hostId := range hostIds {
		result := tx.Model(&rp.Host{}).Where("id = ?", hostId).Update("status", status)
		if result.Error != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_UPDATE_HOST_STATUS_FAIL, "update host [%s] status to %s fail", hostId, status)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_UPDATE_HOST_STATUS_FAIL, "update host [%s] status to %s not affected", hostId, status)
		}
	}
	tx.Commit()
	return nil
}
func (rw *GormResourceReadWrite) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	tx := rw.DB().Begin()
	for _, hostId := range hostIds {
		result := tx.Model(&rp.Host{}).Where("id = ?", hostId).Update("reserved", reserved)
		if result.Error != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESERVE_HOST_FAIL, "update host [%s] reserved status to %v fail", hostId, reserved)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESERVE_HOST_FAIL, "update host [%s] reserved status to %v not affected", hostId, reserved)
		}
	}
	tx.Commit()
	return nil
}
func (rw *GormResourceReadWrite) GetHierarchy(ctx context.Context, filter structs.HostFilter, level int32, depth int32) (root *structs.HierarchyTreeNode, err error) {
	return nil, nil
}
func (rw *GormResourceReadWrite) GetStocks(ctx context.Context, location structs.Location, hostFilter structs.HostFilter, diskFilter structs.DiskFilter) (stocks *structs.Stocks, err error) {
	return nil, nil
}
