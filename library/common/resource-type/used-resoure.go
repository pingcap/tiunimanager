package resource

import (
	"errors"
	"time"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"google.golang.org/grpc/status"
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
	d.ID = uuidutil.GenerateID()
	return nil
}

type UsedPort struct {
	Holder
	ID     string `gorm:"PrimaryKey"`
	HostId string `gorm:"index;not null"`
	Port   int32
}

func (d *UsedPort) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("host_id = ? and port = ?", d.HostId, d.Port).First(&UsedPort{}).Error
	if err == nil {
		return status.Errorf(common.TIEM_RESOURCE_SQL_ERROR, "port %d in host(%s) is already inused", d.Port, d.HostId)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		d.ID = uuidutil.GenerateID()
		return nil
	} else {
		return err
	}
}

type UsedDisk struct {
	Holder
	ID       string `gorm:"PrimaryKey"`
	HostId   string `gorm:"not null"`
	DiskId   string `gorm:"not null"`
	Capacity int32
}

func (d *UsedDisk) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuidutil.GenerateID()
	return nil
}
