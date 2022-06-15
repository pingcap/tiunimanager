/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// QueryRbacRoles
// @Summary query rbac roles
// @Description QueryRoles
// @Tags rbac QueryRoles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role path string true "rbac role"
// @Success 200 {object} controller.CommonResult{data=message.QueryRolesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/role/ [get]
func QueryRbacRoles(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.QueryRolesReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryRoles, &message.QueryRolesResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// CreateRbacRole
// @Summary create rbac role
// @Description CreateRbacRole
// @Tags rbac CreateRbacRole
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param createReq body message.CreateRoleReq true "CreateRole request"
// @Success 200 {object} controller.CommonResult{data=message.CreateRoleResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/role/ [post]
func CreateRbacRole(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CreateRoleReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateRbacRole, &message.CreateRoleResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteRbacRole
// @Summary delete rbac role
// @Description DeleteRbacRole
// @Tags rbac DeleteRbacRole
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role path string true "rbac role"
// @Success 200 {object} controller.CommonResult{data=message.DeleteRoleResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/role/{role} [delete]
func DeleteRbacRole(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteRoleReq{
		Role: c.Param("role"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteRbacRole, &message.DeleteRoleResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// BindRolesForUser
// @Summary bind user with roles
// @Description BindRolesForUser
// @Tags rbac BindRolesForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bindReq body message.BindRolesForUserReq true "BindRolesForUser request"
// @Success 200 {object} controller.CommonResult{data=message.BindRolesForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/role/bind [post]
func BindRolesForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.BindRolesForUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.BindRolesForUser, &message.BindRolesForUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UnbindRoleForUser
// @Summary unbind rbac role from user
// @Description UnbindRoleForUser
// @Tags rbac UnbindRoleForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param deleteRoleForUserReq body message.UnbindRoleForUserReq true "UnbindRoleForUser request"
// @Success 200 {object} controller.CommonResult{data=message.UnbindRoleForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/user_role/delete [delete]
func UnbindRoleForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.UnbindRoleForUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UnbindRoleForUser, &message.UnbindRoleForUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// AddPermissionsForRole
// @Summary add permissions for role
// @Description AddPermissionsForRole
// @Tags rbac AddPermissionsForRole
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param addPermissionsForRoleReq body message.AddPermissionsForRoleReq true "AddPermissionsForRole request"
// @Success 200 {object} controller.CommonResult{data=message.AddPermissionsForRoleResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/permission/add [post]
func AddPermissionsForRole(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.AddPermissionsForRoleReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.AddPermissionsForRole, &message.AddPermissionsForRoleResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeletePermissionsForRole
// @Summary delete permissions for role
// @Description DeletePermissionsForRole
// @Tags rbac DeletePermissionsForRole
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param deletePermissionsForRoleReq body message.DeletePermissionsForRoleReq true "DeleteRoleForUser request"
// @Success 200 {object} controller.CommonResult{data=message.DeletePermissionsForRoleResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/permission/delete [delete]
func DeletePermissionsForRole(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DeletePermissionsForRoleReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeletePermissionsForRole, &message.DeletePermissionsForRoleResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryPermissionsForUser
// @Summary query permissions of user
// @Description QueryPermissionsForUser
// @Tags rbac QueryPermissionsForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "rbac userId"
// @Success 200 {object} controller.CommonResult{data=message.QueryPermissionsForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/permission/{userId} [get]
func QueryPermissionsForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.QueryPermissionsForUserReq{
		UserID: c.Param("userId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryPermissionsForUser, &message.QueryPermissionsForUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// CheckPermissionForUser
// @Summary check permissions of user
// @Description CheckPermissionForUser
// @Tags rbac CheckPermissionForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param checkPermissionForUserReq body message.CheckPermissionForUserReq true "CheckPermissionForUser request"
// @Success 200 {object} controller.CommonResult{data=message.CheckPermissionForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/permission/check [post]
func CheckPermissionForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CheckPermissionForUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CheckPermissionForUser, &message.CheckPermissionForUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
