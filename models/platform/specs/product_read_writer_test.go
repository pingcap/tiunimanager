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

package specs

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/stretchr/testify/assert"
	"testing"
)

var prw *ProductReadWriter

const (
	TiDB                     = "TiDB"
	TiDBVersion50            = "5.0.0"
	TiDBVersion51            = "5.1.0"
	EnterpriseManager        = "EnterpriseManager"
	EnterpriseManagerVersion = "1.0.0"
)

var vendors = []Vendor{
	{VendorID: "Aliyun", Name: "Alibaba cloud"},
}

var regions = []Region{
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", Name: "East China(Hangzhou)"},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", Name: "East China(Beijing)"}}

var zones = []Zone{
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ZoneID: "CN-BEIJING-H", Name: "ZONE(H)"},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ZoneID: "CN-BEIJING-G", Name: "ZONE(G)"},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ZoneID: "CN-HANGZHOU-H", Name: "ZONE(H)"},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ZoneID: "CN-HANGZHOU-G", Name: "ZONE(G)"},
}

var products = []Product{
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "ARM64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "EnterpriseManager", Name: "EnterpriseManager", Version: "1.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes},
}

var nointernalproducts = []Product{
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "ARM64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}

var hangzhouproducts = []Product{
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "ARM64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}
var hangzhouproductctrplan = []Product{
	{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "EnterpriseManager", Name: "EnterpriseManager", Version: "1.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes},
}
var beijingproducts = []Product{
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.0.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Name: "TiDB", Version: "5.1.0", Arch: "x86_64", Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
}

var components = []ProductComponent{
	//TiDB v5.0.0
	{ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "PD", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ComponentID: "TiCDC", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},

	//TiDB v5.1.0
	{ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "PD", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ComponentID: "TiCDC", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},

	//Enterprise Manager v1.0.0
	{ComponentID: "cluster-server", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "cluster-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "openapi-server", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "openapi-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Grafana", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDB50Components = []ProductComponent{
	//TiDB v5.0.0
	{ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "PD", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ComponentID: "TiCDC", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: "5.0.0", Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}
var TiDB51Components = []ProductComponent{
	//TiDB v5.1.0
	{ComponentID: "TiDB", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Compute Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiKV", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "TiFlash", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Column Storage Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ComponentID: "PD", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Schedule Engine", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ComponentID: "TiCDC", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "CDC", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ComponentID: "Grafana", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "NodeExporter", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "NodeExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "BlackboxExporter", ProductID: "TiDB", ProductVersion: "5.1.0", Name: "BlackboxExporter", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}
var EnterpriseManager10Components = []ProductComponent{
	//Enterprise Manager v1.0.0
	{ComponentID: "cluster-server", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "cluster-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "openapi-server", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "openapi-server", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Grafana", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Monitor GUI", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "Prometheus", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Monitor", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ComponentID: "AlertManger", ProductID: "EnterpriseManager", ProductVersion: "1.0.0", Name: "Alert", Status: string(constants.ProductSpecStatusOnline), PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var specs = []ResourceSpec{
	/*CN-BEIJING-H X86_64***/
	{ResourceSpecID: "c2.g.large", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.2xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.large", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.2xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.large", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.xlarge", ZoneID: "CN-BEIJING-H", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	/*CN-BEIJING-G X86_64*/
	{ResourceSpecID: "c2.g.large", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.xlarge", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.2xlarge", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.large", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.xlarge", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.large", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.xlarge", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.large", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.xlarge", ZoneID: "CN-BEIJING-G", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},

	/*CN-HANGZHOU-G X86_64*/
	{ResourceSpecID: "c1.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.2xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd1.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s1.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs1.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.large", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 64, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.xlarge", ZoneID: "CN-HANGZHOU-G", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},

	/*CN-HANGZHOU-H X86_64*/
	{ResourceSpecID: "c1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.2xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "8C16G", CPU: 8, Memory: 64, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "x86_64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},

	/*CN-HANGZHOU-H ARM64*/
	{ResourceSpecID: "c1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c2.g.2xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "c3.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "sd2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "s2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs1.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.large", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "8C16G", CPU: 8, Memory: 64, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ResourceSpecID: "cs2.g.xlarge", ZoneID: "CN-HANGZHOU-H", Arch: "ARM64", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
}

func initAllData() {
	prw.DB(context.TODO()).CreateInBatches(vendors, 512)
	prw.DB(context.TODO()).CreateInBatches(regions, 10)
	prw.DB(context.TODO()).CreateInBatches(zones, 10)
	prw.DB(context.TODO()).CreateInBatches(products, 512)
	prw.DB(context.TODO()).CreateInBatches(components, 512)
	prw.DB(context.TODO()).CreateInBatches(specs, 512)
}

func truncateAllData() {
	prw.DB(context.TODO()).Exec("DELETE FROM vendors")
	prw.DB(context.TODO()).Exec("DELETE FROM regions")
	prw.DB(context.TODO()).Exec("DELETE FROM zones")
	prw.DB(context.TODO()).Exec("DELETE FROM products")
	prw.DB(context.TODO()).Exec("DELETE FROM product_components")
	prw.DB(context.TODO()).Exec("DELETE FROM resource_specs")
}

func TestProductReadWriter_QueryZones(t *testing.T) {
	initAllData()
	tmp := make(map[string]structs.ZoneDetail)
	re, err := prw.QueryZones(context.TODO())
	assert.NoError(t, err)
	//First convert the query results into a map to facilitate comparison of results
	for _, num := range re {
		key := num.RegionID + num.ZoneID
		if _, ok := tmp[key]; !ok {
			tmp[key] = num
		}
	}

	//Compare the data queried from the database for consistency
	for _, item := range zones {
		key := item.RegionID + item.ZoneID
		value, ok := tmp[key]
		assert.Equal(t, true, ok)
		if ok {
			assert.Equal(t, item.Name, value.ZoneName)
			assert.NoError(t, err)
		}
	}

	truncateAllData()
}

func TestProductReadWriter_QueryProductComponentProperty(t *testing.T) {

	initAllData()
	tmp50 := make(map[string]structs.ProductComponentProperty)
	re, err := prw.QueryProductComponentProperty(context.TODO(), TiDB, TiDBVersion50, constants.ProductComponentStatusOnline)
	assert.NoError(t, err)
	//First convert the query results into a map to facilitate comparison of results
	for _, item := range re {
		if _, ok := tmp50[item.ID]; !ok {
			tmp50[item.ID] = item
		}
	}

	//Compare the data queried from the database for consistency
	for _, item := range TiDB50Components {
		value, ok := tmp50[item.ComponentID]
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

	tmp51 := make(map[string]structs.ProductComponentProperty)
	re, err = prw.QueryProductComponentProperty(context.TODO(), TiDB, TiDBVersion51, constants.ProductComponentStatusOnline)
	assert.NoError(t, err)
	//First convert the query results into a map to facilitate comparison of results
	for _, item := range re {
		if _, ok := tmp51[item.ID]; !ok {
			tmp51[item.ID] = item
		}
	}

	//Compare the data queried from the database for consistency
	for _, item := range TiDB51Components {
		value, ok := tmp51[item.ComponentID]
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

	tmp10 := make(map[string]structs.ProductComponentProperty)
	re, err = prw.QueryProductComponentProperty(context.TODO(), EnterpriseManager, EnterpriseManagerVersion, constants.ProductComponentStatusOnline)
	assert.NoError(t, err)
	//First convert the query results into a map to facilitate comparison of results
	for _, item := range re {
		if _, ok := tmp10[item.ID]; !ok {
			tmp10[item.ID] = item
		}
	}

	//Compare the data queried from the database for consistency
	for _, item := range EnterpriseManager10Components {
		value, ok := tmp10[item.ComponentID]
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
	truncateAllData()
}

func TestProductReadWriter_QueryProducts(t *testing.T) {
	type args struct {
		VendorID string
		Internal constants.EMInternalProduct
		Status   constants.ProductStatus
		Products []Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(args args, prw *ProductReadWriter) bool
	}{
		{"QueryProductWithAliyun", args{VendorID: "Aliyun", Status: constants.ProductStatusOnline, Products: nointernalproducts, Internal: constants.EMInternalProductNo},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryProductWithAliyunAndInternal", args{VendorID: "Aliyun", Status: constants.ProductStatusOnline, Products: hangzhouproductctrplan, Internal: constants.EMInternalProductYes},
			false, []func(args args, prw *ProductReadWriter) bool{}},
	}

	initAllData()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prw.QueryProducts(context.TODO(), tt.args.VendorID, tt.args.Status, tt.args.Internal)
			assert.NoError(t, err)
			tmp := make(map[string]structs.Product)
			for _, rt := range result {
				key := rt.RegionID + rt.ID + rt.Version + rt.Arch
				_, ok := tmp[key]
				if !ok {
					tmp[key] = rt
				}
			}
			for _, item := range tt.args.Products {
				key := item.RegionID + item.ProductID + item.Version + item.Arch
				value, ok := tmp[key]
				assert.Equal(t, true, ok)
				assert.Equal(t, item.Name, value.Name)
				assert.Equal(t, item.RegionID, value.RegionID)
				assert.Equal(t, item.Version, value.Version)
				assert.Equal(t, item.Arch, value.Arch)
				assert.Equal(t, item.ProductID, value.ID)
			}
		})
	}

	truncateAllData()
}

func TestProductReadWriter_QueryProductDetail(t *testing.T) {
	type args struct {
		VendorID  string
		RegionID  string
		ProductID string
		Internal  constants.EMInternalProduct
		Status    constants.ProductStatus
		Products  []Product
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(args args, prw *ProductReadWriter) bool
	}{
		{"QueryProductDetailWithHZTiDB", args{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "TiDB", Status: constants.ProductStatusOnline, Products: hangzhouproducts, Internal: constants.EMInternalProductNo},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryProductDetailWithBJTiDB", args{VendorID: "Aliyun", RegionID: "CN-BEIJING", ProductID: "TiDB", Status: constants.ProductStatusOnline, Products: beijingproducts, Internal: constants.EMInternalProductNo},
			false, []func(args args, prw *ProductReadWriter) bool{}},
		{"QueryProductDetailWithHZEnterpriseManager", args{VendorID: "Aliyun", RegionID: "CN-HANGZHOU", ProductID: "EnterpriseManager", Status: constants.ProductStatusOnline, Products: hangzhouproductctrplan, Internal: constants.EMInternalProductYes},
			false, []func(args args, prw *ProductReadWriter) bool{}},
	}
	initAllData()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prw.QueryProductDetail(context.TODO(), tt.args.VendorID, tt.args.RegionID, tt.args.ProductID, tt.args.Status, tt.args.Internal)
			assert.NoError(t, err)
			for _, item := range tt.args.Products {
				value, ok := result[item.ProductID]
				assert.Equal(t, true, ok)
				assert.Equal(t, item.Name, value.Name)
				version, vok := value.Versions[item.Version]
				assert.Equal(t, true, vok)
				if vok {
					assert.Equal(t, version.Arch, item.Arch)
					var components []ProductComponent
					if TiDB == item.ProductID {
						if TiDBVersion50 == item.Version {
							components = TiDB50Components
						} else if TiDBVersion51 == item.Version {
							components = TiDB51Components
						}
					} else if EnterpriseManager == item.ProductID {
						if EnterpriseManagerVersion == item.Version {
							components = EnterpriseManager10Components
						}
					}
					for _, it := range components {
						component, cok := version.Components[it.ComponentID]
						assert.Equal(t, true, cok)
						assert.Equal(t, it.Name, component.Name)
						assert.Equal(t, it.PurposeType, component.PurposeType)
						assert.Equal(t, it.StartPort, component.StartPort)
						assert.Equal(t, it.EndPort, component.EndPort)
						assert.Equal(t, it.MaxPort, component.MaxPort)
						assert.Equal(t, it.MinInstance, component.MinInstance)
						assert.Equal(t, it.MaxInstance, component.MaxInstance)
					}
				}
			}
		})
	}
	truncateAllData()
}
