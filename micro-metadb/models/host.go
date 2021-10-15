package models

import (
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

func (m *DAOResourceManager) getDb() *gorm.DB {
	return m.db
}

func (m *DAOResourceManager) CreateHost(host *rt.Host) (id string, err error) {
	err = m.getDb().Create(host).Error
	if err != nil {
		return
	}
	return host.ID, err
}

func (m *DAOResourceManager) CreateHostsInBatch(hosts []*rt.Host) (ids []string, err error) {
	tx := m.getDb().Begin()
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

func (m *DAOResourceManager) DeleteHost(hostId string) (err error) {
	err = m.getDb().Where("ID = ?", hostId).Delete(&rt.Host{
		ID: hostId,
	}).Error
	return
}

func (m *DAOResourceManager) DeleteHostsInBatch(hostIds []string) (err error) {
	tx := m.getDb().Begin()
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
	Purpose string
	Offset  int
	Limit   int
}

func (m *DAOResourceManager) ListHosts(req ListHostReq) (hosts []rt.Host, err error) {
	db := m.getDb().Table(TABLE_NAME_HOST)
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
	if req.Purpose != "" {
		db = db.Where("purpose = ?", req.Purpose)
	}
	err = db.Offset(req.Offset).Limit(req.Limit).Find(&hosts).Error
	return
}

func (m *DAOResourceManager) FindHostById(hostId string) (*rt.Host, error) {
	host := new(rt.Host)
	err := m.getDb().First(host, "ID = ?", hostId).Error
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
		if host.IsExhaust() {
			host.Stat = int32(rt.HOST_EXHAUST)
		} else {
			host.Stat = int32(rt.HOST_INUSED)
		}
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", resource.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return nil, status.Errorf(codes.Internal, "update host(%s) stat err, %v", resource.HostId, err)
		}
	}
	return
}

func (m *DAOResourceManager) AllocHosts(requests AllocReqs) (resources AllocRsps, err error) {
	log := framework.Log()
	resources = make(AllocRsps)
	tx := m.getDb().Begin()
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

func (m *DAOResourceManager) GetFailureDomain(domain string) (res []FailureDomainResource, err error) {
	selectStr := fmt.Sprintf("%s as FailureDomain, purpose, cpu_cores, memory, count(id) as Count", domain)
	err = m.getDb().Table("hosts").Where("Status = ? and (Stat = ? or Stat = ?)", rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED).Select(selectStr).
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
			return nil, status.Errorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, common.TiEMErrMsg[common.TIEM_RESOURCE_NO_ENOUGH_PORT])
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
					return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "update disk(%s) status err, %v", resource.DiskId, err)
				}
			} else {
				return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "disk %s status not expected(%d)", resource.DiskId, disk.Status)
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
				return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "insert disk(%s) to used_disks table failed: %v", resource.DiskId, err)
			}
		}

		var host rt.Host
		tx.First(&host, "ID = ?", resource.HostId)
		host.FreeCpuCores -= int32(resource.CpuCores)
		host.FreeMemory -= int32(resource.Memory)
		if exclusive {
			host.Stat = rt.HOST_EXCLUSIVE
		} else {
			if host.IsExhaust() {
				host.Stat = rt.HOST_EXHAUST
			} else {
				host.Stat = rt.HOST_INUSED
			}
		}
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", resource.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "update host(%s) stat err, %v", resource.HostId, err)
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
			return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "insert host(%s) to used_disks table failed: %v", resource.HostId, err)
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
					return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "insert host(%s) for port(%d) table failed: %v", resource.HostId, port, err)
				}
			}
		}
	}
	return nil
}

