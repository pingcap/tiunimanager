package domain

import (
	"errors"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
)

func setupMockAdapter() {
	RbacRepo = &MockRepo{}
	TenantRepo = &MockRepo{}
	TokenMNG = &MockRepo{}
	mockData()
}

type MockRepo struct{}

// 拥有两个role
var me *Account

// 没有role
var other *Account

var tokens = make([]*TiEMToken, 0, 2)

var roles []Role
// testPath1 赋给两个role, testPath2
const testPath1 = "testPath1"
const testPath2 = "testPath2"
var permissions []PermissionAggregation

const testMyName = "testMyName"
const testMyPassword = "testMyPassword"
const testOtherName = "testOtherName"
const testOtherPassword = "testOtherPassword"

func mockData() {
	me = &Account{
		Name:     testMyName,
		TenantId: "1",
		Id:       "1",
	}
	me.genSaltAndHash(testMyPassword)

	other = &Account{
		Name:     testOtherName,
		TenantId: "1",
		Id:       "2",
	}
	other.genSaltAndHash(testOtherPassword)
	roles = []Role{{Id: "1", Name: "admin", TenantId: "1"}, {Id: "2", Name: "dba", TenantId: "1"}}

	permissions = []PermissionAggregation{
		{
			Permission{Code: testPath1},
			roles,
		},
		{
			Permission{Code: testPath2},
			[]Role{},
		},
	}
}

func (m MockRepo) AddTenant(tenant *Tenant) error {
	panic("implement me")
}

func (m MockRepo) LoadTenantByName(name string) (Tenant, error) {
	if name == "notExisted" {
		return Tenant{}, errors.New("tenant not existed")
	}
	return Tenant{Name: name}, nil
}

func (m MockRepo) LoadTenantById(id string) (Tenant, error) {
	if id == "notExisted" {
		return Tenant{}, errors.New("tenant not existed")
	}
	return Tenant{Id: id}, nil
}

func (m MockRepo) LoadAccountByName(name string) (Account, error) {
	if name == "" {
		return Account{}, errors.New("name empty")
	}

	if name == me.Name {
		return *me, nil
	}

	if name == other.Name {
		return *other, nil
	}

	return Account{}, errors.New("no account found")
}

func (m MockRepo) LoadAccountAggregation(name string) (AccountAggregation, error) {
	if name == "" {
		return AccountAggregation{}, errors.New("name empty")
	}

	if name == me.Name {
		return AccountAggregation{
			Account: *me,
			Roles:   roles,
		}, nil
	}

	if name == other.Name {
		return AccountAggregation{
			Account: *other,
			Roles:   []Role{},
		}, nil
	}

	return AccountAggregation{}, errors.New("noaccount")
}

func (m MockRepo) LoadAccountById(id string) (Account, error) {
	panic("implement me")
}

func (m MockRepo) LoadRole(tenantId string, name string) (Role, error) {
	if name == "" {
		return Role{}, errors.New("name empty")
	}

	if name == "notExisted" {
		return Role{}, errors.New("no role found")
	}

	return Role{TenantId: tenantId, Id: uuidutil.GenerateID(), Name: name, Status: Valid}, nil
}

func (m MockRepo) LoadPermissionAggregation(tenantId string, code string) (PermissionAggregation, error) {
	for _,p := range permissions {
		if p.Code == code {
			return p, nil
		}
	}
	return PermissionAggregation{}, errors.New("no permission found")
}

func (m MockRepo) LoadPermission(tenantId string, code string) (Permission, error) {
	panic("implement me")
}

func (m MockRepo) LoadAllRolesByAccount(account *Account) ([]Role, error) {
	panic("implement me")
}

func (m MockRepo) LoadAllRolesByPermission(permission *Permission) ([]Role, error) {
	panic("implement me")
}

func (m MockRepo) AddAccount(a *Account) error {
	a.Id = uuidutil.GenerateID()
	return nil
}

func (m MockRepo) AddRole(r *Role) error {
	r.Id = uuidutil.GenerateID()
	return nil
}

func (m MockRepo) AddPermission(r *Permission) error {
	panic("implement me")
}

func (m MockRepo) AddPermissionBindings(bindings []PermissionBinding) error {
	return nil
}

func (m MockRepo) AddRoleBindings(bindings []RoleBinding) error {
	return nil
}

func (m MockRepo) Provide(tiEMToken *TiEMToken) (string, error) {
	tiEMToken.TokenString = uuidutil.GenerateID()

	tokens = append(tokens, tiEMToken)

	return tiEMToken.TokenString, nil
}

func (m MockRepo) Modify(tiEMToken *TiEMToken) error {
	for index,token := range tokens {
		if token.TokenString == tiEMToken.TokenString {
			tokens[index] = tiEMToken
			return nil
		}
	}

	return errors.New("token not exist")
}

func (m MockRepo) GetToken(tokenString string) (TiEMToken, error) {
	for _,token := range tokens {
		if token.TokenString == tokenString {
			return *token, nil
		}
	}
	return TiEMToken{}, errors.New("no token")
}
