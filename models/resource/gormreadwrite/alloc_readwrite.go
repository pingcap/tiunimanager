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

	"github.com/pingcap-inc/tiem/util/bitmap"
	crypto "github.com/pingcap-inc/tiem/util/encrypt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	resource_structs2 "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management2/structs"
	mm "github.com/pingcap-inc/tiem/models/resource/management"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"gorm.io/gorm"
)

func (rw *GormResourceReadWrite) AllocResources2(ctx context.Context, allocReq *resource_structs2.AllocRequest) (results *resource_structs2.AllocResponse, err error) {
	log := framework.LogWithContext(ctx)
	results = new(resource_structs2.AllocResponse)
	results.Applicant.HolderId = allocReq.Applicant.HolderId
	results.Applicant.RequestId = allocReq.Applicant.RequestId
	log.Infof("Applicant %v Alloc Resource by Mode %s in location %s, excludedHosts: %v, arch: %s, vendor: %s, traits: %d, poolType: %s",
		allocReq.Applicant, allocReq.Mode, allocReq.Location, allocReq.Arch, allocReq.Vendor, allocReq.ProductType, allocReq.PoolType)
	hosts := make([]rp.Host, 0)
	var total int64
	tx := rw.DB(ctx).Begin()
	db := tx.Model(&rp.Host{}).
		Where("region = ?", allocReq.Location.Region)
	if allocReq.Mode == "FullPath" {
		zoneCode := structs.GenDomainCodeByName(allocReq.Location.Region, allocReq.Location.Zone)
		rackCode := structs.GenDomainCodeByName(zoneCode, allocReq.Location.Rack)
		db = db.Where("az = ?", zoneCode).Where("rack = ?", rackCode)
	}
	err = db.Where("arch = ?", allocReq.Arch).
		Where("vendor = ?", allocReq.Vendor).
		Where("pool = ?", allocReq.PoolType).
		Where("traits & ? = ?", allocReq.ProductType, allocReq.ProductType).
		Order("az").Order("rack").Order("cpu_cores").Order("memory").
		Count(&total).Find(&hosts).Error
	if err != nil {
		tx.Rollback()
		log.Errorf("Get Resources in location %s (arch %s, vendor %s, traits %d)failed, %s", allocReq.Location, allocReq.Arch, allocReq.Vendor, allocReq.ProductType, err)
		return nil, err
	}
	if total == 0 {
		tx.Rollback()
		log.Errorf("No Resources in location %s (arch %s, vendor %s, traits %d) , %s", allocReq.Location, allocReq.Arch, allocReq.Vendor, allocReq.ProductType, err)
		return nil, err
	}

	rack2hosts, zone2racks, region2zones, err := rw.buildHierarchy(hosts)
	if err != nil {
		tx.Rollback()
		log.Errorf("build hierarchy from hosts %v failed, %v", hosts, err)
		return nil, err
	}

	for i, request := range allocReq.ComponentRequests {
		var result *resource_structs2.ComponentResult
		result, err = rw.allocForComponentRequest(ctx, tx, &request, rack2hosts, zone2racks, region2zones)
		if err != nil {
			tx.Rollback()
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_ALLOCATE_ERROR, "alloc resources failed on %dth component request with %d requires, request: %v, error: %v", i+1, len(request.Requires), request, err)
		}
		results.ComponentResults = append(results.ComponentResults, result)
	}

	tx.Commit()
	return
}

func (rw *GormResourceReadWrite) allocForComponentRequest(ctx context.Context, tx *gorm.DB, req *resource_structs2.ComponentRequest,
	rack2hosts map[string][]*rp.Host, zone2racks map[string][]string, region2zones map[string][]string) (results *resource_structs2.ComponentResult, err error) {
	log := framework.LogWithContext(ctx)
	results = new(resource_structs2.ComponentResult)
	for _, require := range req.Requires {
		result, err := rw.allocForEachRequirement(ctx, tx, &require, rack2hosts, zone2racks, region2zones)
		if err != nil {
			log.Errorf("alloc for requirement %v failed, %v", require, err)
			return nil, err
		}
		results.ZoneResults = append(results.ZoneResults, *result)
	}
	return
}

