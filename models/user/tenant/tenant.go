package tenant

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/gorm"
	"time"
)

type Tenant struct {
	ID        string              `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Name   string                 `gorm:"default:null;not null;uniqueIndex"`
	Type   constants.TenantType   `gorm:"default:0"`
	Status constants.CommonStatus `gorm:"default:0"`
}

func (e *Tenant) BeforeCreate(*gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	e.Status = 0
	return nil
}
