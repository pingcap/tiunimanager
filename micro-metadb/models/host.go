package models

import (
	"errors"
	"time"

	"github.com/asim/go-micro/v3/util/log"
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
	return s == HOST_INUSED
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
	return h.CpuCores == 0 || h.Memory == 0
}

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("IP = ? and Name = ?", h.IP, h.HostName).First(&Host{}).Error
	if err == nil {
		return status.Errorf(codes.AlreadyExists, "Host %s(%s) is Existed", h.HostName, h.IP)
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
			return status.Errorf(codes.NotFound, "HostId %s is not found", h.ID)
		}
	} else {
		if HostStatus(h.Status).IsInused() {
			return status.Errorf(codes.PermissionDenied, "HostId %s is still in used", h.ID)
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
			return nil, status.Errorf(codes.Canceled, "Create %s(%s) err, %v", host.HostName, host.IP, err)
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
			return status.Errorf(codes.FailedPrecondition, "Lock Host %s(%s) error, %v", hostId, host.IP, err)
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
		"disks.status = ? and hosts.az = ? and hosts.status = ? and hosts.cpu_cores >= ? and memory >= ?",
		0, failedDomain, 0, cpuCores, mem).Group("hosts.id").Scan(&resources).Error
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
				log.Errorf("SET FOR UPDATE host %s failed", v.HostId)
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
				"No enough resource for Host %s(status:%d) cpu(%d/%d), mem(%d|%d)",
				v.HostId, host.Status, host.CpuCores, v.RequestCores, host.Memory, v.RequestMem)
		}
		if host.CpuCores+setUpdate[v.HostId].cpuCores == v.OriginCores && host.Memory+setUpdate[v.HostId].mem == v.OriginMem {
			host.CpuCores -= v.RequestCores
			host.Memory -= v.RequestMem
			if host.IsExhaust() {
				host.Status = int32(HOST_EXHAUST)
			}
		} else {
			tx.Rollback()
			return status.Errorf(codes.FailedPrecondition,
				"Host %s was changed concurrently, cpu(%d|%d), mem(%d|%d)",
				v.HostId, host.CpuCores+setUpdate[v.HostId].cpuCores, v.OriginCores, host.Memory+setUpdate[v.HostId].mem, v.OriginMem)
		}
		var disk Disk
		MetaDB.First(&disk, "ID = ?", v.DiskId).First(&disk)
		if DiskStatus(disk.Status).IsAvailable() {
			disk.Status = int32(DISK_INUSED)
		} else {
			tx.Rollback()
			return status.Errorf(codes.FailedPrecondition,
				"Disk %s Status not expected(%d)", v.DiskId, disk.Status)
		}
		setUpdate[v.HostId].cpuCores += v.RequestCores
		setUpdate[v.HostId].mem += v.RequestMem
		tx.Save(&host)
		tx.Save(&disk)
	}
	tx.Commit()
	return nil
}
