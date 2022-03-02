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
 * limitations under the License                                              *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: product.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/25
*******************************************************************************/

package product

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	pt "github.com/pingcap-inc/tiem/models/platform/product"
)

type ProductManager struct{}

func NewProductManager() *ProductManager {
	return &ProductManager{}
}

func (manager *ProductManager) CreateZones(ctx context.Context, request message.CreateZonesReq) (resp message.CreateZonesResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	var zones []pt.Zone
	for _, zone := range request.Zones {
		zones = append(zones, pt.Zone{VendorID: zone.VendorID, VendorName: zone.VendorName,
			RegionID: zone.RegionID, RegionName: zone.RegionName, ZoneID: zone.ZoneID,
			ZoneName: zone.ZoneName, Comment: zone.Comment})
	}
	err = rw.CreateZones(ctx, zones)
	if err != nil {
		log.Warningf("craete zones %v error %v", request.Zones, err)
		return
	}

	log.Debugf("create zones %v success", request.Zones)

	return
}

func (manager *ProductManager) DeleteZones(ctx context.Context, request message.DeleteZoneReq) (resp message.DeleteZoneResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	err = rw.DeleteZones(ctx, request.Zones)
	if err != nil {
		log.Warningf("delete zone %v error %v", request.Zones, err)
		return
	}

	log.Debugf("delete zone %v success", request.Zones)

	return
}

func (manager *ProductManager) QueryZones(ctx context.Context) (resp message.QueryAvailableRegionsResp, err error) {

	log := framework.LogWithContext(ctx)
	zones, queryError := models.GetProductReaderWriter().QueryZones(ctx)

	if queryError != nil {
		err = queryError
		log.Warningf("query all zones error %v", err)
		return
	}

	resp.Vendors = map[string]structs.VendorWithRegion{}
	for _, zone := range zones {
		vendor, vendorExisted := resp.Vendors[zone.VendorID]
		if !vendorExisted {
			vendor = structs.VendorWithRegion {
				VendorInfo: structs.VendorInfo{
					ID: zone.VendorID,
					Name: zone.VendorName,
				},
				Regions: map[string]structs.RegionInfo {},
			}
		}
		region, regionExisted := vendor.Regions[zone.RegionID]
		if !regionExisted {
			region = structs.RegionInfo {
				ID: zone.RegionID,
				Name: zone.RegionName,
			}
		}
		vendor.Regions[zone.RegionID] = region

		resp.Vendors[zone.VendorID] = vendor
	}
	log.Debugf("query all zones success")

	return
}

func (manager *ProductManager) QueryProducts(ctx context.Context, request message.QueryAvailableProductsReq) (resp message.QueryAvailableProductsResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	products, err := rw.QueryProducts(ctx, request.VendorID, constants.ProductStatus(request.Status), constants.EMInternalProduct(request.InternalProduct))
	if err != nil {
		log.Warningf("query all products error: %v,vendorID: %s,status: %s,internalProduct: %d",
			err, request.VendorID, request.Status, request.InternalProduct)
		return
	}

	log.Debugf("query all products success,vendorID: %s,status: %s,internalProduct: %d",
		request.VendorID, request.Status, request.InternalProduct)

	// region arch version
	resp.Products = make(map[string]map[string]map[string]map[string]structs.Product)
	for _, product := range products {
		addToProducts(resp.Products, product)
	}
	return resp, nil
}

func addToProducts(products map[string]map[string]map[string]map[string]structs.Product, product structs.Product) {
	if region, ok := products[product.RegionID]; ok {
		addToRegion(region, product)
	} else {
		products[product.RegionID] = make(map[string]map[string]map[string]structs.Product)
		addToRegion(products[product.RegionID], product)
	}
}

func addToRegion(region map[string]map[string]map[string]structs.Product, product structs.Product) {
	if productType, ok := region[product.ID]; ok {
		addToProductType(productType, product)
	} else {
		region[product.ID] = make(map[string]map[string]structs.Product)
		addToProductType(region[product.ID], product)
	}
}
func addToProductType(region map[string]map[string]structs.Product, product structs.Product) {
	if arch, ok := region[product.Arch]; ok {
		addToArch(arch, product)
	} else {
		region[product.Arch] = make(map[string]structs.Product)
		addToArch(region[product.Arch], product)
	}
}

func addToArch(arch map[string]structs.Product, product structs.Product) {
	arch[product.Version] = product
}

