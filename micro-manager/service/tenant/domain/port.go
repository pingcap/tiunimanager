package domain

var RbacRepo RbacRepository

var TenantRepo TenantRepository

var TokenMNG TokenManager

type TenantRepository interface {
	AddTenant(*Tenant) error

	LoadTenantByName(name string)  (Tenant, error)

	LoadTenantById(id uint)  (Tenant, error)
}

type RbacRepository interface {

	AddAccount(a *Account) error

	LoadAccountByName(name string) (Account, error)

	LoadAccountAggregation(name string) (AccountAggregation, error)

	LoadAccountById(id uint) (Account, error)

	AddRole(r *Role) error

	LoadRole(tenantId uint, name string) (Role, error)

	AddPermission(r *Permission) error

	LoadPermissionAggregation(tenantId uint, code string) (PermissionAggregation, error)

	LoadPermission(tenantId uint, code string) (Permission, error)

	LoadAllRolesByAccount(account *Account) ([]Role, error)

	LoadAllRolesByPermission(permission *Permission) ([]Role, error)

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
