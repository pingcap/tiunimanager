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
	"github.com/pingcap-inc/tiem/library/util/bitmap"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/structs"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	mm "github.com/pingcap-inc/tiem/models/resource/management"
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
	_, err = rw.addTable(ctx, new(mm.UsedCompute))
	if err != nil {
		log.Errorf("create table UsedCompute failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(mm.UsedPort))
	if err != nil {
		log.Errorf("create table UsedPort failed, error: %v", err)
		return err
	}
	_, err = rw.addTable(ctx, new(mm.UsedDisk))
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

func (rw *GormResourceReadWrite) AllocResources(ctx context.Context, batchReq *resource_structs.BatchAllocRequest) (results *resource_structs.BatchAllocResponse, err error) {
	results = new(resource_structs.BatchAllocResponse)
	tx := rw.DB(ctx).Begin()
	for i, request := range batchReq.BatchRequests {
		var result *resource_structs.AllocRsp
		result, err = rw.allocForSingleRequest(ctx, tx, &request)
		if err != nil {
			tx.Rollback()
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NOT_ALL_SUCCEED, "alloc resources in batch failed on request %d, %v", i, err)
		}
		results.BatchResults = append(results.BatchResults, result)
	}
	tx.Commit()
	return
}
func (rw *GormResourceReadWrite) RecycleResources(ctx context.Context, request *resource_structs.RecycleRequest) (err error) {
	return nil
}

func (rw *GormResourceReadWrite) allocForSingleRequest(ctx context.Context, tx *gorm.DB, req *resource_structs.AllocReq) (results *resource_structs.AllocRsp, err error) {
	log := framework.LogWithContext(ctx)
	var choosedHosts []string
	results = new(resource_structs.AllocRsp)
	for i, require := range req.Requires {
		switch resource_structs.AllocStrategy(require.Strategy) {
		case resource_structs.RandomRack:
			res, err := rw.allocResourceWithRR(tx, &req.Applicant, i, &require, choosedHosts)
			if err != nil {
				log.Errorf("alloc resources in random rack strategy for %dth requirement %v failed, %v", i, require, err)
				return nil, err
			}
			for _, result := range res {
				choosedHosts = append(choosedHosts, result.HostIp)
			}
			results.Results = append(results.Results, res...)
		case resource_structs.DiffRackBestEffort:
		case resource_structs.UserSpecifyRack:
		case resource_structs.UserSpecifyHost:
			res, err := rw.allocResourceInHost(tx, &req.Applicant, i, &require)
			if err != nil {
				log.Errorf("alloc resources in specify host strategy for %dth requirement %v failed, %v", i, require, err)
				return nil, err
			}
			results.Results = append(results.Results, res...)
		case resource_structs.ClusterPorts:
			res, err := rw.allocPortsInRegion(tx, &req.Applicant, i, &require)
			if err != nil {
				log.Errorf("alloc resources in cluster port strategy for %dth requirement %v failed, %v", i, require, err)
				return nil, err
			}
			results.Results = append(results.Results, res...)
		default:
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_STRATEGY, "invalid alloc strategy %d", require.Strategy)
		}
	}
	return
}

type Resource struct {
	HostId   string
	HostName string
	Ip       string
	UserName string
	Passwd   string
	CpuCores int
	Memory   int
	DiskId   string
	DiskName string
	Path     string
	Capacity int
	portRes  []*resource_structs.PortResource
}

