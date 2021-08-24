package models

import "errors"

type AccountDO struct {
	Entity

	Name 			string		`gorm:"default:null;not null"`
	Salt 			string		`gorm:"default:null;not null;<-:create"`
	FinalHash 		string		`gorm:"default:null;not null"`
}

func (d AccountDO) TableName() string {
	return "accounts"
}

type RoleDO struct {
	Entity
	Name    		string		`gorm:"default:null;not null"`
	Desc    		string
}

func (d RoleDO) TableName() string {
	return "roles"
}

type PermissionBindingDO struct {
	Entity
	RoleId 			string		`gorm:"default:null;not null"`
	PermissionId	string		`gorm:"default:null;not null"`
}

func (d PermissionBindingDO) TableName() string {
	return "permission_bindings"
}

type RoleBindingDO struct {
	Entity
	RoleId 			string		`gorm:"size:255"`
	AccountId	 	string		`gorm:"size:255"`
}

func (d RoleBindingDO) TableName() string {
	return "role_bindings"
}

type PermissionDO struct {
	Entity
	Name  	 		string		`gorm:"default:null;not null"`
	Type   			int8		`gorm:"default:0"`
	Desc   			string		`gorm:"default:null"`
}

func (d PermissionDO) TableName() string {
	return "permissions"
}

func AddAccount(tenantId string, name string, salt string, finalHash string, status int8) (result AccountDO, err error) {
	result.TenantId = tenantId
	result.Name = name
	result.Salt = salt
	result.FinalHash = finalHash
	result.Status = status

	err = MetaDB.Create(&result).Error
	return
}

func FindAccount(name string) (result AccountDO, err error) {
	err = MetaDB.Where(&AccountDO{Name: name}).First(&result).Error
	return
}

func AddRole(tenantId string, name string, desc string, status int8) (result RoleDO, err error) {
	result.TenantId = tenantId
	result.Name = name
	result.Desc = desc
	result.Status = status

	err = MetaDB.Create(&result).Error
	return
}

func FetchRole(tenantId string, name string) (result RoleDO, err error) {
	err = MetaDB.Where(&RoleDO{Entity: Entity{TenantId: tenantId}, Name: name}).First(&result).Error
	return
}

func AddPermission(tenantId, code, name, desc string, permissionType, status int8) (result PermissionDO, err error) {
	if len(code) == 0 {
		err = errors.New("permission code cannot be empty")
		return
	}
	result.TenantId = tenantId
	result.Code = code
	result.Name = name
	result.Desc = desc
	result.Type = permissionType
	result.Status = status

	err = MetaDB.Create(&result).Error
	return
}

func FetchPermission(tenantId, code string) (result PermissionDO, err error) {
	err = MetaDB.Where(&PermissionDO{Entity: Entity{TenantId: tenantId, Code: code}}).First(&result).Error
	return
}

func FetchAllRolesByAccount(tenantId string, accountId string) (result []RoleDO, err error) {
	//service.DB.Where("account_id = ?", accountId).Limit(50).Find(&result)

	var roleBinds []RoleBindingDO
	err = MetaDB.Where("tenant_id = ? and account_id = ? and status = 0", tenantId, accountId).Limit(50).Find(&roleBinds).Error

	if err != nil {
		return
	}
	var roleIds []string
	for _, v := range roleBinds {
		roleIds = append(roleIds, v.RoleId)
	}

	result, err = FetchRolesByIds(roleIds)
	return
}

func FetchRolesByIds(roleIds []string) (result []RoleDO, err error){
	err = MetaDB.Where("id in ?", roleIds).Find(&result).Error
	return
}

func FetchAllRolesByPermission(tenantId string, permissionId string) (result []RoleDO, err error) {
	var permissionBindings []PermissionBindingDO
	err = MetaDB.Where("tenant_id = ? and permission_id = ? and status = 0", tenantId, permissionId).Limit(50).Find(&permissionBindings).Error

	if err != nil {
		return
	}

	var roleIds []string
	for _, v := range permissionBindings {
		roleIds = append(roleIds, v.RoleId)
	}

	result, err = FetchRolesByIds(roleIds)
	return
}

func AddPermissionBindings(bindings []PermissionBindingDO) error {
	for i,v := range bindings {
		bindings[i].Code= v.RoleId + "_" + v.PermissionId
	}
	return MetaDB.Create(&bindings).Error
}

func AddRoleBindings(bindings []RoleBindingDO) error {
	for i,v := range bindings {
		bindings[i].Code= v.AccountId + "_" + v.RoleId
	}
	return MetaDB.Create(&bindings).Error
}