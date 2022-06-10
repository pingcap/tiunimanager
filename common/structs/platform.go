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
 * @File: platform.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import (
	"github.com/pingcap-inc/tiunimanager/common/constants"
)

//SpecInfo information about spec
type SpecInfo struct {
	ID          string `json:"id"`   //ID of the resource specification
	Name        string `json:"name"` //Name of the resource specification,eg: TiDB.c1.large
	CPU         int    `json:"cpu"`
	Memory      int    `json:"memory"`      //The amount of memory occupied by the instance, in GiB
	DiskType    string `json:"diskType"`    //eg: NVMeSSD/SSD/SATA
	PurposeType string `json:"purposeType"` // eg:Compute/Storage/Schedule
}

// ComponentInstanceResourceSpec Information on the resources required for the product components to run, including: memory, CPU, etc.
type ComponentInstanceResourceSpec struct {
	ID       string `json:"id"`   //ID of the instance resource specification
	Name     string `json:"name"` //Name of the instance resource specification,eg: TiDB.c1.large
	CPU      int    `json:"cpu"`
	Memory   int    `json:"memory"`   //The amount of memory occupied by the instance, in GiB
	DiskType string `json:"diskType"` //eg: NVMeSSD/SSD/SATA
	ZoneID   string `json:"zoneId"`
	ZoneName string `json:"zoneName"`
}

// ProductComponentPropertyWithZones Information about the components of the product, each of which consists of several different types of components
type ProductComponentPropertyWithZones struct {
	ID                      string                           `json:"id"`          //ID of the product component, globally unique
	Name                    string                           `json:"name"`        //Name of the product component, globally unique
	PurposeType             string                           `json:"purposeType"` //The type of resources required by the product component at runtime, e.g. storage class
	StartPort               int32                            `json:"startPort"`
	EndPort                 int32                            `json:"endPort"`
	MaxPort                 int32                            `json:"maxPort"`
	MinInstance             int32                            `json:"minInstance"` //Minimum number of instances of product components at runtime, e.g. at least 1 instance of PD, at least 3 instances of TiKV
	MaxInstance             int32                            `json:"maxInstance"` //Maximum number of instances when the product component is running, e.g. PD can run up to 7 instances, other components have no upper limit
	SuggestedInstancesCount []int32                          `json:"suggestedInstancesCount"`
	AvailableZones          []ComponentInstanceZoneWithSpecs `json:"availableZones"` //Information on the specifications of the resources online for the running of product components,organized by different Zone
}

// ComponentInstanceZoneWithSpecs Specs group by zone
type ComponentInstanceZoneWithSpecs struct {
	ZoneID   string                          `json:"zoneId"`
	ZoneName string                          `json:"zoneName"`
	Specs    []ComponentInstanceResourceSpec `json:"specs"`
}

//ProductVersion Product version and component details, with each product categorized by version and supported CPU architecture
type ProductVersion struct {
	Version string                                         `json:"version"` //Version information of the product, e.g. v5.0.0
	Arch    map[string][]ProductComponentPropertyWithZones `json:"arch"`    //Arch information of the product, e.g. X86/X86_64
	//Components map[string]ProductComponentPropertyWithZones `json:"components"` //Component Info of the product
}

// ProductDetail product information provided by Enterprise Manager
type ProductDetail struct {
	ID       string                    `json:"id"`       //The ID of the product consists of the product ID
	Name     string                    `json:"name"`     //The name of the product consists of the product name and the version
	Versions map[string]ProductVersion `json:"versions"` //Organize product information by version
}

// Product product base information provided by Enterprise Manager
type Product struct {
	ID         string `json:"id"`   // The ID of the product
	Name       string `json:"name"` // the Name of the product
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	RegionID   string `json:"regionId"`
	RegionName string `json:"regionName"`
	VendorID   string `json:"vendorId"`   // the vendor ID of the vendor, e.go AWS
	VendorName string `json:"vendorName"` // the Vendor name of the vendor, e.g AWS/Aliyun
	Status     string `json:"status"`
	Internal   int    `json:"internal"`
}

