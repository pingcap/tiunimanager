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
 * @File: product_read_writer_test.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10
*******************************************************************************/

package product

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/stretchr/testify/assert"
)

var prw *ProductReadWriter

const (
	TiDB                     = "TiDB"
	TiDBVersion50            = "5.0.0"
	TiDBVersion51            = "5.1.0"
	EnterpriseManager        = "EnterpriseManager"
	EnterpriseManagerVersion = "1.0.0"
	AliYun                   = "Aliyun"
	CNHangzhou               = "CN-HANGZHOU"
	CNBeijing                = "CN-BEIJING"
)

var zones = []Zone{
	{VendorID: AliYun, VendorName: AliYun, RegionID: CNBeijing, RegionName: "North China(Beijing)", ZoneID: "CN-BEIJING-G", ZoneName: "ZONE(G)", Comment: "vendor aliyun"},
	{VendorID: AliYun, VendorName: AliYun, RegionID: CNHangzhou, RegionName: "East China(Hangzhou)", ZoneID: "CN-HANGZHOU-H", ZoneName: "ZONE(H)", Comment: "vendor aliyun"},
	{VendorID: AliYun, VendorName: AliYun, RegionID: CNHangzhou, RegionName: "East China(Hangzhou)", ZoneID: "CN-HANGZHOU-G", ZoneName: "ZONE(G)", Comment: "vendor aliyun"},
}
var DeleteZones = []Zone{
	{VendorID: AliYun, VendorName: AliYun, RegionID: CNBeijing, RegionName: "North China(Beijing)", ZoneID: "CN-BEIJING-H", ZoneName: "ZONE(H)", Comment: "vendor aliyun"},
}

var products = []Product{
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "EnterpriseManager", Name: "EnterpriseManager", Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes},
}

var NoInternalProducts = []Product{
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}

var HangZhouProducts = []Product{
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}
var HangZhouProductCtrplan = []Product{
	{VendorID: AliYun, RegionID: CNHangzhou, ProductID: "EnterpriseManager", Name: "EnterpriseManager", Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes},
}
var BeijingProducts = []Product{
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ProductID: "TiDB", Name: "TiDB", Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}

