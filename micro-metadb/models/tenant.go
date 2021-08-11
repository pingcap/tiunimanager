package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Tenant struct {
	ID        string `gorm:"PrimaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Name   string `gorm:"default:null;not null"`
	Type   int8   `gorm:"default:0"`
	Status int8   `gorm:"default:0"`
}

func (e *Tenant) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = GenerateID()
	e.Status = 0
	return nil
}

func AddTenant(name string, tenantType, status int8) (tenant Tenant, err error) {
	if name == "" {
		err = errors.New("tenant name empty")
	}
	tenant.Type = tenantType
	tenant.Name = name
	MetaDB.Create(&tenant)
	// 返回ID
	return
}

func FindTenantById(tenantId string) (tenant Tenant, err error) {
	err = MetaDB.Where("id = ?", tenantId).First(&tenant).Error
	return
}

func FindTenantByName(name string) (tenant Tenant, err error) {
	err = MetaDB.Where("name = ?", name).First(&tenant).Error
	return
}
