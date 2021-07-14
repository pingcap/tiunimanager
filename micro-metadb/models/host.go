package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return status.Errorf(codes.AlreadyExists, "Host %s(%s) is Existed", h.Name, h.IP)
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Errorf(codes.NotFound, "HostId %s is not found", h.ID)
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
			return nil, status.Errorf(codes.Canceled, "Create %s(%s) err, %v", host.Name, host.IP, err)
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
func AllocHosts() (hosts []Host, err error) {
	MetaDB.Find(&hosts)
	return
}
