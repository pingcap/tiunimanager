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

// UpdateVendorZones update vendor zones
// @Summary update vendor zones
// @Description update vendor zones
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateVendorZonesReq body message.UpdateVendorZonesReq true "update vendor zones request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateVendorZonesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/zones [post]
func UpdateVendorZones(c *gin.Context) {

}

// QueryVendorZones query vendor zones
// @Summary query vendor zones
// @Description query vendor zones
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryVendorZonesReq query message.QueryVendorZonesReq true "query vendor zones request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryVendorZonesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/zones [get]
func QueryVendorZones(c *gin.Context) {

}

// UpdateVendorSpecs update vendor specs
// @Summary update vendor specs
// @Description update vendor specs
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateVendorSpecsReq body message.UpdateVendorSpecsReq true "update vendor specs request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateVendorSpecsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/specs [post]
func UpdateVendorSpecs(c *gin.Context) {

}

// QueryVendorSpecs query vendor specs
// @Summary query vendor specs
// @Description query vendor specs
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryVendorSpecsReq query message.QueryVendorSpecsReq true "query vendor specs request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryVendorSpecsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/specs [get]
func QueryVendorSpecs(c *gin.Context) {

}

// QueryVendorRegions query all regions information
// @Summary queries all regions information
// @Description queries all regions information
// @Tags vendor
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryAvailableRegionsReq query message.QueryAvailableRegionsReq true "query region request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryAvailableRegionsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /vendors/regions [get]
func QueryVendorRegions(c *gin.Context) {

}

// QueryProductComponents query product components
// @Summary query product components
// @Description query product components
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryProductComponentReq query message.QueryProductComponentReq true "query product component parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryProductComponentsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/components [get]
func QueryProductComponents(c *gin.Context) {
}

// UpdateProductComponents update product components
// @Summary update product components
// @Description update product components
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateProductComponentReq body message.UpdateProductComponentReq true "update product component properties parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateProductComponentResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/components [post]
func UpdateProductComponents(c *gin.Context) {
}

// QueryProductVersions query online products versions
// @Summary query online products versions
// @Description query online products versions
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryProductVersionsReq body message.QueryProductVersionsReq true "query online products request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryProductVersionsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/versions [get]
func QueryProductVersions(c *gin.Context) {

}

// UpdateProductVersions update online products versions
// @Summary update online products versions
// @Description update online products versions
// @Tags product
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateProductVersionsReq body message.UpdateProductVersionsReq true "product online request parameter"
// @Success 200 {object} controller.CommonResult{data=message.UpdateProductVersionsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/versions [post]
func UpdateProductVersions(c *gin.Context) {

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
// @Router /products/ [get]
func QueryAvailableProducts(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryAvailableProductsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProducts, &message.QueryAvailableProductsResp{},
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