func (rw *GormResourceReadWrite) allocForEachRequirement(ctx context.Context, tx *gorm.DB, require *resource_structs2.AllocRequirement,
	rack2hosts map[string][]*rp.Host, zone2racks map[string][]string, region2zones map[string][]string) (results *resource_structs2.AllocResult, err error) {
	log := framework.LogWithContext(ctx)
	regionName := require.Location.Region
	zoneName := require.Location.Zone
	if regionName == "" || zoneName == "" {
		errMsg := fmt.Sprintf("neither regionName (%s) nor zoneName (%s) should not be empty in require", regionName, zoneName)
		log.Errorln(errMsg)
		return nil, errors.NewError(errors.TIEM_RESOURCE_ALLOCATE_ERROR, errMsg)
	}
	zoneCode := structs.GenDomainCodeByName(regionName, zoneName)
	racks, ok := zone2racks[zoneCode]
	if !ok {
		errMsg := fmt.Sprintf("no racks under zone (%s)", zoneCode)
		log.Errorln(errMsg)
		return nil, errors.NewError(errors.TIEM_RESOURCE_ALLOCATE_ERROR, errMsg)
	}
	count := 0
	var resources []*Resource
	for count < int(require.Count) {
		found := false
		for _, rack := range racks {
			hosts, ok := rack2hosts[rack]
			if !ok {
				errMsg := fmt.Sprintf("no hosts under rack (%s)", rack)
				log.Errorln(errMsg)
				return nil, errors.NewError(errors.TIEM_RESOURCE_ALLOCATE_ERROR, errMsg)
			}
			for i := range hosts {
				if meetRequire(require, hosts[i]) {
					found = true
					resource := updateHost(ctx, tx, require, hosts[i])
					resources = append(resources, resource)
					break
				}
			}
		}
		if !found {
			break
		}
	}
	if count < int(require.Count) {
		errMsg := fmt.Sprintf("no enough resource to meet require %v", *require)
		log.Errorln(errMsg)
		return nil, errors.NewError(errors.TIEM_RESOURCE_ALLOCATE_ERROR, errMsg)
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
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
		result, err := resource.toCompute()
		if err != nil {
			return nil, err
		}
		result.Reqseq = int32(seq)
		results = append(results, *result)
	}
	return
}

func (rw *GormResourceReadWrite) buildHierarchy(hosts []rp.Host) (rack2hosts map[string][]*rp.Host, zone2racks map[string][]string, region2zones map[string][]string, err error) {
	rack2hosts = make(map[string][]*rp.Host)
	zone2racks = make(map[string][]string)
	region2zones = make(map[string][]string)
	// For deduplicated
	tmp_rack_recorded := make(map[string]bool)
	tmp_zone_recorded := make(map[string]bool)

	for i := range hosts {
		rack2hosts[hosts[i].Rack] = append(rack2hosts[hosts[i].Rack], &hosts[i])
		if _, ok := tmp_rack_recorded[hosts[i].Rack]; !ok {
			tmp_rack_recorded[hosts[i].Rack] = true
			zone2racks[hosts[i].AZ] = append(zone2racks[hosts[i].AZ], hosts[i].Rack)
		}
		if _, ok := tmp_zone_recorded[hosts[i].AZ]; !ok {
			tmp_zone_recorded[hosts[i].AZ] = true
			region2zones[hosts[i].Region] = append(region2zones[hosts[i].Region], hosts[i].AZ)
		}
	}
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
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_ALLOCATE_ERROR, "alloc resources in batch failed on %dth request with %d requires, request: %v, error: %v", i+1, len(request.Requires), request, err)
		}
		results.BatchResults = append(results.BatchResults, result)
	}
	tx.Commit()
	return
}

