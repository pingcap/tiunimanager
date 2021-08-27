package models

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type Account struct {
	Entity

	Name 			string		`gorm:"default:null;not null"`
	Salt 			string		`gorm:"default:null;not null;<-:create"`
	FinalHash 		string		`gorm:"default:null;not null"`
}

type Role struct {
	Entity
	Name    		string		`gorm:"default:null;not null"`
	Desc    		string
}

type PermissionBinding struct {
	Entity
	RoleId 			string		`gorm:"default:null;not null"`
	PermissionId	string		`gorm:"default:null;not null"`
}

type RoleBinding struct {
	Entity
	RoleId 			string		`gorm:"size:255"`
	AccountId	 	string		`gorm:"size:255"`
}

type Permission struct {
	Entity
	Name  	 		string		`gorm:"default:null;not null"`
	Type   			int8		`gorm:"default:0"`
	Desc   			string		`gorm:"default:null"`
}

func (*Account) Add(db *gorm.DB,tenantId string, name string, salt string, finalHash string, status int8) (rt* Account, err error) {
	if nil == db || "" == tenantId || "" == name || "" == salt || "" == finalHash {
		return nil, errors.New(fmt.Sprintf("add account failed, has invalid parameter, tenantID: %s, name: %s, salt: %s, finalHash: %s, status: %d", tenantId, name, salt, finalHash,status))
	}
	rt = &Account{
		Entity: Entity{TenantId:tenantId, Status: status},
		Name: name,
		Salt: salt,
		FinalHash: finalHash,
	}
	return rt,db.Create(rt).Error
}

func (*Account) Find(db *gorm.DB, name string) (result *Account, err error) {
	if nil == db || "" == name {
		return nil, errors.New(fmt.Sprintf("find account failed, has invalid parameter, name: %s", name))
	}
	result = &Account{Name:name}
	return result, db.Where(&Account{Name: name}).First(result).Error
}

func (r* Role) AddRole(db *gorm.DB,tenantId string, name string, desc string, status int8) (result *Role, err error) {
	//TODO please add desc and status check
	if nil == db || "" == tenantId || "" == name {
		return nil, errors.New(fmt.Sprintf("add role failed, has invalid parameter, tenantID: %s, name: %s, desc: %s, status: %d", tenantId, name, desc, status))
	}
	result = &Role{Entity: Entity{TenantId:tenantId, Status: status},
		Name:name,
		Desc: desc,
	}
	return result, db.Create(result).Error
}

func (r* Role)FetchRole(db * gorm.DB,tenantId string, name string) (result* Role, err error) {
	if nil == db || "" == tenantId || "" == name {
		return nil, errors.New(fmt.Sprintf("fetch role failed, has invalid parameter, tenantID: %s, name: %s", tenantId, name))
	}
	result = &Role{}
	return result, db.Where(&Role{Entity: Entity{TenantId: tenantId}, Name: name}).First(result).Error
}

func (p * Permission) AddPermission(db *gorm.DB,tenantId, code, name, desc string, permissionType, status int8) (result* Permission, err error) {
	//TODO please add permissionType and status check
	if nil == db || "" == tenantId || "" == name || "" == code || "" == desc {
		return nil, errors.New(fmt.Sprintf("add permission failed, has invalid parameter, tenantID: %s, code: %s name: %s, desc: %s, permission type: %d, status: %d", tenantId, code, name, desc, permissionType,status))
	}
	result = &Permission{
		Entity: Entity{TenantId:tenantId, Status:status, Code:code},
		Name: name,
		Desc: desc,
		Type: permissionType,
	}
	return result, db.Create(result).Error
}

func (p *Permission) FetchPermission(db *gorm.DB,tenantId, code string) (result* Permission, err error) {
	if nil == db || "" == tenantId || "" == code{
		return nil, errors.New(fmt.Sprintf("FetchPermission failed, has invalid parameter, tenantID: %s, name: %s", tenantId, code))
	}
	result = &Permission{}
	return result, db.Where(&Permission{Entity: Entity{TenantId: tenantId, Code: code}}).First(result).Error
}

func (r* Role)FetchAllRolesByAccount(db *gorm.DB, tenantId string, accountId string) (result []Role, err error) {
	if nil == db || "" == tenantId || "" == accountId {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByAccount failed, has invalid parameter, tenantID: %s, accountID: %s", tenantId, accountId))
	}

	var roleBinds []RoleBinding
	err = db.Where("tenant_id = ? and account_id = ? and status = 0", tenantId, accountId).Limit(1024).Find(&roleBinds).Error
	if nil != err {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByAccount, query database failed, tenantID: %s, accountID: %s", tenantId, accountId))
	}

	var roleIds []string
	for _, v := range roleBinds {
		roleIds = append(roleIds, v.RoleId)
	}
	return r.FetchRolesByIds(db,roleIds)
}

func (r *Role) FetchRolesByIds(db *gorm.DB,roleIds []string) (result []Role, err error){
	if nil == db || len(roleIds) <= 0 {
		return nil, errors.New(fmt.Sprintf("FetchRolesByIds failed"))
	}
	return result, db.Where("id in ?", roleIds).Find(&result).Error
}

func (r* Role)FetchAllRolesByPermission(db *gorm.DB,tenantId string, permissionId string) (result []Role, err error) {
	if nil == db || "" == tenantId || "" == permissionId {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByPermission failed, has invalid parameter, tenantID: %s, permissionId: %s", tenantId, permissionId))
	}
	var permissionBindings []PermissionBinding
	err = db.Where("tenant_id = ? and permission_id = ? and status = 0", tenantId, permissionId).Limit(1024).Find(&permissionBindings).Error

	if nil != err {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByPermission, query database failed, tenantID: %s, permissionId: %s", tenantId, permissionId))
	}

	var roleIds []string
	for _, v := range permissionBindings {
		roleIds = append(roleIds, v.RoleId)
	}
	return r.FetchRolesByIds(db, roleIds)
}

func (pb *PermissionBinding) AddPermissionBindings(db* gorm.DB,bindings []PermissionBinding) error {
	for i,v := range bindings {
		bindings[i].Code= v.RoleId + "_" + v.PermissionId
	}
	return db.Create(&bindings).Error
}

func (rb* RoleBinding) AddRoleBindings(db *gorm.DB,bindings []RoleBinding) error {
	for i,v := range bindings {
		bindings[i].Code= v.AccountId + "_" + v.RoleId
	}
	return db.Create(&bindings).Error
}