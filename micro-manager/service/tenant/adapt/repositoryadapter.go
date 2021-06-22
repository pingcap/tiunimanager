package adapt

import (
	"errors"
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
)

type MockRepo struct{}

func InitMock() {
	domain.RbacRepo = &MockRepo{}
	domain.TenantRepo = &MockRepo{}
	domain.TokenMNG = &MockRepo{}
	me.Name = "peijin"
	me.GenSaltAndHash("zhangpeijin")
}

var me = domain.Account{}
func (m MockRepo) AddAccount(a *domain.Account) error {
	panic("implement me")
}

func (m MockRepo) FetchAccountByName(name string) (domain.Account, error) {
	if me.Name == "peijin" {
		return me, nil
	} else {
		return domain.Account{}, errors.New("用户不存在")
	}
}

func (m MockRepo) FetchAccountById(id uint) (domain.Account, error) {
	panic("implement me")
}

func (m MockRepo) AddRole(r *domain.Role) error {
	panic("implement me")
}

func (m MockRepo) FetchRole(tenantId uint, name string) (domain.Role, error) {
	panic("implement me")
}

func (m MockRepo) AddPermission(r *domain.Permission) error {
	panic("implement me")
}

func (m MockRepo) FetchPermission(tenantId uint, code string) (domain.Permission, error) {
	if code == "/api/v1/host/query" {
		return domain.Permission{Code: "/api/v1/host/query"}, nil
	}

	return domain.Permission{Code: "api/v1/host/query"}, errors.New("权限不存在")
}

func (m MockRepo) FetchAllRolesByAccount(account *domain.Account) ([]domain.Role, error) {
	return admin, nil
}

func (m MockRepo) FetchAllRolesByPermission(permission *domain.Permission) ([]domain.Role, error) {
	return admin, nil
}

func (m MockRepo) AddPermissionBindings(bindings []domain.PermissionBinding) error {
	panic("implement me")
}

func (m MockRepo) AddRoleBindings(bindings []domain.RoleBinding) error {
	panic("implement me")
}

var myToken domain.TiCPToken

func (m MockRepo) Provide(tiCPToken *domain.TiCPToken) (string, error) {
	myToken = *tiCPToken
	myToken.TokenString = "mocktoken"

	return myToken.TokenString, nil
}

func (m MockRepo) Modify(tiCPToken *domain.TiCPToken) error {
	myToken = *tiCPToken
	return nil
}

func (m MockRepo) GetToken(tokenString string) (domain.TiCPToken, error) {
	if tokenString == "mocktoken" {
		return myToken, nil
	} else {
		return domain.TiCPToken{}, errors.New("token不合法")
	}
}

func (m MockRepo) AddTenant(tenant *domain.Tenant) error {
	panic("implement me")
}

func (m MockRepo) FetchTenantByName(name string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MockRepo) FetchTenantById(id uint) (domain.Tenant, error) {
	panic("implement me")
}

var admin = []domain.Role{{Id: 1}}