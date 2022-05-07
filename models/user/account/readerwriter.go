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
	CreateUser(ctx context.Context, user *User, name string) (*User, *UserLogin, *UserTenantRelation, error)
	DeleteUser(ctx context.Context, userID string) error
	GetUser(ctx context.Context, userID string) (userInfo structs.UserInfo, err error)
	QueryUsers(ctx context.Context) (userInfos map[string]structs.UserInfo, err error)
	UpdateUserStatus(ctx context.Context, userID string, status string) error
	UpdateUserProfile(ctx context.Context, userID, nickname, email, phone string) error
	UpdateUserPassword(ctx context.Context, userID, salt, finalHash string) error

	GetUserByName(ctx context.Context, name string)(*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)

	CreateTenant(ctx context.Context, tenant *Tenant) (info *structs.TenantInfo, err error)
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context, tenantID string) (tenant structs.TenantInfo, err error)
	QueryTenants(ctx context.Context) (tenants map[string]structs.TenantInfo, err error)
	UpdateTenantStatus(ctx context.Context, tenantID, status string) error
	UpdateTenantProfile(ctx context.Context, tenantID, name string, maxCluster, maxCPU, maxMemory, maxStorage int32) error
	UpdateTenantOnBoardingStatus(ctx context.Context, tenantID, status string) error
}
