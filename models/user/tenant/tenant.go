package tenant

import (
	"gorm.io/gorm"
	"time"
)

type Tenant struct {
	ID        string         `gorm:"PrimaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Name   string            `gorm:"default:null;not null"`
	Type   int8              `gorm:"default:0"`
	Status int8              `gorm:"default:0"`
}