func (rw *GormResourceReadWrite) allocForSingleRequest(ctx context.Context, tx *gorm.DB, req *resource_structs.AllocReq) (results *resource_structs.AllocRsp, err error) {
	log := framework.LogWithContext(ctx)
	var choosedHosts []string
	results = new(resource_structs.AllocRsp)
	results.Applicant = req.Applicant
	for i, require := range req.Requires {
		switch resource_structs.AllocStrategy(require.Strategy) {
		case resource_structs.RandomRack:
			res, err := rw.allocResourceWithRR(ctx, tx, &req.Applicant, i, &require, choosedHosts)
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
			res, err := rw.allocResourceInHost(ctx, tx, &req.Applicant, i, &require)
			if err != nil {
				log.Errorf("alloc resources in specify host strategy for %dth requirement %v failed, %v", i, require, err)
				return nil, err
			}
			results.Results = append(results.Results, res...)
		case resource_structs.ClusterPorts:
			res, err := rw.allocPortsInRegion(ctx, tx, &req.Applicant, i, &require)
			if err != nil {
				log.Errorf("alloc resources in cluster port strategy for %dth requirement %v failed, %v", i, require, err)
				return nil, err
			}
			results.Results = append(results.Results, res...)
		default:
			return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid alloc strategy %d", require.Strategy)
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
	Region   string
	AZ       string
	Rack     string
	portRes  []*resource_structs.PortResource
}

func (resource *Resource) toCompute() (result *resource_structs.Compute, err error) {
	result = &resource_structs.Compute{
		HostId:   resource.HostId,
		HostName: resource.HostName,
		HostIp:   resource.Ip,
		UserName: resource.UserName,
		Passwd:   resource.Passwd,
		Location: structs.Location{
			Region: structs.GetDomainNameFromCode(resource.Region),
			Zone:   structs.GetDomainNameFromCode(resource.AZ),
			Rack:   structs.GetDomainNameFromCode(resource.Rack),
			HostIp: resource.Ip,
		},
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

	result.Passwd, err = crypto.AesDecryptCFB(result.Passwd)
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_DECRYPT_PASSWD_ERROR, "decrypt compute %v password failed, %v", *result, err)
	}
	return result, nil

}

func (rw *GormResourceReadWrite) allocResourceWithRR(ctx context.Context, tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement, choosedHosts []string) (results []resource_structs.Compute, err error) {
	log := framework.LogWithContext(ctx)
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
			"disks.host_id, hosts.host_name, hosts.region, hosts.az, hosts.rack, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", reqCores, reqMem).Joins(
			"left join hosts on disks.host_id = hosts.id").Where("hosts.reserved = 0")
		if excludedHosts == nil {
			db.Count(&count)
		} else {
			db.Not(map[string]interface{}{"hosts.ip": excludedHosts}).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_DISK, "expect disk count %d but only %d after excluded host list", require.Count, count)
		}

		db = db.Where("disks.status = ? and disks.capacity >= ?", constants.DiskAvailable, capacity).Count(&count)
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_DISK, "expect disk count %d but only %d after disk filter", require.Count, count)
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
			"id as host_id, host_name, ip, region, az, rack, user_name, passwd, ? as cpu_cores, ? as memory", reqCores, reqMem).Where("reserved = 0")
		if excludedHosts == nil {
			db.Count(&count)
		} else {
			db.Not(map[string]interface{}{"hosts.ip": excludedHosts}).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "expect host count %d but only %d after excluded host list", require.Count, count)
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
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "select resources failed, %v", err)
	}

	if len(resources) < int(require.Count) {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "hosts in %s, %s is not enough for allocation(%d|%d), arch: %s, traits: %d", regionName, zoneName, len(resources), require.Count, hostArch, hostTraits)
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
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
		result, err := resource.toCompute()
		if err != nil {
			return nil, err
		}
		result.Reqseq = int32(seq)
		results = append(results, *result)
	}
	return
}

func (rw *GormResourceReadWrite) allocResourceInHost(ctx context.Context, tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement) (results []resource_structs.Compute, err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("allocResourceInHost[%d] for application %v, require: %v", seq, *applicant, *require)
	hostIp := require.Location.HostIp
	if require.Count != 1 {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "request host count should be 1 for UserSpecifyHost allocation(%d)", require.Count)
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
			"disks.host_id, hosts.host_name, hosts.region, hosts.az, hosts.rack, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", reqCores, reqMem).Joins(
			"left join hosts on disks.host_id = hosts.id").Where("hosts.ip = ?", hostIp).Count(&count)
		// No Limit in Reserved == false in this strategy for a takeover operation
		if !isTakeOver {
			db = db.Where("hosts.reserved = 0").Count(&count)
		}

		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "disk is not enough(%d|%d) in host (%s), takeover operation (%v)", count, require.Count, hostIp, isTakeOver)
		}
		db = db.Where("hosts.free_cpu_cores >= ? and hosts.free_memory >= ?", reqCores, reqMem).Count(&count)
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "cpucores or memory in host (%s) is not enough", hostIp)
		}
		if !exclusive {
			db = db.Where("hosts.status = ? and (hosts.stat = ? or hosts.stat = ?)", constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed).Count(&count)
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			db = db.Where("hosts.status = ? and hosts.stat = ?", constants.HostOnline, constants.HostLoadLoadLess).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) status/stat is not expected for exlusive(%v) condition", hostIp, exclusive)
		}
		if diskSpecify == "" {
			err = db.Where("disks.type = ? and disks.status = ? and disks.capacity >= ?", diskType, constants.DiskAvailable, capacity).Scan(&resources).Error
		} else {
			err = db.Where("disks.id = ? and disks.status = ?", diskSpecify, constants.DiskAvailable).Scan(&resources).Error
		}
		if err != nil {
			return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "select resources failed, %v", err)
		}
		if len(resources) < int(require.Count) {
			if diskSpecify == "" {
				return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_DISK, "no available disk with type(%s) and capacity(%d) in host(%s) after disk filter", diskType, capacity, hostIp)
			} else {
				return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_DISK, "disk (%s) not existed or it is not available in host(%s)", diskSpecify, hostIp)
			}
		}
	} else {
		db := tx.Model(&rp.Host{}).Select("id as host_id, host_name, region, az, rack, ip, user_name, passwd, ? as cpu_cores, ? as memory", reqCores, reqMem).Where("ip = ?", hostIp).Count(&count)
		// No Limit in Reserved == false in this strategy for a takeover operation
		if !isTakeOver {
			db = db.Where("hosts.reserved = 0").Count(&count)
		}
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) is not existed or reserved, takeover operation(%v)", hostIp, isTakeOver)
		}
		db = db.Where("free_cpu_cores >= ? and free_memory >= ?", reqCores, reqMem).Count(&count)
		if count < int64(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "cpucores or memory in host (%s) is not enough", hostIp)
		}
		if !exclusive {
			err = db.Where("status = ? and (stat = ? or stat = ?)", constants.HostOnline, constants.HostLoadLoadLess, constants.HostLoadInUsed).Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("status = ? and stat = ?", constants.HostOnline, constants.HostLoadLoadLess).Scan(&resources).Error
		}
		if err != nil {
			return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "select resources failed, %v", err)
		}
		if len(resources) < int(require.Count) {
			return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) status/stat is not expected for exlusive(%v) condition", hostIp, exclusive)
		}
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
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
		result, err := resource.toCompute()
		if err != nil {
			return nil, err
		}
		result.Reqseq = int32(seq)
		results = append(results, *result)
	}
	return
}