var TiDBALIYUNHZX8650Components = []ProductComponent{
	//TiDB v5.0.0
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "PD", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "CDC", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNBJX8650Components = []ProductComponent{
	//TiDB v5.0.0
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "PD", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "CDC", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: TiDBVersion50, Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNBJX8651Components = []ProductComponent{
	//TiDB v5.1.0
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "PD", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "CDC", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNBeijing, Arch: string(constants.ArchX8664), ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNHZARM6451Components = []ProductComponent{
	//TiDB v5.1.0
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "PD", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "CDC", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchArm64), ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: TiDBVersion51, Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var EMALIYUNHZX8610Components = []ProductComponent{
	//Enterprise Manager v1.0.0
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "cluster-server", ProductID: EnterpriseManager, ProductVersion: EnterpriseManagerVersion, Name: "cluster-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "openapi-server", ProductID: EnterpriseManager, ProductVersion: EnterpriseManagerVersion, Name: "openapi-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "Grafana", ProductID: EnterpriseManager, ProductVersion: EnterpriseManagerVersion, Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "Prometheus", ProductID: EnterpriseManager, ProductVersion: EnterpriseManagerVersion, Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{VendorID: AliYun, RegionID: CNHangzhou, Arch: string(constants.ArchX8664), ComponentID: "AlertManger", ProductID: EnterpriseManager, ProductVersion: EnterpriseManagerVersion, Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var ComputeSpecs = []Spec{
	{ID: "c1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.large", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.xlarge", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.2xlarge", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c3.g.large", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c3.g.xlarge", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
}
var ScheduleSpecs = []Spec{
	{ID: "sd1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "sd2.g.large", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "sd2.g.xlarge", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
}
var StorageSpecs = []Spec{
	{ID: "s1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "s2.g.large", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "s2.g.xlarge", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs2.g.large", Name: "8C16G", CPU: 8, Memory: 64, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs2.g.xlarge", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
}

func TestProductReadWriter_CreateZones(t *testing.T) {
	t.Run("CreateZones", func(t *testing.T) {
		var err error
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		var results []structs.ZoneFullInfo
		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)
		//First convert the query results into a map to facilitate comparison of results
		tmp := make(map[string]structs.ZoneFullInfo)
		for _, num := range results {
			key := num.VendorID + num.RegionID + num.ZoneID
			if _, ok := tmp[key]; !ok {
				tmp[key] = num
			}
		}
		for _, zone := range zones {
			key := zone.VendorID + zone.RegionID + zone.ZoneID
			val, ok := tmp[key]
			assert.Equal(t, true, ok)
			assert.Equal(t, zone.VendorID, val.VendorID)
			assert.Equal(t, zone.VendorName, val.VendorName)
			assert.Equal(t, zone.RegionID, val.RegionID)
			assert.Equal(t, zone.RegionName, val.RegionName)
			assert.Equal(t, zone.ZoneID, val.ZoneID)
			assert.Equal(t, zone.ZoneName, val.ZoneName)
		}

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)
	})

	t.Run("CreateZonesWithDBError", func(t *testing.T) {
		var err error
		var results []structs.ZoneFullInfo
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)

		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)

		err = prw.CreateZones(context.TODO(), zones)
		assert.Equal(t, errors.CreateZonesError, err.(errors.EMError).GetCode())

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)
	})

	t.Run("CreateZonesWithEmptyParameter", func(t *testing.T) {
		err := prw.CreateZones(context.TODO(), make([]Zone, 0))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_QueryZones(t *testing.T) {
	t.Run("QueryZones", func(t *testing.T) {
		var err error
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		var results []structs.ZoneFullInfo
		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)

		//First convert the query results into a map to facilitate comparison of results
		tmp := make(map[string]structs.ZoneFullInfo)
		for _, num := range results {
			key := num.VendorID + num.RegionID + num.ZoneID
			if _, ok := tmp[key]; !ok {
				tmp[key] = num
			}
		}
		for _, zone := range zones {
			key := zone.VendorID + zone.RegionID + zone.ZoneID
			val, ok := tmp[key]
			assert.Equal(t, true, ok)
			assert.Equal(t, zone.VendorID, val.VendorID)
			assert.Equal(t, zone.VendorName, val.VendorName)
			assert.Equal(t, zone.RegionID, val.RegionID)
			assert.Equal(t, zone.RegionName, val.RegionName)
			assert.Equal(t, zone.ZoneID, val.ZoneID)
			assert.Equal(t, zone.ZoneName, val.ZoneName)
		}
		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)
	})

	t.Run("QueryZonesWithDBError", func(t *testing.T) {
		var err error
		var results []structs.ZoneFullInfo
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)

		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)

		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)
		//assert.Equal(t, errors.TIEM_SQL_ERROR,err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_GetZone(t *testing.T) {
	t.Run("GetZone", func(t *testing.T) {
		var err error
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		var results []structs.ZoneFullInfo

		for i := range zones {
			result, count, err := prw.GetZone(context.TODO(), zones[i].VendorID, zones[i].RegionID, zones[i].ZoneID)
			assert.NoError(t, err)
			assert.Equal(t, zones[i].VendorID, result.VendorID)
			assert.Equal(t, zones[i].RegionID, result.RegionID)
			assert.Equal(t, zones[i].ZoneID, result.ZoneID)
			assert.Equal(t, int64(1), count)

			results = append(results, *result)
		}

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)
	})

	t.Run("QueryZonesWithDBError", func(t *testing.T) {
		result, count, err := prw.GetZone(context.TODO(), "", "", "fakeZoneID")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), count)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_DeleteZones(t *testing.T) {

	t.Run("DeleteZones", func(t *testing.T) {
		var err error
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		var results []structs.ZoneFullInfo
		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)
		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)
	})

	t.Run("DeleteZonesWithDBError", func(t *testing.T) {
		var err error
		var results []structs.ZoneFullInfo
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)

		results, err = prw.QueryZones(context.TODO())
		assert.NoError(t, err)

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)

		err = prw.DeleteZones(context.TODO(), results)
		assert.NoError(t, err)

		//assert.Equal(t, errors.TIEM_SQL_ERROR,err.(errors.EMError).GetCode())
	})

	t.Run("CreateZonesWithEmptyParameter", func(t *testing.T) {
		err := prw.DeleteZones(context.TODO(), make([]structs.ZoneFullInfo, 0))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).CreateInBatches(DeleteZones, 10)
	prw.DB(context.TODO()).CreateInBatches(zones, 10)
	re, err := prw.QueryZones(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, len(zones)+len(DeleteZones), len(re))
	//First convert the query results into a map to facilitate comparison of results
	tmp := make(map[string]structs.ZoneFullInfo)
	for _, num := range re {
		key := num.VendorID + num.RegionID + num.ZoneID
		if _, ok := tmp[key]; !ok {
			tmp[key] = num
		}
	}
	for _, zone := range zones {
		key := zone.VendorID + zone.RegionID + zone.ZoneID
		val, ok := tmp[key]
		assert.Equal(t, true, ok)
		assert.Equal(t, zone.VendorID, val.VendorID)
		assert.Equal(t, zone.VendorName, val.VendorName)
		assert.Equal(t, zone.RegionID, val.RegionID)
		assert.Equal(t, zone.RegionName, val.RegionName)
		assert.Equal(t, zone.ZoneID, val.ZoneID)
		assert.Equal(t, zone.ZoneName, val.ZoneName)
	}

	dzones := []structs.ZoneFullInfo{
		{ZoneID: DeleteZones[0].ZoneID, ZoneName: DeleteZones[0].ZoneName,
			RegionID: DeleteZones[0].RegionID, RegionName: DeleteZones[0].RegionName,
			VendorID: DeleteZones[0].VendorID, VendorName: DeleteZones[0].VendorName},
	}

	err = prw.DeleteZones(context.TODO(), dzones)
	assert.NoError(t, err)

	var zt []structs.ZoneFullInfo
	zt, err = prw.QueryZones(context.TODO())
	assert.Equal(t, len(zones), len(zt))

	dtmp := make(map[string]structs.ZoneFullInfo)
	for _, num := range zt {
		key := num.VendorID + num.RegionID + num.ZoneID
		if _, ok := dtmp[key]; !ok {
			dtmp[key] = num
		}
	}
	for _, zone := range zones {
		key := zone.VendorID + zone.RegionID + zone.ZoneID
		val, ok := dtmp[key]
		assert.Equal(t, true, ok)
		assert.Equal(t, zone.VendorID, val.VendorID)
		assert.Equal(t, zone.VendorName, val.VendorName)
		assert.Equal(t, zone.RegionID, val.RegionID)
		assert.Equal(t, zone.RegionName, val.RegionName)
		assert.Equal(t, zone.ZoneID, val.ZoneID)
		assert.Equal(t, zone.ZoneName, val.ZoneName)
	}

	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_QueryProductComponentProperty(t *testing.T) {
	prw.DB(context.TODO()).CreateInBatches(TiDBALIYUNHZX8650Components, 10)
	prw.DB(context.TODO()).CreateInBatches(TiDBALIYUNHZARM6451Components, 10)
	prw.DB(context.TODO()).CreateInBatches(TiDBALIYUNBJX8650Components, 10)
	prw.DB(context.TODO()).CreateInBatches(TiDBALIYUNBJX8651Components, 10)
	prw.DB(context.TODO()).CreateInBatches(EMALIYUNHZX8610Components, 10)
	type args struct {
		VendorID   string
		RegionID   string
		ProductID  string
		Version    string
		Arch       string
		Status     string
		Components []ProductComponent
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(args args, prw *ProductReadWriter) bool
	}{
		{"QueryTiDBProductComponentsWithAliYunHangzhou50x864", args{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Components: TiDBALIYUNHZX8650Components},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryTiDBProductComponentsWithAliYunHangzhou51ARM64", args{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Components: TiDBALIYUNHZARM6451Components},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryTiDBProductComponentsWithAliYunBeijing50x864", args{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Components: TiDBALIYUNBJX8650Components},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryTiDBProductComponentsWithAliYunBeijing51x864", args{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Components: TiDBALIYUNBJX8651Components},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryEMProductComponentsWithAliYunHangzhou10x864", args{VendorID: AliYun, RegionID: CNHangzhou, ProductID: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Components: EMALIYUNHZX8610Components},
			false, []func(args args, prw *ProductReadWriter) bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prw.QueryProductComponentProperty(context.TODO(), tt.args.VendorID, tt.args.RegionID, tt.args.ProductID,
				tt.args.Version, tt.args.Arch, constants.ProductComponentStatus(tt.args.Status))
			assert.NoError(t, err)

			tmp := make(map[string]structs.ProductComponentPropertyWithZones)

			//First convert the query results into a map to facilitate comparison of results
			for _, component := range result {
				if _, ok := tmp[component.ID]; !ok {
					tmp[component.ID] = component
				}
			}

			for _, item := range tt.args.Components {
				value, ok := tmp[item.ComponentID]
				assert.Equal(t, true, ok)
				if ok {
					assert.Equal(t, item.Name, value.Name)
					assert.Equal(t, item.PurposeType, value.PurposeType)
					assert.Equal(t, item.ComponentID, value.ID)
					assert.Equal(t, item.StartPort, value.StartPort)
					assert.Equal(t, item.EndPort, value.EndPort)
					assert.Equal(t, item.MaxPort, value.MaxPort)
					assert.Equal(t, item.MaxInstance, value.MaxInstance)
					assert.Equal(t, item.MinInstance, value.MinInstance)
				}
			}
		})
	}
	t.Run("QueryProductComponentPropertyWithEmptyParameter", func(t *testing.T) {
		_, err := prw.QueryProductComponentProperty(context.TODO(), "", "", "", "", "", constants.ProductComponentStatus(""))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM product_components")
}

func TestProductReadWriter_CreateProducts(t *testing.T) {
	t.Run("CreateProductsWithEmptyParameter", func(t *testing.T) {
		err := prw.CreateProduct(context.TODO(), Product{}, make([]ProductComponent, 0))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})
	//CreateProducts
	{
		products := make(map[string]Product)
		type args struct {
			Product    Product
			Components []ProductComponent
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
			wants   []func(args args, prw *ProductReadWriter) bool
		}{
			{"CreateTiDBProductWithAliYunHangzhou51ARM64", args{Product: Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}, Components: TiDBALIYUNHZARM6451Components},
				false, []func(args args, prw *ProductReadWriter) bool{}},
			{"CreateTiDBProductWithAliYunHangzhou50x8664", args{Product: Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}, Components: TiDBALIYUNHZX8650Components},
				false, []func(args args, prw *ProductReadWriter) bool{}},
			{"CreateTiDBProductWithAliYunBeijing50x8664", args{Product: Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}, Components: TiDBALIYUNBJX8650Components},
				false, []func(args args, prw *ProductReadWriter) bool{}},
			{"CreateTiDBProductWithAliYunBeijing51x8664", args{Product: Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}, Components: TiDBALIYUNBJX8651Components},
				false, []func(args args, prw *ProductReadWriter) bool{}},
			{"CreateEMProductWithAliYunHangzhou10x8664", args{Product: Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Name: EnterpriseManager, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes}, Components: EMALIYUNHZX8610Components},
				false, []func(args args, prw *ProductReadWriter) bool{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := prw.CreateProduct(context.TODO(), tt.args.Product, tt.args.Components)
				assert.NoError(t, err)
				if nil != err {
					info := tt.args.Product
					key := info.VendorID + info.RegionID + info.ProductID + info.Name + info.Arch + info.Version
					_, ok := products[key]
					if !ok {
						products[key] = tt.args.Product
					}
				}
			})
		}

		var tmp []structs.Product
		all_products, err := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, err)
		tmp, err = prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, err)
		all_products = append(all_products, tmp...)

		dmap := make(map[string]structs.Product)
		for _, it := range all_products {
			key := it.VendorID + it.RegionID + it.ID + it.Name + it.Arch + it.Version
			_, ok := dmap[key]
			if !ok {
				dmap[key] = it
			}
		}

		for key, value := range products {
			item, ok := dmap[key]
			assert.Equal(t, true, ok)
			if ok {
				assert.Equal(t, item.Name, value.Name)
				assert.Equal(t, item.ID, value.ProductID)
				assert.Equal(t, item.VendorID, value.VendorID)
				assert.Equal(t, item.RegionID, value.RegionID)
				assert.Equal(t, item.Version, value.Version)
				assert.Equal(t, item.Arch, value.Arch)
				assert.Equal(t, item.Internal, value.Internal)
				assert.Equal(t, item.Status, value.Status)
			}
		}
		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		prw.DB(context.TODO()).Exec("DELETE FROM products")
		/*for _,pd := range all_products {
			err = prw.DeleteProduct(context.TODO(),pd)
			assert.NoError(t, err)
		}*/
	}

	t.Run("CreateProductsWithDBError", func(t *testing.T) {
		product := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err := prw.CreateProduct(context.TODO(), product, TiDBALIYUNHZARM6451Components)
		assert.NoError(t, err)
		err = prw.CreateProduct(context.TODO(), product, TiDBALIYUNHZARM6451Components)
		assert.Equal(t, errors.CreateProductError, err.(errors.EMError).GetCode())
		newProduct := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Name: TiDB, Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), newProduct, TiDBALIYUNHZARM6451Components)
		assert.Equal(t, errors.CreateProductError, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM product_components")
	prw.DB(context.TODO()).Exec("DELETE FROM products")
}

