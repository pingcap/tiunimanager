package resource

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiskType string

const (
	NvmeSSD DiskType = "nvme_ssd"
	SSD     DiskType = "ssd"
	Sata    DiskType = "sata"
)

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
	ID        string         `json:"diskId" gorm:"PrimaryKey"`
	HostId    string         `json:"omitempty"`
	Name      string         `json:"name" gorm:"size:255"`          // [sda/sdb/nvmep0...]
	Capacity  int32          `json:"capacity"`                      // Disk size, Unit: GB
	Path      string         `json:"path" gorm:"size:255;not null"` // Disk mount path: [/data1]
	Type      string         `json:"type"`                          // Disk type: [nvme-ssd/ssd/sata]
	Status    int32          `json:"status" gorm:"index"`           // Disk Status, 0 for available, 1 for inused
	UsedBy    string         `json:"usedby" gorm:"index"`           // Disk is used by which cluster
	RequestId string         `json:"-" gorm:"index"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (d Disk) TableName() string {
	return "disks"
}

func (d *Disk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}