func (rw *GormResourceReadWrite) allocResourceWithRR(tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement, choosedHosts []string) (results []resource_structs.Compute, err error) {
	log := framework.Log()
	regionName := require.Location.Region
	zoneName := require.Location.Zone
	zoneCode := structs.GenDomainCodeByName(regionName, zoneName)
	var excludedHosts []string
	if require.HostExcluded.Hosts != nil {
		excludedHosts = require.HostExcluded.Hosts
	}
	// excluded choosed hosts in one request
	excludedHosts = append(excludedHosts, choosedHosts...)
	hostArch := require.HostFilter.Arch
	hostTraits := require.HostFilter.HostTraits
	exclusive := require.Require.Exclusive
	reqCores := require.Require.ComputeReq.CpuCores
	reqMem := require.Require.ComputeReq.Memory
	capacity := require.Require.DiskReq.Capacity
	needDisk := require.Require.DiskReq.NeedDisk
	log.Infof("Alloc Resource With RR: zoneCode: %s, excludedHosts: %v, arch: %s, traits: %d, excludsive: %v, cpuCores: %d, memory: %d, needDisk: %v, diskCapacity: %d\n",
		zoneCode, excludedHosts, hostArch, hostTraits, exclusive, reqCores, reqMem, needDisk, capacity)
	// 1. Choose Host/Disk List
	var resources []*Resource
	if needDisk {
		var count int64
		db := tx.Order("hosts.free_cpu_cores desc").Order("hosts.free_memory desc").Limit(int(require.Count)).Model(&rp.Disk{}).Select(
			"disks.host_id, hosts.host_name, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", reqCores, reqMem).Joins(
			"left join hosts on disks.host_id = hosts.id").Where("hosts.reserved = 0")
		if excludedHosts == nil {
			db.Count(&count)
		} else {
			db.Not(map[string]interface{}{"hosts.ip": excludedHosts}).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED, "expect disk count %d but only %d after excluded host list", require.Count, count)
		}

		db = db.Where("disks.status = ? and disks.capacity >= ?", constants.DiskAvailable, capacity).Count(&count)
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "expect disk count %d but only %d after disk filter", require.Count, count)
		}

		if !exclusive {
			err = db.Where("hosts.az = ? and hosts.arch = ? and hosts.traits & ? = ? and hosts.status = ? and (hosts.stat = ? or hosts.stat = ?) and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("hosts.az = ? and hosts.arch = ? and hosts.traits & ? = ? and hosts.status = ? and hosts.stat = ? and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, constants.HostOnline, constants.HostLoadLoadLess, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		}
	} else {
		var count int64
		db := tx.Order("hosts.free_cpu_cores desc").Order("hosts.free_memory desc").Limit(int(require.Count)).Model(&rp.Host{}).Select(
			"id as host_id, host_name, ip, user_name, passwd, ? as cpu_cores, ? as memory", reqCores, reqMem).Where("reserved = 0")
		if excludedHosts == nil {
			db.Count(&count)
		} else {
			db.Not(map[string]interface{}{"hosts.ip": excludedHosts}).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "expect host count %d but only %d after excluded host list", require.Count, count)
		}
		if !exclusive {
			err = db.Where("az = ? and arch = ? and traits & ? = ? and status = ? and (stat = ? or stat = ?) and free_cpu_cores >= ? and free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed, reqCores, reqMem).Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("az = ? and arch = ? and traits & ? = ? and status = ? and stat = ? and free_cpu_cores >= ? and free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, constants.HostOnline, constants.HostLoadLoadLess, reqCores, reqMem).Scan(&resources).Error
		}
	}
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
	}

	if len(resources) < int(require.Count) {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "hosts in %s, %s is not enough for allocation(%d|%d), arch: %s, traits: %d", regionName, zoneName, len(resources), require.Count, hostArch, hostTraits)
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
			}
			resource.portRes = append(resource.portRes, res)
		}
	}

	// 3. Mark Resources in used
	err = rw.markResourcesForUsed(tx, applicant, resources, exclusive)
	if err != nil {
		return nil, err
	}

	// 4. make Results and Complete one Requirement
	for _, resource := range resources {
		result := resource_structs.Compute{
			Reqseq:   int32(seq),
			HostId:   resource.HostId,
			HostName: resource.HostName,
			HostIp:   resource.Ip,
			UserName: resource.UserName,
			Passwd:   resource.Passwd,
		}
		result.ComputeRes.CpuCores = int32(resource.CpuCores)
		result.ComputeRes.Memory = int32(resource.Memory)
		result.DiskRes.DiskId = resource.DiskId
		result.DiskRes.DiskName = resource.DiskName
		result.DiskRes.Path = resource.Path
		result.DiskRes.Capacity = int32(resource.Capacity)
		for _, portRes := range resource.portRes {
			result.PortRes = append(result.PortRes, *portRes)
		}

		results = append(results, result)
	}
	return
}