func TestProductReadWriter_QueryProducts(t *testing.T) {
	t.Run("QueryProductsWithEmptyParameter", func(t *testing.T) {
		_, err := prw.QueryProducts(context.TODO(), "", constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QueryProducts", func(t *testing.T) {

		err := prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		TiDBHZ50X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBHZ50X8664, TiDBALIYUNHZX8650Components)
		assert.NoError(t, err)
		TiDBHZ51ARM64 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBHZ51ARM64, TiDBALIYUNHZARM6451Components)
		assert.NoError(t, err)
		TiDBBJ50X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBBJ50X8664, TiDBALIYUNBJX8650Components)
		assert.NoError(t, err)
		TiDBBJ51X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBBJ51X8664, TiDBALIYUNBJX8651Components)
		assert.NoError(t, err)
		EMHHZ10X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes}
		err = prw.CreateProduct(context.TODO(), EMHHZ10X8664, EMALIYUNHZX8610Components)
		assert.NoError(t, err)
		var allSourceProducts []Product
		allSourceProducts = append(allSourceProducts, TiDBHZ50X8664)
		allSourceProducts = append(allSourceProducts, TiDBHZ51ARM64)
		allSourceProducts = append(allSourceProducts, TiDBBJ50X8664)
		allSourceProducts = append(allSourceProducts, TiDBBJ51X8664)
		allSourceProducts = append(allSourceProducts, EMHHZ10X8664)

		allProducts, err := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, err)
		tmp, er := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductYes)
		assert.NoError(t, er)
		allProducts = append(allProducts, tmp...)

		temp := make(map[string]structs.Product)
		for _, it := range allProducts {
			key := it.VendorID + it.RegionID + it.ID + it.Name + it.Arch + it.Version
			_, ok := temp[key]
			if !ok {
				temp[key] = it
			}
		}

		for _, value := range allSourceProducts {
			key := value.VendorID + value.RegionID + value.ProductID + value.Name + value.Arch + value.Version
			item, ok := temp[key]
			assert.Equal(t, true, ok)
			if ok {
				assert.Equal(t, item.Name, value.Name)
				assert.Equal(t, item.ID, value.ProductID)
				assert.Equal(t, item.VendorID, value.VendorID)
				assert.Equal(t, item.RegionID, value.RegionID)
				assert.Equal(t, item.Version, value.Version)
				assert.Equal(t, item.Arch, value.Arch)
			}
		}
		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		prw.DB(context.TODO()).Exec("DELETE FROM products")
		prw.DB(context.TODO()).Exec("DELETE FROM zones")
	})

	t.Run("QueryProductsWithDBError", func(t *testing.T) {

	})

	prw.DB(context.TODO()).Exec("DELETE FROM product_components")
	prw.DB(context.TODO()).Exec("DELETE FROM products")
	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_DeleteProduct(t *testing.T) {
	t.Run("DeleteProductWithEmptyParameter", func(t *testing.T) {
		err := prw.DeleteProduct(context.TODO(), structs.Product{})
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})
	t.Run("DeleteProduct", func(t *testing.T) {
		err := prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		TiDBHZ50X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBHZ50X8664, TiDBALIYUNHZX8650Components)
		assert.NoError(t, err)
		TiDBHZ51ARM64 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBHZ51ARM64, TiDBALIYUNHZARM6451Components)
		assert.NoError(t, err)
		TiDBBJ50X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBBJ50X8664, TiDBALIYUNBJX8650Components)
		assert.NoError(t, err)
		TiDBBJ51X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBBJ51X8664, TiDBALIYUNBJX8651Components)
		assert.NoError(t, err)
		EMHHZ10X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes}
		err = prw.CreateProduct(context.TODO(), EMHHZ10X8664, EMALIYUNHZX8610Components)
		assert.NoError(t, err)

		allProducts, e := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, e)
		tmp, er := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductYes)
		assert.NoError(t, er)
		allProducts = append(allProducts, tmp...)

		for _, value := range allProducts {
			err = prw.DeleteProduct(context.TODO(), value)
			assert.NoError(t, err)
		}

		tmp, errr := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, errr)
		assert.Equal(t, true, len(tmp) == 0)

		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		prw.DB(context.TODO()).Exec("DELETE FROM products")
		prw.DB(context.TODO()).Exec("DELETE FROM zones")
	})

	t.Run("CreateProductsWithDBError", func(t *testing.T) {
		err := prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)
		TiDBHZ50X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		err = prw.CreateProduct(context.TODO(), TiDBHZ50X8664, TiDBALIYUNHZX8650Components)
		assert.NoError(t, err)

		tmp, er := prw.QueryProducts(context.TODO(), AliYun, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, er)
		assert.Equal(t, 1, len(tmp))

		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		err = prw.DeleteProduct(context.TODO(), tmp[0])
		assert.Equal(t, errors.DeleteProductError, err.(errors.EMError).GetCode())

		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		prw.DB(context.TODO()).Exec("DELETE FROM products")
		prw.DB(context.TODO()).Exec("DELETE FROM zones")
	})

	prw.DB(context.TODO()).Exec("DELETE FROM product_components")
	prw.DB(context.TODO()).Exec("DELETE FROM products")
	prw.DB(context.TODO()).Exec("DELETE FROM zones")
}

