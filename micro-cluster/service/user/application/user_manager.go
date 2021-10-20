
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package application

import (
	"fmt"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/ports"
)

type UserManager struct {
	rbacRepo ports.RbacRepository
}

func NewUserManager(rbacRepo ports.RbacRepository) *UserManager {
	return &UserManager{rbacRepo : rbacRepo}
}

// CreateAccount CreateAccount
func (p *UserManager) CreateAccount(tenant *domain.Tenant, name, passwd string) (*domain.Account, error) {
	if tenant == nil || !tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := p.FindAccountByName(name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("account already exist")
	}

	account := domain.Account{Name: name, Status: domain.Valid}

	account.GenSaltAndHash(passwd)
	p.rbacRepo.AddAccount(&account)

	return &account, nil
}

// FindAccountByName FindAccountByName
func (p *UserManager) FindAccountByName(name string) (*domain.Account, error) {
	a, err := p.rbacRepo.LoadAccountByName(name)
	if err != nil {
		return nil, err
	}

	return &a, err
}

func (p *UserManager) Empower(role *domain.Role, permissions []domain.Permission) error {
	bindings := make([]domain.PermissionBinding, len(permissions), len(permissions))

	for index, r := range permissions {
		bindings[index] = domain.PermissionBinding{Role: role, Permission: &r, Status: domain.Valid}
	}
	return p.rbacRepo.AddPermissionBindings(bindings)
}

func (p *UserManager) CreateRole(tenant *domain.Tenant, name string, desc string) (*domain.Role, error) {
	if tenant == nil || !tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}
	if name == "" {
		return nil, fmt.Errorf("empty role name")
	}

	existed, e := p.FindRoleByName(tenant, name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("role already exist")
	}

	role := domain.Role{TenantId: tenant.Id, Name: name, Desc: desc, Status: domain.Valid}

	p.rbacRepo.AddRole(&role)
	return &role, nil

}

func (p *UserManager) FindRoleByName(tenant *domain.Tenant, name string) (*domain.Role, error) {
	r, e := p.rbacRepo.LoadRole(tenant.Id, name)
	return &r, e
}

func (p *UserManager) createPermission(tenant *domain.Tenant, code, name ,desc string, permissionType domain.PermissionType) (*domain.Permission, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := p.findPermissionByCode(tenant.Id, code)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("permission already exist")
	}

	permission := domain.Permission{
		TenantId: tenant.Id,
		Code:     code,
		Name:     name,
		Type:     permissionType,
		Desc:     desc,
		Status:   domain.Valid,
	}

	p.rbacRepo.AddPermission(&permission)
	return &permission, nil
}

func (p *UserManager) findPermissionByCode(tenantId string, code string) (*domain.Permission, error) {
	a,e := p.rbacRepo.LoadPermission(tenantId, code)
	return &a, e
}

// findAccountExtendInfo 根据名称获取账号及扩展信息
func (p *UserManager) findAccountAggregation(name string) (*domain.AccountAggregation, error) {
	a, err := p.rbacRepo.LoadAccountAggregation(name)
	if err != nil {
		return nil, err
	}

	return &a, err
}

func (p *UserManager) findPermissionAggregationByCode(tenantId string, code string) (*domain.PermissionAggregation, error) {
	a, e := p.rbacRepo.LoadPermissionAggregation(tenantId, code)
	return &a, e
}

func (p *UserManager) RbacRepo() ports.RbacRepository {
	return p.rbacRepo
}

func (p *UserManager) SetRbacRepo(rbacRepo ports.RbacRepository) {
	p.rbacRepo = rbacRepo
}
