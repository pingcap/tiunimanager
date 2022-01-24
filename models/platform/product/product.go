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
 * @File: product.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package product

import (
	"gorm.io/gorm"
	"time"
)

//Product Enterprise Manager product information
// Enterprise Manager provides TiDB and TiDB Data Migration products
type Product struct {
	VendorID  string         `gorm:"primaryKey;"`
	RegionID  string         `gorm:"primaryKey;"`
	ProductID string         `gorm:"primaryKey;"`
	Version   string         `gorm:"primaryKey;size 16;comment: original version of the product"`
	Arch      string         `gorm:"primaryKey;size 16;comment: original cpu arch of the product"`
	Name      string         `gorm:"not null;size 16;comment: original name of the product"`
	Status    string         `gorm:"not null;size:32;"` //Online、Offline
	Internal  int            `gorm:"default:0"`         //Whether it is an internal product, the value is 1 for Enterprise Manager products only
	CreatedAt time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:""`
}

//ProductComponent Enterprise Manager offers two products, TiDB and TiDB Data Migration,
//each consisting of multiple components, each with multiple ports enabled and a limit on the number of instances each component can start.
type ProductComponent struct {
	VendorID       string         `gorm:"primaryKey;"`
	RegionID       string         `gorm:"primaryKey;"`
	ProductID      string         `gorm:"primaryKey;"`
	ProductVersion string         `gorm:"primaryKey;"`
	Arch           string         `gorm:"primaryKey;size 16;comment: original cpu arch of the product"`
	ComponentID    string         `gorm:"primaryKey;"`
	Name           string         `gorm:"primaryKey;comment:'original name of the product'"`
	Status         string         `gorm:"not null;size:32;"` //Online,Offline
	PurposeType    string         `gorm:"size:32;comment:Compute/Storage/Schedule"`
	StartPort      int32          `gorm:"comment:'starting value of the port opened by the component'"`
	EndPort        int32          `gorm:"comment:'ending value of the port opened by the component'"`
	MaxPort        int32          `gorm:"comment:'maximum number of ports that can be opened by the component'"`
	MinInstance    int32          `gorm:"default:0;comment:'minimum number of instances started by the component'"`
	MaxInstance    int32          `gorm:"default:128;comment:'PD: 1、3、5、7，Prometheus：1，AlertManger：1，Grafana'"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:""`
}

//Spec product specification information provided by Enterprise Manager
type Spec struct {
	ID          string         `gorm:"primaryKey;size:32"`
	Name        string         `gorm:"not null;size 16;comment: original spec of the product, eg 8C16G"`
	CPU         int            `gorm:"comment: unit: vCPU"`
	Memory      int            `gorm:"comment: unit: GiB"`
	DiskType    string         `gorm:"comment:NVMeSSD/SSD/SATA"`
	PurposeType string         `gorm:"comment:Compute/Storage/Schedule"`
	Status      string         `gorm:"default:'Online';not null;size:32;"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;<-:create;->;"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:""`
}

/**
SELECT t2.zone_id,t2.name,t1.product_id,t1.name,t1.version,t1.arch,
t4.resource_spec_id,t4.name,t4.disk_type,t4.cpu,t4.memory,t3.component_id,t3.name,t3.purpose_type,t3.start_port,t3.end_port,
t3.max_port,t3.max_instance,t3.min_instance FROM products t1,zones t2,product_components t3,resource_specs t4
WHERE t1.region_id=t2.region_id AND t1.region_id='CN-HANGZHOU' AND
t1.version=t3.product_version AND t1.product_id=t3.product_id AND t3.purpose_type=t4.purpose_type
AND t1.arch=t4.arch AND t2.zone_id =t4.zone_id AND t1.status = 'Online' AND t3.status = 'Online' AND t4.status = 'Online';
*/