func (rw *GormResourceReadWrite) allocPortsInRegion(ctx context.Context, tx *gorm.DB, applicant *resource_structs.Applicant, seq int, require *resource_structs.AllocRequirement) (results []resource_structs.Compute, err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("allocPortsInRegion[%d] for application: %v, require: %v", seq, *applicant, *require)
	regionCode := require.Location.Region
	hostArch := require.HostFilter.Arch
	if regionCode == "" {
		return nil, errors.NewError(errors.TIEM_RESOURCE_INVALID_LOCATION, "no valid region")
	}
	if hostArch == "" {
		return nil, errors.NewError(errors.TIEM_RESOURCE_INVALID_ARCH, "no valid arch")
	}
	if len(require.Require.PortReq) != 1 {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, "require portReq len should be 1 for RegionUniformPorts allocation(%d)", len(require.Require.PortReq))
	}
	portReq := require.Require.PortReq[0]
	var regionHosts []string
	err = tx.Model(&rp.Host{}).Select("id").Where("region = ? and arch = ?", regionCode, hostArch).Where("reserved = 0").Scan(&regionHosts).Error
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "select %s hosts in region %s failed, %v", hostArch, regionCode, err)
	}
	if len(regionHosts) == 0 {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_HOST, "no %s host in region %s", hostArch, regionCode)
	}

	var usedPorts []int32
	err = tx.Order("port").Model(&mm.UsedPort{}).Select("port").Where("host_id in ?", regionHosts).Group("port").Having("port >= ? and port < ?", portReq.Start, portReq.End).Scan(&usedPorts).Error
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "select used port range %d - %d in region %s failed, %v", portReq.Start, portReq.End, regionCode, err)
	}
	log.Infof("region used port list: %v, hosts: %v, start: %d, end: %d", usedPorts, regionHosts, portReq.Start, portReq.End)

	res, err := rw.getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, "Region %s has no enough %d ports on range [%d, %d]", regionCode, portReq.PortCnt, portReq.Start, portReq.End)
	}
	log.Infof("get region ports: %v, [%d, %d, %d]", res, portReq.Start, portReq.End, portReq.PortCnt)

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
			return nil, errors.NewError(errors.TIEM_RESOURCE_NO_ENOUGH_PORT, errors.TIEM_RESOURCE_NO_ENOUGH_PORT.Explain())
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
					return errors.NewErrorf(errors.TIEM_SQL_ERROR, "update disk(%s) status err, %v", resource.DiskId, err)
				}
			} else {
				return errors.NewErrorf(errors.TIEM_SQL_ERROR, "disk %s status not expected(%s)", resource.DiskId, disk.Status)
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
				return errors.NewErrorf(errors.TIEM_SQL_ERROR, "insert disk(%s) to used_disks table failed: %v", resource.DiskId, err)
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
			return errors.NewErrorf(errors.TIEM_SQL_ERROR, "update host(%s) stat err, %v", resource.HostId, err)
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
			return errors.NewErrorf(errors.TIEM_SQL_ERROR, "insert host(%s) to used_computes table failed: %v", resource.HostId, err)
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
					return errors.NewErrorf(errors.TIEM_SQL_ERROR, "insert host(%s) for port(%d) table failed: %v", resource.HostId, port, err)
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
				return errors.NewErrorf(errors.TIEM_SQL_ERROR, "insert host(%s) for port(%d) table failed: %v", host, port, err)
			}
		}
	}
	return
}
