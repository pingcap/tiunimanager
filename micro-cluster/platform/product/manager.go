/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 * @File: manager.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/product"
	"sync"
)

type Manager struct{}
var once sync.Once

var manager *Manager

func NewManager() *Manager {
	once.Do(func() {
		if manager == nil {
			manager = &Manager{}
		}
	})
	return manager
}

func (p *Manager) UpdateVendors(ctx context.Context, req message.UpdateVendorInfoReq) (resp message.UpdateVendorInfoResp, err error){
	err = models.Transaction(ctx, func(transactionCtx context.Context) error {
		for _, vendor := range req.Vendors {
			vendorInfo, zones, specs := convertVendorRequest(vendor)
			return models.GetProductReaderWriter().SaveVendor(transactionCtx, vendorInfo, zones, specs)
		}
		return nil
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update vendors failed, req = %v, err = %s", req, err.Error())
	} else {
		framework.LogWithContext(ctx).Infof("update vendors succeed, req = %v", req)
	}
	return
}

func (p *Manager) QueryVendors(ctx context.Context, req message.QueryVendorInfoReq) (resp message.QueryVendorInfoResp, err error){
	if len(req.VendorIDs) == 0 {
		err = errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "vendorIDs is empty")
		return
	}

	resp.Vendors = make([]structs.VendorConfigInfo, 0)
	for _, vendorID := range req.VendorIDs {
		vendorInfo, zones, specs, innerErr := models.GetProductReaderWriter().GetVendor(ctx, vendorID)
		if innerErr != nil {
			framework.LogWithContext(ctx).Errorf("get vendors failed, req = %v, err = %s", req, innerErr.Error())
			return
		}
		resp.Vendors = append(resp.Vendors, convertToVendorResponse(vendorInfo, zones, specs))
	}
	return
}

func (p *Manager) QueryAvailableVendors(ctx context.Context, req message.QueryAvailableVendorsReq) (message.QueryAvailableVendorsResp, error){
	return message.QueryAvailableVendorsResp{}, nil
}

func (p *Manager) UpdateProducts(ctx context.Context, req message.UpdateProductsInfoReq) (resp message.UpdateProductsInfoResp, err error){
	err = models.Transaction(ctx, func(transactionCtx context.Context) error {
		for _, product := range req.Products {
			productInfo, versions, components := convertProductRequest(product)
			return models.GetProductReaderWriter().SaveProduct(transactionCtx, productInfo, versions, components)
		}
		return nil
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update products failed, req = %v, err = %s", req, err.Error())
	} else {
		framework.LogWithContext(ctx).Infof("update products succeed, req = %v", req)
	}
	return
}

func (p *Manager) QueryProducts(ctx context.Context, req message.QueryProductsInfoReq) (resp message.QueryProductsInfoResp, err error){
	if len(req.ProductIDs) == 0 {
		err = errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "ProductIDs is empty")
		return
	}

	resp.Products = make([]structs.ProductConfigInfo, 0)
	for _, productId := range req.ProductIDs {
		productInfo, versions, components, innerErr := models.GetProductReaderWriter().GetProduct(ctx, productId)
		if innerErr != nil {
			framework.LogWithContext(ctx).Errorf("get products failed, req = %v, err = %s", req, innerErr.Error())
			return
		}
		resp.Products = append(resp.Products, convertToProductResponse(productInfo, versions, components))
	}
	return
}

func (p *Manager) QueryAvailableProducts(ctx context.Context, req message.QueryAvailableProductsReq) (message.QueryAvailableProductsResp, error){
	return message.QueryAvailableProductsResp{}, nil
}

func (p *Manager) QueryProductDetail(ctx context.Context, req message.QueryProductDetailReq) (message.QueryProductDetailResp, error){
	return message.QueryProductDetailResp{}, nil
}

func convertVendorRequest(reqConfig structs.VendorConfigInfo) (*product.Vendor, []*product.VendorZone, []*product.VendorSpec){
	vendorInfo := &product.Vendor {
		VendorID: reqConfig.ID,
		VendorName: reqConfig.Name,
	}
	zones := make([]*product.VendorZone, 0)
	for _, region := range reqConfig.Regions {
		for _, zone := range region.Zones {
			zones = append(zones, &product.VendorZone{
				VendorID: reqConfig.ID,
				RegionID: region.ID,
				RegionName: region.Name,
				ZoneID: zone.ZoneID,
				ZoneName: zone.ZoneName,
				Comment: zone.Comment,
			})
		}
	}
	specs := make([]*product.VendorSpec, 0)
	for _, spec := range reqConfig.Specs {
		specs = append(specs, &product.VendorSpec {
			VendorID: reqConfig.ID,
			SpecID: spec.ID,
			SpecName: spec.Name,
			CPU: spec.CPU,
			Memory: spec.Memory,
			DiskType: spec.DiskType,
			PurposeType: spec.PurposeType,
		})
	}
	return vendorInfo, zones, specs
}

