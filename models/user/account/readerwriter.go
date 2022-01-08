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
 ******************************************************************************/

package account

import (
	"context"
	"github.com/pingcap-inc/tiem/common/structs"
)

type ReaderWriter interface {
	AddAccount(ctx context.Context, tenantId string, name string, salt string, finalHash string, status int8) (*Account, error)
	FindAccountByName(ctx context.Context, name string) (*Account, error)
	FindAccountById(ctx context.Context, id string) (*Account, error)

	CreateUser(ctx context.Context, user User) (info *structs.UserInfo, err error)
	DeleteUser(ctx context.Context, tenantID, userID string) error
	QueryUsers(ctx context.Context) (userInfos map[string]structs.UserInfo, err error)
	GetUser(ctx context.Context, tenantID, userID string) (userInfo structs.UserInfo, err error)
	UpdateUserStatus(ctx context.Context, tenantID, userID string, status string) error
	UpdateUserProfile(ctx context.Context, tenantID, userID, email, phone string) error
	UpdateUserPassword(ctx context.Context, tenantID, userID, salt, finalHash string) error

	CreateTenant(ctx context.Context, tenant Tenant) (info *structs.TenantInfo, err error)
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context, tenantID string) (tenant structs.TenantInfo, err error)
	QueryTenants(ctx context.Context) (tenants map[string]structs.TenantInfo, err error)
	UpdateTenantStatus(ctx context.Context, tenantID, status string) error
	UpdateTenantProfile(ctx context.Context, tenantID, name string, maxCluster, maxCPU, maxMemory, maxStorage int32) error
	UpdateTenantOnBoardingStatus(ctx context.Context, tenantID, status string) error
}