// ZoneFullInfo vendor & region & zone information provided by Enterprise Manager
type ZoneFullInfo struct {
	ZoneID     string `json:"zoneId" form:"zoneId"`         //The value of the ZoneID is similar to CN-HANGZHOU-H
	ZoneName   string `json:"zoneName" form:"zoneName"`     //The value of the Name is similar to Hangzhou(H)
	RegionID   string `json:"regionId" form:"regionId"`     //The value of the RegionID is similar to CN-HANGZHOU
	RegionName string `json:"regionName" form:"regionName"` //The value of the Name is similar to East China(Hangzhou)
	VendorID   string `json:"vendorId" form:"vendorId"`     //The value of the VendorID is similar to AWS
	VendorName string `json:"vendorName" form:"vendorName"` //The value of the Name is similar to AWS
	Comment    string `json:"comment" form:"comment"`
}

// ZoneInfo zone information
type ZoneInfo struct {
	ZoneID   string `json:"zoneId" form:"zoneId"`     //The value of the ZoneID is similar to CN-HANGZHOU-H
	ZoneName string `json:"zoneName" form:"zoneName"` //The value of the Name is similar to Hangzhou(H)
	Comment  string `json:"comment" form:"comment"`
}

// SystemConfig system config of platform
type SystemConfig struct {
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
}

// SystemInfo system info of platform
type SystemInfo struct {
	SystemName       string `json:"systemName"`
	SystemLogo       string `json:"systemLogo"`
	CurrentVersionID string `json:"currentVersionId"`
	LastVersionID    string `json:"lastVersionId"`
	State            string `json:"state"`

	SupportedVendors  []VendorInfo          `json:"supportedVendors"`
	SupportedProducts []ProductWithVersions `json:"supportedProducts"`

	VendorZonesInitialized       bool `json:"vendorZonesInitialized"`
	VendorSpecsInitialized       bool `json:"vendorSpecsInitialized"`
	ProductComponentsInitialized bool `json:"productComponentsInitialized"`
	ProductVersionsInitialized   bool `json:"productVersionsInitialized"`
}

type ProductWithVersions struct {
	ProductID   string                   `json:"productId"`
	ProductName string                   `json:"productName"`
	Versions    []SpecificVersionProduct `json:"versions"`
}

type SpecificVersionProduct struct {
	ProductID string   `json:"productId"`
	Arch      string   `json:"arch"`
	Version   string   `json:"version"`
}

type SystemVersionInfo struct {
	VersionID   string `json:"versionId"`
	Desc        string `json:"desc"`
	ReleaseNote string `json:"releaseNote"`
}

type VendorConfigInfo struct {
	VendorInfo
	Regions []RegionConfigInfo `json:"regions" form:"regions"`
	Specs   []SpecInfo         `json:"specs" form:"specs"`
}

type RegionConfigInfo struct {
	RegionInfo
	Zones []ZoneInfo `json:"zones" form:"zones"`
}

type ProductConfigInfo struct {
	ProductID   string                              `json:"productId" form:"productId"`
	ProductName string                              `json:"productName" form:"productName"`
	Components  []ProductComponentPropertyWithZones `json:"components" form:"components"`
	Versions    []SpecificVersionProduct            `json:"versions" form:"versions"`
}

// DBUserRole role information of the DBUser
type DBUserRole struct {
	ClusterType constants.EMProductIDType
	RoleName    string
	RoleType    constants.DBUserRoleType
	Permission  []string
}

var DBUserRoleRecords = map[constants.DBUserRoleType]DBUserRole{
	constants.Root: {
		ClusterType: constants.EMProductIDTiDB,
		RoleName:    "root",
		RoleType:    constants.Root,
		Permission:  constants.DBUserPermission[constants.Root],
	},
	constants.DBUserBackupRestore: {
		ClusterType: constants.EMProductIDTiDB,
		RoleName:    "backup_restore",
		RoleType:    constants.DBUserBackupRestore,
		Permission:  constants.DBUserPermission[constants.DBUserBackupRestore],
	},
	constants.DBUserParameterManagement: {
		ClusterType: constants.EMProductIDTiDB,
		RoleName:    "parameter_management",
		RoleType:    constants.DBUserParameterManagement,
		Permission:  constants.DBUserPermission[constants.DBUserParameterManagement],
	},
	constants.DBUserCDCDataSync: {
		ClusterType: constants.EMProductIDTiDB,
		RoleName:    "data_sync",
		RoleType:    constants.DBUserCDCDataSync,
		Permission:  constants.DBUserPermission[constants.DBUserCDCDataSync],
	},
}
