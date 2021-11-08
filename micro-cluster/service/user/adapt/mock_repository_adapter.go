
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

package adapt

import (
	"context"
	"errors"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
)

type MockRepo struct{}

var Me *domain.Account

var Other *domain.Account

var Tokens = make([]*domain.TiEMToken, 0, 2)

var Roles []domain.Role

const TestPath1 = "testPath1"
const TestPath2 = "testPath2"

var permissions []domain.PermissionAggregation

const TestMyName = "testMyName"
const TestMyPassword = "testMyPassword"
const TestOtherName = "testOtherName"
const TestOtherPassword = "testOtherPassword"

func NewMockRepo() *MockRepo {
	mockData()
	return &MockRepo{}
}

func mockData() {
	Me = &domain.Account{
		Name:     TestMyName,
		TenantId: "1",
		Id:       "1",
	}
	Me.GenSaltAndHash(TestMyPassword)

	Other = &domain.Account{
		Name:     TestOtherName,
		TenantId: "1",
		Id:       "2",
	}
	Other.GenSaltAndHash(TestOtherPassword)
	Roles = []domain.Role{{Id: "1", Name: "admin", TenantId: "1"}, {Id: "2", Name: "dba", TenantId: "1"}}

	permissions = []domain.PermissionAggregation{
		{
			Permission: domain.Permission{Code: TestPath1},
			Roles: Roles,
		},
		{
			Permission: domain.Permission{Code: TestPath2},
			Roles: []domain.Role{},
		},
	}
}

func (m MockRepo) AddTenant(ctx context.Context, tenant *domain.Tenant) error {
	panic("implement me")
}

func (m MockRepo) LoadTenantByName(ctx context.Context, name string) (domain.Tenant, error) {
	if name == "notExisted" {
		return domain.Tenant{}, errors.New("tenant not existed")
	}
	return domain.Tenant{Name: name}, nil
}

func (m MockRepo) LoadTenantById(ctx context.Context, id string) (domain.Tenant, error) {
	if id == "notExisted" {
		return domain.Tenant{}, errors.New("tenant not existed")
	}
	return domain.Tenant{Id: id}, nil
}

func (m MockRepo) LoadAccountByName(ctx context.Context, name string) (domain.Account, error) {
	if name == "" {
		return domain.Account{}, errors.New("name empty")
	}

	if name == Me.Name {
		return *Me, nil
	}

	if name == Other.Name {
		return *Other, nil
	}

	return domain.Account{}, errors.New("no account found")
}

func (m MockRepo) LoadAccountAggregation(ctx context.Context, name string) (domain.AccountAggregation, error) {
	if name == "" {
		return domain.AccountAggregation{}, errors.New("name empty")
	}

	if name == Me.Name {
		return domain.AccountAggregation{
			Account: *Me,
			Roles:   Roles,
		}, nil
	}

	if name == Other.Name {
		return domain.AccountAggregation{
			Account: *Other,
			Roles:   []domain.Role{},
		}, nil
	}

	return domain.AccountAggregation{}, errors.New("noaccount")
}

func (m MockRepo) LoadAccountById(ctx context.Context, id string) (domain.Account, error) {
	panic("implement me")
}

func (m MockRepo) LoadRole(ctx context.Context, tenantId string, name string) (domain.Role, error) {
	if name == "" {
		return domain.Role{}, errors.New("name empty")
	}

	if name == "notExisted" {
		return domain.Role{}, errors.New("no role found")
	}

	return domain.Role{TenantId: tenantId, Id: uuidutil.GenerateID(), Name: name, Status: domain.Valid}, nil
}

func (m MockRepo) LoadPermissionAggregation(ctx context.Context, tenantId string, code string) (domain.PermissionAggregation, error) {
	for _, p := range permissions {
		if p.Code == code {
			return p, nil
		}
	}
	return domain.PermissionAggregation{}, errors.New("no permission found")
}

func (m MockRepo) LoadPermission(ctx context.Context, tenantId string, code string) (domain.Permission, error) {
	panic("implement me")
}

func (m MockRepo) LoadAllRolesByAccount(ctx context.Context, account *domain.Account) ([]domain.Role, error) {
	panic("implement me")
}

func (m MockRepo) LoadAllRolesByPermission(ctx context.Context, permission *domain.Permission) ([]domain.Role, error) {
	panic("implement me")
}

func (m MockRepo) AddAccount(ctx context.Context, a *domain.Account) error {
	a.Id = uuidutil.GenerateID()
	return nil
}

func (m MockRepo) AddRole(ctx context.Context, r *domain.Role) error {
	r.Id = uuidutil.GenerateID()
	return nil
}

func (m MockRepo) AddPermission(ctx context.Context, r *domain.Permission) error {
	panic("implement me")
}

func (m MockRepo) AddPermissionBindings(ctx context.Context, bindings []domain.PermissionBinding) error {
	return nil
}

func (m MockRepo) AddRoleBindings(ctx context.Context, bindings []domain.RoleBinding) error {
	return nil
}

func (m MockRepo) Provide(ctx context.Context, tiEMToken *domain.TiEMToken) (string, error) {
	tiEMToken.TokenString = uuidutil.GenerateID()

	Tokens = append(Tokens, tiEMToken)

	return tiEMToken.TokenString, nil
}

func (m MockRepo) Modify(ctx context.Context, tiEMToken *domain.TiEMToken) error {
	for index, token := range Tokens {
		if token.TokenString == tiEMToken.TokenString {
			Tokens[index] = tiEMToken
			return nil
		}
	}

	return errors.New("token not exist")
}

func (m MockRepo) GetToken(ctx context.Context, tokenString string) (domain.TiEMToken, error) {
	for _, token := range Tokens {
		if token.TokenString == tokenString {
			return *token, nil
		}
	}
	return domain.TiEMToken{}, errors.New("no token")
}
