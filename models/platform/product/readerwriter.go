/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: readerwriter.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import "context"

type ReaderWriter interface {
	//
	// QueryAllVendors
	//  @Description: query all vendors
	//  @param ctx
	//  @return []*Vendor
	//  @return error
	//
	QueryAllVendors(ctx context.Context) ([]*Vendor, error)
	//
	// SaveVendor
	// @Description: save all vendor info, including vendor name, zones, specs
	// @param ctx
	// @param vendor
	// @param zones
	// @param specs
	// @return error
	//
	SaveVendor(ctx context.Context, vendor *Vendor, zones []*VendorZone, specs []*VendorSpec) error
	//
	// GetVendor
	//  @Description: get all vendor info by vendorID, including vendor name, zones, specs
	//  @param ctx
	//  @param vendorID
	//  @return vendor
	//  @return zones
	//  @return specs
	//  @return err
	//
	GetVendor(ctx context.Context, vendorID string) (vendor *Vendor, zones []*VendorZone, specs []*VendorSpec, err error)

	//
	// DeleteVendor
	//  @Description: delete all vendor info by vendorID, including vendor name, zones, specs
	//  @param ctx
	//  @param vendorID
	//  @return error
	//
	DeleteVendor(ctx context.Context, vendorID string) error

	//
	// QueryAllProducts
	//  @Description:
	//  @param ctx
	//  @return []*ProductInfo
	//  @return error
	//
	QueryAllProducts(ctx context.Context) ([]*ProductInfo, error)
	//
	// SaveProduct
	//  @Description: save all product info, including product name, versions, components
	//  @param ctx
	//  @param product
	//  @param versions
	//  @param components
	//  @return error
	//
	SaveProduct(ctx context.Context, product *ProductInfo, versions []*ProductVersion, components []*ProductComponentInfo) error
	//
	// GetProduct
	//  @Description: get all product info by productID, including product name, versions, components
	//  @param ctx
	//  @param productID
	//  @return product
	//  @return versions
	//  @return components
	//  @return err
	//
	GetProduct(ctx context.Context, productID string) (product *ProductInfo, versions []*ProductVersion, components []*ProductComponentInfo, err error)
	//
	// DeleteProduct
	//  @Description: delete all product info by productID, including product name, versions, components
	//  @param ctx
	//  @param productID
	//  @return error
	//
	DeleteProduct(ctx context.Context, productID string) error
}
