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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	"github.com/pingcap-inc/tiem/library/common"
	rt "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/util/bitmap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type DAOResourceManager struct {
	db *gorm.DB
}

func NewDAOResourceManager(d *gorm.DB) *DAOResourceManager {
	m := new(DAOResourceManager)
	m.SetDb(d)
	return m
}

func (m *DAOResourceManager) SetDb(d *gorm.DB) {
	m.db = d
}

func (m *DAOResourceManager) getDb(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *DAOResourceManager) InitSystemDefaultLabels(ctx context.Context) (err error) {
	db := m.getDb(ctx)
	for _, v := range rt.DefaultLabelTypes {
		err = db.Create(&v).Error
		if err != nil {
			return err
		}
	}
	return err
}

func (m *DAOResourceManager) CreateHost(ctx context.Context, host *rt.Host) (id string, err error) {
	err = m.getDb(ctx).Create(host).Error
	if err != nil {
		return
	}
	return host.ID, err
}

func (m *DAOResourceManager) CreateHostsInBatch(ctx context.Context, hosts []*rt.Host) (ids []string, err error) {
	tx := m.getDb(ctx).Begin()
	for _, host := range hosts {
		err = tx.Create(host).Error
		if err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Canceled, "create %s(%s) err, %v", host.HostName, host.IP, err)
		}
		ids = append(ids, host.ID)
	}
	err = tx.Commit().Error
	return
}

func (m *DAOResourceManager) DeleteHost(ctx context.Context, hostId string) (err error) {
	err = m.getDb(ctx).Where("ID = ?", hostId).Delete(&rt.Host{
		ID: hostId,
	}).Error
	return
}

func (m *DAOResourceManager) DeleteHostsInBatch(ctx context.Context, hostIds []string) (err error) {
	tx := m.getDb(ctx).Begin()
	for _, hostId := range hostIds {
		var host rt.Host
		if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&host, "ID = ?", hostId).Error; err != nil {
			tx.Rollback()
			return status.Errorf(codes.FailedPrecondition, "lock host %s(%s) error, %v", hostId, host.IP, err)
		}
		err = tx.Delete(&host).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit().Error
	return
}

type ListHostReq struct {
	Status  rt.HostStatus
	Stat    rt.HostStat
	Purpose string
	Offset  int
	Limit   int
}

func (m *DAOResourceManager) ListHosts(ctx context.Context, req ListHostReq) (hosts []rt.Host, err error) {
	db := m.getDb(ctx).Table(TABLE_NAME_HOST)
	if err = db.Error; err != nil {
		return nil, err
	}
	if req.Status != rt.HOST_WHATEVER {
		if req.Status != rt.HOST_DELETED {
			db = db.Where("status = ?", req.Status)
		} else {
			db = db.Unscoped().Where("status = ?", req.Status)
		}
	}
	if req.Stat != rt.HOST_STAT_WHATEVER {
		db = db.Where("stat = ?", req.Stat)
	}
	if req.Purpose != "" {
		db = db.Where("purpose = ?", req.Purpose)
	}
	err = db.Offset(req.Offset).Limit(req.Limit).Find(&hosts).Error
	return
}

func (m *DAOResourceManager) FindHostById(ctx context.Context, hostId string) (*rt.Host, error) {
	host := new(rt.Host)
	err := m.getDb(ctx).First(host, "ID = ?", hostId).Error
	return host, err
}

type DiskResource struct {
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
}

// For each Request for one compoent in the same FailureDomain, we should alloc disks in different hosts
type HostAllocReq struct {
	FailureDomain string
	CpuCores      int32
	Memory        int32
	Count         int
}

type AllocReqs map[string][]*HostAllocReq
type AllocRsps map[string][]*DiskResource

