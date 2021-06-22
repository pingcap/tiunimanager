package domain

var RbacRepo RbacRepository

var TenantRepo TenantRepository

var TokenMNG TokenManager

type TenantRepository interface {
	AddTenant(*Tenant) error

	FetchTenantByName(name string)  (Tenant, error)

	FetchTenantById(id uint)  (Tenant, error)
}

type RbacRepository interface {

	AddAccount(a *Account) error

	FetchAccountByName(name string) (Account, error)

	FetchAccountById(id uint) (Account, error)

	AddRole(r *Role) error

	FetchRole(tenantId uint, name string) (Role, error)

	AddPermission(r *Permission) error

	FetchPermission(tenantId uint, code string) (Permission, error)

	FetchAllRolesByAccount(account *Account) ([]Role, error)

	FetchAllRolesByPermission(permission *Permission) ([]Role, error)

	AddPermissionBindings(bindings []PermissionBinding) error

	AddRoleBindings(bindings []RoleBinding) error
}

type TokenManager interface {

	// Provide 提供一个有效的token
	Provide  (tiCPToken *TiCPToken) (string, error)

	// Modify 修改token
	Modify (tiCPToken *TiCPToken) error

	// GetToken 获取一个token
	GetToken(tokenString string) (TiCPToken, error)
}
