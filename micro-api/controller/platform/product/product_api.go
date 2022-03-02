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

// CreateZones create zones interface
// @Summary created  zones
// @Description created  zones
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.CreateZonesReq true "create zones request parameter"
// @Success 200 {object} controller.CommonResult{data=message.CreateZonesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /zones/ [post]
func CreateZones(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CreateZonesReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateZones, &message.CreateZonesResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteZones delete zones
// @Summary deleted zones
// @Description deleted zones
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateZoneReq body message.DeleteZoneReq true "delete zone request parameter"
// @Success 200 {object} controller.CommonResult{data=message.DeleteZoneResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /zones/ [delete]
func DeleteZones(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DeleteZoneReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteZone, &message.DeleteZoneResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryRegions query all regions information
// @Summary queries all regions information
// @Description queries all regions information
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryRegionsReq query message.QueryRegionsReq true "query region request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryRegionsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /zones/regions [get]
func QueryRegions(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryRegionsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryZones, &message.QueryRegionsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryZones query zones
// @Summary query zones
// @Description query zones
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryZonesReq query message.QueryZonesReq true "query zones"
// @Success 200 {object} controller.CommonResult{data=message.QueryZonesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /zones/ [get]
func QueryZones(c *gin.Context) {
	// todo
}

// CreateSpecs create specs interface
// @Summary created  specs
// @Description created specs
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateSpecsReq body message.CreateSpecsReq true "create specs request parameter"
// @Success 200 {object} controller.CommonResult{data=message.CreateSpecsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /specs/ [post]
func CreateSpecs(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CreateSpecsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateSpecs, &message.CreateSpecsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteSpecs delete specs interface
// @Summary deleted  specs
// @Description deleted specs
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param DeleteSpecsReq body message.DeleteSpecsReq true "delete specs request parameter"
// @Success 200 {object} controller.CommonResult{data=message.DeleteSpecsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /specs/ [delete]
func DeleteSpecs(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DeleteSpecsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteSpecs, &message.DeleteSpecsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QuerySpecs query all specs information
// @Summary queries all specs information
// @Description queries all specs information
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QuerySpecsReq query message.QuerySpecsReq true "query specs request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QuerySpecsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /specs/ [get]
func QuerySpecs(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QuerySpecsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QuerySpecs, &message.QuerySpecsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryComponentProperties query product component properties
// @Summary query product component properties
// @Description query product component properties
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param productId path string true "product id"
// @Success 200 {object} controller.CommonResult{data=message.QueryComponentPropertiesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/{productID}/components [get]
func QueryComponentProperties(c *gin.Context) {
}

// UpdateComponentProperties update product component properties
// @Summary update product component properties
// @Description update product component properties
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateComponentPropertiesReq body message.UpdateComponentPropertiesReq true "update product component properties parameter"
// @Param productId path string true "product id"
// @Success 200 {object} controller.CommonResult{data=message.UpdateComponentPropertiesResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/{productID}/components [post]
func UpdateComponentProperties(c *gin.Context) {
}

// UpdateOnlineProducts product online
// @Summary  product online
// @Description product online
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param UpdateOnlineProductsReq body message.UpdateOnlineProductsReq true "product online request parameter"
// @Param productId path string true "product id"
// @Success 200 {object} controller.CommonResult{data=message.UpdateOnlineProductsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/{productID}/online [post]
func UpdateOnlineProducts(c *gin.Context) {

}

// QueryOnlineProducts query online products
// @Summary query online products
// @Description query online products
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param productId path string true "product id"
// @Param QueryOnlineProductsReq body message.QueryOnlineProductsReq true "query online products request parameter"
// @Success 200 {object} controller.CommonResult{data=message.QueryOnlineProductsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/{productID}/online [get]
func QueryOnlineProducts(c *gin.Context) {

}

// CreateProduct create product interface
// @Summary created product
// @Description created product
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateProductReq body message.CreateProductReq true "create product request parameter"
// @Success 200 {object} controller.CommonResult{data=message.CreateProductResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/ [post]
func CreateProduct(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.CreateProductReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateProduct, &message.CreateProductResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DeleteProduct delete product interface
// @Summary delete product
// @Description delete product
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param CreateProductReq body message.DeleteProductReq true "create product request parameter"
// @Success 200 {object} controller.CommonResult{data=message.DeleteProductResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/ [delete]
func DeleteProduct(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DeleteProductReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteProduct, &message.DeleteProductResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryProducts query all products' information
// @Summary queries all products' information
// @Description queries all products' information
// @Tags platform
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param QueryProducts query message.QueryProductsReq true "query products request"
// @Success 200 {object} controller.CommonResult{data=message.QueryProductsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /products/ [get]
func QueryProducts(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &message.QueryProductsReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryProducts, &message.QueryProductsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryProductDetail query all product detail
// @Summary query all product detail
// @Description query all product detail
// @Tags platform
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
