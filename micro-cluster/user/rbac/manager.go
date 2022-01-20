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
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
)

var rbacService RBACService

func GetRBACService() RBACService {
	if rbacService == nil {
		rbacService = NewRBACManager()
	}
	return rbacService
}

func MockRBACService(service RBACService) {
	rbacService = service
}

type RBACManager struct {
	enforcer *casbin.SyncedEnforcer // use synced enforcer to fit concurrent case
}

func NewRBACManager() *RBACManager {
	// casbin RBAC conf
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && (r.act == p.act || p.act == \"*\")")
	adapter, err := models.GetRBACReaderWriter().GetRBACAdapter(context.Background())
	if err != nil {
		framework.LogWithContext(context.Background()).Fatalf("get casbin gorm adapter failed, %s", err.Error())
		return nil
	}
	e, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		framework.LogWithContext(context.Background()).Fatalf("new casbin enforcer failed, %s", err.Error())
		return nil
	}
	if err := e.LoadPolicy(); err != nil {
		framework.LogWithContext(context.Background()).Fatalf("load rbac policy failed, %s", err.Error())
		return nil
	}
	mgr := &RBACManager{
		enforcer: e,
	}
	mgr.initDefaultRBAC(context.Background())
	return mgr
}

/*
	1. init role and permission
	2. bind admin role for admin user
*/
func (mgr *RBACManager) initDefaultRBAC(ctx context.Context) {
	framework.LogWithContext(ctx).Infof("begin init default rbac...")
	// 1. init role and permission
	// init admin role
	framework.LogWithContext(ctx).Infof("begin init default rbac role %s ...", constants.RbacRoleAdmin)
	mgr.CreateRole(ctx, message.CreateRoleReq{Role: string(constants.RbacRoleAdmin)}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceCluster),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceResource),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceParameter),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceUser),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceCDC),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceProduct),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceSystem),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)
	mgr.AddPermissionsForRole(ctx, message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{
		{
			Resource: string(constants.RbacResourceWorkflow),
			Action:   string(constants.RbacActionAll),
		},
	}}, true)

	// 2. bind admin role for admin user
	// todo: replace new account
	adminUser, _ := models.GetAccountReaderWriter().GetUserByName(ctx, "admin")
	if adminUser != nil && adminUser.ID != "" {
		mgr.BindRolesForUser(ctx, message.BindRolesForUserReq{UserID: adminUser.ID, Roles: []string{string(constants.RbacRoleAdmin)}})
		framework.LogWithContext(ctx).Infof("bind admin role for userId %s", adminUser.ID)
	} else {
		framework.LogWithContext(ctx).Errorf("get empty adminUser %+v", adminUser)
	}
}

func (mgr *RBACManager) CheckPermissionForUser(ctx context.Context, request message.CheckPermissionForUserReq) (resp message.CheckPermissionForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin CheckPermissionForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end CheckPermissionForUser")

	for _, permission := range request.Permissions {
		result, err := mgr.enforcer.Enforce(request.UserID, permission.Resource, permission.Action)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("user %s check permission error: %s", request.UserID, err.Error())
			return message.CheckPermissionForUserResp{Result: false}, errors.WrapError(errors.TIEM_RBAC_PERMISSION_CHECK_FAILED, fmt.Sprintf("user %s check permission failed", request.UserID), err)
		}
		if !result {
			framework.LogWithContext(ctx).Infof("user %s check permission %+v failed", request.UserID, permission)
			return message.CheckPermissionForUserResp{Result: false}, nil
		}
	}
	return message.CheckPermissionForUserResp{Result: true}, nil
}

func (mgr *RBACManager) QueryRoles(ctx context.Context, request message.QueryRolesReq) (resp message.QueryRolesResp, err error) {
	framework.LogWithContext(ctx).Infof("begin QueryRoles, request: %+v", request)
	framework.LogWithContext(ctx).Info("end QueryRoles")

	if request.UserID == "" {
		return message.QueryRolesResp{
			Roles: mgr.enforcer.GetAllRoles(),
		}, nil
	} else {
		roles, err := mgr.enforcer.GetRolesForUser(request.UserID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("call enforcer GetRolesForUser failed %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_RBAC_ROLE_QUERY_FAILED, fmt.Sprintf("call enforcer GetRolesForUser failed"), err)
		}
		return message.QueryRolesResp{
			Roles: roles,
		}, nil
	}
}

func (mgr *RBACManager) DeleteRole(ctx context.Context, request message.DeleteRoleReq, system bool) (resp message.DeleteRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeleteRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end DeleteRole")

	if !system {
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("default role %s can not delete", request.Role))
		}
	}

	if _, err = mgr.enforcer.DeletePermissionsForUser(request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeletePermissionsForUser failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_PERMISSION_DELETE_FAILED, fmt.Sprintf("call enforcer DeletePermissionsForUser failed"), err)
	}

	if _, err = mgr.enforcer.DeleteRole(request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeleteRole failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_ROLE_DELETE_FAILED, fmt.Sprintf("call enforcer DeleteRole failed"), err)
	}

	return
}

