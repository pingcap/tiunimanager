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
 * @File: product_read_writer.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package product

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type ProductReadWriterInterface interface {
	// QueryZones
	// @Description: Query region & zone information provided by Enterprise Manager
	// @param ctx
	// @return zones
	// @return err
	QueryZones(ctx context.Context) (zones []structs.ZoneInfo, err error)

	// CreateZones
	// @Description: batch create zones interface
	// @param ctx
	// @param zones
	// @return err
	CreateZones(ctx context.Context, zones []Zone) (er error)

	// DeleteZones
	// @Description: batch delete zones
	// @param ctx
	// @param zones
	// @return err
	DeleteZones(ctx context.Context, zones []structs.ZoneInfo) (er error)

	// QueryProductDetail
	// @Description: Query the product detail information provided by EM
	// @param ctx
	// @param vendorID
	// @param regionID
	// @param productID
	// @param status
	// @return products
	// @return err
	QueryProductDetail(ctx context.Context, vendorID, regionID, productID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products map[string]structs.ProductDetail, er error)

	// QueryProducts
	// @Description: Query all product information by vendor_id,region_id
	// @param ctx
	// @param regionID
	// @param status
	// @return err
	// @return products
	QueryProducts(ctx context.Context, vendorID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products []structs.Product, err error)

	// CreateProduct
	// @Description: create a product by product information and components
	// @param ctx
	// @param product
	// @param components
	// @return err
	CreateProduct(ctx context.Context, product Product, components []ProductComponent) (err error)

	// DeleteProduct
	// @Description: delete a product by product information
	// @param ctx
	// @param product
	// @return err
	DeleteProduct(ctx context.Context, product structs.Product) (err error)

	// QueryProductComponentProperty
	// @Description: Query the properties of a product component
	// @param ctx
	// @param productID
	// @param productVersion
	// @return err
	// @return property[]
	QueryProductComponentProperty(ctx context.Context, vendorID, regionID, productID, productVersion, arch string, productStatus constants.ProductComponentStatus) (property []structs.ProductComponentProperty, er error)

	// CreateSpecs
	// @Description: batch create specs
	// @param ctx
	// @param specs
	// @return err
	CreateSpecs(ctx context.Context, specs []Spec) (er error)

	// DeleteSpecs batch delete specs
	// @param ctx
	// @param specs
	// @return err
	DeleteSpecs(ctx context.Context, specs []string) (er error)

	// QuerySpecs query all specs
	// @return specs
	// @return err
	QuerySpecs(ctx context.Context) (specs []structs.SpecInfo, er error)
}

func NewProductReadWriter(db *gorm.DB) *ProductReadWriter {
	m := &ProductReadWriter{
		dbCommon.WrapDB(db),
	}
	return m
}

type ProductReadWriter struct {
	dbCommon.GormDB
}

//CreateZones batch create zones
func (p *ProductReadWriter) CreateZones(ctx context.Context, zones []Zone) (er error) {
	if len(zones) <= 0 {
		framework.LogWithContext(ctx).Warningf("create zones %v, len: %d, parameter invalid", zones, len(zones))
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "create zones %v, parameter invalid", zones)
	}

	db := p.DB(ctx).CreateInBatches(zones, 1024)
	if db.Error != nil {
		framework.LogWithContext(ctx).Errorf("create zones %v, errors: %v", zones, db.Error)
		return errors.NewErrorf(errors.CreateZonesError, "create zones %v, error: %v", zones, db.Error)
	}
	return nil
}

//DeleteZones batch delete zones
func (p *ProductReadWriter) DeleteZones(ctx context.Context, zones []structs.ZoneInfo) (er error) {
	if len(zones) <= 0 {
		framework.LogWithContext(ctx).Warningf("delete zones %v, len: %d, parameter invalid", zones, len(zones))
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "delete zones %v, parameter invalid", zones)
	}
	p.DB(ctx).Begin()
	for _, zone := range zones {
		err := p.DB(ctx).Where("vendor_id = ? AND  region_id = ? AND zone_id = ?", zone.VendorID, zone.RegionID, zone.ZoneID).Unscoped().Delete(&Zone{}).Error
		if err != nil {
			framework.LogWithContext(ctx).Errorf("delete zone, VendorID:%s RegeionID: %s, ZoneID:%s, errors: %v", zone.VendorID, zone.RegionID, zone.ZoneID, err)
			p.DB(ctx).Rollback()
			return errors.NewErrorf(errors.DeleteZonesError, "delete zone, VendorID:%s RegionID: %s, ZoneID:%s, errors: %v", zone.VendorID, zone.RegionID, zone.ZoneID, err)
		}
	}
	p.DB(ctx).Commit()
	return nil
}

