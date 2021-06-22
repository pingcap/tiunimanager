package domain

import (
	"fmt"
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
	RbacRepo.AddPermission(permission)
	return nil
}

func createPermission(tenant *Tenant, code, name ,desc string, permissionType PermissionType) (*Permission, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := findPermissionByCode(tenant.Id, code)

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

func findPermissionByCode(tenantId uint, code string) (*Permission, error) {
	a,e := RbacRepo.FetchPermission(tenantId, code)
	return &a, e
}

func (permission *Permission) listAllRoles() ([]Role, error){
	return RbacRepo.FetchAllRolesByPermission(permission)
}