func (mgr *RBACManager) CreateRole(ctx context.Context, request message.CreateRoleReq, system bool) (resp message.CreateRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin CreateRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end CreateRole")

	if request.Role == "" {
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid input empty role")
	}

	if !system {
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("default role %s can not modify permission", request.Role))
		}
	}

	if _, err = mgr.enforcer.AddRoleForUser("", request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer AddRoleForUser failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_ROLE_CREATE_FAILED, fmt.Sprintf("call enforcer AddRoleForUser failed"), err)
	}
	return
}

func (mgr *RBACManager) BindRolesForUser(ctx context.Context, request message.BindRolesForUserReq) (resp message.BindRolesForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin BindRolesForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end BindRolesForUser")

	if len(request.Roles) == 0 || request.UserID == "" {
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid input empty userId or roles")
	}
	roles := mgr.enforcer.GetAllRoles()
	for _, role := range roles {
		if role == request.UserID {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("conflict userId %s with role", request.UserID))
		}
	}
	for _, reqRole := range request.Roles {
		exist := false
		for _, role := range roles {
			if reqRole == role {
				exist = true
				break
			}
		}
		if !exist {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("bind role %s not exist", reqRole))
		}
	}

	if _, err = mgr.enforcer.AddRolesForUser(request.UserID, request.Roles); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer AddRolesForUser failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_ROLE_BIND_FAILED, fmt.Sprintf("call enforcer AddRolesForUser failed"), err)
	}
	return resp, nil
}

func (mgr *RBACManager) UnbindRoleForUser(ctx context.Context, request message.UnbindRoleForUserReq) (resp message.UnbindRoleForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin UnBindRoleForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end UnBindRoleForUser")

	if _, err = mgr.enforcer.DeleteRoleForUser(request.UserID, request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeleteRoleForUser failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_ROLE_UNBIND_FAILED, "call enforcer DeleteRoleForUser failed", err)
	}
	return resp, nil
}

func (mgr *RBACManager) AddPermissionsForRole(ctx context.Context, request message.AddPermissionsForRoleReq, system bool) (resp message.AddPermissionsForRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin AddPermissionsForRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end AddPermissionsForRole")

	if request.Role == "" {
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, "invalid input empty role", err)
	}

	if !system {
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("default role %s can not modify permission", request.Role))
		}
	}

	var permissionList [][]string
	for _, permission := range request.Permissions {
		if !permission.CheckInvalid() {
			err = fmt.Errorf("permission %+v is invaild", permission)
			framework.LogWithContext(ctx).Errorf(err.Error())
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, err.Error())
		}
		permissionList = append(permissionList, []string{permission.Resource, permission.Action})
	}

	if _, err = mgr.enforcer.AddPermissionsForUser(request.Role, permissionList...); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer AddPermissionsForUser failed %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_PERMISSION_ADD_FAILED, "call enforcer AddPermissionsForUser failed", err)
	}
	return
}

func (mgr *RBACManager) DeletePermissionsForRole(ctx context.Context, request message.DeletePermissionsForRoleReq) (resp message.DeletePermissionsForRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeletePermissionsForRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end DeletePermissionsForRole")

	if request.Role == "" {
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, "invalid input empty role", err)
	}

	if _, ok := constants.RbacRoleMap[request.Role]; ok {
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("default role %s can not modify permission", request.Role))
	}

	for _, permission := range request.Permissions {
		if !permission.CheckInvalid() {
			err = fmt.Errorf("permission %+v is invaild", permission)
			framework.LogWithContext(ctx).Errorf(err.Error())
			return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, "input parameters invalid", err)
		}
	}
	for _, permission := range request.Permissions {
		if _, err = mgr.enforcer.DeletePermissionForUser(request.Role, permission.Resource, permission.Action); err != nil {
			framework.LogWithContext(ctx).Errorf("call enforcer DeletePermissionForUser failed %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_RBAC_PERMISSION_DELETE_FAILED, fmt.Sprintf("delete permissions of role %s failed", request.Role), err)
		}
	}

	return
}

func (mgr *RBACManager) QueryPermissionsForUser(ctx context.Context, request message.QueryPermissionsForUserReq) (resp message.QueryPermissionsForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin QueryPermissionsForUserResp, request: %+v", request)
	framework.LogWithContext(ctx).Info("end QueryPermissionsForUserResp")

	roles, err := mgr.enforcer.GetRolesForUser(request.UserID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer GetRolesForUser roles failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_RBAC_PERMISSION_QUERY_FAILED, fmt.Sprintf("query permissions of userId %s failed", request.UserID), err)
	}

	for _, role := range roles {
		rbacPermissions := mgr.enforcer.GetPermissionsForUser(role)
		framework.LogWithContext(ctx).Infof("call enforcer GetPermissionsForUser by role %s result %+v", role, rbacPermissions)

		resp.UserID = request.UserID
		for index := 0; index < len(rbacPermissions); index++ {
			permission := structs.RbacPermission{
				Resource: rbacPermissions[index][ResourceIndex],
				Action:   rbacPermissions[index][ActionIndex],
			}
			resp.Permissions = append(resp.Permissions, permission)
		}
	}

	return
}