func getHostsFromFailureDomain(tx *gorm.DB, failureDomain string, numReps int, cpuCores int32, mem int32) (resources []*DiskResource, err error) {
	err = tx.Order("hosts.free_cpu_cores desc").Order("hosts.free_memory desc").Limit(numReps).Model(&rt.Disk{}).Select(
		"disks.host_id, hosts.host_name, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", cpuCores, mem).Joins(
		"left join hosts on disks.host_id = hosts.id").Where(
		"hosts.az = ? and hosts.status = ? and (hosts.stat = ? or hosts.stat = ?) and hosts.free_cpu_cores >= ? and hosts.free_memory >= ? and disks.status = ?",
		failureDomain, rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED, cpuCores, mem, rt.DISK_AVAILABLE).Group("hosts.id").Scan(&resources).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "select resources failed, %v", err)
	}

	if len(resources) < numReps {
		return nil, status.Errorf(codes.Internal, "hosts in %s is not enough for allocation(%d|%d)", failureDomain, len(resources), numReps)
	}

	for _, resource := range resources {
		var disk rt.Disk
		tx.First(&disk, "ID = ?", resource.DiskId).First(&disk)
		if rt.DiskStatus(disk.Status).IsAvailable() {
			err = tx.Model(&disk).Update("Status", int32(rt.DISK_INUSED)).Error
			if err != nil {
				return nil, status.Errorf(codes.Internal, "update disk(%s) status err, %v", resource.DiskId, err)
			}
		} else {
			return nil, status.Errorf(codes.FailedPrecondition, "disk %s status not expected(%d)", resource.DiskId, disk.Status)
		}

		var host rt.Host
		tx.First(&host, "ID = ?", resource.HostId)
		host.FreeCpuCores -= cpuCores
		host.FreeMemory -= mem
		host.Stat = int32(rt.HOST_INUSED)
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", resource.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return nil, status.Errorf(codes.Internal, "update host(%s) stat err, %v", resource.HostId, err)
		}
	}
	return
}

func (m *DAOResourceManager) AllocHosts(ctx context.Context, requests AllocReqs) (resources AllocRsps, err error) {
	log := framework.LogWithContext(ctx)
	resources = make(AllocRsps)
	tx := m.getDb(ctx).Begin()
	for component, reqs := range requests {
		for _, eachReq := range reqs {
			log.Infof("alloc resources for component %s in %s (%dC%dG) x %d", component, eachReq.FailureDomain, eachReq.CpuCores, eachReq.Memory, eachReq.Count)
			disks, err := getHostsFromFailureDomain(tx, eachReq.FailureDomain, eachReq.Count, eachReq.CpuCores, eachReq.Memory)
			if err != nil {
				log.Errorf("failed to alloc host info for %s in %s, %v", component, eachReq.FailureDomain, err)
				tx.Rollback()
				return nil, err
			}
			resources[component] = append(resources[component], disks...)
		}
	}
	tx.Commit()
	return resources, nil
}

type FailureDomainResource struct {
	FailureDomain string
	Purpose       string
	CpuCores      int
	Memory        int
	Count         int
}

func (m *DAOResourceManager) GetFailureDomain(ctx context.Context, domain string) (res []FailureDomainResource, err error) {
	selectStr := fmt.Sprintf("%s as FailureDomain, purpose, cpu_cores, memory, count(id) as Count", domain)
	err = m.getDb(ctx).Table("hosts").Where("Status = ? and (Stat = ? or Stat = ?)", rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED).Select(selectStr).
		Group(domain).Group("purpose").Group("cpu_cores").Group("memory").Scan(&res).Error
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
	portRes  []*rt.PortResource
}

