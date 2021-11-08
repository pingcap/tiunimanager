
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
	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"time"
)

type MicroMetaDbRepo struct{}

func (m MicroMetaDbRepo) LoadPermissionAggregation(ctx context.Context, tenantId string, code string) (p domain.PermissionAggregation, err error) {
	req := dbpb.DBFindRolesByPermissionRequest{
		TenantId: tenantId,
		Code:     code,
	}

	resp, err := client.DBClient.FindRolesByPermission(ctx, &req)
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

func (m MicroMetaDbRepo) LoadPermission(ctx context.Context, tenantId string, code string) (domain.Permission, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountAggregation(ctx context.Context, name string) (account domain.AccountAggregation, err error) {
	req := dbpb.DBFindAccountRequest{
		Name:     name,
		WithRole: true,
	}

	resp, err := client.DBClient.FindAccount(ctx, &req)
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

func (m MicroMetaDbRepo) Provide(ctx context.Context, tiEMToken *domain.TiEMToken) (tokenString string, err error) {
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

	_, err = client.DBClient.SaveToken(ctx, &req)

	return
}

func (m MicroMetaDbRepo) Modify(ctx context.Context, tiEMToken *domain.TiEMToken) error {
	req := dbpb.DBSaveTokenRequest{
		Token: &dbpb.DBTokenDTO{
			TenantId:       tiEMToken.TenantId,
			AccountId:      tiEMToken.AccountId,
			AccountName:    tiEMToken.AccountName,
			ExpirationTime: tiEMToken.ExpirationTime.Unix(),
			TokenString:    tiEMToken.TokenString,
		},
	}

	_, err := client.DBClient.SaveToken(ctx, &req)

	return err
}

func (m MicroMetaDbRepo) GetToken(ctx context.Context, tokenString string) (token domain.TiEMToken, err error) {
	req := dbpb.DBFindTokenRequest{
		TokenString: tokenString,
	}

	resp, err := client.DBClient.FindToken(ctx, &req)
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

func (m MicroMetaDbRepo) AddTenant(ctx context.Context, tenant *domain.Tenant) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantByName(ctx context.Context, name string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadTenantById(ctx context.Context, id string) (domain.Tenant, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddAccount(ctx context.Context, a *domain.Account) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAccountByName(ctx context.Context, name string) (account domain.Account, err error) {
	req := dbpb.DBFindAccountRequest{
		Name:     name,
		WithRole: false,
	}

	resp, err := client.DBClient.FindAccount(ctx, &req)
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

func (m MicroMetaDbRepo) LoadAccountById(ctx context.Context, id string) (domain.Account, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRole(ctx context.Context, r *domain.Role) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadRole(ctx context.Context, tenantId string, name string) (domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermission(ctx context.Context, r *domain.Permission) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByAccount(ctx context.Context, account *domain.Account) ([]domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) LoadAllRolesByPermission(ctx context.Context, permission *domain.Permission) ([]domain.Role, error) {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddPermissionBindings(ctx context.Context, bindings []domain.PermissionBinding) error {
	panic("implement me")
}

func (m MicroMetaDbRepo) AddRoleBindings(ctx context.Context, bindings []domain.RoleBinding) error {
	panic("implement me")
}
