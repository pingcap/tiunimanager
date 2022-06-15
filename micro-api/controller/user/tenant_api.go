/*
 * Copyright (c)  2022 PingCAP
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: tenant_api.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/7
*******************************************************************************/

package user

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// CreateTenant create tenant interface
// @Summary created  tenant
// @Description created tenant
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateTenantReq body message.CreateTenantReq true "create tenant request parameter"
// @Success 200 {object} controller.CommonResult{data=message.CreateTenantResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenants/ [post]
func CreateTenant(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CreateTenantReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateTenant, &message.CreateTenantResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteTenant delete tenant interface
// @Summary delete tenant
// @Description delete tenant
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param tenantId path string true "tenant id"
// @Param DeleteTenantReq body message.DeleteTenantReq true "delete tenant request parameter"
// @Success 200 {object} controller.CommonResult{data=message.DeleteTenantResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenants/{tenantId} [delete]
func DeleteTenant(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteTenantReq{
		ID: c.Param("tenantId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteTenant, &message.DeleteTenantResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetTenant get tenant profile interface
// @Summary get tenant profile
// @Description get tenant profile
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param tenantId path string true "tenant id"
// @Param GetTenantReq body message.GetTenantReq true "get tenant profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.GetTenantResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenant/{tenantId} [get]
func GetTenant(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.GetTenantReq{
		ID: c.Param("tenantId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetTenant, &message.GetTenantResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryTenants query all tenant profile interface
// @Summary queries all tenant profile
// @Description query all tenant profile
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryTenantReq body message.QueryTenantReq true "query tenant profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryTenantResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenant [get]
func QueryTenants(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryTenantReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryTenants, &message.QueryTenantResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UpdateTenantProfile update tenant profile interface
// @Summary update tenant profile
// @Description update tenant profile
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param tenantId path string true "tenant id"
// @Param UpdateTenantProfileReq body message.UpdateTenantProfileReq true "query tenant profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateTenantProfileResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenants/{tenantId}/update_profile [post]
func UpdateTenantProfile(c *gin.Context) {
	var req message.UpdateTenantProfileReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.UpdateTenantProfileReq).ID = c.Param("tenantId")
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateTenantProfile, &message.UpdateTenantProfileResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UpdateTenantOnBoardingStatus update tenant onboarding status interface
// @Summary update tenant onboarding status
// @Description update tenant onboarding status
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param tenantId path string true "tenant id"
// @Param UpdateTenantOnBoardingStatusReq body message.UpdateTenantOnBoardingStatusReq true "query tenant profile request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateTenantOnBoardingStatusResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /tenant [post]
func UpdateTenantOnBoardingStatus(c *gin.Context) {
	var req message.UpdateTenantOnBoardingStatusReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*message.UpdateTenantOnBoardingStatusReq).ID = c.Param("tenantId")
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateTenantOnBoardingStatus, &message.UpdateTenantOnBoardingStatusResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
