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

type QueryVendorZonesReq struct {
	Vendors []string  `json:"vendors" form:"vendors"`
}

type QueryVendorZonesResp struct {
	Vendors map[string]structs.VendorWithRegion  `json:"vendors"`
}

type UpdateVendorZonesReq struct {
	Vendors map[string]structs.VendorWithRegion  `json:"vendors"`
}

type UpdateVendorZonesResp struct {
}

type QueryVendorSpecsReq struct {
	Vendors []string  `json:"vendors" form:"vendors"`
}

type QueryVendorSpecsResp struct {
	Vendors map[string][]structs.SpecInfo  `json:"vendors"`
}

type UpdateVendorSpecsReq struct {
	Vendors map[string][]structs.SpecInfo  `json:"vendors"`
}

type UpdateVendorSpecsResp struct {
}

type UpdateProductComponentReq struct {
	Products map[string][]structs.ProductComponentProperty `json:"products"`
}

type UpdateProductComponentResp struct {
}

type QueryProductComponentReq struct {
	Products []string  `json:"productIDs"`
}

type QueryProductComponentsResp struct {
	Products map[string][]structs.ProductComponentProperty `json:"products"`
}

type UpdateProductVersionsReq struct {
	Products map[string][]structs.SpecificVersionProduct `json:"products"`
}

type UpdateProductVersionsResp struct {
}

type QueryProductVersionsReq struct {
	Products []string  `json:"productIDs"`
}

type QueryProductVersionsResp struct {
	Products map[string][]structs.SpecificVersionProduct `json:"products"`
}

//QueryAvailableRegionsReq query all zone information message, include vendor、region、zone
type QueryAvailableRegionsReq struct {
}

type QueryAvailableRegionsResp struct {
	Vendors map[string][]structs.VendorWithRegion `json:"vendors" form:"vendors"`
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

type GetSystemConfigReq struct {
	ConfigKey string `json:"configKey" form:"configKey"`
}

type GetSystemConfigResp struct {
	structs.SystemConfig
}

type GetSystemInfoReq struct {
	WithVersionDetail bool `json:"withVersionDetail" form:"withVersionDetail" `
}

type GetSystemInfoResp struct {
	Info           structs.SystemInfo        `json:"info"`
	CurrentVersion structs.SystemVersionInfo `json:"currentVersion"`
	LastVersion    structs.SystemVersionInfo `json:"lastVersion"`
}