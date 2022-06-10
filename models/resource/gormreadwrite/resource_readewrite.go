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

package gormreadwrite

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/common/structs"
	dbCommon "github.com/pingcap-inc/tiunimanager/models/common"
	resource_models "github.com/pingcap-inc/tiunimanager/models/resource"
	rp "github.com/pingcap-inc/tiunimanager/models/resource/resourcepool"
	"gorm.io/gorm"
)

type GormResourceReadWrite struct {
	dbCommon.GormDB
}

func NewGormResourceReadWrite(db *gorm.DB) resource_models.ReaderWriter {
	m := &GormResourceReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

/*
func (rw *GormResourceReadWrite) addTable(ctx context.Context, tableModel interface{}) (newTable bool, err error) {
	if !rw.DB(ctx).Migrator().HasTable(tableModel) {
		err := rw.DB(ctx).Migrator().CreateTable(tableModel)
		if nil != err {
			return true, errors.NewErrorf(errors.TIEM_RESOURCE_ADD_TABLE_ERROR, "crete table %v failed, error: %v", tableModel, err)
		}
		return true, nil
	} else {
		return false, nil
	}
}
*/
func (rw *GormResourceReadWrite) Create(ctx context.Context, hosts []rp.Host) (hostIds []string, err error) {
	tx := rw.DB(ctx).Begin()
	for _, host := range hosts {
		err = tx.Create(&host).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_CREATE_HOST_ERROR, "create %s(%s) error, %v", host.HostName, host.IP, err)
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
		//if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&host, "ID = ?", hostId).Error; err != nil {
		if err = tx.First(&host, "ID = ?", hostId).Error; err != nil {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_SQL_ERROR, "lock host %s(%s) error, %v", hostId, host.IP, err)
		}
		err = tx.Delete(&host).Error
		if err != nil {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_RESOURCE_DELETE_HOST_ERROR, "delete host %s(%s) error, %v", hostId, host.IP, err)
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
			db = db.Where("hosts.status = ?", filter.Status)
		} else {
			db = db.Unscoped().Where("hosts.status = ?", filter.Status)
		}
	}
	if filter.Stat != "" {
		db = db.Where("stat = ?", filter.Stat)
	}

	if filter.HostName != "" {
		db = db.Where("host_name = ?", filter.HostName)
	}

	var labels int64
	if filter.ClusterType != "" {
		label, err := structs.GetTraitByName(filter.ClusterType)
		if err != nil {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, "query host use a invalid clusterType name %s, %v", filter.ClusterType, err)
		}
		labels |= label
	}
	if filter.HostDiskType != "" {
		label, err := structs.GetTraitByName(filter.HostDiskType)
		if err != nil {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, "query host use a invalid diskType name %s, %v", filter.HostDiskType, err)
		}
		labels |= label
	}

	if filter.Purpose != "" {
		label, err := structs.GetTraitByName(filter.Purpose)
		if err != nil {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, "query host use a invalid purpose name %s, %v", filter.Purpose, err)
		}
		labels |= label
	}

	if labels != 0 {
		db = db.Where("traits & ? = ?", labels, labels)
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

	if location.HostIp != "" {
		db = db.Where("hosts.ip = ?", location.HostIp)
	}

	// Region field should be required for follower filter
	if location.Region == "" {
		return db, nil
	}
	regionCode = location.Region
	db = db.Where("hosts.region = ?", regionCode)

	//  Zone field should be required for follower filter
	if location.Zone == "" {
		return db, nil
	}
	zoneCode = structs.GenDomainCodeByName(regionCode, location.Zone)
	db = db.Where("hosts.az = ?", zoneCode)

	// Rack field is optional by now
	if location.Rack != "" {
		rackCode = structs.GenDomainCodeByName(zoneCode, location.Rack)
		db = db.Where("hosts.rack = ?", rackCode)
	}

	return db, nil
}

func (rw *GormResourceReadWrite) Query(ctx context.Context, location *structs.Location, filter *structs.HostFilter, offset int, limit int) (hosts []rp.Host, total int64, err error) {
	hosts = make([]rp.Host, 0)
	db := rw.DB(ctx).Model(&rp.Host{})
	// Check Host Detail
	if filter.HostID != "" {
		err = db.Where("id = ?", filter.HostID).Count(&total).Find(&hosts).Error
		if err != nil {
			return nil, 0, errors.NewErrorf(errors.TIEM_RESOURCE_HOST_NOT_FOUND, "query host %s error, %v", filter.HostID, err)
		}
		return
	}
	db, err = rw.locationFiltered(db, location)
	if err != nil {
		return nil, 0, err
	}
	db, err = rw.hostFiltered(db, filter)
	if err != nil {
		return nil, 0, err
	}
	err = db.Count(&total).Offset(offset).Limit(limit).Find(&hosts).Error
	return
}

func (rw *GormResourceReadWrite) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	tx := rw.DB(ctx).Begin()
	for _, hostId := range hostIds {
		result := tx.Model(&rp.Host{}).Where("id = ?", hostId).Update("status", status)
		if result.Error != nil {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_UPDATE_HOST_STATUS_FAIL, "update host %s status to %s fail, %v", hostId, status, result.Error)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_UPDATE_HOST_STATUS_FAIL, "update host %s status to %s not affected", hostId, status)
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
			return errors.NewErrorf(errors.TIEM_RESERVE_HOST_FAIL, "update host %s reserved status to %v fail, %v", hostId, reserved, result.Error)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_RESERVE_HOST_FAIL, "update host %s reserved status to %v not affected", hostId, reserved)
		}
	}
	tx.Commit()
	return nil
}