func (manager *ProductManager) CreateProduct(ctx context.Context, request message.CreateProductReq) (resp message.CreateProductResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	product := pt.Product{VendorID: request.ProductInfo.VendorID,
		RegionID:  request.ProductInfo.RegionID,
		ProductID: request.ProductInfo.ID,
		Version:   request.ProductInfo.Version,
		Arch:      request.ProductInfo.Arch,
		Name:      request.ProductInfo.Name,
		Status:    request.ProductInfo.Status,
		Internal:  request.ProductInfo.Internal,
	}
	var components []pt.ProductComponent
	for _, item := range request.Components {
		components = append(components, pt.ProductComponent{
			VendorID:                request.ProductInfo.VendorID,
			RegionID:                request.ProductInfo.RegionID,
			ProductID:               request.ProductInfo.ID,
			ProductVersion:          request.ProductInfo.Version,
			Arch:                    request.ProductInfo.Arch,
			ComponentID:             item.ID,
			Name:                    item.Name,
			Status:                  request.ProductInfo.Status,
			PurposeType:             item.PurposeType,
			StartPort:               item.StartPort,
			EndPort:                 item.EndPort,
			MaxPort:                 item.MaxPort,
			MinInstance:             item.MinInstance,
			MaxInstance:             item.MaxInstance,
		})
	}
	err = rw.CreateProduct(ctx, product, components)
	if err != nil {
		log.Warningf("craete product: %v, components: %v, error %v", request.ProductInfo, request.Components, err)
		return
	}

	log.Debugf("craete product: %v, components: %v success", request.ProductInfo, request.Components)

	return
}

func (manager *ProductManager) DeleteProduct(ctx context.Context, request message.DeleteProductReq) (resp message.DeleteProductResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	err = rw.DeleteProduct(ctx, request.ProductInfo)
	if err != nil {
		log.Warningf("delete product: %v error %v", request.ProductInfo, err)
		return
	}

	log.Debugf("delete product: %v sucess", request.ProductInfo)

	return
}

func (manager *ProductManager) QueryProductDetail(ctx context.Context, request message.QueryProductDetailReq) (resp message.QueryProductDetailResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()

	resp.Products, err = rw.QueryProductDetail(ctx, request.VendorID, request.RegionID, request.ProductID,
		constants.ProductStatus(request.Status), constants.EMInternalProduct(request.InternalProduct))
	if err != nil {
		log.Warningf("query product detail information error: %v ,vendorID: %s, regionID: %s, productID: %s, status: %s,internalProduct:%d",
			err, request.VendorID, request.RegionID, request.ProductID, request.Status, request.InternalProduct)
		return
	}

	log.Debugf("query product detail information success,vendorID: %s,status: %s,internalProduct:%d",
		request.VendorID, request.Status, request.InternalProduct)

	return resp, nil
}

func (manager *ProductManager) CreateSpecs(ctx context.Context, request message.CreateSpecsReq) (resp message.CreateSpecsResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	var sps []pt.Spec
	for _, spec := range request.Specs {
		sps = append(sps, pt.Spec{ID: spec.ID, Name: spec.Name, CPU: spec.CPU, Memory: spec.Memory,
			DiskType: spec.DiskType, PurposeType: spec.PurposeType, Status: spec.Status})
	}
	err = rw.CreateSpecs(ctx, sps)
	if err != nil {
		log.Warningf("craete specs %v error %v", request.Specs, err)
		return
	}

	log.Debugf("create specs %v success", request.Specs)

	return
}

func (manager *ProductManager) DeleteSpecs(ctx context.Context, request message.DeleteSpecsReq) (resp message.DeleteSpecsResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	err = rw.DeleteSpecs(ctx, request.SpecIDs)
	if err != nil {
		log.Warningf("delete specs %v error %v", request.SpecIDs, err)
		return
	}

	log.Debugf("delete sepecs %v success", request.SpecIDs)

	return
}

func (manager *ProductManager) QuerySpecs(ctx context.Context) (resp message.QuerySpecsResp, err error) {

	log := framework.LogWithContext(ctx)
	rw := models.GetProductReaderWriter()
	resp.Specs, err = rw.QuerySpecs(ctx)
	if err != nil {
		log.Warningf("query specs error %v", err)
		return
	}

	log.Debugf("query sepecs %v success", resp.Specs)

	return
}