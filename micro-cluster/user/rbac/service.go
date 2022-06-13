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

package rbac

import (
	"context"
	"github.com/pingcap/tiunimanager/message"
)

// RBACService RBAC service interface
type RBACService interface {
	// CheckPermissionForUser
	// @Description: check permission for user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.CheckPermissionForUserResp
	// @Return error
	CheckPermissionForUser(ctx context.Context, request message.CheckPermissionForUserReq) (resp message.CheckPermissionForUserResp, err error)

	// DeleteRole
	// @Description: delete rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter system
	// @Return message.DeleteRoleResp
	// @Return error
	DeleteRole(ctx context.Context, request message.DeleteRoleReq, system bool) (resp message.DeleteRoleResp, err error)

	// QueryRoles
	// @Description: query rbac roles by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryRolesResp
	// @Return error
	QueryRoles(ctx context.Context, request message.QueryRolesReq) (resp message.QueryRolesResp, err error)

	// CreateRole
	// @Description: create rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter system
	// @Return message.CreateRoleResp
	// @Return error
	CreateRole(ctx context.Context, request message.CreateRoleReq, system bool) (resp message.CreateRoleResp, err error)

	// BindRolesForUser
	// @Description: bind rbac roles for user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.BindRolesForUserResp
	// @Return error
	BindRolesForUser(ctx context.Context, request message.BindRolesForUserReq) (resp message.BindRolesForUserResp, err error)

	// UnbindRoleForUser
	// @Description: unbind rbac role for user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.UnbindRoleForUserResp
	// @Return error
	UnbindRoleForUser(ctx context.Context, request message.UnbindRoleForUserReq) (resp message.UnbindRoleForUserResp, err error)

	// AddPermissionsForRole
	// @Description: add permissions for rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter system
	// @Return message.AddPermissionsForRoleResp
	// @Return error
	AddPermissionsForRole(ctx context.Context, request message.AddPermissionsForRoleReq, system bool) (resp message.AddPermissionsForRoleResp, err error)

	// DeletePermissionsForRole
	// @Description: delete permissions for rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DeletePermissionsForRoleResp
	// @Return error
	DeletePermissionsForRole(ctx context.Context, request message.DeletePermissionsForRoleReq) (resp message.DeletePermissionsForRoleResp, err error)

	// QueryPermissionsForUser
	// @Description: query permissions for rbac user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryPermissionsForUserResp
	// @Return error
	QueryPermissionsForUser(ctx context.Context, request message.QueryPermissionsForUserReq) (resp message.QueryPermissionsForUserResp, err error)
}