func allocResourceWithRR(tx *gorm.DB, applicant *dbpb.DBApplicant, seq int, require *dbpb.DBAllocRequirement, choosedHosts []string) (results []rt.HostResource, err error) {
	region := require.Location.Region
	zone := require.Location.Zone
	var excludedHosts []string
	if require.HostExcluded != nil {
		excludedHosts = require.HostExcluded.Hosts
	}
	// excluded choosed hosts in one request
	excludedHosts = append(excludedHosts, choosedHosts...)
	hostArch := require.HostFilter.Arch
	hostPurpose := require.HostFilter.Purpose
	hostDiskType := require.HostFilter.DiskType
	exclusive := require.Require.Exclusive
	reqCores := require.Require.ComputeReq.CpuCores
	reqMem := require.Require.ComputeReq.Memory
	diskType := rt.DiskType(require.Require.DiskReq.DiskType)
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
			return nil, status.Errorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED, "expect disk count %d but only %d after excluded host list", require.Count, count)
		}

		db = db.Where("disks.type = ? and disks.status = ? and disks.capacity >= ?", diskType, rt.DISK_AVAILABLE, capacity).Count(&count)
		if count < int64(require.Count) {
			return nil, status.Errorf(common.TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "expect disk count %d but only %d after disk filter", require.Count, count)
		}

		if !exclusive {
			err = db.Where("hosts.region = ? and hosts.az = ? and hosts.arch = ? and hosts.purpose = ? and hosts.disk_type = ? and hosts.status = ? and (hosts.stat = ? or hosts.stat = ?) and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				region, zone, hostArch, hostPurpose, hostDiskType, rt.HOST_ONLINE, rt.HOST_LOADLESS, rt.HOST_INUSED, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		} else {
			// If need exclusive resource, only choosing from loadless hosts
			err = db.Where("hosts.region = ? and hosts.az = ? and hosts.arch = ? and hosts.purpose = ? and hosts.disk_type = ? and hosts.status = ? and hosts.stat = ? and hosts.free_cpu_cores >= ? and hosts.free_memory >= ?",
				region, zone, hostArch, hostPurpose, hostDiskType, rt.HOST_ONLINE, rt.HOST_LOADLESS, reqCores, reqMem).Group("hosts.id").Scan(&resources).Error
		}
	} else {
		err = nil
	}
	if err != nil {
		return nil, status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "select resources failed, %v", err)
	}

	if len(resources) < int(require.Count) {
		return nil, status.Errorf(common.TIEM_RESOURCE_NO_ENOUGH_HOST, "hosts in %s,%s is not enough for allocation(%d|%d)", region, zone, len(resources), require.Count)
	}

	// 2. Choose Ports in Hosts
	for _, resource := range resources {
		var usedPorts []int32
		tx.Order("port").Model(&rt.UsedPort{}).Select("port").Where("host_id = ?", resource.HostId).Scan(&usedPorts)
		for _, portReq := range require.Require.PortReq {
			res, err := getPortsInRange(usedPorts, portReq.Start, portReq.End, int(portReq.PortCnt))
			if err != nil {
				return nil, status.Errorf(common.TIEM_RESOURCE_NO_ENOUGH_PORT, "host %s(%s) has no enough ports on range [%d, %d]", resource.HostId, resource.Ip, portReq.Start, portReq.End)
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
		default:
			return nil, status.Errorf(common.TIEM_RESOURCE_INVALID_STRATEGY, "invalid alloc strategy %d", require.Strategy)
		}
	}
	return
}

