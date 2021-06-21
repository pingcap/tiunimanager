package port

import (
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
)

var RbacRepo RbacRepository

var TenantRepo TenantRepository

type TenantRepository interface {
	AddTenant(*domain.Tenant) error

	FetchTenantByName(name string)  (domain.Tenant, error)

	FetchTenantById(id uint)  (domain.Tenant, error)
}

type RbacRepository interface {

	AddAccount(a *domain.Account) error

	FetchAccountByName(name string) (domain.Account, error)

	FetchAccountById(id uint) (domain.Account, error)

	AddRole(r *domain.Role) error

	FetchRole(tenantId uint, name string) (domain.Role, error)

	AddPermission(r *domain.Permission) error

	FetchPermission(tenantId uint, code string) (domain.Permission, error)

	FetchAllRolesByAccount(account *domain.Account) ([]domain.Role, error)

	FetchAllRolesByPermission(permission *domain.Permission) ([]domain.Role, error)

	AddPermissionBindings(bindings []domain.PermissionBinding) error

	AddRoleBindings(bindings []domain.RoleBinding) error
}