func getPortsInRange(usedPorts []int32, start int32, end int32, count int) (*rt.PortResource, error) {
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
	result := &rt.PortResource{
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

func markResourcesForUsed(tx *gorm.DB, applicant *dbpb.DBApplicant, resources []*Resource, exclusive bool) (err error) {
	for _, resource := range resources {
		if resource.DiskId != "" {
			var disk rt.Disk
			tx.First(&disk, "ID = ?", resource.DiskId).First(&disk)
			if rt.DiskStatus(disk.Status).IsAvailable() {
				err = tx.Model(&disk).Update("Status", int32(rt.DISK_EXHAUST)).Error
				if err != nil {
					return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update disk(%s) status err, %v", resource.DiskId, err)
				}
			} else {
				return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "disk %s status not expected(%d)", resource.DiskId, disk.Status)
			}
			usedDisk := rt.UsedDisk{
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

		var host rt.Host
		tx.First(&host, "ID = ?", resource.HostId)
		host.FreeCpuCores -= int32(resource.CpuCores)
		host.FreeMemory -= int32(resource.Memory)
		if exclusive {
			host.Stat = int32(rt.HOST_EXCLUSIVE)
		} else {
			host.Stat = int32(rt.HOST_INUSED)
		}
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", resource.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update host(%s) stat err, %v", resource.HostId, err)
		}
		usedCompute := rt.UsedCompute{
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
				usedPort := rt.UsedPort{
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

func allocResourceInHost(tx *gorm.DB, applicant *dbpb.DBApplicant, seq int, require *dbpb.DBAllocRequirement) (results []rt.HostResource, err error) {
	hostIp := require.Location.Host
	if require.Count != 1 {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "request host count should be 1 for UserSpecifyHost allocation(%d)", require.Count)
	}
	reqCores := require.Require.ComputeReq.CpuCores
	reqMem := require.Require.ComputeReq.Memory
	exclusive := require.Require.Exclusive
	isTakeOver := applicant.TakeoverOperation

	needDisk := require.Require.DiskReq.NeedDisk
	diskSpecify := require.Require.DiskReq.DiskSpecify
	diskType := rt.DiskType(require.Require.DiskReq.DiskType)
	capacity := require.Require.DiskReq.Capacity

	var resources []*Resource
	var count int64

	if needDisk {
		db := tx.Order("disks.capacity").Limit(int(require.Count)).Model(&rt.Disk{}).Select(
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
			db = db.Where("hosts.status = ? and (hosts.stat = ? or hosts.stat = ?)", rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED).Count(&count)
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			db = db.Where("hosts.status = ? and hosts.stat = ?", rt.HOST_ONLINE, rt.HOST_LOADLESS).Count(&count)
		}
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "host(%s) status/stat is not expected for exlusive(%v) condition", hostIp, exclusive)
		}
		if diskSpecify == "" {
			err = db.Where("disks.type = ? and disks.status = ? and disks.capacity >= ?", diskType, rt.DISK_AVAILABLE, capacity).Scan(&resources).Error
		} else {
			err = db.Where("disks.id = ? and disks.status = ?", diskSpecify, rt.DISK_AVAILABLE).Scan(&resources).Error
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
		db := tx.Model(&rt.Host{}).Select("id as host_id, host_name, ip, user_name, passwd, ? as cpu_cores, ? as memory", reqCores, reqMem).Where("ip = ?", hostIp).Count(&count)
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
			err = db.Where("status = ? and (stat = ? or stat = ?)", rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED).Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("status = ? and stat = ?", rt.HOST_ONLINE, rt.HOST_LOADLESS).Scan(&resources).Error
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
		tx.Order("port").Model(&rt.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
			}
			resource.portRes = append(resource.portRes, res)
		}
	}

	// 3. Mark Resources in used
	err = markResourcesForUsed(tx, applicant, resources, exclusive)
	if err != nil {
		return nil, err
	}

	// 4. make Results and Complete one Requirement
	for _, resource := range resources {
		result := rt.HostResource{
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

func allocResourceWithRR(tx *gorm.DB, applicant *dbpb.DBApplicant, seq int, require *dbpb.DBAllocRequirement, choosedHosts []string) (results []rt.HostResource, err error) {
	regionName := require.Location.Region
	zoneName := require.Location.Zone
	zoneCode := rt.GenDomainCodeByName(regionName, zoneName)
	var excludedHosts []string
	if require.HostExcluded != nil {
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
	// 1. Choose Host/Disk List
	var resources []*Resource
	if needDisk {
		var count int64
		db := tx.Order("hosts.free_cpu_cores desc").Order("hosts.free_memory desc").Limit(int(require.Count)).Model(&rt.Disk{}).Select(
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

		db = db.Where("disks.status = ? and disks.capacity >= ?", rt.DISK_AVAILABLE, capacity).Count(&count)
		if count < int64(require.Count) {
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "expect disk count %d but only %d after disk filter", require.Count, count)
		}

		if !exclusive {
			err = db.Where("hosts.az = ? and hosts.arch = ? and hosts.traits & ? = ? and hosts.status = ? and (hosts.stat = ? or hosts.stat = ?) and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("hosts.az = ? and hosts.arch = ? and hosts.traits & ? = ? and hosts.status = ? and hosts.stat = ? and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, rt.HOST_ONLINE, rt.HOST_LOADLESS, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		}
	} else {
		var count int64
		db := tx.Order("hosts.free_cpu_cores desc").Order("hosts.free_memory desc").Limit(int(require.Count)).Model(&rt.Host{}).Select(
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
				zoneCode, hostArch, hostTraits, hostTraits, rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED, reqCores, reqMem).Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("az = ? and arch = ? and traits & ? = ? and status = ? and stat = ? and free_cpu_cores >= ? and free_memory >= ?",
				zoneCode, hostArch, hostTraits, hostTraits, rt.HOST_ONLINE, rt.HOST_LOADLESS, reqCores, reqMem).Scan(&resources).Error
		}
	}
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
	}

	if len(resources) < int(require.Count) {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "hosts in %s,%s is not enough for allocation(%d|%d)", regionName, zoneName, len(resources), require.Count)
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&rt.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
			}
			resource.portRes = append(resource.portRes, res)
		}
	}

	// 3. Mark Resources in used
	err = markResourcesForUsed(tx, applicant, resources, exclusive)
	if err != nil {
		return nil, err
	}

	// 4. make Results and Complete one Requirement
	for _, resource := range resources {
		result := rt.HostResource{
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

func markPortsInRegion(tx *gorm.DB, applicant *dbpb.DBApplicant, regionHosts []string, portResources *rt.PortResource) (err error) {
	for _, host := range regionHosts {
		for _, port := range portResources.Ports {
			usedPort := rt.UsedPort{
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

func allocPortsInRegion(tx *gorm.DB, applicant *dbpb.DBApplicant, seq int, require *dbpb.DBAllocRequirement) (results []rt.HostResource, err error) {
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
	err = tx.Model(&rt.Host{}).Select("id").Where("region = ? and arch = ?", regionCode, hostArch).Where("reserved = 0").Scan(&regionHosts).Error
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
	}
	if len(regionHosts) == 0 {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "no %s host in region %s", hostArch, regionCode)
	}

	var usedPorts []int32
	tx.Order("port").Model(&rt.UsedPort{}).Select("port").Where("host_id in ?", regionHosts).Group("port").Having("port >= ? and port < ?", portReq.Start, portReq.End).Scan(&usedPorts)

	res, err := getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
	if err != nil {
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "Region %s has no enough %d ports on range [%d, %d]", regionCode, portReq.PortCnt, portReq.Start, portReq.End)
	}

	err = markPortsInRegion(tx, applicant, regionHosts, res)
	if err != nil {
		return nil, err
	}

	var result rt.HostResource
	result.Reqseq = int32(seq)
	result.PortRes = append(result.PortRes, *res)

	results = append(results, result)
	return
}

func (m *DAOResourceManager) doAlloc(tx *gorm.DB, req *dbpb.DBAllocRequest) (results *rt.AllocRsp, err error) {
	var choosedHosts []string
	results = new(rt.AllocRsp)
	for i, require := range req.Requires {
		switch rt.AllocStrategy(require.Strategy) {
		case rt.RandomRack:
			res, err := allocResourceWithRR(tx, req.Applicant, i, require, choosedHosts)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "alloc with RandomRack %dth require failed, %v", i, err)
			}
			for _, result := range res {
				choosedHosts = append(choosedHosts, result.HostIp)
			}
			results.Results = append(results.Results, res...)
		case rt.DiffRackBestEffort:
		case rt.UserSpecifyRack:
		case rt.UserSpecifyHost:
			res, err := allocResourceInHost(tx, req.Applicant, i, require)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "alloc resources in host %s with %dth require failed, %v", require.Location.Host, i, err)
			}
			results.Results = append(results.Results, res...)
		case rt.ClusterPorts:
			res, err := allocPortsInRegion(tx, req.Applicant, i, require)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "alloc region port range in region %s with %dth require failed, %v", require.Location.Region, i, err)
			}
			results.Results = append(results.Results, res...)
		default:
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_STRATEGY, "invalid alloc strategy %d", require.Strategy)
		}
	}
	return
}

