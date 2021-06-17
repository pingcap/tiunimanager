package service

import (
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model

	Name   string		`gorm:"size:255"`
	Id     uint			`gorm:"size:255"`
	Type   int8			`gorm:"size:255"`
	Status int8			`gorm:"size:255"`
}

func AddTenant(name string, tenantType, status int8) (tenant Tenant, err error) {
	tenant.Status = tenantType
	tenant.Type = tenantType
	tenant.Name = name

	DB.Create(&tenant)
	// 返回ID
	return
}

func FetchTenant(tenantId int) (tenant Tenant, err error){
	DB.First(&tenant, tenantId)
	return
}