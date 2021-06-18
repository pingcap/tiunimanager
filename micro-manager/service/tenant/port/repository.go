package port

import (
	"github.com/pingcap/ticp/micro-manager/service/tenant"
)

var RbacRepo RbacRepository

var TenantRepo TenantRepository

type TenantRepository interface {
	AddTenant(*tenant.Tenant) error

	FetchTenantByName(name string)  (tenant.Tenant, error)

	FetchTenantById(id uint)  (tenant.Tenant, error)
}

type RbacRepository interface {

	AddAccount(a *tenant.Account) error

	FetchAccountByName(name string) (tenant.Account, error)

	FetchAccountById(id uint) (tenant.Account, error)

	AddRole(r *tenant.Role) error

	FetchRole(tenantId uint, name string) (tenant.Role, error)

	AddPermission(r *tenant.Permission) error

	FetchPermission(tenantId uint, code string) (tenant.Permission, error)

	FetchAllRolesByAccount(account *tenant.Account) ([]tenant.Role, error)

	FetchAllRolesByPermission(permission *tenant.Permission) ([]tenant.Role, error)

	AddPermissionBindings(bindings []tenant.PermissionBinding) error

	AddRoleBindings(bindings []tenant.RoleBinding) error
}
