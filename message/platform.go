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

//UpdateVendorInfoReq update vendor info request
type UpdateVendorInfoReq struct {
	Vendors []structs.VendorConfigInfo `json:"vendors"`
}

//UpdateVendorInfoResp update vendor info response
type UpdateVendorInfoResp struct {
}

//QueryVendorInfoReq query vendor info request
type QueryVendorInfoReq struct {
	VendorIDs []string `json:"vendorIDs"`
}

//QueryVendorInfoResp query vendor info response
type QueryVendorInfoResp struct {
	Vendors []structs.VendorConfigInfo `json:"vendors"`
}

//UpdateProductsInfoReq update product info request
type UpdateProductsInfoReq struct {
	Products []structs.ProductConfigInfo `json:"products"`
}

//UpdateProductsInfoResp update product info response
type UpdateProductsInfoResp struct {
}

//QueryProductsInfoReq query product info request
type QueryProductsInfoReq struct {
	ProductIDs []string `json:"productIDs"`
}

//QueryProductsInfoResp query product info response
type QueryProductsInfoResp struct {
	Products []structs.ProductConfigInfo `json:"products"`
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