func TestProductReadWriter_QueryProductDetail(t *testing.T) {

	t.Run("QueryProductDetailWithEmptyParameter", func(t *testing.T) {
		_, err := prw.QueryProductDetail(context.TODO(), "", "", "", constants.ProductStatusOffline, constants.EMInternalProductYes)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QueryProductDetail", func(t *testing.T) {
		var specs []Spec
		specs = append(specs, StorageSpecs...)
		specs = append(specs, ComputeSpecs...)
		specs = append(specs, ScheduleSpecs...)
		err := prw.CreateSpecs(context.TODO(), specs)
		assert.NoError(t, err)

		//init zones
		err = prw.CreateZones(context.TODO(), zones)
		assert.NoError(t, err)

		//storage all products

		TiDBHZ50X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		TiDBHZ51ARM64 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		TiDBBJ50X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		TiDBBJ51X8664 := Product{VendorID: AliYun, RegionID: CNBeijing, ProductID: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo}
		EMHHZ10X8664 := Product{VendorID: AliYun, RegionID: CNHangzhou, ProductID: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes}

		//init specs
		type Args struct {
			Product    Product
			Components []ProductComponent
		}
		args := []Args{{Product: TiDBHZ50X8664, Components: TiDBALIYUNHZX8650Components},
			{Product: TiDBHZ51ARM64, Components: TiDBALIYUNHZARM6451Components},
			{Product: TiDBBJ50X8664, Components: TiDBALIYUNBJX8650Components},
			{Product: TiDBBJ51X8664, Components: TiDBALIYUNBJX8651Components},
			{Product: EMHHZ10X8664, Components: EMALIYUNHZX8610Components}}

		for _, arg := range args {
			err = prw.CreateProduct(context.TODO(), arg.Product, arg.Components)
			assert.NoError(t, err)

		}

		actualDetail, er := prw.QueryProductDetail(context.TODO(), AliYun, CNHangzhou, TiDB, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, er)
		assert.Equal(t, 1, len(actualDetail))

		actualBJDetail, e := prw.QueryProductDetail(context.TODO(), AliYun, CNBeijing, TiDB, constants.ProductStatusOnline, constants.EMInternalProductNo)
		assert.NoError(t, e)
		assert.Equal(t, 1, len(actualBJDetail))

		EMDetail, er := prw.QueryProductDetail(context.TODO(), AliYun, CNHangzhou, EnterpriseManager, constants.ProductStatusOnline, constants.EMInternalProductYes)
		assert.NoError(t, er)
		assert.Equal(t, 1, len(EMDetail))

		prw.DB(context.TODO()).Exec("DELETE FROM specs")
		prw.DB(context.TODO()).Exec("DELETE FROM product_components")
		prw.DB(context.TODO()).Exec("DELETE FROM products")
		prw.DB(context.TODO()).Exec("DELETE FROM zones")
	})
}

