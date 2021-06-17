package service

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model

	TenantId 		uint		`gorm:"size:255"`
	Name 			string		`gorm:"size:255"`
	Salt 			string		`gorm:"size:255"`
	FinalHash 		string		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
}

type Role struct {
	gorm.Model

	TenantId 		uint		`gorm:"size:255"`
	Name    		string		`gorm:"size:255"`
	Desc    		string		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
}

type PermissionBinding struct {
	gorm.Model

	TenantId 		uint		`gorm:"size:255"`
	RoleId 			uint		`gorm:"size:255"`
	PermissionId	uint		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
}

type RoleBinding struct {
	gorm.Model

	TenantId 		uint		`gorm:"size:255"`
	RoleId 			uint		`gorm:"size:255"`
	AccountId	 	uint		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
}

type Permission struct {
	TenantId 		uint		`gorm:"size:255"`
	Code   			string		`gorm:"size:255"`
	Name  	 		string		`gorm:"size:255"`
	Type   			int8		`gorm:"size:255"`
	Desc   			string		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
}

func AddAccount(tenantId uint, name string, salt string, finalHash string, status int8) (result Account, err error) {
	result.TenantId = tenantId
	result.Name = name
	result.Salt = salt
	result.FinalHash = finalHash
	result.Status = status

	DB.Create(&result)
	return
}

func FetchAccount(tenantId uint, name string) (result Account, err error) {
	DB.Where(&Account{TenantId: tenantId, Name: name}).First(&result)
	return
}

func AddRole(tenantId uint, name string, desc string, status int8) (result Role, err error) {
	result.TenantId = tenantId
	result.Name = name
	result.Desc = desc
	result.Status = status

	DB.Create(&result)
	return
}

func FetchRole(tenantId uint, name string) (result Role, err error) {
	DB.Where(&Role{TenantId: tenantId, Name: name}).First(&result)
	return
}
func AddPermission(tenantId uint, code, name, desc string, permissionType, status int8) (result Permission, err error) {
	result.TenantId = tenantId
	result.Code = code
	result.Name = name
	result.Desc = desc
	result.Type = permissionType
	result.Status = status

	DB.Create(&result)
	return
}

func FetchPermission(tenantId uint, code string) (result Permission, err error) {
	DB.Where(&Permission{TenantId: tenantId, Code: code}).First(&result)
	return
}

func FetchAllRolesByAccount(accountId uint) (result []Role, err error) {
	DB.Where("account_id = ?", accountId).Limit(50).Find(&result)
	return
}

func FetchAllRolesByPermission(permissionId uint) (result []Role, err error) {
	DB.Where("permission_id = ?", permissionId).Limit(50).Find(&result)
	return
}

func AddPermissionBindings(bindings []PermissionBinding) error {
	DB.Create(&bindings)
	return nil
}

func AddRoleBindings(bindings []RoleBinding) error{
	DB.Create(&bindings)
	return nil
}