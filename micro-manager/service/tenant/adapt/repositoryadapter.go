package adapt

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
	"github.com/pingcap/ticp/micro-metadb/client"
	db "github.com/pingcap/ticp/micro-metadb/proto"
	"time"
)

type MockRepo struct{}
type MicroMetaDbRepo struct {}

func (m MicroMetaDbRepo) LoadPermissionAggregation(tenantId uint, code string) (p domain.PermissionAggregation, err error) {
	req := db.DBFindRolesByPermissionRequest{
		TenantId: int32(tenantId),
		Code: code,
	}

	resp, err := client.DBClient.FindRolesByPermission(context.TODO(), &req)
	if err != nil {
		return
	}

	permissionDTO := resp.Permission

	p.Permission = domain.Permission{
		Code: permissionDTO.GetCode(),
		TenantId: uint(permissionDTO.GetTenantId()),
		Name: permissionDTO.GetName(),
		Type: domain.PermissionTypeFromType(permissionDTO.GetType()),
		Desc: permissionDTO.GetDesc(),
		Status: domain.CommonStatusFromStatus(permissionDTO.GetStatus()),
	}

	rolesDTOs := resp.GetRoles()

	if rolesDTOs == nil {
		p.Roles = nil
	} else {
		p.Roles = make([]domain.Role, len(rolesDTOs), cap(rolesDTOs))
		for index, r := range rolesDTOs {
			p.Roles[index] = domain.Role{
				TenantId: uint(r.TenantId),
				Name: r.GetName(),
				Desc: r.GetDesc(),
				Status: domain.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}
	return
}

func (m MicroMetaDbRepo) LoadPermission(tenantId uint, code string) (domain.Permission, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountAggregation(name string) (account domain.AccountAggregation, err error) {
	req := db.DBFindAccountRequest{
		Name: name,
		WithRole: true,
	}

	resp, err := client.DBClient.FindAccount(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Account

	account.Id = uint(dto.Id)
	account.Name = dto.Name
	account.TenantId = uint(dto.TenantId)
	account.Salt = dto.Salt
	account.FinalHash = dto.FinalHash

	if dto.Roles == nil {
		account.Roles = nil
	} else {
		account.Roles = make([]domain.Role, len(dto.Roles), cap(dto.Roles))
		for index, r := range dto.Roles {
			account.Roles[index] = domain.Role{
				TenantId: uint(r.TenantId),
				Name: r.GetName(),
				Desc: r.GetDesc(),
				Status: domain.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}

	return
}

func (m MicroMetaDbRepo) Provide(tiCPToken *domain.TiCPToken) (tokenString string, err error) {
	// 提供token
	tokenString = uuid.New().String()

	req := db.DBSaveTokenRequest{
		Token: &db.DBTokenDTO{
			TenantId: int32(tiCPToken.TenantId),
			AccountId: int32(tiCPToken.AccountId),
			AccountName: tiCPToken.AccountName,
			ExpirationTime: tiCPToken.ExpirationTime.Unix(),
			TokenString: tokenString,
		},
	}

	_, err = client.DBClient.SaveToken(context.TODO(), &req)

	return
}

func (m MicroMetaDbRepo) Modify(tiCPToken *domain.TiCPToken) error {
	req := db.DBSaveTokenRequest{
		Token: &db.DBTokenDTO{
			TenantId: int32(tiCPToken.TenantId),
			AccountId: int32(tiCPToken.AccountId),
			AccountName: tiCPToken.AccountName,
			ExpirationTime: tiCPToken.ExpirationTime.Unix(),
			TokenString: tiCPToken.TokenString,
		},
	}

	_, err := client.DBClient.SaveToken(context.TODO(), &req)

	return err
}

func (m MicroMetaDbRepo) GetToken(tokenString string) (token domain.TiCPToken, err error) {
	req := db.DBFindTokenRequest{
		TokenString: tokenString,
	}

	resp, err := client.DBClient.FindToken(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Token

	token.TokenString = dto.TokenString
	token.AccountId =  uint(dto.AccountId)
	token.AccountName = dto.AccountName
	token.TenantId = uint(dto.TenantId)
	token.ExpirationTime = time.Unix(dto.ExpirationTime, 0)

	return
}

func (m MicroMetaDbRepo) AddTenant(tenant *domain.Tenant) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantByName(name string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantById(id uint) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddAccount(a *domain.Account) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountByName(name string) (account domain.Account, err error) {
	req := db.DBFindAccountRequest{
		Name: name,
		WithRole: false,
	}

	resp, err := client.DBClient.FindAccount(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Account
	account.Id = uint(dto.Id)
	account.Name = dto.Name
	account.TenantId = uint(dto.TenantId)
	account.Salt = dto.Salt
	account.FinalHash = dto.FinalHash

	return
}

func (m MicroMetaDbRepo) LoadAccountById(id uint) (domain.Account, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRole(r *domain.Role) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadRole(tenantId uint, name string) (domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermission(r *domain.Permission) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByAccount(account *domain.Account) ([]domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByPermission(permission *domain.Permission) ([]domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermissionBindings(bindings []domain.PermissionBinding) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRoleBindings(bindings []domain.RoleBinding) error {
	panic("implement me")
}

func InitMock() {
	domain.RbacRepo = &MicroMetaDbRepo{}
	domain.TenantRepo = &MicroMetaDbRepo{}
	domain.TokenMNG = &MicroMetaDbRepo{}
}


var me = domain.Account{}
var myToken domain.TiCPToken

var admin = []domain.Role{{Id: 1}}
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

func (m MockRepo) LoadTenantByName(name string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MockRepo) LoadTenantById(id uint) (domain.Tenant, error) {
	panic("implement me")
}