func (rw *GormResourceReadWrite) allocResourceInHost(tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement) (results []resource_structs.Compute, err error) {
	hostIp := require.Location.HostIp
	if require.Count != 1 {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "request host count should be 1 for UserSpecifyHost allocation(%d)", require.Count)
	}
	reqCores := require.Require.ComputeReq.CpuCores
	reqMem := require.Require.ComputeReq.Memory
	exclusive := require.Require.Exclusive
	isTakeOver := applicant.TakeoverOperation

	needDisk := require.Require.DiskReq.NeedDisk
	diskSpecify := require.Require.DiskReq.DiskSpecify
	diskType := constants.DiskType(require.Require.DiskReq.DiskType)
	capacity := require.Require.DiskReq.Capacity

	var resources []*Resource
	var count int64

	if needDisk {
		db := tx.Order("disks.capacity").Limit(int(require.Count)).Model(&rp.Disk{}).Select(
			"disks.host_id, hosts.host_name, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", reqCores, reqMem).Joins(
			"left join hosts on disks.host_id = hosts.id").Where("hosts.ip = ?", hostIp).Count(&count)
		// No Limit in Reserved == false in this strategy for a takeover operation
		if !isTakeOver {
			db = db.Where("hosts.reserved = 0").Count(&count)
		}

		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "disk is not enough(%d|%d) in host (%s), takeover operation (%v)", count, require.Count, hostIp, isTakeOver)
		}
		db = db.Where("hosts.free_cpu_cores >= ? and hosts.free_memory >= ?", reqCores, reqMem).Count(&count)
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "cpucores or memory in host (%s) is not enough", hostIp)
		}
		if !exclusive {
			db = db.Where("hosts.status = ? and (hosts.stat = ? or hosts.stat = ?)", constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed).Count(&count)
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			db = db.Where("hosts.status = ? and hosts.stat = ?", constants.HostOnline, constants.HostLoadLoadLess).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) status/stat is not expected for exlusive(%v) condition", hostIp, exclusive)
		}
		if diskSpecify == "" {
			err = db.Where("disks.type = ? and disks.status = ? and disks.capacity >= ?", diskType, constants.DiskAvailable, capacity).Scan(&resources).Error
		} else {
			err = db.Where("disks.id = ? and disks.status = ?", diskSpecify, constants.DiskAvailable).Scan(&resources).Error
		}
		if err != nil {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
		}
		if len(resources) < int(require.Count) {
			if diskSpecify == "" {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "no available disk with type(%s) and capacity(%d) in host(%s) after disk filter", diskType, capacity, hostIp)
			} else {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "disk (%s) not existed or it is not available in host(%s)", diskSpecify, hostIp)
			}
		}
	} else {
		db := tx.Model(&rp.Host{}).Select("id as host_id, host_name, ip, user_name, passwd, ? as cpu_cores, ? as memory", reqCores, reqMem).Where("ip = ?", hostIp).Count(&count)
		// No Limit in Reserved == false in this strategy for a takeover operation
		if !isTakeOver {
			db = db.Where("hosts.reserved = 0").Count(&count)
		}
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) is not existed or reserved, takeover operation(%v)", hostIp, isTakeOver)
		}
		db = db.Where("free_cpu_cores >= ? and free_memory >= ?", reqCores, reqMem).Count(&count)
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "cpucores or memory in host (%s) is not enough", hostIp)
		}
		if !exclusive {
			err = db.Where("status = ? and (stat = ? or stat = ?)", constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed).Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("status = ? and stat = ?", constants.HostOnline, constants.HostLoadLoadLess).Scan(&resources).Error
		}
		if err != nil {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
		}
		if len(resources) < int(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) status/stat is not expected for exlusive(%v) condition", hostIp, exclusive)
		}
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
			}
			resource.portRes = append(resource.portRes, res)
		}
	}

	// 3. Mark Resources in used
	err = rw.markResourcesForUsed(tx, applicant, resources, exclusive)
	if err != nil {
		return nil, err
	}

	// 4. make Results and Complete one Requirement
	for _, resource := range resources {
		result := resource_structs.Compute{
			Reqseq:   int32(seq),
			HostId:   resource.HostId,
			HostName: resource.HostName,
			HostIp:   resource.Ip,
			UserName: resource.UserName,
			Passwd:   resource.Passwd,
		}
		result.ComputeRes.CpuCores = int32(resource.CpuCores)
		result.ComputeRes.Memory = int32(resource.Memory)
		result.DiskRes.DiskId = resource.DiskId
		result.DiskRes.DiskName = resource.DiskName
		result.DiskRes.Path = resource.Path
		result.DiskRes.Capacity = int32(resource.Capacity)
		for _, portRes := range resource.portRes {
			result.PortRes = append(result.PortRes, *portRes)
		}

		results = append(results, result)
	}
	return
}

func (rw *GormResourceReadWrite) allocPortsInRegion(tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement) (results []resource_structs.Compute, err error) {
	regionCode := require.Location.Region
	hostArch := require.HostFilter.Arch
	if regionCode == "" {
		return nil, framework.NewTiEMError(common.TIEM_RESOURCE_INVALID_LOCATION, "no valid region")
	}
	if hostArch == "" {
		return nil, framework.NewTiEMError(common.TIEM_RESOURCE_INVALID_ARCH, "no valid arch")
	}
	if len(require.Require.PortReq) != 1 {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "require portReq len should be 1 for RegionUniformPorts allocation(%d)", len(require.Require.PortReq))
	}
	portReq := require.Require.PortReq[0]
	var regionHosts []string
	err = tx.Model(&rp.Host{}).Select("id").Where("region = ? and arch = ?", regionCode, hostArch).Where("reserved = 0").Scan(&regionHosts).Error
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select %s hosts in region %s failed, %v", hostArch, regionCode, err)
	}
	if len(regionHosts) == 0 {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "no %s host in region %s", hostArch, regionCode)
	}

	var usedPorts []int32
	err = tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id in ?", regionHosts).Group("port").Having("port >= ? and port < ?", portReq.Start, portReq.End).Scan(&usedPorts).Error
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select used port range %d - %d in region %s failed, %v", portReq.Start, portReq.End, regionCode, err)
	}

	res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "Region %s has no enough %d ports on range [%d, %d]", regionCode, portReq.PortCnt, portReq.Start, portReq.End)
	}

	err = rw.markPortsInRegion(tx, applicant, regionHosts, res)
	if err != nil {
		return nil, err
	}

	var result resource_structs.Compute
	result.Reqseq = int32(seq)
	result.PortRes = append(result.PortRes, *res)

	results = append(results, result)
	return
}

