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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/product"
	"sort"
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
			if innerError := models.GetProductReaderWriter().SaveVendor(transactionCtx, vendorInfo, zones, specs);innerError != nil{
				return innerError
			}
		}
		if innerError := models.GetSystemReaderWriter().VendorInitialized(transactionCtx); innerError != nil {
			return innerError
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
	vendorIDs := req.VendorIDs
	// empty vendorIDs means query all vendors
	if len(vendorIDs) == 0 {
		vendors, innerErr := models.GetProductReaderWriter().QueryAllVendors(ctx)
		if innerErr != nil {
			err = innerErr
			return
		}
		for _, v := range vendors {
			vendorIDs = append(vendorIDs, v.VendorID)
		}
	}

	resp.Vendors = make([]structs.VendorConfigInfo, 0)
	for _, vendorID := range vendorIDs {
		vendorInfo, zones, specs, innerErr := models.GetProductReaderWriter().GetVendor(ctx, vendorID)
		if innerErr != nil {
			framework.LogWithContext(ctx).Errorf("get vendors failed, req = %v, err = %s", req, innerErr.Error())
			err = innerErr
			return
		}
		resp.Vendors = append(resp.Vendors, convertToVendorResponse(vendorInfo, zones, specs))
	}
	return
}

func (p *Manager) QueryAvailableVendors(ctx context.Context, req message.QueryAvailableVendorsReq) (resp message.QueryAvailableVendorsResp, err error){
	vendorInfos, queryVendorsError := models.GetProductReaderWriter().QueryAllVendors(ctx)
	if queryVendorsError != nil {
		framework.LogWithContext(ctx).Errorf("query available vendors failed, err = %s", queryVendorsError.Error())
		return resp, queryVendorsError
	}

	resp.Vendors = make(map[string]structs.VendorWithRegion)
	for _, v := range vendorInfos {
		vendorInfo, zones, specs, getVendorError := models.GetProductReaderWriter().GetVendor(ctx, v.VendorID)
		if getVendorError != nil {
			framework.LogWithContext(ctx).Errorf("query available vendors failed, err = %s", getVendorError.Error())
			return resp, getVendorError
		}
		// ignore empty vendor
		if len(zones) == 0 || len(specs) == 0 {
			continue
		} else {
			resp.Vendors[v.VendorID] = structs.VendorWithRegion {
				VendorInfo: structs.VendorInfo{
					ID: vendorInfo.VendorID,
					Name: vendorInfo.VendorName,
				},
				Regions: map[string]structs.RegionInfo{},
			}
		}
		for _, z := range zones {
			resp.Vendors[v.VendorID].Regions[z.RegionID] = structs.RegionInfo{
				ID: z.RegionID,
				Name: z.RegionName,
			}
		}
	}
	return
}

