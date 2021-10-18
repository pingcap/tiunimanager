package application

import (
	"fmt"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/ports"
)

type TenantManager struct {
	tenantRepo ports.TenantRepository
}

func NewTenantManager(tenantRepo ports.TenantRepository) *TenantManager {
	return &TenantManager{tenantRepo : tenantRepo}
}

// CreateTenant 创建租户
func (p *TenantManager) CreateTenant(name string) (*domain.Tenant, error) {
	existed, e := p.FindTenant(name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("tenant already exist")
	}

	tenant := domain.Tenant{Name: name, Type: domain.InstanceWorkspace, Status: domain.Valid}

	return &tenant, nil
}

// FindTenant 查找租户
func (p *TenantManager) FindTenant(name string) (*domain.Tenant, error) {
	tenant, err := p.tenantRepo.LoadTenantByName(name)
	return &tenant, err
}

func (p *TenantManager) FindTenantById(id string) (*domain.Tenant, error) {
	tenant, err := p.tenantRepo.LoadTenantById(id)
	return &tenant, err
}
