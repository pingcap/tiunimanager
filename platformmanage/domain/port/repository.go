package port

import (
	"github.com/pingcap/ticp/platformmanage/domain/entity"
)

var RbacRepo RbacRepository

var TenantRepo TenantRepository

var ResourceRepo ResourceRepository

type TenantRepository interface {

	AddTenant(*entity.Tenant) error

	FetchTenant(name string)  (entity.Tenant, error)

}

type ResourceRepository interface {
	// List 查询主机信息
	List() ([]entity.Host, error)

	// Preempt 查询一批可用机器供业务使用
	Preempt(count int) ([]entity.Host, error)

	// Allocate 将一批主机(或资源)分配给一个集群
	Allocate(host []entity.Host) error
}

type RbacRepository interface {

	AddAccount(a *entity.Account) error

	FetchAccount(tenantId int, name string) (entity.Account, error)

	AddRole(r *entity.Role) error

	FetchRole(tenantId int, name string) (entity.Role, error)

	AddPermission(r *entity.Permission) error

	FetchPermission(tenantId int, code string) (entity.Permission, error)

	FetchAllRolesByAccount(account *entity.Account) ([]entity.Role, error)

	FetchAllRolesByPermission(permission *entity.Permission) ([]entity.Role, error)

	AddPermissionBindings(bindings []entity.PermissionBinding) error

	AddRoleBindings(bindings []entity.RoleBinding) error

}
