package entity

import (
	"fmt"
	"github.com/pingcap/ticp/platformmanage/domain/port"
)

type Role struct {
	Tenant 		*Tenant

	Id      	int
	Name    	string
	Desc    	string
	Status 		CommonStatus
}

func CreateRole(tenant *Tenant, name string, desc string) (*Role, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := FindRoleByName(tenant, name)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("role already exist")
	}

	role := Role{Tenant: tenant, Name: name, Desc: desc, Status: Valid}

	role.persist()
	return &role, nil

}

func (role *Role) persist() error{
	port.RbacRepo.AddRole(role)
	return nil
}

func FindRoleByName(tenant *Tenant, name string) (*Role, error) {
	r,e := port.RbacRepo.FetchRole(tenant.Id, name)
	return &r, e
}

type PermissionBinding struct {
	Role 			*Role
	Permission	 	*Permission
	Status 			CommonStatus
}

type RoleBinding struct {
	Role 			*Role
	Account	 	    *Account
	Status 			CommonStatus
}

// Empower 给一个角色赋予权限
func (role *Role) Empower(permissions []Permission) error {
	bindings := make([]PermissionBinding, len(permissions), len(permissions))

	for index,r := range permissions {
		bindings[index] = PermissionBinding{Role: role, Permission: &r, Status: Valid}
	}
	return port.RbacRepo.AddPermissionBindings(bindings)
}