func convertToVendorResponse(vendor *product.Vendor, zones []*product.VendorZone, specs []*product.VendorSpec) structs.VendorConfigInfo {
	return structs.VendorConfigInfo {
		VendorInfo: structs.VendorInfo {
			ID: vendor.VendorID,
			Name: vendor.VendorName,
		},
		Regions: convertToRegionResponse(zones),
		Specs:   convertToSpecResponse(specs),
	}
}

func convertToRegionResponse(zones []*product.VendorZone) []structs.RegionConfigInfo {
	regionMap := map[string]*structs.RegionConfigInfo{}
	for _, zone := range zones {
		if _, ok := regionMap[zone.RegionID]; !ok {
			regionMap[zone.RegionID] = &structs.RegionConfigInfo {
				RegionInfo: structs.RegionInfo {
					ID: zone.RegionID,
					Name: zone.RegionName,
				},
				Zones: make([]structs.ZoneInfo, 0),
			}
		}

		newZones := append(regionMap[zone.RegionID].Zones, structs.ZoneInfo{
			ZoneID: zone.ZoneID,
			ZoneName: zone.ZoneName,
			Comment: zone.Comment,
		})
		regionMap[zone.RegionID].Zones = newZones
	}
	result := make([]structs.RegionConfigInfo, 0)
	for _, r := range regionMap {
		result = append(result, *r)
	}
	return result
}

func convertToSpecResponse(specs []*product.VendorSpec) []structs.SpecInfo {
	result := make([]structs.SpecInfo, 0)
	for _, spec := range specs {
		result = append(result, structs.SpecInfo {
			ID: spec.SpecID,
			Name: spec.SpecName,
			CPU: spec.CPU,
			Memory: spec.Memory,
			DiskType: spec.DiskType,
			PurposeType: spec.PurposeType,
		})
	}
	return result
}

func convertProductRequest(reqConfig structs.ProductConfigInfo) (*product.ProductInfo, []*product.ProductVersion, []*product.ProductComponentInfo){
	productInfo := &product.ProductInfo{
		ProductID: reqConfig.ProductID,
		ProductName: reqConfig.ProductName,
	}
	versions := make([]*product.ProductVersion, 0)
	for _, version := range reqConfig.Versions {
		versions = append(versions, &product.ProductVersion {
			ProductID: reqConfig.ProductID,
			Arch: version.Arch,
			Version: version.Version,
		})
	}
	components := make([]*product.ProductComponentInfo, 0)
	for _, component := range reqConfig.Components {
		components = append(components, &product.ProductComponentInfo {
			ProductID: reqConfig.ProductID,
			ComponentID: component.ID,
			PurposeType: component.PurposeType,
			StartPort: component.StartPort,
			EndPort: component.EndPort,
			MaxPort: component.MaxPort,
			MinInstance: component.MinInstance,
			MaxInstance: component.MaxInstance,
			SuggestedInstancesCount: component.SuggestedInstancesCount,
		})
	}

	return productInfo, versions, components
}

func convertToProductResponse(productInfo *product.ProductInfo, versions []*product.ProductVersion, components []*product.ProductComponentInfo) structs.ProductConfigInfo {
	productConfigInfo := structs.ProductConfigInfo {
		ProductID: productInfo.ProductID,
		ProductName: productInfo.ProductName,
	}
	productConfigInfo.Components = make([]structs.ProductComponentPropertyWithZones, 0)
	for _, component := range components {
		productConfigInfo.Components = append(productConfigInfo.Components, structs.ProductComponentPropertyWithZones{
			ID: component.ComponentID,
			Name: component.ComponentName,
			PurposeType: component.PurposeType,
			StartPort: component.StartPort,
			EndPort: component.EndPort,
			MaxPort: component.MaxPort,
			MinInstance: component.MinInstance,
			MaxInstance: component.MaxInstance,
			SuggestedInstancesCount: component.SuggestedInstancesCount,
		})
	}
	productConfigInfo.Versions = make([]structs.SpecificVersionProduct, 0)
	for _, version := range versions {
		productConfigInfo.Versions = append(productConfigInfo.Versions, structs.SpecificVersionProduct{
			ProductID: version.ProductID,
			Arch: version.Arch,
			Version: version.Version,
		})
	}
	return productConfigInfo
}