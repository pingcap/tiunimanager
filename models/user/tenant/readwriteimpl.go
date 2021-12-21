package tenant

import (
	"context"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
)

//todo: where to definite
const (
	TENANT_STATUS_NORMAL int8 = iota
	TENANT_STATUS_DEACTIVATE
)

type TenantReadWrite struct {
	dbCommon.GormDB
}

func (g *TenantReadWrite) AddTenant(ctx context.Context, name string, tenantType, status int8) (*Tenant, error) {
	if name == " " || status < TENANT_STATUS_NORMAL || status > TENANT_STATUS_NORMAL {
		return nil, errors.Errorf("add tenant has invalid parameter,name: %s, type: %d, status: %d", name, tenantType, status)
	}
	t := &Tenant{Type: tenantType, Name: name}
	return t, g.DB(ctx).Create(t).Error
}

func (g *TenantReadWrite) FindTenantByName(ctx context.Context, name string) (*Tenant, error) {
	if name == "" {
		return nil, errors.Errorf("FindTenantByName has invalid parameter, tenantId: %s", name)
	}
	t := &Tenant{}
	return t, g.DB(ctx).Where("name = ?", name).First(t).Error //todo: name? TenantName
}

func (g *TenantReadWrite) FindTenantById(ctx context.Context, tenantId string) (*Tenant, error) {
	if tenantId == "" {
		return nil, errors.Errorf("FindTenantById has invalid parameter, tenantId: %s", tenantId)
	}
	t := &Tenant{}
	return t, g.DB(ctx).Where("id = ?", tenantId).First(t).Error
}

func NewTenantReadWrite(db *gorm.DB) *TenantReadWrite {
	return &TenantReadWrite{
		dbCommon.WrapDB(db),
	}
}