func (rw *GormResourceReadWrite) getPortsInRange(usedPorts []int32, start int32, end int32, count int) (*resource_structs.PortResource, error) {
	bitlen := int(end - start)
	bm := bitmap.NewConcurrentBitmap(bitlen)
	for _, used := range usedPorts {
		if used < start {
			continue
		} else if used > end {
			break
		} else {
			bm.Set(int(used - start))
		}
	}
	result := &resource_structs.PortResource{
		Start: start,
		End:   end,
	}
	for i := 0; i < count; i++ {
		found := false
		for j := 0; j < bitlen; j++ {
			if !bm.UnsafeIsSet(j) {
				bm.Set(j)
				result.Ports = append(result.Ports, start+int32(j))
				found = true
				break
			}
		}
		if !found {
			return nil, framework.NewTiEMError(common.TIEM_RESOURCE_NO_ENOUGH_PORT, common.TIEM_RESOURCE_NO_ENOUGH_PORT.Explain())
		}
	}
	return result, nil
}

func (rw *GormResourceReadWrite) markResourcesForUsed(tx *gorm.DB, applicant *resource_structs.Applicant, resources []*Resource, exclusive bool) (err error) {
	for _, resource := range resources {
		if resource.DiskId != "" {
			var disk rp.Disk
			tx.First(&disk, "ID = ?", resource.DiskId).First(&disk)
			if constants.DiskStatus(disk.Status).IsAvailable() {
				err = tx.Model(&disk).Update("Status", string(constants.DiskExhaust)).Error
				if err != nil {
					return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update disk(%s) status err, %v", resource.DiskId, err)
				}
			} else {
				return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "disk %s status not expected(%d)", resource.DiskId, disk.Status)
			}
			usedDisk := mm.UsedDisk{
				DiskId:   resource.DiskId,
				HostId:   resource.HostId,
				Capacity: int32(resource.Capacity),
			}
			usedDisk.HolderId = applicant.HolderId
			usedDisk.RequestId = applicant.RequestId
			err = tx.Create(&usedDisk).Error
			if err != nil {
				return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "insert disk(%s) to used_disks table failed: %v", resource.DiskId, err)
			}
		}

		var host rp.Host
		tx.First(&host, "ID = ?", resource.HostId)
		host.FreeCpuCores -= int32(resource.CpuCores)
		host.FreeMemory -= int32(resource.Memory)
		if exclusive {
			host.Stat = string(constants.HostLoadExclusive)
		} else {
			host.Stat = string(constants.HostLoadInUsed)
		}
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", resource.HostId).Updates(rp.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update host(%s) stat err, %v", resource.HostId, err)
		}
		usedCompute := mm.UsedCompute{
			HostId:   resource.HostId,
			CpuCores: int32(resource.CpuCores),
			Memory:   int32(resource.Memory),
		}
		usedCompute.Holder.HolderId = applicant.HolderId
		usedCompute.RequestId = applicant.RequestId
		err = tx.Create(&usedCompute).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "insert host(%s) to used_computes table failed: %v", resource.HostId, err)
		}

		for _, ports := range resource.portRes {
			for _, port := range ports.Ports {
				usedPort := mm.UsedPort{
					HostId: resource.HostId,
					Port:   port,
				}
				usedPort.HolderId = applicant.HolderId
				usedPort.RequestId = applicant.RequestId
				err = tx.Create(&usedPort).Error
				if err != nil {
					return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "insert host(%s) for port(%d) table failed: %v", resource.HostId, port, err)
				}
			}
		}
	}
	return nil
}

func (rw *GormResourceReadWrite) markPortsInRegion(tx *gorm.DB, applicant *resource_structs.Applicant, regionHosts []string, portResources *resource_structs.PortResource) (err error) {
	for _, host := range regionHosts {
		for _, port := range portResources.Ports {
			usedPort := mm.UsedPort{
				HostId: host,
				Port:   port,
			}
			usedPort.HolderId = applicant.HolderId
			usedPort.RequestId = applicant.RequestId
			err = tx.Create(&usedPort).Error
			if err != nil {
				return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "insert host(%s) for port(%d) table failed: %v", host, port, err)
			}
		}
	}
	return
}
