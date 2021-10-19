package adapt

import (
	"context"
	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"time"
)

type MicroMetaDbRepo struct{}

func (m MicroMetaDbRepo) LoadPermissionAggregation(tenantId string, code string) (p domain.PermissionAggregation, err error) {
	req := dbpb.DBFindRolesByPermissionRequest{
		TenantId: tenantId,
		Code:     code,
	}

	resp, err := client.DBClient.FindRolesByPermission(context.TODO(), &req)
	if err != nil {
		return
	}

	permissionDTO := resp.Permission

	p.Permission = domain.Permission{
		Code:     permissionDTO.GetCode(),
		TenantId: permissionDTO.GetTenantId(),
		Name:     permissionDTO.GetName(),
		Type:     domain.PermissionTypeFromType(permissionDTO.GetType()),
		Desc:     permissionDTO.GetDesc(),
		Status:   domain.CommonStatusFromStatus(permissionDTO.GetStatus()),
	}

	rolesDTOs := resp.GetRoles()

	if rolesDTOs == nil {
		p.Roles = nil
	} else {
		p.Roles = make([]domain.Role, len(rolesDTOs), cap(rolesDTOs))
		for index, r := range rolesDTOs {
			p.Roles[index] = domain.Role{
				TenantId: r.TenantId,
				Name:     r.GetName(),
				Desc:     r.GetDesc(),
				Status:   domain.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}
	return
}

func (m MicroMetaDbRepo) LoadPermission(tenantId string, code string) (domain.Permission, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountAggregation(name string) (account domain.AccountAggregation, err error) {
	req := dbpb.DBFindAccountRequest{
		Name:     name,
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
		account.Roles = make([]domain.Role, len(dto.Roles), cap(dto.Roles))
		for index, r := range dto.Roles {
			account.Roles[index] = domain.Role{
				TenantId: r.TenantId,
				Name:     r.GetName(),
				Desc:     r.GetDesc(),
				Status:   domain.CommonStatusFromStatus(r.GetStatus()),
			}
		}
	}

	return
}

func (m MicroMetaDbRepo) Provide(tiEMToken *domain.TiEMToken) (tokenString string, err error) {
	// 提供token，简单地使用UUID
	tokenString = uuid.New().String()

	req := dbpb.DBSaveTokenRequest{
		Token: &dbpb.DBTokenDTO{
			TenantId:       tiEMToken.TenantId,
			AccountId:      tiEMToken.AccountId,
			AccountName:    tiEMToken.AccountName,
			ExpirationTime: tiEMToken.ExpirationTime.Unix(),
			TokenString:    tokenString,
		},
	}

	_, err = client.DBClient.SaveToken(context.TODO(), &req)

	return
}

func (m MicroMetaDbRepo) Modify(tiEMToken *domain.TiEMToken) error {
	req := dbpb.DBSaveTokenRequest{
		Token: &dbpb.DBTokenDTO{
			TenantId:       tiEMToken.TenantId,
			AccountId:      tiEMToken.AccountId,
			AccountName:    tiEMToken.AccountName,
			ExpirationTime: tiEMToken.ExpirationTime.Unix(),
			TokenString:    tiEMToken.TokenString,
		},
	}

	_, err := client.DBClient.SaveToken(context.TODO(), &req)

	return err
}

func (m MicroMetaDbRepo) GetToken(tokenString string) (token domain.TiEMToken, err error) {
	req := dbpb.DBFindTokenRequest{
		TokenString: tokenString,
	}

	resp, err := client.DBClient.FindToken(context.TODO(), &req)
	if err != nil {
		return
	}

	dto := resp.Token

	token.TokenString = dto.TokenString
	token.AccountId = dto.AccountId
	token.AccountName = dto.AccountName
	token.TenantId = dto.TenantId
	token.ExpirationTime = time.Unix(dto.ExpirationTime, 0)

	return
}

func (m MicroMetaDbRepo) AddTenant(tenant *domain.Tenant) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantByName(name string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantById(id string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddAccount(a *domain.Account) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountByName(name string) (account domain.Account, err error) {
	req := dbpb.DBFindAccountRequest{
		Name:     name,
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

func (m MicroMetaDbRepo) LoadAccountById(id string) (domain.Account, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRole(r *domain.Role) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadRole(tenantId string, name string) (domain.Role, error) {
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
