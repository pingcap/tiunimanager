package models

import (
	"context"
	"time"

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

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}

type Host struct {
	ID        string `gorm:"PrimaryKey"`
	IP        string `gorm:"size:32"`
	Name      string `gorm:"size:255"`
	Status    int32  `gorm:"size:32"`
	OS        string `gorm:"size:32"`
	Kernel    string `gorm:"size:32"`
	CpuCores  int
	Memory    int
	Nic       string `gorm:"size:32"`
	AZ        string `gorm:"size:32"`
	Rack      string `gorm:"size:32"`
	Purpose   string `gorm:"size:32"`
	Disks     []Disk
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	h.ID = uuid.New().String()
	return nil
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
	MetaDB.Where("UUID = ?", hostId).Delete(&Host{})
	return nil
}

func ListHosts() (hosts []Host, err error) {
	MetaDB.Find(&hosts)
	return
}

func FindHostById(hostId string) (*Host, error) {
	host := new(Host)
	MetaDB.First(host, hostId)

	return host, nil
}

// TODO: Just a trick demo function
func AllocHosts() (hosts []Host, err error) {
	MetaDB.Find(&hosts)
	return hosts, err
}
