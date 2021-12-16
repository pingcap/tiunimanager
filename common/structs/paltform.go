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
 * @File: paltform.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

// ComponentInstanceResourceSpec Information on the resources required for the product components to run, including: memory, CPU, etc.
type ComponentInstanceResourceSpec struct {
	ID       string //ID of the instance resource specification
	Name     string //Name of the instance resource specification,eg: TiDB.c1.large
	CPU      int
	Memory   int    //The amount of memory occupied by the instance, in GiB
	DiskType string //eg: NVMeSSD/SSD/SATA
	ZoneID   string
	ZoneName string
}

// ProductComponentProperty Information about the components of the product, each of which consists of several different types of components
type ProductComponentProperty struct {
	ID          string //ID of the product component, globally unique
	Name        string //Name of the product component, globally unique
	PurposeType string //The type of resources required by the product component at runtime, e.g. storage class
	StartPort   int32
	EndPort     int32
	MaxPort     int32
	MinInstance int32                                    //Minimum number of instances of product components at runtime, e.g. at least 1 instance of PD, at least 3 instances of TiKV
	MaxInstance int32                                    //Maximum number of instances when the product component is running, e.g. PD can run up to 7 instances, other components have no upper limit
	Spec        map[string]ComponentInstanceResourceSpec //Information on the specifications of the resources online for the running of product components,organized by different Zone
}

//ProductVersion Product version and component details, with each product categorized by version and supported CPU architecture
type ProductVersion struct {
	Version    string                              //Version information of the product, e.g. v5.0.0
	Arch       string                              //Arch information of the product, e.g. X86/X86_64
	Components map[string]ProductComponentProperty //Component Info of the product
}

// ProductDetail product information provided by Enterprise Manager
type ProductDetail struct {
	ID       string                    //The ID of the product consists of the product ID
	Name     string                    //The name of the product consists of the product name and the version
	Versions map[string]ProductVersion //Organize product information by version
}

// Product product base information provided by Enterprise Manager
type Product struct {
	ID         string // The ID of the product
	Name       string // the Name of the product
	Version    string
	Arch       string
	RegionID   string
	RegionName string
	VendorID   string // the vendor ID of the vendor, e.go AWS
	VendorName string // the Vendor name of the vendor, e.g AWS/Aliyun
}

//ZoneDetail vendor & region & zone information provided by Enterprise Manager
type ZoneDetail struct {
	ZoneID     string //The value of the ZoneID is similar to CN-HANGZHOU-H
	ZoneName   string //The value of the Name is similar to Hangzhou(H)
	RegionID   string //The value of the RegionID is similar to CN-HANGZHOU
	RegionName string //The value of the Name is similar to East China(Hangzhou)
	VendorID   string //The value of the VendorID is similar to AWS
	VendorName string //The value of the Name is similar to AWS
}

// SystemConfig system config of platform
type SystemConfig struct {
	ConfigKey   string
	ConfigValue string
}