//QueryZones Query vendor & region & zone information provided by Enterprise Manager
func (p *ProductReadWriter) QueryZones(ctx context.Context) (zones []structs.ZoneInfo, err error) {
	var info structs.ZoneInfo
	SQL := "SELECT t1.zone_id,t1.zone_name,t1.region_id,t1.region_name,t1.vendor_id,t1.vendor_name FROM zones t1;"
	rows, err := p.DB(ctx).Raw(SQL).Rows()
	defer rows.Close()
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "query all zones error: %v, SQL: %s", err, SQL)
	}
	for rows.Next() {
		err = rows.Scan(&info.ZoneID, &info.ZoneName, &info.RegionID, &info.RegionName, &info.VendorID, &info.VendorName)
		if err == nil {
			zones = append(zones, info)
		} else {
			return nil, errors.NewErrorf(errors.QueryZoneScanRowError, "query all zones, scan data error: %v, SQL: %s", err, SQL)
		}
	}
	return zones, err
}

//CreateProduct create a product by product information and components
func (p *ProductReadWriter) CreateProduct(ctx context.Context, product Product, components []ProductComponent) (er error) {
	if len(components) <= 0 {
		framework.LogWithContext(ctx).Warningf("create product invalid parameter,product: %v, components: %v", product, components)
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "create product invalid parameter, product: %v, components: %v",
			product, components)
	}
	err := p.DB(ctx).Create(&product).Error
	if err != nil {
		p.DB(ctx).Rollback()
		framework.LogWithContext(ctx).Warningf("create product, insert product information failed,product: %v, components: %v", product, components)
		return errors.NewErrorf(errors.CreateProductError, "create product, insert product information into product table failed, rollback,product: %v, components: %v", product, components)
	}
	for _, component := range components {
		err = p.DB(ctx).Create(&component).Error
		if err != nil {
			p.DB(ctx).Rollback()
			framework.LogWithContext(ctx).Warningf("create product, insert component information into components table failed,product: %v, components: %v", product, component)
			return errors.NewErrorf(errors.CreateProductError, "create product, insert component information into components table failed, rollback,product: %v, components: %v", product, component)
		}
	}
	p.DB(ctx).Commit()

	return nil
}

//DeleteProduct delete a product by product information
func (p *ProductReadWriter) DeleteProduct(ctx context.Context, product structs.Product) (er error) {
	if "" == product.ID || "" == product.VendorID || "" == product.RegionID ||
		"" == product.Version || "" == product.Arch {
		framework.LogWithContext(ctx).Warningf("delete product invalid parameter,product: %v", product)
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "delete product invalid parameter, product: %v", product)
	}
	components, err := p.QueryProductComponentProperty(ctx, product.VendorID, product.RegionID, product.ID, product.Version, product.Arch, constants.ProductComponentStatus(product.Status))
	if err != nil || len(components) == 0 {
		framework.LogWithContext(ctx).Warningf("delete product, query product component form product_component table failed,product: %v,components: %v", product, components)
		return errors.NewErrorf(errors.DeleteProductError, "delete product, query product component form product_component table failed,product: %v,components:%v", product, components)
	}

	p.DB(ctx).Begin()
	err = p.DB(ctx).Where("vendor_id = ? AND  region_id = ? AND product_id = ? AND version = ? AND arch = ?",
		product.VendorID, product.RegionID, product.ID, product.Version, product.Arch).Unscoped().Delete(&Product{}).Error
	framework.LogWithContext(ctx).Warningf("delete product, delete product information form product table failed,product: %v", product)
	if err != nil {
		p.DB(ctx).Rollback()
		framework.LogWithContext(ctx).Warningf("delete product, delete product information form product table failed,product: %v", product)
		return errors.NewErrorf(errors.DeleteProductError, "delete product, delete product information form product table failed,product: %v", product)
	}

	for _, component := range components {
		err = p.DB(ctx).Where("vendor_id = ? AND  region_id = ? AND product_id = ? AND product_version = ? AND arch = ? AND component_id AND name = ?",
			product.VendorID, product.RegionID, product.ID, product.Version, product.Arch, component.ID, component.Name).Unscoped().Delete(&ProductComponent{}).Error
		if err != nil {
			p.DB(ctx).Rollback()
			framework.LogWithContext(ctx).Warningf("delete product, delete product component form product_component table failed,product: %v, component: %v", product, component)
			return errors.NewErrorf(errors.DeleteProductError, "delete product, delete product component form product_component table failed,product: %v, component: %v", product, component)
		}
	}
	p.DB(ctx).Commit()

	return nil
}

