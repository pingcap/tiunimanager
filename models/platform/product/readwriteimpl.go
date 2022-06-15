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
 * @File: readwriteimpl.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"fmt"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
)

type ProductReadWrite struct {
	dbCommon.GormDB
}

func NewProductReadWrite(db *gorm.DB) *ProductReadWrite {
	return &ProductReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (p *ProductReadWrite) QueryAllVendors(ctx context.Context) ([]*Vendor, error) {
	vendors := make([]*Vendor, 0)
	return vendors, dbCommon.WrapDBError(p.DB(ctx).Model(&Vendor{}).Scan(&vendors).Error)
}

func (p *ProductReadWrite) SaveVendor(ctx context.Context, vendor *Vendor, zones []*VendorZone, specs []*VendorSpec) error {
	if vendor == nil || len(vendor.VendorID) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "vendor is empty")
	}
	existed := int64(0)
	err := p.DB(ctx).Model(&Vendor{}).Where("vendor_id = ?", vendor.VendorID).Count(&existed).Error

	// delete all infos of vendor, then rebuild it
	return dbCommon.WrapDBError(errors.OfNullable(err).BreakIf(func() error {
		if existed != 0 {
			return p.DeleteVendor(ctx, vendor.VendorID)
		} else {
			return nil
		}
	}).BreakIf(func() error {
		return p.DB(ctx).Create(vendor).Error
	}).BreakIf(func() error {
		return p.DB(ctx).CreateInBatches(&zones, len(zones)).Error
	}).BreakIf(func() error {
		return p.DB(ctx).CreateInBatches(&specs, len(specs)).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("save vendor info failed, vendor = %v, zones = %v, specs = %v, err = %s", vendor, zones, specs, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("save vendor info succeed, vendor = %v, zones = %v, specs = %v", vendor, zones, specs)
	}).Present())
}

func (p *ProductReadWrite) GetVendor(ctx context.Context, vendorID string) (vendor *Vendor, zones []*VendorZone, specs []*VendorSpec, err error) {
	if len(vendorID) == 0 {
		err = errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "vendorID is empty")
	}

	vendor = &Vendor{}
	zones = make([]*VendorZone, 0)
	specs = make([]*VendorSpec, 0)
	whereClause := fmt.Sprintf("vendor_id = '%s'", vendorID)
	err = errors.OfNullable(err).BreakIf(func() error {
		return p.DB(ctx).Model(&Vendor{}).Where(whereClause).First(vendor).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Model(&VendorZone{}).Where(whereClause).Find(&zones).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Model(&VendorSpec{}).Where(whereClause).Find(&specs).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("get vendor info failed, vendorID = %s, err = %s", vendorID, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("get vendor info succeed, vendorID = %s", vendorID)
	}).Present()
	return
}

func (p *ProductReadWrite) DeleteVendor(ctx context.Context, vendorID string) error {
	if len(vendorID) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "vendorID is empty")
	}
	whereClause := fmt.Sprintf("vendor_id = '%s'", vendorID)
	// delete from vendors、specs、zones
	return dbCommon.WrapDBError(errors.OfNullable(nil).BreakIf(func() error {
		return p.DB(ctx).Model(&Vendor{}).Where(whereClause).First(&Vendor{}).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&Vendor{}, whereClause).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&VendorSpec{}, whereClause).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&VendorZone{}, whereClause).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("delete vendor info failed, vendorID = %s, err = %s", vendorID, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("delete vendor info succeed, vendorID = %s", vendorID)
	}).Present())
}

func (p *ProductReadWrite) QueryAllProducts(ctx context.Context) ([]*ProductInfo, error) {
	products := make([]*ProductInfo, 0)
	return products, dbCommon.WrapDBError(p.DB(ctx).Model(&ProductInfo{}).Scan(&products).Error)
}

func (p *ProductReadWrite) SaveProduct(ctx context.Context, product *ProductInfo, versions []*ProductVersion, components []*ProductComponentInfo) error {
	if product == nil || len(product.ProductID) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "product is empty")
	}
	existed := int64(0)
	err := p.DB(ctx).Model(&ProductInfo{}).Where("product_id = ?", product.ProductID).Count(&existed).Error

	// delete all infos of product, then rebuild it
	return dbCommon.WrapDBError(errors.OfNullable(err).BreakIf(func() error {
		if existed != 0 {
			return p.DeleteProduct(ctx, product.ProductID)
		} else {
			return nil
		}
	}).BreakIf(func() error {
		return p.DB(ctx).Create(product).Error
	}).BreakIf(func() error {
		return p.DB(ctx).CreateInBatches(&versions, len(versions)).Error
	}).BreakIf(func() error {
		return p.DB(ctx).CreateInBatches(&components, len(components)).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("save product info failed, product = %v, versions = %v, components = %v, err = %s", product, versions, components, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("save product info succeed, product = %v, versions = %v, components = %v", product, versions, components)
	}).Present())
}

func (p *ProductReadWrite) GetProduct(ctx context.Context, productID string) (product *ProductInfo, versions []*ProductVersion, components []*ProductComponentInfo, err error) {
	if len(productID) == 0 {
		err = errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "productID is empty")
	}

	product = &ProductInfo{}
	versions = make([]*ProductVersion, 0)
	components = make([]*ProductComponentInfo, 0)
	whereClause := fmt.Sprintf("product_id = '%s'", productID)
	// get product, versions, components
	err = errors.OfNullable(err).BreakIf(func() error {
		return p.DB(ctx).Model(&ProductInfo{}).Where(whereClause).First(product).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Model(&ProductVersion{}).Where(whereClause).Find(&versions).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Model(&ProductComponentInfo{}).Where(whereClause).Find(&components).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("get product info failed, productID = %s, err = %s", productID, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("get product info succeed, productID = %s", productID)
	}).Present()
	return
}

func (p *ProductReadWrite) DeleteProduct(ctx context.Context, productID string) error {
	if len(productID) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "productID is empty")
	}
	whereClause := fmt.Sprintf("product_id = '%s'", productID)
	// delete from products, product_versions, product_components
	return dbCommon.WrapDBError(errors.OfNullable(nil).BreakIf(func() error {
		return p.DB(ctx).Model(&ProductInfo{}).Where(whereClause).First(&ProductInfo{}).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&ProductInfo{}, whereClause).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&ProductVersion{}, whereClause).Error
	}).BreakIf(func() error {
		return p.DB(ctx).Delete(&ProductComponentInfo{}, whereClause).Error
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("delete product info failed, productID = %s, err = %s", productID, err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("delete product info succeed, productID = %s", productID)
	}).Present())
}