func (m *DAOResourceManager) AllocResources(ctx context.Context, req *dbpb.DBAllocRequest) (result *rt.AllocRsp, err error) {
	tx := m.getDb(ctx).Begin()
	result, err = m.doAlloc(tx, req)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func (m *DAOResourceManager) AllocResourcesInBatch(ctx context.Context, batchReq *dbpb.DBBatchAllocRequest) (results *rt.BatchAllocResponse, err error) {
	results = new(rt.BatchAllocResponse)
	tx := m.getDb(ctx).Begin()
	for i, req := range batchReq.BatchRequests {
		var result *rt.AllocRsp
		result, err = m.doAlloc(tx, req)
		if err != nil {
			tx.Rollback()
			return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_NOT_ALL_SUCCEED, "alloc resources in batch failed on request %d, %v", i, err)
		}
		results.BatchResults = append(results.BatchResults, result)
	}
	tx.Commit()
	return
}

func recycleUsedTablesByClusterId(tx *gorm.DB, clusterId string) (err error) {
	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedCompute{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedCompute for cluster %s failed, %v", clusterId, err)
	}

	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedDisk{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedDisk for cluster %s failed, %v", clusterId, err)
	}

	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedPort{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedPort for cluster %s failed, %v", clusterId, err)
	}
	return nil
}