//QueryProducts Query the all product information provided by Enterprise Manager
func (p *ProductReadWriter) QueryProducts(ctx context.Context, vendorID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products []structs.Product, er error) {
	if "" == vendorID || "" == status {
		framework.LogWithContext(ctx).Warningf("query product invalid parameter, vendorID: %s, status: %s, internal: %d",
			vendorID, status, internal)
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "query product invalid parameter, vendorID: %s, status: %s, internal: %d",
			vendorID, status, internal)
	}
	SQL := `SELECT DISTINCT t2.vendor_id,t2.vendor_name,t2.region_id,t2.region_name,t1.product_id,t1.name,t1.version,t1.arch,t1.status FROM products t1 INNER JOIN zones t2 ON
 t1.vendor_id=t2.vendor_id AND t1.region_id=t2.region_id AND t1.vendor_id = ? AND t1.status = ? AND t1.internal = ?;`
	var info structs.Product
	rows, err := p.DB(ctx).Raw(SQL, vendorID, status, internal).Rows()
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			//Read a row of data and store it in a temporary variable
			err = rows.Scan(&info.VendorID, &info.VendorName, &info.RegionID, &info.RegionName, &info.ID, &info.Name, &info.Version, &info.Arch, &info.Status)
			if err != nil {
				framework.LogWithContext(ctx).Warningf("query product error: %v, vendorID: %s, status: %s, internal: %d",
					err, vendorID, status, internal)
				return nil, errors.NewErrorf(errors.QueryProductsScanRowError, "query product error: %v, vendorID: %s, status: %s, internal: %d",
					err, vendorID, status, internal)
			}
			products = append(products, info)
		}
	}
	return products, err
}