func (rw *GormResourceReadWrite) UpdateHostInfo(ctx context.Context, host rp.Host) (err error) {
	if host.ID == "" {
		return errors.NewError(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update host info but no host id specified")
	}
	var originHost rp.Host
	originHost.ID = host.ID
	tx := rw.DB(ctx).Begin()
	if err = tx.First(&originHost, "ID = ?", originHost.ID).Error; err != nil {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_SQL_ERROR, "get origin host info before update (%s) error, %v", originHost.ID, err)
	}
	updates, err := originHost.PrepareForUpdate(&host)
	if err != nil {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "prepare for update host %s %s failed, %v", originHost.HostName, originHost.IP, err)
	}
	result := tx.Model(&originHost).Omit("Reserved", "Status", "Stat").Updates(updates)
	if result.Error != nil {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update host %s %s info failed, %v", originHost.HostName, originHost.IP, result.Error)
	}
	if result.RowsAffected != 1 {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update host %s %s not affected one row, rows affected %d", originHost.HostName, originHost.IP, result.RowsAffected)
	}
	err = tx.Commit().Error
	return
}

func (rw *GormResourceReadWrite) GetHostItems(ctx context.Context, filter *structs.HostFilter, level int32, depth int32) (items []resource_models.HostItem, err error) {
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
		err = errors.NewError(errors.TIEM_PARAMETER_INVALID, errMsg)
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "get hierarchy failed, %v", err)
	}
	tx.Commit()
	return
}
func (rw *GormResourceReadWrite) GetHostStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error) {
	tx := rw.DB(ctx).Begin()
	db := tx.Model(&rp.Host{}).Select(
		"hosts.az as zone, hosts.free_cpu_cores as free_cpu_cores, hosts.free_memory as free_memory, count(disks.id) as free_disk_count, sum(disks.capacity) as free_disk_capacity").Joins(
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
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "get stocks failed, %v", err)
	}
	tx.Commit()
	return
}

func (rw *GormResourceReadWrite) CreateDisks(ctx context.Context, hostId string, disks []rp.Disk) (diskIds []string, err error) {
	if hostId == "" {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_CREATE_DISK_ERROR, "batch create %d disks error, no hostId specified", len(disks))
	}

	tx := rw.DB(ctx).Begin()
	var total int64
	var host rp.Host
	host.ID = hostId
	// get host info
	err = tx.Model(&host).First(&host).Count(&total).Error
	if err != nil || total == 0 {
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_HOST_NOT_FOUND, "query host(%d) %s failed before creating disks, %v", total, hostId, err)
	}
	for i := range disks {
		if err = disks[i].ValidateDisk(hostId, host.DiskType); err != nil {
			tx.Rollback()
			return nil, err
		}
		if disks[i].HostID == "" {
			disks[i].HostID = hostId
		}
		if disks[i].Type == "" {
			disks[i].Type = host.DiskType
		}
		err = tx.Create(&disks[i]).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_CREATE_DISK_ERROR, "create disk %s for host %s failed, %v", disks[i].Name, hostId, err)
		}
		diskIds = append(diskIds, disks[i].ID)
	}
	err = tx.Commit().Error
	return
}

func (rw *GormResourceReadWrite) DeleteDisks(ctx context.Context, diskIds []string) (err error) {
	tx := rw.DB(ctx).Begin()
	for _, diskId := range diskIds {
		err = tx.Delete(&rp.Disk{ID: diskId}).Error
		if err != nil {
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_RESOURCE_DELETE_DISK_ERROR, "delete disk %s error, %v", diskId, err)
		}
	}
	err = tx.Commit().Error
	return
}

func (rw *GormResourceReadWrite) UpdateDisk(ctx context.Context, disk rp.Disk) (err error) {
	if disk.ID == "" {
		return errors.NewError(errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update disk failed, no disk id specified")
	}
	if disk.Status != "" && disk.Status != string(constants.DiskError) {
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "disk status %s is not supported for update, only support set to %s by now",
			disk.Status, string(constants.DiskError))
	}
	var originDisk rp.Disk
	originDisk.ID = disk.ID
	tx := rw.DB(ctx).Begin()
	if err = tx.First(&originDisk, "ID = ?", originDisk.ID).Error; err != nil {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_SQL_ERROR, "get origin disk info before update (%s) error, %v", originDisk.ID, err)
	}
	patch := originDisk
	if err = patch.PrepareForUpdate(&disk); err != nil {
		tx.Rollback()
		return err
	}
	result := tx.Model(&originDisk).Updates(patch)
	if result.Error != nil {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update disk %s %s info failed, %v", originDisk.Name, originDisk.Path, result.Error)
	}
	if result.RowsAffected != 1 {
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update host %s %s not affected one row, rows affected %d", originDisk.Name, originDisk.Path, result.RowsAffected)
	}
	err = tx.Commit().Error
	return
}
