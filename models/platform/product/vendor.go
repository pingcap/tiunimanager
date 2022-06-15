/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 * @File: vendor.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package product

// Vendor information provided by Enterprise Manager
type Vendor struct {
	VendorID   string `gorm:"primaryKey;size:32"`
	VendorName string `gorm:"size:32"`
	// accessKey and secretKey
}

// VendorZone information provided by Enterprise Manager
type VendorZone struct {
	VendorID   string `gorm:"uniqueIndex:vendor_region_zone;size:32"`
	RegionID   string `gorm:"uniqueIndex:vendor_region_zone;size:32"`
	RegionName string `gorm:"size:32"`
	ZoneID     string `gorm:"uniqueIndex:vendor_region_zone;size:32"`
	ZoneName   string `gorm:"size:32"`
	Comment    string `gorm:"size:1024;"`
}

//VendorSpec specification information provided by Enterprise Manager
type VendorSpec struct {
	VendorID    string `gorm:"uniqueIndex:vendor_spec;size:32"`
	SpecID      string `gorm:"uniqueIndex:vendor_spec;size:32"`
	SpecName    string `gorm:"not null;size 16;comment: original spec of the product, eg 8C16G"`
	CPU         int    `gorm:"comment: unit: vCPU"`
	Memory      int    `gorm:"comment: unit: GiB"`
	DiskType    string `gorm:"comment:NVMeSSD/SSD/SATA"`
	PurposeType string `gorm:"comment:Compute/Storage/Schedule"`
}
