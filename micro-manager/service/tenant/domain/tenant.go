package domain

import (
	"fmt"
	port2 "github.com/pingcap/ticp/micro-manager/service/tenant/port"
)

// Tenant 租户
type Tenant struct {
	Name   string
	Id     uint
	Type   TenantType
	Status CommonStatus
}

func (t *Tenant) persist() error{
	return nil
}

// CreateTenant 创建租户
func CreateTenant(name string) (*Tenant, error) {
	existed, e := FindTenant(name)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("tenant already exist")
	}

	tenant := Tenant{Name: name, Type: InstanceWorkspace, Status: Valid}
	tenant.persist()

	return &tenant, nil
}

// FindTenant 查找租户
func FindTenant(name string) (*Tenant, error) {
	tenant,err := port2.TenantRepo.FetchTenantByName(name)
	return &tenant, err
}

func FindTenantById(id uint) (*Tenant, error) {
	tenant,err := port2.TenantRepo.FetchTenantById(id)
	return &tenant, err
}
type TenantType int

const (
	SystemManagement  TenantType = 0
	InstanceWorkspace TenantType = 1
	PluginAccess      TenantType = 2
)

type CommonStatus int

const (
	Valid   CommonStatus = 0
	Invalid CommonStatus = 1
	Deleted CommonStatus = 2
)

func (s CommonStatus) IsValid() bool {
	return s == Valid
}