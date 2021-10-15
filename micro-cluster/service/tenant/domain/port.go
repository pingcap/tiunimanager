
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package domain

var RbacRepo RbacRepository

var TenantRepo TenantRepository

var TokenMNG TokenManager

type TenantRepository interface {
	AddTenant(*Tenant) error

	LoadTenantByName(name string)  (Tenant, error)

	LoadTenantById(id string)  (Tenant, error)
}

type RbacRepository interface {

	AddAccount(a *Account) error

	LoadAccountByName(name string) (Account, error)

	LoadAccountAggregation(name string) (AccountAggregation, error)

	LoadAccountById(id string) (Account, error)

	AddRole(r *Role) error

	LoadRole(tenantId string, name string) (Role, error)

	AddPermission(r *Permission) error

	LoadPermissionAggregation(tenantId string, code string) (PermissionAggregation, error)

	LoadPermission(tenantId string, code string) (Permission, error)

	LoadAllRolesByAccount(account *Account) ([]Role, error)

	LoadAllRolesByPermission(permission *Permission) ([]Role, error)

	AddPermissionBindings(bindings []PermissionBinding) error

	AddRoleBindings(bindings []RoleBinding) error
}

type TokenManager interface {

	// Provide 提供一个有效的token
	Provide  (tiEMToken *TiEMToken) (string, error)

	// Modify 修改token
	Modify (tiEMToken *TiEMToken) error

	// GetToken 获取一个token
	GetToken(tokenString string) (TiEMToken, error)
}
