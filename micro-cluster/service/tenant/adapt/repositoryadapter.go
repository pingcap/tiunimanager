package adapt

import (
	"context"
	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/library/client"
	tenant "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/domain"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"time"
)

type MicroMetaDbRepo struct {}

func InjectionMetaDbRepo () {
	tenant.RbacRepo = &MicroMetaDbRepo{}
	tenant.TenantRepo = &MicroMetaDbRepo{}
	tenant.TokenMNG = &MicroMetaDbRepo{}
}

func (m MicroMetaDbRepo) LoadPermissionAggregation(tenantId string, code string) (p tenant.PermissionAggregation, err error) {
	req := db.DBFindRolesByPermissionRequest{
		TenantId: tenantId,
		Code: code,
	}

	resp, err := client.DBClient.FindRolesByPermission(context.TODO(), &req)
	if err != nil {
		return
	}

	permissionDTO := resp.Permission

	p.Permission = tenant.Permission{
		Code:     permissionDTO.GetCode(),
		TenantId: permissionDTO.GetTenantId(),
		Name:     permissionDTO.GetName(),
		Type:     tenant.PermissionTypeFromType(permissionDTO.GetType()),
		Desc:     permissionDTO.GetDesc(),
		Status:   tenant.CommonStatusFromStatus(permissionDTO.GetStatus()),
	}

	rolesDTOs := resp.GetRoles()

	if rolesDTOs == nil {
		p.Roles = nil
	} else {
		p.Roles = make([]tenant.Role, len(rolesDTOs), cap(rolesDTOs))
		for index, r := range rolesDTOs {
			p.Roles[index] = tenant.Role{
				TenantId: r.TenantId,
				Name:     r.GetName(),
				Desc:     r.GetDesc(),
				Status:   tenant.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}
	return
}

func (m MicroMetaDbRepo) LoadPermission(tenantId string, code string) (tenant.Permission, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountAggregation(name string) (account tenant.AccountAggregation, err error) {
	req := db.DBFindAccountRequest{
		Name: name,
		WithRole: true,
	}

	resp, err := client.DBClient.FindAccount(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Account

	account.Id = dto.Id
	account.Name = dto.Name
	account.TenantId = dto.TenantId
	account.Salt = dto.Salt
	account.FinalHash = dto.FinalHash

	if dto.Roles == nil {
		account.Roles = nil
	} else {
		account.Roles = make([]tenant.Role, len(dto.Roles), cap(dto.Roles))
		for index, r := range dto.Roles {
			account.Roles[index] = tenant.Role{
				TenantId: r.TenantId,
				Name:     r.GetName(),
				Desc:     r.GetDesc(),
				Status:   tenant.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}

	return
}

func (m MicroMetaDbRepo) Provide(tiEMToken *tenant.TiEMToken) (tokenString string, err error) {
	// 提供token，简单地使用UUID
	tokenString = uuid.New().String()

	req := db.DBSaveTokenRequest{
		Token: &db.DBTokenDTO{
			TenantId: tiEMToken.TenantId,
			AccountId: tiEMToken.AccountId,
			AccountName: tiEMToken.AccountName,
			ExpirationTime: tiEMToken.ExpirationTime.Unix(),
			TokenString: tokenString,
		},
	}

	_, err = client.DBClient.SaveToken(context.TODO(), &req)

	return
}

func (m MicroMetaDbRepo) Modify(tiEMToken *tenant.TiEMToken) error {
	req := db.DBSaveTokenRequest{
		Token: &db.DBTokenDTO{
			TenantId: tiEMToken.TenantId,
			AccountId: tiEMToken.AccountId,
			AccountName: tiEMToken.AccountName,
			ExpirationTime: tiEMToken.ExpirationTime.Unix(),
			TokenString: tiEMToken.TokenString,
		},
	}

	_, err := client.DBClient.SaveToken(context.TODO(), &req)

	return err
}

func (m MicroMetaDbRepo) GetToken(tokenString string) (token tenant.TiEMToken, err error) {
	req := db.DBFindTokenRequest{
		TokenString: tokenString,
	}

	resp, err := client.DBClient.FindToken(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Token

	token.TokenString = dto.TokenString
	token.AccountId =  dto.AccountId
	token.AccountName = dto.AccountName
	token.TenantId = dto.TenantId
	token.ExpirationTime = time.Unix(dto.ExpirationTime, 0)

	return
}

func (m MicroMetaDbRepo) AddTenant(tenant *tenant.Tenant) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantByName(name string) (tenant.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantById(id string) (tenant.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddAccount(a *tenant.Account) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountByName(name string) (account tenant.Account, err error) {
	req := db.DBFindAccountRequest{
		Name: name,
		WithRole: false,
	}

	resp, err := client.DBClient.FindAccount(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Account
	account.Id = dto.Id
	account.Name = dto.Name
	account.TenantId = dto.TenantId
	account.Salt = dto.Salt
	account.FinalHash = dto.FinalHash

	return
}

func (m MicroMetaDbRepo) LoadAccountById(id string) (tenant.Account, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRole(r *tenant.Role) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadRole(tenantId string, name string) (tenant.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermission(r *tenant.Permission) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByAccount(account *tenant.Account) ([]tenant.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByPermission(permission *tenant.Permission) ([]tenant.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermissionBindings(bindings []tenant.PermissionBinding) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRoleBindings(bindings []tenant.RoleBinding) error {
	panic("implement me")
}