func recycleUsedTablesByRequestId(tx *gorm.DB, requestId string) (err error) {
	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedCompute{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedCompute for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedDisk{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedDisk for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedPort{}).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedPort for request %s failed, %v", requestId, err)
	}
	return nil
}

func recycleUsedTablesBySpecify(tx *gorm.DB, holderId, requestId, hostId string, cpuCores int32, memory int32, diskIds []string, ports []int32) (err error) {
	recycleCompute := rt.UsedCompute{
		HostId:   hostId,
		CpuCores: -cpuCores,
		Memory:   -memory,
	}
	recycleCompute.Holder.HolderId = holderId
	recycleCompute.RequestId = requestId
	err = tx.Create(&recycleCompute).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle host(%s) to used_computes table failed: %v", hostId, err)
	}
	var usedCompute ComputeStatistic
	err = tx.Model(&rt.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where(
		"holder_id = ?", holderId).Group("host_id").Having("host_id = ?", hostId).Scan(&usedCompute).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get host %s total used compute for cluster %s failed, %v", hostId, holderId, err)
	}
	if usedCompute.TotalCpuCores == 0 && usedCompute.TotalMemory == 0 {
		err = tx.Where("host_id = ? and holder_id = ?", hostId, holderId).Delete(&rt.UsedCompute{}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "clean up UsedCompute in host %s for cluster %s failed, %v", hostId, holderId, err)
		}
	}

	for _, diskId := range diskIds {
		err = tx.Where("host_id = ? and holder_id = ? and disk_id = ?", hostId, holderId, diskId).Delete(&rt.UsedDisk{}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedDisk for disk %s in host %s failed, %v", diskId, hostId, err)
		}
	}

	for _, port := range ports {
		err = tx.Where("host_id = ? and holder_id = ? and port = ?", hostId, holderId, port).Delete(&rt.UsedPort{}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedPort for %d in host %s failed, %v", port, hostId, err)
		}
	}
	return nil
}

type ComputeStatistic struct {
	HostId        string
	TotalCpuCores int
	TotalMemory   int
}

func recycleResourcesInHosts(tx *gorm.DB, usedCompute []ComputeStatistic, usedDisks []string) (err error) {
	for _, diskId := range usedDisks {
		var disk rt.Disk
		tx.First(&disk, "ID = ?", diskId).First(&disk)
		if rt.DiskStatus(disk.Status).IsExhaust() {
			err = tx.Model(&disk).Update("Status", int32(rt.DISK_AVAILABLE)).Error
			if err != nil {
				return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update disk(%s) status while recycle failed, %v", diskId, err)
			}
		} else {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "disk %s status not expected(%d) while recycle", diskId, disk.Status)
		}
	}
	for _, usedCompute := range usedCompute {
		var host rt.Host
		tx.First(&host, "ID = ?", usedCompute.HostId)
		host.FreeCpuCores += int32(usedCompute.TotalCpuCores)
		host.FreeMemory += int32(usedCompute.TotalMemory)

		if host.IsLoadless() {
			host.Stat = int32(rt.HOST_LOADLESS)
		} else {
			host.Stat = int32(rt.HOST_INUSED)
		}

		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", usedCompute.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "update host(%s) stat while recycle failed, %v", usedCompute.HostId, err)
		}
	}
	return nil
}

func recycleHolderResource(tx *gorm.DB, clusterId string) (err error) {
	var usedCompute []ComputeStatistic
	err = tx.Model(&rt.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get cluster %s total used compute failed, %v", clusterId, err)
	}

	var usedDisks []string
	err = tx.Model(&rt.UsedDisk{}).Select("disk_id").Where("holder_id = ?", clusterId).Scan(&usedDisks).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get cluster %s total used disks failed, %v", clusterId, err)
	}

	// update stat for hosts and disks
	err = recycleResourcesInHosts(tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	// Drop used resources in used_computes/used_disks/used_ports
	err = recycleUsedTablesByClusterId(tx, clusterId)
	if err != nil {
		return err
	}

	return nil
}

