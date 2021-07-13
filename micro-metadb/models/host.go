package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

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
	Name      string `gorm:"size:255"`
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
	err = tx.Where("IP = ? and Name = ?", h.IP, h.Name).First(&Host{}).Error
	if err == nil {
		errMsg := fmt.Sprintf("Host %s(%s) is Existed", h.Name, h.IP)
		return errors.New(errMsg)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.ID = uuid.New().String()
		return nil
	} else {
		return err
	}
}

func (h *Host) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("ID = ?", h.ID).First(&Host{}).Error
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

// TODO: Check Record before delete
func DeleteHost(hostId string) (err error) {
	err = MetaDB.Where("ID = ?", hostId).Delete(&Host{
		ID: hostId,
	}).Error
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
func AllocHosts() (hosts []Host, err error) {
	MetaDB.Find(&hosts)
	return
}
