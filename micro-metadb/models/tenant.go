package models

import (
	"time"

	"gorm.io/gorm"
)

type Tenant struct {
	ID        string `gorm:"PrimaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Name   string `gorm:"size:255"`
	Type   int8   `gorm:"size:255"`
	Status int8   `gorm:"size:255"`
}

func AddTenant(name string, tenantType, status int8) (tenant Tenant, err error) {
	tenant.Status = tenantType
	tenant.Type = tenantType
	tenant.Name = name
	tenant.ID = "d5624ef9-43b6-411f-b06a-01422bdfb0e5"
	MetaDB.Create(&tenant)
	// 返回ID
	return
}

func FindTenantById(tenantId string) (tenant Tenant, err error) {
	MetaDB.First(&tenant, tenantId)
	return
}

func FindTenantByName(name string) (tenant Tenant, err error) {
	MetaDB.Where("name = ?", name).First(&tenant)
	return
}