func TestProductReadWriter_CreateSpecs(t *testing.T) {
	t.Run("CreateSpecs", func(t *testing.T) {
		var specs []Spec
		specs = append(specs, StorageSpecs...)
		specs = append(specs, ComputeSpecs...)
		specs = append(specs, ScheduleSpecs...)

		err := prw.CreateSpecs(context.TODO(), specs)
		assert.NoError(t, err)

		re, er := prw.QuerySpecs(context.TODO())
		assert.NoError(t, er)
		assert.Equal(t, len(StorageSpecs)+len(ComputeSpecs)+len(ScheduleSpecs), len(re))

		//First convert the query results into a map to facilitate comparison of results
		tmp := make(map[string]structs.SpecInfo)
		for _, value := range re {
			if _, ok := tmp[value.ID]; !ok {
				tmp[value.ID] = value
			}
		}
		for _, spec := range specs {
			val, ok := tmp[spec.ID]
			assert.Equal(t, true, ok)
			assert.Equal(t, spec.ID, val.ID)
			assert.Equal(t, spec.Name, val.Name)
			assert.Equal(t, spec.Memory, val.Memory)
			assert.Equal(t, spec.CPU, val.CPU)
			assert.Equal(t, spec.DiskType, val.DiskType)
			assert.Equal(t, spec.PurposeType, val.PurposeType)
			assert.Equal(t, spec.Status, val.Status)
		}
	})

	t.Run("CreateSpecsWithEmptyParameter", func(t *testing.T) {
		err := prw.CreateSpecs(context.TODO(), make([]Spec, 0))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM specs")
}

func TestProductReadWriter_QuerySpecs(t *testing.T) {
	t.Run("QuerySpecs", func(t *testing.T) {
		var specs []Spec
		specs = append(specs, StorageSpecs...)
		specs = append(specs, ComputeSpecs...)
		specs = append(specs, ScheduleSpecs...)
		prw.DB(context.TODO()).CreateInBatches(specs, 10)
		re, err := prw.QuerySpecs(context.TODO())
		assert.NoError(t, err)
		//First convert the query results into a map to facilitate comparison of results
		tmp := make(map[string]structs.SpecInfo)
		for _, value := range re {
			if _, ok := tmp[value.ID]; !ok {
				tmp[value.ID] = value
			}
		}
		for _, spec := range specs {
			val, ok := tmp[spec.ID]
			assert.Equal(t, true, ok)
			assert.Equal(t, spec.ID, val.ID)
			assert.Equal(t, spec.Name, val.Name)
			assert.Equal(t, spec.Memory, val.Memory)
			assert.Equal(t, spec.CPU, val.CPU)
			assert.Equal(t, spec.DiskType, val.DiskType)
			assert.Equal(t, spec.PurposeType, val.PurposeType)
			assert.Equal(t, spec.Status, val.Status)
		}
	})

	t.Run("CreateSpecsWithDBError", func(t *testing.T) {
		_, err := prw.QuerySpecs(context.TODO())
		assert.NoError(t, err)
		//assert.Equal(t,errors.TIEM_PARAMETER_INVALID,err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM specs")
}

func TestProductReadWriter_DeleteSpecs(t *testing.T) {
	t.Run("DeleteSpecs", func(t *testing.T) {
		var err error
		var specs []Spec
		specs = append(specs, StorageSpecs...)
		specs = append(specs, ComputeSpecs...)
		specs = append(specs, ScheduleSpecs...)

		err = prw.CreateSpecs(context.TODO(), specs)
		assert.NoError(t, err)

		re, er := prw.QuerySpecs(context.TODO())
		assert.NoError(t, er)
		assert.Equal(t, len(StorageSpecs)+len(ComputeSpecs)+len(ScheduleSpecs), len(re))

		var specsID []string
		for _, item := range specs {
			specsID = append(specsID, item.ID)
		}

		err = prw.DeleteSpecs(context.TODO(), specsID)
		assert.NoError(t, err)
		err = prw.DeleteSpecs(context.TODO(), specsID)
		assert.NoError(t, err)
	})

	t.Run("DeleteSpecsWithEmptyParameter", func(t *testing.T) {
		err := prw.DeleteSpecs(context.TODO(), make([]string, 0))
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	prw.DB(context.TODO()).Exec("DELETE FROM specs")
}
