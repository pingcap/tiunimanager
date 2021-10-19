package models

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type Account struct {
	Entity

	Name      string `gorm:"default:null;not null"`
	Salt      string `gorm:"default:null;not null;<-:create"`
	FinalHash string `gorm:"default:null;not null"`
}

type Role struct {
	Entity
	Name string `gorm:"default:null;not null"`
	Desc string
}

type PermissionBinding struct {
	Entity
	RoleId       string `gorm:"default:null;not null"`
	PermissionId string `gorm:"default:null;not null"`
}

type RoleBinding struct {
	Entity
	RoleId    string `gorm:"size:255"`
	AccountId string `gorm:"size:255"`
}

type Permission struct {
	Entity
	Name string `gorm:"default:null;not null"`
	Type int8   `gorm:"default:0"`
	Desc string `gorm:"default:null"`
}

type DAOAccountManager struct {
	db *gorm.DB
}

func NewDAOAccountManager(d *gorm.DB) *DAOAccountManager {
	m := new(DAOAccountManager)
	m.SetDb(d)
	return m
}

func (m *DAOAccountManager) Db(ctx context.Context, ) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *DAOAccountManager) SetDb(db *gorm.DB) {
	m.db = db
}

func (m *DAOAccountManager) Add(ctx context.Context, tenantId string, name string, salt string, finalHash string, status int8) (rt *Account, err error) {
	if "" == tenantId || "" == name || "" == salt || "" == finalHash {
		return nil, errors.New(fmt.Sprintf("add account failed, has invalid parameter, tenantID: %s, name: %s, salt: %s, finalHash: %s, status: %d", tenantId, name, salt, finalHash, status))
	}
	rt = &Account{
		Entity:    Entity{TenantId: tenantId, Status: status},
		Name:      name,
		Salt:      salt,
		FinalHash: finalHash,
	}
	return rt, m.Db(ctx).Create(rt).Error
}

func (m *DAOAccountManager) Find(ctx context.Context, name string) (result *Account, err error) {
	if "" == name {
		return nil, errors.New(fmt.Sprintf("find account failed, has invalid parameter, name: %s", name))
	}
	result = &Account{Name: name}
	return result, m.Db(ctx).Where(&Account{Name: name}).First(result).Error
}

func (m *DAOAccountManager) FindById(ctx context.Context, id string) (result *Account, err error) {
	if "" == id {
		return nil, errors.New(fmt.Sprintf("find account failed, has invalid parameter, id: %s", id))
	}
	result = &Account{Entity: Entity{ID: id}}
	return result, m.Db(ctx).Where(&Account{Entity: Entity{ID: id}}).First(result).Error
}

func (m *DAOAccountManager) AddRole(ctx context.Context, tenantId string, name string, desc string, status int8) (result *Role, err error) {
	//TODO please add desc and status check
	if "" == tenantId || "" == name {
		return nil, errors.New(fmt.Sprintf("add role failed, has invalid parameter, tenantID: %s, name: %s, desc: %s, status: %d", tenantId, name, desc, status))
	}
	result = &Role{Entity: Entity{TenantId: tenantId, Status: status},
		Name: name,
		Desc: desc,
	}
	return result, m.Db(ctx).Create(result).Error
}

func (m *DAOAccountManager) FetchRole(ctx context.Context, tenantId string, name string) (result *Role, err error) {
	if "" == tenantId || "" == name {
		return nil, errors.New(fmt.Sprintf("fetch role failed, has invalid parameter, tenantID: %s, name: %s", tenantId, name))
	}
	result = &Role{}
	return result, m.Db(ctx).Where(&Role{Entity: Entity{TenantId: tenantId}, Name: name}).First(result).Error
}

func (m *DAOAccountManager) AddPermission(ctx context.Context, tenantId, code, name, desc string, permissionType, status int8) (result *Permission, err error) {
	//TODO please add permissionType and status check
	if "" == tenantId || "" == name || "" == code {
		return nil, errors.New(fmt.Sprintf("add permission failed, has invalid parameter, tenantID: %s, code: %s name: %s, permission type: %d, status: %d", tenantId, code, name, permissionType, status))
	}
	result = &Permission{
		Entity: Entity{TenantId: tenantId, Status: status, Code: code},
		Name:   name,
		Desc:   desc,
		Type:   permissionType,
	}
	return result, m.Db(ctx).Create(result).Error
}

func (m *DAOAccountManager) FetchPermission(ctx context.Context, tenantId, code string) (result *Permission, err error) {
	if "" == tenantId || "" == code {
		return nil, errors.New(fmt.Sprintf("FetchPermission failed, has invalid parameter, tenantID: %s, name: %s", tenantId, code))
	}
	result = &Permission{}
	return result, m.Db(ctx).Where(&Permission{Entity: Entity{TenantId: tenantId, Code: code}}).First(result).Error
}

func (m *DAOAccountManager) FetchAllRolesByAccount(ctx context.Context, tenantId string, accountId string) (result []Role, err error) {
	if "" == tenantId || "" == accountId {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByAccount failed, has invalid parameter, tenantID: %s, accountID: %s", tenantId, accountId))
	}

	var roleBinds []RoleBinding
	err = m.Db(ctx).Where("tenant_id = ? and account_id = ? and status = 0", tenantId, accountId).Limit(1024).Find(&roleBinds).Error
	if nil != err {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByAccount, query database failed, tenantID: %s, accountID: %s", tenantId, accountId))
	}

	if len(roleBinds) == 0 {
		return result, nil
	}

	var roleIds []string
	for _, v := range roleBinds {
		roleIds = append(roleIds, v.RoleId)
	}
	return m.FetchRolesByIds(ctx, roleIds)
}

func (m *DAOAccountManager) FetchRolesByIds(ctx context.Context, roleIds []string) (result []Role, err error) {
	if len(roleIds) <= 0 {
		return nil, errors.New(fmt.Sprintf("FetchRolesByIds failed, roleIds: %v", roleIds))
	}
	return result, m.Db(ctx).Where("id in ?", roleIds).Find(&result).Error
}

func (m *DAOAccountManager) FetchAllRolesByPermission(ctx context.Context, tenantId string, permissionId string) (result []Role, err error) {
	if "" == tenantId || "" == permissionId {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByPermission failed, has invalid parameter, tenantID: %s, permissionId: %s", tenantId, permissionId))
	}
	var permissionBindings []PermissionBinding
	err = m.Db(ctx).Where("tenant_id = ? and permission_id = ? and status = 0", tenantId, permissionId).Limit(1024).Find(&permissionBindings).Error

	if nil != err {
		return nil, errors.New(fmt.Sprintf("FetchAllRolesByPermission, query database failed, tenantID: %s, permissionId: %s", tenantId, permissionId))
	}

	var roleIds []string
	for _, v := range permissionBindings {
		roleIds = append(roleIds, v.RoleId)
	}
	return m.FetchRolesByIds(ctx, roleIds)
}

func (m *DAOAccountManager) AddPermissionBindings(ctx context.Context, bindings []PermissionBinding) error {
	for i, v := range bindings {
		bindings[i].Code = v.RoleId + "_" + v.PermissionId
	}
	return m.Db(ctx).Create(&bindings).Error
}

func (m *DAOAccountManager) AddRoleBindings(ctx context.Context, bindings []RoleBinding) error {
	for i, v := range bindings {
		bindings[i].Code = v.AccountId + "_" + v.RoleId
	}
	return m.Db(ctx).Create(&bindings).Error
}
