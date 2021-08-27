package models

import (
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

const (
	TENANT_STATUS_NORMAL int8 = iota
	TENANT_STATUS_DEACTIVATE
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

func (e *Tenant) BeforeCreate( *gorm.DB) (err error) {
	e.ID = GenerateID()
	e.Status = TENANT_STATUS_NORMAL
	return nil
}

func ( *Tenant) AddTenant(db *gorm.DB, name string, tenantType, status int8) (t *Tenant, err error) {
	if nil == db || name == " " || status < TENANT_STATUS_NORMAL || status > TENANT_STATUS_NORMAL {
		return nil , errors.Errorf("add tenant has invalid parameter,name: %s, type: %d, status: %d", name, tenantType, status)
	}
	t= &Tenant { Type: tenantType, Name: name }
	return t,db.Create(t).Error
}

func ( *Tenant) FindTenantById(db *gorm.DB,tenantId string) (t *Tenant, err error) {
	if nil == db || tenantId == "" {
		return nil , errors.Errorf("FindTenantByDd has invalid parameter, tenantId: %s", tenantId)
	}
	t = &Tenant{}
	return t, db.Where("id = ?", tenantId).First(t).Error
}

func ( *Tenant) FindTenantByName(db *gorm.DB, name string) (t *Tenant, err error) {
	if nil == db || name == "" {
		return nil, errors.Errorf("FindTenantByName has invalid parameter, name: %s", name)
	}
	t = &Tenant{}
	return t, db.Where("name = ?", name).First(t).Error
}