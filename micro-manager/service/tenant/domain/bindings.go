package domain

import (
	"fmt"
)

type Role struct {
	TenantId 	uint
	Id     		int
	Name   		string
	Desc   		string
	Status 		CommonStatus
}

func createRole(tenant *Tenant, name string, desc string) (*Role, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := findRoleByName(tenant, name)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("role already exist")
	}

	role := Role{TenantId: tenant.Id, Name: name, Desc: desc, Status: Valid}

	role.persist()
	return &role, nil

}

func (role *Role) persist() error{
	RbacRepo.AddRole(role)
	return nil
}

func findRoleByName(tenant *Tenant, name string) (*Role, error) {
	r,e := RbacRepo.LoadRole(tenant.Id, name)
	return &r, e
}

type PermissionBinding struct {
	Role       *Role
	Permission *Permission
	Status     CommonStatus
}

type RoleBinding struct {
	Role    *Role
	Account *Account
	Status  CommonStatus
}

func (role *Role) empower(permissions []Permission) error {
	bindings := make([]PermissionBinding, len(permissions), len(permissions))

	for index,r := range permissions {
		bindings[index] = PermissionBinding{Role: role, Permission: &r, Status: Valid}
	}
	return RbacRepo.AddPermissionBindings(bindings)
}
