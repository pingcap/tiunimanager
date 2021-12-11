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
	"errors"
	"fmt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/resource/management"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"gorm.io/gorm"
)

type GormResourceReadWrite struct {
	dbCommon.GormDB
}

func NewGormResourceReadWrite(db *gorm.DB) ReaderWriter {
	m := &GormResourceReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (rw *GormResourceReadWrite) addTable(ctx context.Context, tableModel interface{}) (newTable bool, err error) {
	if !rw.DB(ctx).Migrator().HasTable(tableModel) {
		err := rw.DB(ctx).Migrator().CreateTable(tableModel)
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
	newTable, err := rw.addTable(ctx, new(structs.Label))
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
	for _, v := range structs.DefaultLabelTypes {
		err = rw.DB(ctx).Create(&v).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INIT_LABELS_ERROR, "init default label table failed, error: %v", err)
		}
	}
	return nil
}

func (rw *GormResourceReadWrite) Create(ctx context.Context, hosts []rp.Host) (hostIds []string, err error) {
	tx := rw.DB(ctx).Begin()
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
	tx := rw.DB(ctx).Begin()
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

func (rw *GormResourceReadWrite) hostFiltered(db *gorm.DB, filter *structs.HostFilter) (*gorm.DB, error) {
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
	return db, nil
}

func (rw *GormResourceReadWrite) diskFiltered(db *gorm.DB, filter *structs.DiskFilter) (*gorm.DB, error) {
	if filter.DiskStatus != "" {
		db = db.Where("disks.status = ?", filter.DiskStatus)
	}
	if filter.DiskType != "" {
		db = db.Where("disks.type = ?", filter.DiskType)
	}
	if filter.Capacity >= 0 {
		db = db.Where("disks.capacity >= ?", filter.Capacity)
	}
	return db, nil
}

func (rw *GormResourceReadWrite) locationFiltered(db *gorm.DB, location *structs.Location) (*gorm.DB, error) {
	var regionCode, zoneCode, rackCode string

	// Region field should be required for follower filter
	if location.Region == "" {
		return db, nil
	}
	regionCode = location.Region
	db = db.Where("hosts.region = ?", regionCode)

	//  Zone field should be required for follower filter
	if location.Zone != "" {
		return db, nil
	}
	zoneCode = structs.GenDomainCodeByName(regionCode, location.Zone)
	db = db.Where("hosts.az = ?", zoneCode)

	// Rack field is optional by now
	if location.Rack != "" {
		rackCode = structs.GenDomainCodeByName(zoneCode, location.Rack)
		db = db.Where("hosts.rack = ?", rackCode)
	}

	if location.HostIp != "" {
		db = db.Where("hosts.ip = ?", location.HostIp)
	}

	return db, nil
}

func (rw *GormResourceReadWrite) Query(ctx context.Context, filter *structs.HostFilter, offset int, limit int) (hosts []rp.Host, err error) {
	db := rw.DB(ctx)
	// Check Host Detail
	if filter.HostID != "" {
		err = db.Where("id = ?", filter.HostID).Find(&hosts).Error
		if err != nil {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_HOST_NOT_FOUND, "query host %s error, %v", filter.HostID, err)
		}
		return
	}
	db, err = rw.hostFiltered(db, filter)
	if err != nil {
		return nil, err
	}
	err = db.Offset(offset).Limit(limit).Find(&hosts).Error
	return
}

func (rw *GormResourceReadWrite) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	tx := rw.DB(ctx).Begin()
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
	tx := rw.DB(ctx).Begin()
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

func (rw *GormResourceReadWrite) GetHostItems(ctx context.Context, filter *structs.HostFilter, level int32, depth int32) (items []HostItem, err error) {
	leafLevel := level + depth
	tx := rw.DB(ctx).Begin()
	db := tx.Model(&rp.Host{}).Select("region, az, rack, ip, host_name as name")
	db, err = rw.hostFiltered(db, filter)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	switch constants.HierarchyTreeNodeLevel(leafLevel) {
	case constants.REGION:
		err = db.Group("region").Scan(&items).Error
	case constants.ZONE:
		err = db.Group("region").Group("az").Scan(&items).Error
	case constants.RACK:
		err = db.Group("region").Group("az").Group("rack").Scan(&items).Error
	case constants.HOST:
		// Build whole tree with hosts
		err = db.Order("region").Order("az").Order("rack").Order("ip").Scan(&items).Error
	default:
		errMsg := fmt.Sprintf("invalid leaf level %d, level = %d, depth = %d", leafLevel, level, depth)
		err = errors.New(errMsg)
	}
	if err != nil {
		tx.Rollback()
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get hierarchy failed, %v", err)
	}
	tx.Commit()
	return
}
func (rw *GormResourceReadWrite) GetHostStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error) {
	tx := rw.DB(ctx).Begin()
	db := tx.Model(&rp.Host{}).Select(
		"hosts.free_cpu_cores as free_cpu_cores, hosts.free_memory as free_memory, count(disks.id) as free_disk_count, sum(disks.capacity) as free_disk_capacity").Joins(
		"left join disks on disks.host_id = hosts.id")
	db, err = rw.locationFiltered(db, location)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	db, err = rw.hostFiltered(db, hostFilter)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	db, err = rw.diskFiltered(db, diskFilter)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	db = db.Group("hosts.id")

	err = db.Scan(&stocks).Error
	if err != nil {
		tx.Rollback()
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get stocks failed, %v", err)
	}
	tx.Commit()
	return
}