func (p *Manager) UpdateProducts(ctx context.Context, req message.UpdateProductsInfoReq) (resp message.UpdateProductsInfoResp, err error){
	err = models.Transaction(ctx, func(transactionCtx context.Context) error {
		for _, product := range req.Products {
			productInfo, versions, components := convertProductRequest(product)
			if innerError := models.GetProductReaderWriter().SaveProduct(transactionCtx, productInfo, versions, components); innerError != nil {
				return innerError
			}
		}
		if innerError := models.GetSystemReaderWriter().ProductInitialized(transactionCtx); innerError != nil {
			return innerError
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
	productIDs := req.ProductIDs
	// empty productIDs means query all products
	if len(productIDs) == 0 {
		products, innerErr := models.GetProductReaderWriter().QueryAllProducts(ctx)
		if innerErr != nil {
			err = innerErr
			return
		}
		for _, v := range products {
			productIDs = append(productIDs, v.ProductID)
		}
	}

	resp.Products = make([]structs.ProductConfigInfo, 0)
	for _, productId := range productIDs {
		productInfo, versions, components, innerErr := models.GetProductReaderWriter().GetProduct(ctx, productId)
		if innerErr != nil {
			framework.LogWithContext(ctx).Errorf("get products failed, req = %v, err = %s", req, innerErr.Error())
			err = innerErr
			return
		}
		resp.Products = append(resp.Products, convertToProductResponse(productInfo, versions, components))
	}
	return
}

func (p *Manager) QueryAvailableProducts(ctx context.Context, req message.QueryAvailableProductsReq) (resp message.QueryAvailableProductsResp, err error){
	productResp, err := p.QueryProducts(ctx, message.QueryProductsInfoReq {
		ProductIDs: []string{},
	})

	if err != nil {
		return
	}

	vendorResp, err := p.QueryVendors(ctx, message.QueryVendorInfoReq{VendorIDs: []string{req.VendorID}})
	if err != nil {
		return
	}
	
	productsMap := make(map[string]map[string]map[string]map[string]structs.Product)

	for _, vendor := range vendorResp.Vendors {
		for _, region := range vendor.Regions {
			productsMap[region.ID] = make(map[string]map[string]map[string]structs.Product)
			for _, product := range productResp.Products {
				productsMap[region.ID][product.ProductID] = make(map[string]map[string]structs.Product)
				for _, version := range product.Versions {
					if _, ok := productsMap[region.ID][product.ProductID][version.Arch]; !ok {
						productsMap[region.ID][product.ProductID][version.Arch] = make(map[string]structs.Product)
					}
					if _, ok := productsMap[region.ID][product.ProductID][version.Arch][version.Version]; !ok {
						productsMap[region.ID][product.ProductID][version.Arch][version.Version] = structs.Product{
							ID:         product.ProductID,
							Name:       product.ProductName,
							Arch:       version.Arch,
							Version:    version.Version,
							RegionID:   region.ID,
							RegionName: region.Name,
							VendorID:   vendor.ID,
							VendorName: vendor.Name,
							Status:     string(constants.ProductStatusOnline),
							Internal:   constants.EMInternalProductNo,
						}
					}
				}
			}
		}
	}

	resp.Products = productsMap
	return
}

func (p *Manager) QueryProductDetail(ctx context.Context, req message.QueryProductDetailReq) (resp message.QueryProductDetailResp, err error){
	productResp, err := p.QueryProducts(ctx, message.QueryProductsInfoReq {
		ProductIDs: []string{req.ProductID},
	})
	if err != nil || len(productResp.Products) == 0 {
		return resp, errors.WrapError(errors.TIEM_UNSUPPORT_PRODUCT, "query products failed", err)
	}

	vendorResp, err := p.QueryVendors(ctx, message.QueryVendorInfoReq{
		VendorIDs: []string{req.VendorID},
	})
	if err != nil || len(vendorResp.Vendors) == 0 {
		return resp, errors.WrapError(errors.TIEM_UNSUPPORT_PRODUCT, "query vendors failed", err)
	}

	products := make(map[string]structs.ProductDetail)

	product := productResp.Products[0]
	vendor := vendorResp.Vendors[0]
	var region structs.RegionConfigInfo
	for _, r := range vendor.Regions {
		if r.ID == req.RegionID {
			region = r
		}
	}
	if region.ID != req.RegionID {
		return resp, errors.WrapError(errors.TIEM_UNSUPPORT_PRODUCT, "query region failed", nil)
	}
	products[product.ProductID] = structs.ProductDetail {
		ID: product.ProductID,
		Name: product.ProductName,
		Versions: make(map[string]structs.ProductVersion),
	}

	productComponentPropertyWithZones := make([]structs.ProductComponentPropertyWithZones, 0)
	for _, component := range product.Components {
		componentInstanceZoneWithSpecs := make([]structs.ComponentInstanceZoneWithSpecs, 0)
		for _, zone := range region.Zones {
			componentInstanceZoneWithSpecs = append(componentInstanceZoneWithSpecs, structs.ComponentInstanceZoneWithSpecs {
				ZoneID: zone.ZoneID,
				ZoneName: zone.ZoneName,
				Specs: convertComponentInstanceResourceSpecs(vendor, component, zone.ZoneID, zone.ZoneName),
			})
		}
		productComponentPropertyWithZones = append(productComponentPropertyWithZones, structs.ProductComponentPropertyWithZones {
			ID: component.ID,
			Name: component.Name,
			PurposeType: component.PurposeType,
			StartPort: component.StartPort,
			EndPort: component.EndPort,
			MaxPort: component.MaxPort,
			MinInstance: component.MinInstance,
			MaxInstance: component.MaxInstance,
			SuggestedInstancesCount: component.SuggestedInstancesCount,
			AvailableZones: componentInstanceZoneWithSpecs,
		})
	}

	sort.Slice(productComponentPropertyWithZones, func(i, j int) bool {
		return constants.EMProductComponentIDType(productComponentPropertyWithZones[i].ID).SortWeight() > constants.EMProductComponentIDType(productComponentPropertyWithZones[j].ID).SortWeight()
	})

	for _, version := range product.Versions {
		if _, ok := products[product.ProductID].Versions[version.Version]; !ok {
			products[product.ProductID].Versions[version.Version] = structs.ProductVersion {
				Version: version.Version,
				Arch: make(map[string][]structs.ProductComponentPropertyWithZones),
			}
		}
		if _, ok := products[product.ProductID].Versions[version.Version].Arch[version.Arch]; !ok {
			products[product.ProductID].Versions[version.Version].Arch[version.Arch] = productComponentPropertyWithZones
		}
	}

	resp.Products = products
	return
}

func convertComponentInstanceResourceSpecs(vendor structs.VendorConfigInfo, component structs.ProductComponentPropertyWithZones, zoneID, zoneName string) []structs.ComponentInstanceResourceSpec {
	componentInstanceResourceSpecs := make([]structs.ComponentInstanceResourceSpec, 0)
	for _, spec := range vendor.Specs {
		if spec.PurposeType == component.PurposeType {
			componentInstanceResourceSpecs = append(componentInstanceResourceSpecs, structs.ComponentInstanceResourceSpec{
				ID: spec.ID,
				Name: spec.Name,
				CPU: spec.CPU,
				Memory: spec.Memory,
				DiskType: spec.DiskType,
				ZoneID: zoneID,
				ZoneName: zoneName,
			})
		}
	}
	return componentInstanceResourceSpecs
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