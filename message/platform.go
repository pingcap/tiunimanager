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
 ******************************************************************************/

/*******************************************************************************
 * @File: platform.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type GetSystemConfigReq struct {
	ConfigKey string `json:"configKey" form:"configKey"`
}

type GetSystemConfigResp struct {
	structs.SystemConfig
}

//CreateZonesReq create zone message, include vendor、region、zone
type CreateZonesReq struct {
	Zones []structs.ZoneInfo `json:"zones"`
}
type CreateZonesResp struct {
}

//QueryRegionsReq query all zone information message, include vendor、region、zone
type QueryRegionsReq struct {
}

type QueryRegionsResp struct {
	Vendors map[string]structs.VendorWithRegion `json:"vendors" form:"vendors"`
}

//DeleteZoneReq delete a zone message
type DeleteZoneReq struct {
	Zones []structs.ZoneInfo `json:"zones"`
}
type DeleteZoneResp struct {
}

//QueryZonesReq query all zone
type QueryZonesReq struct {
	VendorID string `json:"vendorID" form:"vendorID"`
	RegionID string `json:"regionID" form:"regionID"`
}

type QueryZonesResp struct {
	Zones []structs.ZoneInfo `json:"zones"`
}

//CreateProductReq create a product message
type CreateProductReq struct {
	ProductInfo structs.Product                    `json:"productInfo"`
	Components  []structs.ProductComponentProperty `json:"components"`
}

type CreateProductResp struct {
}

type UpdateComponentPropertiesReq struct {
	ProductID           string                             `json:"productID" swaggerignore:"true"`
	ComponentProperties []structs.ProductComponentProperty `json:"componentProperties"`
}

type UpdateComponentPropertiesResp struct {
}

type QueryComponentPropertiesReq struct {
	ProductID           string                             `json:"productID" swaggerignore:"true"`
}

type QueryComponentPropertiesResp struct {
	ProductID           string                             `json:"productID"`
	ComponentProperties []structs.ProductComponentProperty `json:"componentProperties"`
}

type UpdateOnlineProductsReq struct {
	ProductID           string                             `json:"productID" swaggerignore:"true"`
	SpecificVersionProducts []structs.SpecificVersionProduct `json:"specificVersionProducts"`
}

type UpdateOnlineProductsResp struct {
}

type QueryOnlineProductsReq struct {
	ProductID           string                             `json:"productID" swaggerignore:"true"`
}

type QueryOnlineProductsResp struct {
	ProductID           string                             `json:"productID"`
	SpecificVersionProducts []structs.SpecificVersionProduct `json:"specificVersionProducts"`
}

//DeleteProductReq delete a product message
type DeleteProductReq struct {
	ProductInfo structs.Product `json:"productInfo"`
}
type DeleteProductResp struct {
}

//QueryAvailableProductsReq query all products message
type QueryAvailableProductsReq struct {
	VendorID        string `json:"vendorId" form:"vendorId"`
	Status          string `json:"status" form:"status"`
	InternalProduct int    `json:"internalProduct" form:"internalProduct"`
}

type QueryAvailableProductsResp struct {
	// arch version
	Products map[string]map[string]map[string]map[string]structs.Product `json:"products"`
}

//QueryProductDetailReq query product detail message
type QueryProductDetailReq struct {
	VendorID        string `json:"vendorId" form:"vendorId"`
	RegionID        string `json:"regionId" form:"regionId"`
	ProductID       string `json:"productId" form:"productId"`
	Status          string `json:"status" form:"status"`
	InternalProduct int    `json:"internalProduct" form:"internalProduct"`
}
type QueryProductDetailResp struct {
	Products map[string]structs.ProductDetail `json:"products"`
}

// CreateSpecsReq component instance resource spec message
//CreateSpecsReq create spec message
type CreateSpecsReq struct {
	Specs []structs.SpecInfo `json:"specs"`
}

type CreateSpecsResp struct {
}

//DeleteSpecsReq delete spec message
type DeleteSpecsReq struct {
	SpecIDs []string `json:"specIds"`
}
type DeleteSpecsResp struct {
}

//QuerySpecsReq query spec message
type QuerySpecsReq struct {
}

type QuerySpecsResp struct {
	Specs []structs.SpecInfo `json:"specs"`
}

type GetSystemInfoReq struct {
	WithVersionDetail bool `json:"withVersionDetail" form:"withVersionDetail" `
}

type GetSystemInfoResp struct {
	Info           structs.SystemInfo        `json:"info"`
	CurrentVersion structs.SystemVersionInfo `json:"currentVersion"`
	LastVersion    structs.SystemVersionInfo `json:"lastVersion"`
}