func recycleResourceForRequest(tx *gorm.DB, requestId string) (err error) {
	var usedCompute []ComputeStatistic
	err = tx.Model(&rt.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("request_id = ?", requestId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get request %s total used compute failed, %v", requestId, err)
	}

	var usedDisks []string
	err = tx.Model(&rt.UsedDisk{}).Select("disk_id").Where("request_id = ?", requestId).Scan(&usedDisks).Error
	if err != nil {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get request %s total used disks failed, %v", requestId, err)
	}

	// update stat for hosts and disks
	err = recycleResourcesInHosts(tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	// Drop used resources in used_computes/used_disks/used_ports
	err = recycleUsedTablesByRequestId(tx, requestId)
	if err != nil {
		return err
	}

	return nil
}

func recycleHostResource(tx *gorm.DB, clusterId string, requestId string, hostId string, computeReq *dbpb.DBComputeRequirement, diskReq []*dbpb.DBDiskResource, portReqs []*dbpb.DBPortResource) (err error) {
	var usedCompute []ComputeStatistic
	var usedDisks []string
	var usedPorts []int32
	if computeReq != nil {
		usedCompute = append(usedCompute, ComputeStatistic{
			HostId:        hostId,
			TotalCpuCores: int(computeReq.CpuCores),
			TotalMemory:   int(computeReq.Memory),
		})
	}

	for _, diskResouce := range diskReq {
		usedDisks = append(usedDisks, diskResouce.DiskId)
	}

	// Update Host Status and Disk Status
	err = recycleResourcesInHosts(tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	for _, portReq := range portReqs {
		usedPorts = append(usedPorts, portReq.Ports...)
	}

	err = recycleUsedTablesBySpecify(tx, clusterId, requestId, hostId, computeReq.CpuCores, computeReq.Memory, usedDisks, usedPorts)
	if err != nil {
		return err
	}

	return nil
}

func (m *DAOResourceManager) doRecycle(tx *gorm.DB, req *dbpb.DBRecycleRequire) (err error) {
	switch rt.RecycleType(req.RecycleType) {
	case rt.RecycleHolder:
		return recycleHolderResource(tx, req.HolderId)
	case rt.RecycleOperate:
		return recycleResourceForRequest(tx, req.RequestId)
	case rt.RecycleHost:
		return recycleHostResource(tx, req.HolderId, req.RequestId, req.HostId, req.ComputeReq, req.DiskReq, req.PortReq)
	default:
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVAILD_RECYCLE_TYPE, "invalid recycle resource type %d", req.RecycleType)
	}
}

func (m *DAOResourceManager) RecycleAllocResources(ctx context.Context, request *dbpb.DBRecycleRequest) (err error) {
	tx := m.getDb(ctx).Begin()
	for i, req := range request.RecycleReqs {
		err = m.doRecycle(tx, req)
		if err != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESOURCE_NOT_ALL_SUCCEED, "recycle resources failed on request %d, %v", i, err)
		}
	}
	tx.Commit()
	return
}

func (m *DAOResourceManager) UpdateHostStatus(ctx context.Context, request *dbpb.DBUpdateHostStatusRequest) (err error) {
	tx := m.getDb(ctx).Begin()
	for _, hostId := range request.HostIds {
		result := tx.Model(&rt.Host{}).Where("id = ?", hostId).Update("status", request.Status)
		if result.Error != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_UPDATE_HOST_STATUS_FAIL, "update host [%s] status to %d fail", hostId, request.Status)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_UPDATE_HOST_STATUS_FAIL, "update invaild host [%s] status", hostId)
		}
	}
	tx.Commit()
	return nil
}

func (m *DAOResourceManager) ReserveHost(ctx context.Context, request *dbpb.DBReserveHostRequest) (err error) {
	tx := m.getDb(ctx).Begin()
	for _, hostId := range request.HostIds {
		result := tx.Model(&rt.Host{}).Where("id = ?", hostId).Update("reserved", request.Reserved)
		if result.Error != nil {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESERVE_HOST_FAIL, "set host [%s] reserved status to %v fail", hostId, request.Reserved)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return framework.NewTiEMErrorf(common.TIEM_RESERVE_HOST_FAIL, "set reserved status for invaild host [%s]", hostId)
		}
	}
	tx.Commit()
	return nil
}

