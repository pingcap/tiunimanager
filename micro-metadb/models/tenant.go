package models

import (
	"context"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
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

func (e *Tenant) BeforeCreate(*gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	e.Status = TENANT_STATUS_NORMAL
	return nil
}

func (m *DAOAccountManager) AddTenant(ctx context.Context, name string, tenantType, status int8) (t *Tenant, err error) {
	if name == " " || status < TENANT_STATUS_NORMAL || status > TENANT_STATUS_NORMAL {
		return nil, errors.Errorf("add tenant has invalid parameter,name: %s, type: %d, status: %d", name, tenantType, status)
	}
	t = &Tenant{Type: tenantType, Name: name}
	return t, m.Db(ctx).Create(t).Error
}

func (m *DAOAccountManager) FindTenantById(ctx context.Context, tenantId string) (t *Tenant, err error) {
	if tenantId == "" {
		return nil, errors.Errorf("FindTenantByDd has invalid parameter, tenantId: %s", tenantId)
	}
	t = &Tenant{}
	return t, m.Db(ctx).Where("id = ?", tenantId).First(t).Error
}

func (m *DAOAccountManager) FindTenantByName(ctx context.Context, name string) (t *Tenant, err error) {
	if name == "" {
		return nil, errors.Errorf("FindTenantByName has invalid parameter, name: %s", name)
	}
	t = &Tenant{}
	return t, m.Db(ctx).Where("name = ?", name).First(t).Error
}