// QueryProductDetail Query the product information provided by EM, if you pass in RegionID,productID,Status, only query all the commodities under RegionID,
// on the contrary, query all the product information
func (p *ProductReadWriter) QueryProductDetail(ctx context.Context, vendorID, regionID, productID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products map[string]structs.ProductDetail, er error) {
	if "" == vendorID || "" == regionID || "" == productID {
		framework.LogWithContext(ctx).Warningf("query product detail invalid parameter, vendorID: %s, regionID:%s, productID: %s,status: %s, internal: %d",
			vendorID, regionID, productID, status, internal)
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "query product detail invalid parameter, vendorID: %s, regionID:%s, productID: %s,status: %s, internal: %d",
			vendorID, regionID, productID, status, internal)
	}
	/*	SQL := `SELECT t2.zone_id,t2.zone_name,t1.name,t1.version,t1.arch,
		t4.id,t4.name,t4.disk_type,t4.cpu,t4.memory,t3.component_id,t3.name,t3.purpose_type,t3.start_port,t3.end_port,
		t3.max_port,t3.max_instance,t3.min_instance FROM products t1,zones t2,product_components t3,specs t4,vendors t5
		WHERE t1.vendor_id = t5.vendor_id AND t1.vendor_id = ?
		AND t1.region_id=t2.region_id AND t1.region_id= ?
		AND t1.Internal = ?
		AND t1.version=t3.product_version
		AND t1.product_id=t3.product_id AND t1.product_id = ?
		AND t3.purpose_type=t4.purpose_type
		AND t1.arch=t4.arch
		AND t2.zone_id =t4.zone_id
		AND t1.status = ? AND t3.status = ? AND t4.status = ?;`*/
	SQL := `SELECT t2.zone_id,t2.zone_name,t1.name,t1.version,t1.arch,
t4.id,t4.name,t4.disk_type,t4.cpu,t4.memory,t3.component_id,t3.name,t3.purpose_type,t3.start_port,t3.end_port,
t3.max_port,t3.max_instance,t3.min_instance FROM products t1,zones t2,product_components t3,specs t4
WHERE t1.vendor_id = t2.vendor_id AND t1.vendor_id = ? 
AND t1.region_id=t2.region_id AND t1.region_id= ? 
AND t1.Internal = ? 
AND t1.version=t3.product_version 
AND t1.product_id=t3.product_id AND t1.product_id = ? 
AND t3.purpose_type=t4.purpose_type
AND t1.status = ? AND t3.status = ? AND t4.status = ?;`

	var ok bool
	var detail structs.ProductDetail
	var productVersion structs.ProductVersion
	var spec structs.ComponentInstanceResourceSpec
	var info, productComponentInfo structs.ProductComponentProperty
	var productName, version, arch string
	var components []structs.ProductComponentProperty
	products = make(map[string]structs.ProductDetail)
	rows, err := p.DB(ctx).Raw(SQL, vendorID, regionID, internal, productID, status, constants.ProductSpecStatusOnline, constants.ProductSpecStatusOnline).Rows()
	defer rows.Close()
	log := framework.LogWithContext(ctx)
	log.Debugf("QueryProductDetail SQL: %s vendorID:%s, regionID: %s productID: %s, internal: %d, status: %s, execute result, error: %v",
		SQL, vendorID, regionID, productID, internal, status, err)

	if err == nil {
		for rows.Next() {
			//Read a row of data and store it in a temporary variable
			err = rows.Scan(&spec.ZoneID, &spec.ZoneName, &productName, &version, &arch,
				&spec.ID, &spec.Name, &spec.DiskType, &spec.CPU, &spec.Memory,
				&info.ID, &info.Name, &info.PurposeType, &info.StartPort,
				&info.EndPort, &info.MaxPort, &info.MaxInstance, &info.MinInstance)
			//fmt.Printf("zoneID:%s,zoneName:%s,productName:%s,version:%s,arch:%s,specs:%v,info: %v\n",
			//	spec.ZoneID, spec.ZoneName, productName, version, arch, spec,info)
			if err != nil {
				framework.LogWithContext(ctx).Warningf("query product detail scan data error: %v, vendorID: %s, regionID:%s, productID: %s,status: %s, internal: %d",
					err, vendorID, regionID, productID, status, internal)
				return nil, errors.NewErrorf(errors.QueryProductsScanRowError, "query product detail scan data error: %v, vendorID: %s, regionID:%s, productID: %s,status: %s, internal: %d",
					err, vendorID, regionID, productID, status, internal)
			}
			//Query whether the product information is already in commodity,
			//if it already exists, then directly modify the relevant data structure
			detail, ok = products[productID]
			if !ok {
				products[productID] = structs.ProductDetail{ID: productID, Name: productName, Versions: make(map[string]structs.ProductVersion)}
				detail, _ = products[productID]
			}

			//Query whether the product version information is already in Versions,
			//if it already exists, then directly modify the relevant data structure
			productVersion, ok = detail.Versions[version]
			if !ok {
				detail.Versions[version] = structs.ProductVersion{Version: version, Arch: make(map[string][]structs.ProductComponentProperty)}
				productVersion, _ = detail.Versions[version]
			}

			//Query whether the product arch information is already in archs,
			//if it already exists, then directly modify the relevant data structure
			components, ok = productVersion.Arch[arch]
			if !ok {
				productVersion.Arch[arch] = make([]structs.ProductComponentProperty, 0)
				components, _ = productVersion.Arch[arch]
			}

			//Query whether the product component information is already in Components,
			//if it already exists, then directly modify the relevant data structure

			//productComponentInfo := structs.ProductComponentProperty{}
			//for _, v := range components {
			//	if i == 0 {
			//
			//	}
			//}
			//productComponentInfo, ok = components[info.ID]
			//if !ok {
			//	components = append(components, structs.ProductComponentProperty{ID: info.ID, Name: info.Name, PurposeType: info.PurposeType,
			//		StartPort: info.StartPort, EndPort: info.EndPort, MaxPort: info.MaxPort, MinInstance: info.MinInstance, MaxInstance: info.MaxInstance, Spec: make(map[string]structs.ComponentInstanceResourceSpec)})
			//	components[info.ID] =
			//	productComponentInfo, _ = components[info.ID]
			//}
			//
			////Query whether the product component specification information is already in Specifications,
			////if it already exists, then directly modify the relevant data structure
			//_, ok = productComponentInfo.Spec[spec.ID]
			//if !ok {
			//	productComponentInfo.Spec[spec.ID] = spec
			//}
		}
	}
	return products, err
}

