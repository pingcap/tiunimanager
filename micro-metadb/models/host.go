package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asim/go-micro/v3/util/log"
	"github.com/google/uuid"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/config"

	"gorm.io/gorm"
)

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

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	h.ID = uuid.New().String()
	return nil
}

func (h *Host) AfterDelete(tx *gorm.DB) (err error) {
	tx.Where("host_id = ?", h.ID).Delete(&Disk{})
	return nil
}

func (h *Host) AfterFind(tx *gorm.DB) (err error) {
	tx.Find(&(h.Disks), "HOST_ID = ?", h.ID)
	return
}

func CreateHostTable() (int32, error) {
	var err error
	var tablebuilt int32 = 0
	dbFile := config.GetSqliteFilePath()
	log := logger.WithContext(context.TODO()).WithField("dbFile", dbFile)
	if MetaDB.Migrator().HasTable(&Host{}) {
		log.Info("Host Table Has Already Created")
	} else {
		err = MetaDB.Migrator().CreateTable(&Host{})
		if err != nil {
			log.Fatalf("sqlite create host table failed: %v", err)
		}
		tablebuilt++
	}
	if MetaDB.Migrator().HasTable(&Disk{}) {
		log.Info("Disk Table Has Already Created")
	} else {
		err = MetaDB.Migrator().CreateTable(&Disk{})
		if err != nil {
			log.Fatalf("sqlite create disk table failed: %v", err)
		}
		tablebuilt++
	}
	return tablebuilt, err
}

func CreateHost(host *Host) (id string, err error) {
	err = MetaDB.Create(host).Error
	if err != nil {
		return
	}
	return host.ID, err
}

// TODO: Check Record before delete
func DeleteHost(hostId string) (err error) {
	MetaDB.Where("ID = ?", hostId).Delete(&Host{
		ID: hostId,
	})
	return nil
}

func ListHosts() (hosts []Host, err error) {
	MetaDB.Find(&hosts)
	return
}

func FindHostById(hostId string) (*Host, error) {
	host := new(Host)
	MetaDB.First(host, "ID = ?", hostId)
	return host, nil
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
	var errmsg string
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
		if host.Status != 0 || host.CpuCores < v.RequestCores || host.Memory < v.RequestMem {
			tx.Rollback()
			errmsg = fmt.Sprintf("No enough resource for Host %s cpu(%d/%d), mem(%d|%d)", v.HostId, host.CpuCores, v.RequestCores, host.Memory, v.RequestMem)
			return errors.New(errmsg)
		}
		if host.CpuCores+setUpdate[v.HostId].cpuCores == v.OriginCores && host.Memory+setUpdate[v.HostId].mem == v.OriginMem {
			host.CpuCores -= v.RequestCores
			host.Memory -= v.RequestMem
			if host.CpuCores == 0 || host.Memory == 0 {
				host.Status = 2
			}
		} else {
			tx.Rollback()
			errmsg = fmt.Sprintf("Host %s was changed concurrently, cpu(%d|%d), mem(%d|%d)", v.HostId, host.CpuCores+setUpdate[v.HostId].cpuCores, v.OriginCores, host.Memory+setUpdate[v.HostId].mem, v.OriginMem)
			return errors.New(errmsg)
		}
		var disk Disk
		MetaDB.First(&disk, "ID = ?", v.DiskId).First(&disk)
		if disk.Status == 0 {
			disk.Status = 2
		} else {
			tx.Rollback()
			errmsg = fmt.Sprintf("Disk %s Status not expected(%d)", v.DiskId, disk.Status)
			return errors.New(errmsg)
		}
		setUpdate[v.HostId].cpuCores += v.RequestCores
		setUpdate[v.HostId].mem += v.RequestMem
		tx.Save(&host)
		tx.Save(&disk)
	}
	tx.Commit()
	return nil
}