type Item struct {
	Region string
	Az     string
	Rack   string
	Ip     string
	Name   string
}

func (m *DAOResourceManager) GetHostItems(ctx context.Context, filter rt.Filter, level int32, depth int32) (Items []Item, err error) {
	leafLevel := level + depth
	tx := m.getDb(ctx).Begin()
	db := tx.Model(&rt.Host{}).Select("region, az, rack, ip, host_name as name")
	if filter.Arch != "" {
		db = db.Where("arch = ?", filter.Arch)
	}
	if filter.Purpose != "" {
		db = db.Where("purpose = ?", filter.Purpose)
	}
	if filter.DiskType != "" {
		db = db.Where("disk_type = ?", filter.DiskType)
	}
	switch rt.FailureDomain(leafLevel) {
	case rt.REGION:
		err = db.Group("region").Scan(&Items).Error
	case rt.ZONE:
		err = db.Group("region").Group("az").Scan(&Items).Error
	case rt.RACK:
		err = db.Group("region").Group("az").Group("rack").Scan(&Items).Error
	case rt.HOST:
		// Build whole tree with hosts
		err = db.Order("region").Order("az").Order("rack").Order("ip").Scan(&Items).Error
	default:
		errMsg := fmt.Sprintf("invaild leaf level %d, level = %d, depth = %d", leafLevel, level, depth)
		err = errors.New(errMsg)
	}
	if err != nil {
		tx.Rollback()
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get hierarchy failed, %v", err)
	}
	tx.Commit()
	return
}

type HostCondition struct {
	Status *int32
	Stat   *int32
	Arch   *string
}
type DiskCondition struct {
	Type     *string
	Capacity *int32
	Status   *int32
}
type StockCondition struct {
	Location      rt.Location
	HostCondition HostCondition
	DiskCondition DiskCondition
}

type Stock struct {
	FreeCpuCores     int
	FreeMemory       int
	FreeDiskCount    int
	FreeDiskCapacity int
}

func (m *DAOResourceManager) GetStocks(ctx context.Context, stockCondition StockCondition) (stocks []Stock, err error) {
	tx := m.getDb(ctx).Begin()
	db := tx.Model(&rt.Host{}).Select(
		"hosts.free_cpu_cores as free_cpu_cores, hosts.free_memory as free_memory, count(disks.id) as free_disk_count, sum(disks.capacity) as free_disk_capacity").Joins(
		"left join disks on disks.host_id = hosts.id")
	if stockCondition.DiskCondition.Status != nil {
		db = db.Where("disks.status = ?", stockCondition.DiskCondition.Status)
	}
	if stockCondition.DiskCondition.Type != nil {
		db = db.Where("disks.type = ?", stockCondition.DiskCondition.Type)
	}
	if stockCondition.DiskCondition.Capacity != nil {
		db = db.Where("disks.capacity >= ?", stockCondition.DiskCondition.Capacity)
	}
	db = db.Group("hosts.id")
	// Filter by Location
	if stockCondition.Location.Host != "" {
		db = db.Having("hosts.ip = ?", stockCondition.Location.Host)
	}
	if stockCondition.Location.Rack != "" {
		db = db.Having("hosts.rack = ?", stockCondition.Location.Rack)
	}
	if stockCondition.Location.Zone != "" {
		db = db.Having("hosts.az = ?", stockCondition.Location.Zone)
	}
	if stockCondition.Location.Region != "" {
		db = db.Having("hosts.region = ?", stockCondition.Location.Region)
	}
	// Filter by host fields
	if stockCondition.HostCondition.Arch != nil {
		db = db.Having("hosts.arch = ?", stockCondition.HostCondition.Arch)
	}
	if stockCondition.HostCondition.Status != nil {
		db = db.Having("hosts.status = ?", stockCondition.HostCondition.Status)
	}
	if stockCondition.HostCondition.Stat != nil {
		db = db.Having("hosts.stat = ?", stockCondition.HostCondition.Stat)
	}

	err = db.Scan(&stocks).Error
	if err != nil {
		tx.Rollback()
		return nil, framework.NewTiEMErrorf(common.TIEM_RESOURCE_SQL_ERROR, "get stocks failed, %v", err)
	}
	tx.Commit()
	return
}
