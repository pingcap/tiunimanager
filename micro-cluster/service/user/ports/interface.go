
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

package ports

import (
	"context"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
)

type TenantRepository interface {
	AddTenant(ctx context.Context, tenant *domain.Tenant) error

	LoadTenantByName(ctx context.Context, name string)  (domain.Tenant, error)

	LoadTenantById(ctx context.Context, id string)  (domain.Tenant, error)
}

type RbacRepository interface {

	AddAccount(ctx context.Context, a *domain.Account) error

	LoadAccountByName(ctx context.Context, name string) (domain.Account, error)

	LoadAccountAggregation(ctx context.Context, name string) (domain.AccountAggregation, error)

	LoadAccountById(ctx context.Context, id string) (domain.Account, error)

	AddRole(ctx context.Context, r *domain.Role) error

	LoadRole(ctx context.Context, tenantId string, name string) (domain.Role, error)

	AddPermission(ctx context.Context, r *domain.Permission) error

	LoadPermissionAggregation(ctx context.Context, tenantId string, code string) (domain.PermissionAggregation, error)

	LoadPermission(ctx context.Context, tenantId string, code string) (domain.Permission, error)

	LoadAllRolesByAccount(ctx context.Context, account *domain.Account) ([]domain.Role, error)

	LoadAllRolesByPermission(ctx context.Context, permission *domain.Permission) ([]domain.Role, error)

	AddPermissionBindings(ctx context.Context, bindings []domain.PermissionBinding) error

	AddRoleBindings(ctx context.Context, bindings []domain.RoleBinding) error
}

type TokenHandler interface {

	// Provide provide a valid token string
	Provide (ctx context.Context, tiEMToken *domain.TiEMToken) (string, error)

	// Modify
	Modify (ctx context.Context, tiEMToken *domain.TiEMToken) error

	// GetToken get token by tokenString
	GetToken(ctx context.Context, tokenString string) (domain.TiEMToken, error)
}