// QueryProductComponentProperty Query the properties of a product component,For creating, expanding and shrinking clusters only
func (p *ProductReadWriter) QueryProductComponentProperty(ctx context.Context, vendorID, regionID, productID, productVersion, arch string, productStatus constants.ProductComponentStatus) (property []structs.ProductComponentProperty, er error) {
	if "" == vendorID || "" == regionID || "" == productID || "" == productVersion || "" == arch {
		framework.LogWithContext(ctx).Warningf("query product component property invalid parameter, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
			vendorID, regionID, productID, productVersion, arch, string(productStatus))
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "query product component property invalid parameter, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
			vendorID, regionID, productID, productVersion, arch, string(productStatus))
	}
	var info structs.ProductComponentProperty
	SQL := "SELECT component_id,name,purpose_type,start_port,end_port,max_port,max_instance,min_instance FROM product_components " +
		"WHERE vendor_id = ? AND region_id = ? AND product_id =? AND product_version = ? AND arch = ? AND status=?"
	//The number of components of the product does not exceed 20
	rows, err := p.DB(ctx).Raw(SQL, vendorID, regionID, productID, productVersion, arch, productStatus).Rows()
	defer rows.Close()
	if err != nil {
		framework.LogWithContext(ctx).Warningf("query product component property scan data error: %v, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
			err, vendorID, regionID, productID, productVersion, arch, string(productStatus))
		return nil, errors.NewErrorf(errors.QueryProductComponentProperty, "query product component property scan data error: %v, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
			err, vendorID, regionID, productID, productVersion, arch, string(productStatus))
	}
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.PurposeType, &info.StartPort, &info.EndPort, &info.MaxPort, &info.MaxInstance, &info.MinInstance)
		if err == nil {
			property = append(property, info)
		} else {
			framework.LogWithContext(ctx).Warningf("query product component property scan data error: %v, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
				err, vendorID, regionID, productID, productVersion, arch, productStatus)
			return nil, errors.NewErrorf(errors.QueryProductComponentProperty, "query product component property scan data error: %v, vendorID: %s, regionID: %s, productID: %s, version: %s, arch: %s,status: %s",
				err, vendorID, regionID, productID, productVersion, arch, productStatus)
		}
	}
	return property, err
}

//CreateSpecs batch create specs
func (p *ProductReadWriter) CreateSpecs(ctx context.Context, specs []Spec) (er error) {
	if len(specs) <= 0 || len(specs) > 1024 {
		framework.LogWithContext(ctx).Warningf("create specs %v, len: %d, parameter invalid", specs, len(specs))
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "create specs %v, parameter invalid", specs)
	}

	db := p.DB(ctx).CreateInBatches(specs, 10)
	if db.Error != nil {
		framework.LogWithContext(ctx).Errorf("create specs %v, errors: %v", specs, db.Error)
		return errors.NewErrorf(errors.CreateSpecsError, "create specs %v, error: %v", specs, db.Error)
	}
	return nil
}

//DeleteSpecs batch delete specs
func (p *ProductReadWriter) DeleteSpecs(ctx context.Context, specs []string) (er error) {
	if len(specs) <= 0 || len(specs) > 1024 {
		framework.LogWithContext(ctx).Warningf("delete specs %v, len: %d, parameter invalid", specs, len(specs))
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "delete specs %v, parameter invalid", specs)
	}
	p.DB(ctx).Begin()
	for _, spec := range specs {
		err := p.DB(ctx).Where("id= ?", spec).Delete(&Spec{}).Error
		if err != nil {
			framework.LogWithContext(ctx).Errorf("delete specs, spec: %s errors: %v", spec, err)
			p.DB(ctx).Rollback()
			return errors.NewErrorf(errors.DeleteSpecsError, "delete spec: %s, errors: %v", spec, err)
		}
	}
	p.DB(ctx).Commit()
	return nil
}

//QuerySpecs query all specs
func (p *ProductReadWriter) QuerySpecs(ctx context.Context) (specs []structs.SpecInfo, er error) {

	var info structs.SpecInfo
	SQL := "SELECT id,name,cpu,memory,disk_type,purpose_type,status FROM specs;"
	rows, err := p.DB(ctx).Raw(SQL).Rows()
	defer rows.Close()
	log := framework.LogWithContext(ctx)
	log.Debugf("QuerySpecs SQL: %s, execute result, error: %v", SQL, err)
	if err != nil {
		return nil, errors.NewErrorf(errors.TIEM_SQL_ERROR, "query all specs error: %v, SQL: %s", err, SQL)
	}
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.CPU, &info.Memory, &info.DiskType, &info.PurposeType, &info.Status)
		if err == nil {
			specs = append(specs, info)
		} else {
			return nil, errors.NewErrorf(errors.QuerySpecScanRowError, "query all specs, scan data error: %v, SQL: %s", err, SQL)
		}
	}
	return specs, err
}
