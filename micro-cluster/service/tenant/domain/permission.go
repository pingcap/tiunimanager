
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

import (
	"fmt"
)

type Permission struct {
	TenantId string
	Code     string
	Name     string
	Type     PermissionType
	Desc     string
	Status   CommonStatus
}

type PermissionType int

const (
	UnrecognizedType PermissionType = 0
	Path             PermissionType = 1
	Act              PermissionType = 2
	Data             PermissionType = 3
)

func PermissionTypeFromType(pType int32) PermissionType {
	switch pType {
		case 1: return Path
		case 2: return Act
		case 3: return Data
		default: return UnrecognizedType
	}
}

func (permission *Permission) persist() error{
	RbacRepo.AddPermission(permission)
	return nil
}

func createPermission(tenant *Tenant, code, name ,desc string, permissionType PermissionType) (*Permission, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := findPermissionByCode(tenant.Id, code)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("permission already exist")
	}

	permission := Permission{
		TenantId: tenant.Id,
		Code:     code,
		Name:     name,
		Type:     permissionType,
		Desc:     desc,
		Status:   Valid,
	}

	permission.persist()

	return &permission, nil
}

func findPermissionByCode(tenantId string, code string) (*Permission, error) {
	a,e := RbacRepo.LoadPermission(tenantId, code)
	return &a, e
}


func (permission *Permission) listAllRoles() ([]Role, error){
	return RbacRepo.LoadAllRolesByPermission(permission)
}