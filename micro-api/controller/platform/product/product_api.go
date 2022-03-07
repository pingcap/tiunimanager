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

package product

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// UpdateVendors update vendors
// @Summary update vendors
// @Description update vendors
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateVendorInfoReq body message.UpdateVendorInfoReq true "update vendor info request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateVendorInfoReq}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/ [post]
func UpdateVendors(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.UpdateVendorInfoReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateVendors, &message.UpdateVendorInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryVendors query vendors
// @Summary query vendors
// @Description query vendors
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryVendorInfoReq query message.QueryVendorInfoReq true "query vendor info request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryVendorInfoResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/ [get]
func QueryVendors(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryVendorInfoReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryVendors, &message.QueryVendorInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UpdateProducts update products
// @Summary update products
// @Description update products
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateProductsInfoReq body message.UpdateProductsInfoReq true "update products info request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateProductsInfoResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/ [post]
func UpdateProducts(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.UpdateProductsInfoReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateProducts, &message.UpdateProductsInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryProducts query products
// @Summary query products
// @Description query products
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryProductsInfoReq query message.QueryProductsInfoReq true "query products info request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryProductsInfoResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/ [get]
func QueryProducts(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryProductsInfoReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProducts, &message.QueryProductsInfoResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryAvailableVendors query available vendors and regions
// @Summary query available vendors and regions
// @Description query available vendors and regions
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryAvailableVendorsReq query message.QueryAvailableVendorsReq true "query available vendors parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryAvailableVendorsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/available [get]
func QueryAvailableVendors(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryAvailableVendorsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryAvailableVendors, &message.QueryAvailableVendorsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryAvailableProducts query all products' information
// @Summary queries all products' information
// @Description queries all products' information
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryAvailableProducts query message.QueryAvailableProductsReq true "query products request"
// @Success 200 {object} controller.CommonResult{data=message.QueryAvailableProductsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/available [get]
func QueryAvailableProducts(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryAvailableProductsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryAvailableProducts, &message.QueryAvailableProductsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryProductDetail query all product detail
// @Summary query all product detail
// @Description query all product detail
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryProductDetail query message.QueryProductDetailReq true "query product detail request"
// @Success 200 {object} controller.CommonResult{data=message.QueryProductDetailResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/detail [get]
func QueryProductDetail(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryProductDetailReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProductDetail, &message.QueryProductDetailResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
