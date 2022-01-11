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
	enforcer *casbin.Enforcer
}

func NewRBACManager() *RBACManager {
	// casbin RBAC conf: https://github.com/casbin/casbin/blob/master/examples/rbac_model.conf
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")
	adapter, err := models.GetRBACReaderWriter().GetRBACAdapter(context.Background())
	if err != nil {
		framework.LogWithContext(context.Background()).Fatalf("get casbin gorm adapter failed, %s", err.Error())
		return nil
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		framework.LogWithContext(context.Background()).Fatalf("new casbin enforcer failed, %s", err.Error())
		return nil
	}
	if err := e.LoadPolicy(); err != nil {
		framework.LogWithContext(context.Background()).Fatalf("load rbac policy failed, %s", err.Error())
		return nil
	}
	return &RBACManager{
		enforcer: e,
	}
}

func (mgr *RBACManager) CheckPermissionForUser(ctx context.Context, request message.CheckPermissionForUserReq) (resp message.CheckPermissionForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin CheckPermissionForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end CheckPermissionForUser")

	for _, permission := range request.Permissions {
		result, err := mgr.enforcer.Enforce(request.UserID, permission.Resource, permission.Action)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("user %s check permission error: %s", request.UserID, err.Error())
			return message.CheckPermissionForUserResp{Result: result}, err
		}
		if !result {
			framework.LogWithContext(ctx).Infof("user %s check permission %+v failed", request.UserID, permission)
			return message.CheckPermissionForUserResp{Result: result}, nil
		}
	}
	return message.CheckPermissionForUserResp{Result: true}, nil
}

func (mgr *RBACManager) GetRoles(ctx context.Context, request message.GetRolesReq) (resp message.GetRolesResp, err error) {
	framework.LogWithContext(ctx).Infof("begin GetRoles, request: %+v", request)
	framework.LogWithContext(ctx).Info("end GetRoles")

	return message.GetRolesResp{
		Roles: mgr.enforcer.GetAllRoles(),
	}, nil
}

func (mgr *RBACManager) DeleteRole(ctx context.Context, request message.DeleteRoleReq) (resp message.DeleteRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeleteRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end DeleteRole")

	if _, err = mgr.enforcer.DeleteRole(request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeleteRole failed %s", err.Error())
	}
	return
}

func (mgr *RBACManager) DeleteUser(ctx context.Context, request message.DeleteUserReq) (resp message.DeleteUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeleteUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end DeleteUser")

	if _, err = mgr.enforcer.DeleteUser(request.UserID); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeleteUser failed %s", err.Error())
	}
	return
}

func (mgr *RBACManager) AddRoleForUser(ctx context.Context, request message.AddRoleForUserReq) (resp message.AddRoleForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin BindRoleForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end BindRoleForUser")

	//todo: check userId valid
	if request.Role == "" {
		return resp, fmt.Errorf("invalid input empty role")
	}
	/*
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, fmt.Errorf("default role %s can not modify permission", request.Role)
		}
	*/
	if _, err = mgr.enforcer.AddRoleForUser(request.UserID, request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer AddRoleForUser failed %s", err.Error())
	}
	return
}

func (mgr *RBACManager) DeleteRoleForUser(ctx context.Context, request message.DeleteRoleForUserReq) (resp message.DeleteRoleForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin UnBindRoleForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end UnBindRoleForUser")

	if _, err = mgr.enforcer.DeleteRoleForUser(request.UserID, request.Role); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer DeleteRoleForUser failed %s", err.Error())
	}
	return
}

func (mgr *RBACManager) AddPermissionsForRole(ctx context.Context, request message.AddPermissionsForRoleReq) (resp message.AddPermissionsForRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin AddPermissionsForRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end AddPermissionsForRole")

	if request.Role == "" {
		return resp, fmt.Errorf("invalid input empty role")
	}
	/*
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, fmt.Errorf("default role %s can not modify permission", request.Role)
		}
	*/

	var permissionList [][]string
	for _, permission := range request.Permissions {
		if !permission.CheckInvalid() {
			err = fmt.Errorf("permission %+v is invaild", permission)
			framework.LogWithContext(ctx).Errorf(err.Error())
			return
		}
		permissionList = append(permissionList, []string{permission.Resource, permission.Action})
	}

	if _, err = mgr.enforcer.AddPermissionsForUser(request.Role, permissionList...); err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer AddPermissionsForUser failed %s", err.Error())
	}
	return
}

func (mgr *RBACManager) DeletePermissionsForRole(ctx context.Context, request message.DeletePermissionsForRoleReq) (resp message.DeletePermissionsForRoleResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeletePermissionsForRole, request: %+v", request)
	framework.LogWithContext(ctx).Info("end DeletePermissionsForRole")

	if request.Role == "" {
		return resp, fmt.Errorf("invalid input empty role")
	}
	/*
		if _, ok := constants.RbacRoleMap[request.Role]; ok {
			return resp, fmt.Errorf("default role %s can not modify permission", request.Role)
		}
	*/

	for _, permission := range request.Permissions {
		if !permission.CheckInvalid() {
			err = fmt.Errorf("permission %+v is invaild", permission)
			framework.LogWithContext(ctx).Errorf(err.Error())
			return
		}
	}
	for _, permission := range request.Permissions {
		if _, err = mgr.enforcer.DeletePermissionForUser(request.Role, permission.Resource, permission.Action); err != nil {
			framework.LogWithContext(ctx).Errorf("call enforcer DeletePermissionForUser failed %s", err.Error())
			return
		}
	}

	return
}

func (mgr *RBACManager) GetPermissionsForUser(ctx context.Context, request message.GetPermissionsForUserReq) (resp message.GetPermissionsForUserResp, err error) {
	framework.LogWithContext(ctx).Infof("begin GetPermissionsForUser, request: %+v", request)
	framework.LogWithContext(ctx).Info("end GetPermissionsForUser")

	roles, err := mgr.enforcer.GetRolesForUser(request.UserID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call enforcer GetRolesForUser roles failed, %s", err.Error())
		return
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
