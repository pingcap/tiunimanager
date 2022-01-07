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
 *                                                                            *
 ******************************************************************************/

package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// AddRoleForUser
// @Summary add user with role
// @Description AddRoleForUser
// @Tags rbac AddRoleForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param addReq body message.AddRoleForUserReq true "AddRoleForUser request"
// @Success 200 {object} controller.CommonResult{data=message.AddRoleForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/user_role/add [post]
func AddRoleForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.AddRoleForUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.AddRoleForUser, &message.AddRoleForUserResp{},
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

// DeleteRbacUser
// @Summary delete rbac user
// @Description DeleteRbacUser
// @Tags rbac DeleteRbacUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "rbac userId"
// @Success 200 {object} controller.CommonResult{data=message.DeleteUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/user/{userId} [delete]
func DeleteRbacUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteUserReq{
		UserID: c.Param("userId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteRbacUser, &message.DeleteUserResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteRoleForUser
// @Summary unbind rbac role from user
// @Description DeleteRoleForUser
// @Tags rbac DeleteRoleForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param deleteRoleForUserReq body message.DeleteRoleForUserReq true "DeleteRoleForUser request"
// @Success 200 {object} controller.CommonResult{data=message.DeleteRoleForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/user_role/delete [delete]
func DeleteRoleForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DeleteRoleForUserReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteRoleForUser, &message.DeleteRoleForUserResp{},
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

// GetPermissionsForUser
// @Summary get permissions of user
// @Description GetPermissionsForUser
// @Tags rbac GetPermissionsForUser
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "rbac userId"
// @Success 200 {object} controller.CommonResult{data=message.GetPermissionsForUserResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /rbac/permission/{userId} [get]
func GetPermissionsForUser(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.GetPermissionsForUserReq{
		UserID: c.Param("userId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetPermissionsForUser, &message.GetPermissionsForUserResp{},
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
