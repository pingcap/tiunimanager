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

package specs

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
	QueryZones(ctx context.Context) (zones []structs.ZoneDetail, err error)

	// CreateRegion
	// @Description: add a region
	// @param ctx
	// @param id
	// @param name
	// @return err
	CreateRegion(ctx context.Context, id, name string) (err error)

	// DeleteRegion
	// @Description: delete a region
	// @param ctx
	// @param id
	// @return err
	DeleteRegion(ctx context.Context, id string) (err error)

	// CreateZone
	// @Description: add a zone
	// @param ctx
	// @param zoneID
	// @param zoneName
	// @param regionID
	// @return er
	CreateZone(ctx context.Context, zoneID, zoneName, regionID string) (er error)

	// DeleteZone
	// @Description: delete zone
	// @param ctx
	// @param zoneID
	// @return er
	DeleteZone(ctx context.Context, zoneID string) (er error)

	// QueryProductDetail
	// @Description: Query the product detail information provided by EM
	// @param ctx
	// @param vendorID
	// @param regionID
	// @param productID
	// @param status
	// @return products
	// @return err
	QueryProductDetail(ctx context.Context, vendorID, regionID, productID string, status constants.ProductStatus) (products map[string]structs.ProductDetail, er error)

	// QueryProducts
	// @Description: Query all product information by vendor_id,region_id
	// @param ctx
	// @param regionID
	// @param status
	// @return err
	// @return products
	QueryProducts(ctx context.Context, vendorID string, status constants.ProductStatus, internal constants.EMInternalProduct) (err error, products []structs.Product)

	// QueryProductComponentProperty
	// @Description: Query the properties of a product component
	// @param ctx
	// @param productID
	// @param productVersion
	// @return err
	// @return property[]
	QueryProductComponentProperty(ctx context.Context, productID, productVersion string, productStatus constants.ProductComponentStatus) (err error, property []structs.ProductComponentProperty)
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

func (p *ProductReadWriter) QueryProducts(ctx context.Context, vendorID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products []structs.Product, er error) {
	if "" == vendorID || "" == status {
		return nil, errors.NewError(errors.QueryProductsScanRowError, "")
	}
	SQL := `SELECT t1.name,t2.region_id,t2.name,t3.product_id,t3.name,t3.version,t3.arch FROM products t3 join vendors t1 join regions t2
WHERE t1.vendor_id=t2.vendor_id AND t1.vendor_id= t3.vendor_id and t3.region_id = t2.region_id AND t1.vendor_id = ? AND t3.status = ? AND t3.internal = ?;`
	var info structs.Product
	rows, err := p.DB(ctx).Raw(SQL, vendorID, status, internal).Rows()
	log := framework.LogWithContext(ctx)
	log.Debugf("QueryProducts SQL: %s, vendorID: %s,status: %s, internal: %d, execute result, error: %v", SQL, vendorID, status, internal, err)
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			//Read a row of data and store it in a temporary variable
			err = rows.Scan(&info.VendorName, &info.RegionID, &info.RegionName, &info.ID, &info.Name, &info.Version, &info.Arch)
			if err != nil {
				return nil, errors.WrapError(errors.QueryProductsScanRowError, "", err)
			}
			products = append(products, info)
		}
	}
	return products, err
}

//QueryZones Query vendor & region & zone information provided by Enterprise Manager
func (p *ProductReadWriter) QueryZones(ctx context.Context) (zones []structs.ZoneDetail, err error) {
	var info structs.ZoneDetail
	SQL := "SELECT t1.zone_id,t1.name,t2.region_id,t2.name,t3.vendor_id,t3.name FROM zones t1,regions t2,vendors t3 where t1.region_id = t2.region_id AND t1.vendor_id = t3.vendor_id"
	rows, err := p.DB(ctx).Raw(SQL).Rows()
	defer rows.Close()
	log := framework.LogWithContext(ctx)
	log.Debugf("QueryZones SQL: %s, execute result, error: %v", SQL, err)
	if err != nil {
		return nil, errors.WrapError(errors.QueryZoneScanRowError, "", err)
	}
	for rows.Next() {
		err = rows.Scan(&info.ZoneID, &info.ZoneName, &info.RegionID, &info.RegionName, &info.VendorID, &info.VendorName)
		if err == nil {
			zones = append(zones, info)
		} else {
			return nil, errors.WrapError(errors.QueryZoneScanRowError, "", err)
		}
	}
	return zones, err
}

