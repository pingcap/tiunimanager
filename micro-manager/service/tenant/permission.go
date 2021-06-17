package tenant

import (
	"fmt"
	port2 "github.com/pingcap/ticp/micro-manager/service/tenant/port"
)

type Permission struct {
	Tenant *Tenant
	Code   string
	Name   string
	Type   PermissionType
	Desc   string
	Status CommonStatus
}

type PermissionType int

const (
	Path PermissionType = 1
	Act  PermissionType = 2
	Data PermissionType = 3
)

func (permission *Permission) persist() error{
	port2.RbacRepo.AddPermission(permission)
	return nil
}

func CreatePermission(tenant *Tenant, code, name ,desc string, permissionType PermissionType) (*Permission, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := FindPermissionByCode(tenant, code)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("permission already exist")
	}

	permission := Permission{
		Tenant: tenant,
		Code:   code,
		Name:   name,
		Type:   permissionType,
		Desc:   desc,
		Status: Valid,
	}

	permission.persist()

	return &permission, nil
}

func FindPermissionByCode(tenant *Tenant, code string) (*Permission, error) {
	a,e := port2.RbacRepo.FetchPermission(tenant.Id, code)
	return &a, e
}

func (permission *Permission) ListAllRoles() ([]Role, error){
	return port2.RbacRepo.FetchAllRolesByPermission(permission)
}