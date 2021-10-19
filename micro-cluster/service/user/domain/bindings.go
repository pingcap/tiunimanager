package domain

type Role struct {
	TenantId string
	Id       string
	Name     string
	Desc     string
	Status   CommonStatus
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