// QueryProductDetail Query the product information provided by EM, if you pass in RegionID,productID,Status, only query all the commodities under RegionID,
// on the contrary, query all the product information
func (p *ProductReadWriter) QueryProductDetail(ctx context.Context, vendorID, regionID, productID string, status constants.ProductStatus, internal constants.EMInternalProduct) (products map[string]structs.ProductDetail, er error) {
	if "" == regionID {
		return nil, errors.NewError(errors.QueryProductsScanRowError, "")
	}
	SQL := `SELECT t2.zone_id,t2.name,t1.name,t1.version,t1.arch,
t4.resource_spec_id,t4.name,t4.disk_type,t4.cpu,t4.memory,t3.component_id,t3.name,t3.purpose_type,t3.start_port,t3.end_port,
t3.max_port,t3.max_instance,t3.min_instance FROM products t1,zones t2,product_components t3,resource_specs t4,vendors t5
WHERE t1.vendor_id = t5.vendor_id AND t1.vendor_id = ? 
AND t1.region_id=t2.region_id AND t1.region_id= ? 
AND t1.Internal = ? 
AND t1.version=t3.product_version 
AND t1.product_id=t3.product_id AND t1.product_id = ? 
AND t3.purpose_type=t4.purpose_type
AND t1.arch=t4.arch 
AND t2.zone_id =t4.zone_id 
AND t1.status = ? AND t3.status = ? AND t4.status = ?;`
	var ok bool
	var detail structs.ProductDetail
	var productVersion structs.ProductVersion
	var spec structs.ComponentInstanceResourceSpec
	var info, productComponentInfo structs.ProductComponentProperty
	var productName, version, arch string
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
			if err != nil {
				return nil, errors.WrapError(errors.QueryProductsScanRowError, "", err)
			}
			//Query whether the product information is already in commodity,
			//if it already exists, then directly modify the relevant data structure
			detail, ok = products[productID]
			if !ok {
				products[productID] = structs.ProductDetail{ID: productID, Name: productName, Versions: make(map[string]structs.ProductVersion)}
				detail = products[productID]
			}

			//Query whether the product version information is already in Versions,
			//if it already exists, then directly modify the relevant data structure
			productVersion, ok = detail.Versions[version]
			if !ok {
				detail.Versions[version] = structs.ProductVersion{Version: version, Arch: arch, Components: make(map[string]structs.ProductComponentProperty)}
				productVersion = detail.Versions[version]
			}

			//Query whether the product component information is already in Components,
			//if it already exists, then directly modify the relevant data structure
			productComponentInfo, ok = productVersion.Components[info.ID]
			if !ok {
				productVersion.Components[info.ID] = structs.ProductComponentProperty{ID: info.ID, Name: info.Name, PurposeType: info.PurposeType,
					StartPort: info.StartPort, EndPort: info.EndPort, MaxPort: info.MaxPort, MinInstance: info.MinInstance, MaxInstance: info.MaxInstance, Spec: make(map[string]structs.ComponentInstanceResourceSpec)}
				productComponentInfo = productVersion.Components[info.ID]
			}

			//Query whether the product component specification information is already in Specifications,
			//if it already exists, then directly modify the relevant data structure
			_, ok = productComponentInfo.Spec[spec.ID]
			if !ok {
				productComponentInfo.Spec[spec.ID] = spec
			}
		}
	}
	return products, err
}

// QueryProductComponentProperty Query the properties of a product component,For creating, expanding and shrinking clusters only
func (p *ProductReadWriter) QueryProductComponentProperty(ctx context.Context, productID, productVersion string, productStatus constants.ProductComponentStatus) (property []structs.ProductComponentProperty, er error) {
	if "" == productID || "" == productVersion {
		//TODO wait for Error modify
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "")
	}
	var info structs.ProductComponentProperty
	SQL := "SELECT component_id,name,purpose_type,start_port,end_port,max_port,max_instance,min_instance FROM product_components " +
		"WHERE product_id =? AND product_version = ? AND status=?"
	//The number of components of the product does not exceed 20
	rows, err := p.DB(ctx).Raw(SQL, productID, productVersion, productStatus).Rows()
	defer rows.Close()
	if err != nil {
		return nil, errors.WrapError(errors.QueryProductComponentProperty, "", err)
	}
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.PurposeType, &info.StartPort, &info.EndPort, &info.MaxPort, &info.MaxInstance, &info.MinInstance)
		if err == nil {
			property = append(property, info)
		} else {
			return nil, errors.WrapError(errors.QueryProductComponentProperty, "", err)
		}
	}
	return property, err
}
