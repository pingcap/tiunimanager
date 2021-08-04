package models

import (
	"errors"
	"fmt"
	"github.com/pingcap/ticp/library/thirdparty/logger"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type HostStatus int32

const (
	HOST_ONLINE HostStatus = iota
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

func (h Host) TableName() string {
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
	return
}

func (h *Host) AfterFind(tx *gorm.DB) (err error) {
	err = tx.Find(&(h.Disks), "HOST_ID = ?", h.ID).Error
	return
}

func CreateHost(host *Host) (id string, err error) {
	err = MetaDB.Create(host).Error
	if err != nil {
		return
	}
	return host.ID, err
}

func CreateHostsInBatch(hosts []*Host) (ids []string, err error) {
	tx := MetaDB.Begin()
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

func DeleteHost(hostId string) (err error) {
	err = MetaDB.Where("ID = ?", hostId).Delete(&Host{
		ID: hostId,
	}).Error
	return
}

func DeleteHostsInBatch(hostIds []string) (err error) {
	tx := MetaDB.Begin()
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

func ListHosts() (hosts []Host, err error) {
	err = MetaDB.Find(&hosts).Error
	return
}

func FindHostById(hostId string) (*Host, error) {
	host := new(Host)
	err := MetaDB.First(host, "ID = ?", hostId).Error
	return host, err
}

// TODO: Just a trick demo function
type Resource struct {
	HostId   string
	Id       string
	CpuCores int
	Memory   int
	HostName string
	Ip       string
	Name     string
	Path     string
	Capacity int
}

func PreAllocHosts(failedDomain string, numReps int, cpuCores int, mem int) (resources []Resource, err error) {
	err = MetaDB.Order("hosts.cpu_cores").Limit(numReps).Model(&Disk{}).Select(
		"disks.host_id, disks.id, hosts.cpu_cores, hosts.memory, hosts.host_name, hosts.ip, disks.name, disks.path, disks.capacity").Joins("left join hosts on disks.host_id = hosts.id").Where(
		"disks.status = ? and hosts.az = ? and (hosts.status = ? or hosts.status = ?) and hosts.cpu_cores >= ? and memory >= ?",
		DISK_AVAILABLE, failedDomain, HOST_ONLINE, HOST_INUSED, cpuCores, mem).Group("hosts.id").Scan(&resources).Error
	return
}

type ResourceLock struct {
	HostId       string
	OriginCores  int
	OriginMem    int
	RequestCores int
	RequestMem   int
	DiskId       string
}

type HostLocked struct {
	cpuCores int
	mem      int
}

func LockHosts(resources []ResourceLock) (err error) {
	var setUpdate map[string]*HostLocked = make(map[string]*HostLocked)
	tx := MetaDB.Begin()
	for _, v := range resources {
		var host Host
		MetaDB.First(&host, "ID = ?", v.HostId)
		if _, ok := setUpdate[v.HostId]; !ok {
			if err = tx.Set("gorm:query_option", "FOR UPDATE").First(&host).Error; err != nil {
				tx.Rollback()
				logger.GetLogger().Errorf("set for update host %s failed", v.HostId)
				return err
			}
			setUpdate[v.HostId] = &HostLocked{
				cpuCores: 0,
				mem:      0,
			}
		}
		if !HostStatus(host.Status).IsAvailable() || host.CpuCores < v.RequestCores || host.Memory < v.RequestMem {
			tx.Rollback()
			return status.Errorf(codes.ResourceExhausted,
				"no enough resource for host %s(status:%d) cpu(%d/%d), mem(%d|%d)",
				v.HostId, host.Status, host.CpuCores, v.RequestCores, host.Memory, v.RequestMem)
		}
		if host.CpuCores+setUpdate[v.HostId].cpuCores == v.OriginCores && host.Memory+setUpdate[v.HostId].mem == v.OriginMem {
			host.CpuCores -= v.RequestCores
			host.Memory -= v.RequestMem
		} else {
			tx.Rollback()
			return status.Errorf(codes.FailedPrecondition,
				"host %s was changed concurrently, cpu(%d|%d), mem(%d|%d)",
				v.HostId, host.CpuCores+setUpdate[v.HostId].cpuCores, v.OriginCores, host.Memory+setUpdate[v.HostId].mem, v.OriginMem)
		}
		var disk Disk
		MetaDB.First(&disk, "ID = ?", v.DiskId).First(&disk)
		if DiskStatus(disk.Status).IsAvailable() {
			disk.Status = int32(DISK_INUSED)
			// Set Disk Status in host - for exaust judgement
			host.SetDiskStatus(v.DiskId, DISK_INUSED)
		} else {
			tx.Rollback()
			return status.Errorf(codes.FailedPrecondition,
				"disk %s status not expected(%d)", v.DiskId, disk.Status)
		}
		setUpdate[v.HostId].cpuCores += v.RequestCores
		setUpdate[v.HostId].mem += v.RequestMem
		if host.IsExhaust() {
			host.Status = int32(HOST_EXHAUST)
		} else {
			host.Status = int32(HOST_INUSED)
		}
		tx.Model(&host).Select("CpuCores", "Memory", "Status").Updates(Host{CpuCores: host.CpuCores, Memory: host.Memory, Status: host.Status})
		tx.Model(&disk).Update("Status", disk.Status)
	}
	tx.Commit()
	return nil
}

type FailureDomainResource struct {
	FailureDomain string
	Purpose       string
	CpuCores      int
	Memory        int
	Count         int
}

func GetFailureDomain(domain string) (res []FailureDomainResource, err error) {
	selectStr := fmt.Sprintf("%s as FailureDomain, purpose, cpu_cores, memory, count(id) as Count", domain)
	err = MetaDB.Table("hosts").Where("Status = ? or Status = ?", HOST_ONLINE, HOST_INUSED).Select(selectStr).
		Group(domain).Group("purpose").Group("cpu_cores").Group("memory").Scan(&res).Error
	return
}
