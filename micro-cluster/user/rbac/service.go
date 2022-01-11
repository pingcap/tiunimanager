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
	"github.com/pingcap-inc/tiem/message"
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
	// @Return message.DeleteRoleResp
	// @Return error
	DeleteRole(ctx context.Context, request message.DeleteRoleReq) (resp message.DeleteRoleResp, err error)

	// DeleteUser
	// @Description: delete rbac user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DeleteUserResp
	// @Return error
	DeleteUser(ctx context.Context, request message.DeleteUserReq) (resp message.DeleteUserResp, err error)

	// GetRoles
	// @Description: get rbac roles
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.GetRolesResp
	// @Return error
	GetRoles(ctx context.Context, request message.GetRolesReq) (resp message.GetRolesResp, err error)

	// CreateRole
	// @Description: create rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.CreateRoleResp
	// @Return error
	CreateRole(ctx context.Context, request message.CreateRoleReq) (resp message.CreateRoleResp, err error)

	// BindRoleForUser
	// @Description: bind rbac role for user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.BindRoleForUserResp
	// @Return error
	BindRoleForUser(ctx context.Context, request message.BindRoleForUserReq) (resp message.BindRoleForUserResp, err error)

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
	// @Return message.AddPermissionsForRoleResp
	// @Return error
	AddPermissionsForRole(ctx context.Context, request message.AddPermissionsForRoleReq) (resp message.AddPermissionsForRoleResp, err error)

	// DeletePermissionsForRole
	// @Description: delete permissions for rbac role
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DeletePermissionsForRoleResp
	// @Return error
	DeletePermissionsForRole(ctx context.Context, request message.DeletePermissionsForRoleReq) (resp message.DeletePermissionsForRoleResp, err error)

	// GetPermissionsForUser
	// @Description: get permissions for rbac user
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.GetPermissionsForUserResp
	// @Return error
	GetPermissionsForUser(ctx context.Context, request message.GetPermissionsForUserReq) (resp message.GetPermissionsForUserResp, err error)
}
