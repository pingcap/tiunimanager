package resource

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Holder struct {
	HolderId  string `gorm:"index"` // who(clusterId) hold the resource
	RequestId string `gorm:"index"` // the resource is allocated in which request
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UsedCompute struct {
	Holder
	ID       string `gorm:"PrimaryKey"`
	HostId   string `gorm:"index;not null"`
	CpuCores int32
	Memory   int32
}

func (d *UsedCompute) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}

type UsedPort struct {
	Holder
	ID     string `gorm:"PrimaryKey"`
	HostId string `gorm:"index;not null"`
	Port   int32
}

func (d *UsedPort) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}

type UsedDisk struct {
	Holder
	ID       string `gorm:"PrimaryKey"`
	HostId   string `gorm:"not null"`
	DiskId   string `gorm:"not null"`
	Capacity int32
}

func (d *UsedDisk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New().String()
	return nil
}
