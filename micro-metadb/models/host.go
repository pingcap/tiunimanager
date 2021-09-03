package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

type HostStatus int32

const (
	HOST_WHATEVER HostStatus = iota - 1
	HOST_ONLINE
	HOST_OFFLINE
	HOST_INUSED
	HOST_EXHAUST
	HOST_DELETED
)

func (s HostStatus) IsInused() bool {
	return s == HOST_INUSED || s == HOST_EXHAUST
}

func (s HostStatus) IsAvailable() bool {
	return (s == HOST_ONLINE || s == HOST_INUSED)
}

type DiskStatus int32

const (
	DISK_AVAILABLE DiskStatus = iota
	DISK_INUSED
)

func (s DiskStatus) IsInused() bool {
	return s == DISK_INUSED
}

func (s DiskStatus) IsAvailable() bool {
	return s == DISK_AVAILABLE
}

type Disk struct {
	ID        string `gorm:"PrimaryKey"`
	HostId    string
	Name      string `gorm:"size:255"`
	Capacity  int32
	Path      string `gorm:"size:255"`
	Status    int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (d Disk) TableName() string {
	return "disks"
}

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}

type Host struct {
	ID        string `gorm:"PrimaryKey"`
	IP        string `gorm:"size:32"`
	UserName  string `gorm:"size:32"`
	Passwd    string `gorm:"size:32"`
	HostName  string `gorm:"size:255"`
	Status    int32  `gorm:"size:32"`
	OS        string `gorm:"size:32"`
	Kernel    string `gorm:"size:32"`
	CpuCores  int
	Memory    int
	Spec      string
	Nic       string `gorm:"size:32"`
	DC        string `gorm:"size:32"`
	AZ        string `gorm:"size:32"`
	Rack      string `gorm:"size:32"`
	Purpose   string `gorm:"size:32"`
	Disks     []Disk
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func HostTableName() string {
	return "hosts"
}

func (h Host) IsExhaust() bool {
	diskExaust := true
	for _, disk := range h.Disks {
		if disk.Status == int32(DISK_AVAILABLE) {
			diskExaust = false
			break
		}
	}
	return diskExaust || h.CpuCores == 0 || h.Memory == 0
}

func (h *Host) SetDiskStatus(diskId string, s DiskStatus) {
	for i := range h.Disks {
		if h.Disks[i].ID == diskId {
			h.Disks[i].Status = int32(s)
			break
		}
	}
}

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("IP = ? and HOST_NAME = ?", h.IP, h.HostName).First(&Host{}).Error
	if err == nil {
		return status.Errorf(codes.AlreadyExists, "host %s(%s) is existed", h.HostName, h.IP)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.ID = uuid.New().String()
		return nil
	} else {
		return err
	}
}

func (h *Host) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("ID = ?", h.ID).First(h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "host %s is not found", h.ID)
		}
	} else {
		if HostStatus(h.Status).IsInused() {
			return status.Errorf(codes.PermissionDenied, "host %s is still in used", h.ID)
		}
	}

	return err
}

func (h *Host) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where("host_id = ?", h.ID).Delete(&Disk{}).Error
	if err != nil {
		return
	}
	h.Status = int32(HOST_DELETED)
	err = tx.Model(&h).Update("Status", h.Status).Error
	return
}

func (h *Host) AfterFind(tx *gorm.DB) (err error) {
	err = tx.Find(&(h.Disks), "HOST_ID = ?", h.ID).Error
	return
}

func (m *DAOResourceManager) CreateHost(host *Host) (id string, err error) {
	err = m.getDb().Create(host).Error
	if err != nil {
		return
	}
	return host.ID, err
}

func (m *DAOResourceManager) CreateHostsInBatch(hosts []*Host) (ids []string, err error) {
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
	err = m.getDb().Where("ID = ?", hostId).Delete(&Host{
		ID: hostId,
	}).Error
	return
}

func (m *DAOResourceManager) DeleteHostsInBatch(hostIds []string) (err error) {
	tx := m.getDb().Begin()
	for _, hostId := range hostIds {
		var host Host
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
	Status  HostStatus
	Purpose string
	Offset  int
	Limit   int
}

func (m *DAOResourceManager) ListHosts(req ListHostReq) (hosts []Host, err error) {
	db := m.getDb().Table(HostTableName())
	if err = db.Error; err != nil {
		return nil, err
	}
	if req.Status != HOST_WHATEVER {
		if req.Status != HOST_DELETED {
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

func (m *DAOResourceManager) FindHostById(hostId string) (*Host, error) {
	host := new(Host)
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
	CpuCores      int
	Memory        int
	Count         int
}

type AllocReqs map[string][]*HostAllocReq
type AllocRsps map[string][]*DiskResource

func getHostsFromFailureDomain(tx *gorm.DB, failureDomain string, numReps int, cpuCores int, mem int) (resources []*DiskResource, err error) {
	err = tx.Order("hosts.cpu_cores desc").Order("hosts.memory desc").Limit(numReps).Model(&Disk{}).Select(
		"disks.host_id, hosts.host_name, hosts.ip, hosts.user_name, hosts.passwd, ? as cpu_cores, ? as memory, disks.id as disk_id, disks.name as disk_name, disks.path, disks.capacity", cpuCores, mem).Joins(
		"left join hosts on disks.host_id = hosts.id").Where(
		"hosts.az = ? and (hosts.status = ? or hosts.status = ?) and hosts.cpu_cores >= ? and memory >= ? and disks.status = ?",
		failureDomain, HOST_ONLINE, HOST_INUSED, cpuCores, mem, DISK_AVAILABLE).Group("hosts.id").Scan(&resources).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "select resources failed, %v", err)
	}

	if len(resources) < numReps {
		return nil, status.Errorf(codes.Internal, "hosts in %s is not enough for allocation(%d|%d)", failureDomain, len(resources), numReps)
	}

	for _, resource := range resources {
		var disk Disk
		tx.First(&disk, "ID = ?", resource.DiskId).First(&disk)
		if DiskStatus(disk.Status).IsAvailable() {
			err = tx.Model(&disk).Update("Status", int32(DISK_INUSED)).Error
			if err != nil {
				return nil, status.Errorf(codes.Internal, "update disk(%s) status err, %v", resource.DiskId, err)
			}
		} else {
			return nil, status.Errorf(codes.FailedPrecondition, "disk %s status not expected(%d)", resource.DiskId, disk.Status)
		}

		var host Host
		tx.First(&host, "ID = ?", resource.HostId)
		host.CpuCores -= cpuCores
		host.Memory -= mem
		if host.IsExhaust() {
			host.Status = int32(HOST_EXHAUST)
		} else {
			host.Status = int32(HOST_INUSED)
		}
		err = tx.Model(&host).Select("CpuCores", "Memory", "Status").Where("id = ?", resource.HostId).Updates(Host{CpuCores: host.CpuCores, Memory: host.Memory, Status: host.Status}).Error
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
	err = m.getDb().Table("hosts").Where("Status = ? or Status = ?", HOST_ONLINE, HOST_INUSED).Select(selectStr).
		Group(domain).Group("purpose").Group("cpu_cores").Group("memory").Scan(&res).Error
	return
}
