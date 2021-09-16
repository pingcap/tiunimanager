package models

import (
	"fmt"

	rt "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"

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
		"hosts.az = ? and (hosts.status = ? or hosts.status = ?) and hosts.free_cpu_cores >= ? and free_memory >= ? and disks.status = ?",
		failureDomain, rt.HOST_ONLINE, rt.HOST_INUSED, cpuCores, mem, rt.DISK_AVAILABLE).Group("hosts.id").Scan(&resources).Error
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
			host.Status = int32(rt.HOST_EXHAUST)
		} else {
			host.Status = int32(rt.HOST_INUSED)
		}
		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Status").Where("id = ?", resource.HostId).Updates(rt.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Status: host.Status}).Error
		if err != nil {
			return nil, status.Errorf(codes.Internal, "update host(%s) status err, %v", resource.HostId, err)
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
	err = m.getDb().Table("hosts").Where("Status = ? or Status = ?", rt.HOST_ONLINE, rt.HOST_INUSED).Select(selectStr).
		Group(domain).Group("purpose").Group("cpu_cores").Group("memory").Scan(&res).Error
	return
}
