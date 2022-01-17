/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package message

import "github.com/pingcap-inc/tiem/common/structs"

type CheckPermissionForUserReq struct {
	UserID      string                   `json:"userId"`
	Permissions []structs.RbacPermission `json:"permissions"`
}

type CheckPermissionForUserResp struct {
	Result bool `json:"result"`
}

type CreateRoleReq struct {
	Role string `json:"role"`
}

type CreateRoleResp struct {
}

type DeleteRoleReq struct {
	Role string `json:"role"`
}

type DeleteRoleResp struct {
}

type QueryRolesReq struct {
	UserID string `json:"userId"`
}

type QueryRolesResp struct {
	Roles []string
}

type AddPermissionsForRoleReq struct {
	Role        string                   `json:"role"`
	Permissions []structs.RbacPermission `json:"permissions"`
}

type AddPermissionsForRoleResp struct {
}

type DeletePermissionsForRoleReq struct {
	Role        string                   `json:"role"`
	Permissions []structs.RbacPermission `json:"permissions"`
}

type DeletePermissionsForRoleResp struct {
}

type QueryPermissionsForUserReq struct {
	UserID string `json:"userId"`
}

type QueryPermissionsForUserResp struct {
	UserID      string                   `json:"userId"`
	Permissions []structs.RbacPermission `json:"permissions"`
}

type BindRolesForUserReq struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
}

type BindRolesForUserResp struct {
}

type UnbindRoleForUserReq struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type UnbindRoleForUserResp struct {
}