func (m *DAOResourceManager) AllocResources(req *dbpb.DBAllocRequest) (result *rt.AllocRsp, err error) {
	tx := m.getDb().Begin()
	result, err = m.doAlloc(tx, req)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func (m *DAOResourceManager) AllocResourcesInBatch(batchReq *dbpb.DBBatchAllocRequest) (results *rt.BatchAllocResponse, err error) {
	results = new(rt.BatchAllocResponse)
	tx := m.getDb().Begin()
	for i, req := range batchReq.BatchRequests {
		var result *rt.AllocRsp
		result, err = m.doAlloc(tx, req)
		if err != nil {
			tx.Rollback()
			return nil, status.Errorf(common.TIEM_RESOURCE_NOT_ALL_SUCCEED, "alloc resources in batch failed on request %d, %v", i, err)
		}
		results.BatchResults = append(results.BatchResults, result)
	}
	tx.Commit()
	return
}

func recycleUsedTablesByClusterId(tx *gorm.DB, clusterId string) (err error) {
	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedCompute{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedCompute for cluster %s failed, %v", clusterId, err)
	}

	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedDisk{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedDisk for cluster %s failed, %v", clusterId, err)
	}

	err = tx.Where("holder_id = ?", clusterId).Delete(&rt.UsedPort{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedPort for cluster %s failed, %v", clusterId, err)
	}
	return nil
}

func recycleUsedTablesByRequestId(tx *gorm.DB, requestId string) (err error) {
	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedCompute{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedCompute for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedDisk{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedDisk for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&rt.UsedPort{}).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "recycle UsedPort for request %s failed, %v", requestId, err)
	}
	return nil
}

type UsedComputeStatistic struct {
	HostId        string
	TotalCpuCores int
	TotalMemory   int
}

func recycleResourcesInHosts(tx *gorm.DB, usedCompute []UsedComputeStatistic, usedDisks []string) (err error) {
	for _, diskId := range usedDisks {
		var disk rt.Disk
		tx.First(&disk, "ID = ?", diskId).First(&disk)
		if rt.DiskStatus(disk.Status).IsExhaust() {
			err = tx.Model(&disk).Update("Status", int32(rt.DISK_AVAILABLE)).Error
			if err != nil {
				return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "update disk(%s) status while recycle failed, %v", diskId, err)
			}
		} else {
			return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "disk %s status not expected(%d) while recycle", diskId, disk.Status)
		}
	}
	for _, usedCompute := range usedCompute {
		var host rt.Host
		tx.First(&host, "ID = ?", usedCompute.HostId)
		host.FreeCpuCores += int32(usedCompute.TotalCpuCores)
		host.FreeMemory += int32(usedCompute.TotalMemory)

		if host.IsLoadless() {
			host.Stat = rt.HOST_LOADLESS
		} else {
			host.Stat = rt.HOST_INUSED
		}

		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", usedCompute.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "update host(%s) stat while recycle failed, %v", usedCompute.HostId, err)
		}
	}
	return nil
}

func recycleHolderResource(tx *gorm.DB, clusterId string) (err error) {
	var usedCompute []UsedComputeStatistic
	err = tx.Model(&rt.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "get cluster %s total used compute failed, %v", clusterId, err)
	}

	var usedDisks []string
	err = tx.Model(&rt.UsedDisk{}).Select("disk_id").Where("holder_id = ?", clusterId).Scan(&usedDisks).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "get cluster %s total used disks failed, %v", clusterId, err)
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
	var usedCompute []UsedComputeStatistic
	err = tx.Model(&rt.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("request_id = ?", requestId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "get request %s total used compute failed, %v", requestId, err)
	}

	var usedDisks []string
	err = tx.Model(&rt.UsedDisk{}).Select("disk_id").Where("request_id = ?", requestId).Scan(&usedDisks).Error
	if err != nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "get request %s total used disks failed, %v", requestId, err)
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

func (m *DAOResourceManager) doRecycle(tx *gorm.DB, req *dbpb.DBRecycleRequire) (err error) {
	switch rt.RecycleType(req.RecycleType) {
	case rt.RecycleHolder:
		return recycleHolderResource(tx, req.HolderId)
	case rt.RecycleOperate:
		return recycleResourceForRequest(tx, req.RequestId)
	case rt.RecycleCompute:
	case rt.RecycleDisk:
	default:
		return status.Errorf(common.TIEM_RESOURCE_INVAILD_RECYCLE_TYPE, "invalid recycle resource type %d", req.RecycleType)
	}
	return nil
}

func (m *DAOResourceManager) RecycleAllocResources(request *dbpb.DBRecycleRequest) (err error) {
	tx := m.getDb().Begin()
	for i, req := range request.RecycleReqs {
		err = m.doRecycle(tx, req)
		if err != nil {
			tx.Rollback()
			return status.Errorf(common.TIEM_RESOURCE_NOT_ALL_SUCCEED, "recycle resources failed on request %d, %v", i, err)
		}
	}
	tx.Commit()
	return
}
