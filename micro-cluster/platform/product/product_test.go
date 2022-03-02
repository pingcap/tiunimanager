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
 * limitations under the License                                              *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: product_test.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/30
*******************************************************************************/

package product

import (
	"context"
	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"reflect"
	"testing"
)

func init() {
	models.MockDB()
}

const (
	TiDB                     = "TiDB"
	TiDBVersion50            = "5.0.0"
	TiDBVersion51            = "5.1.0"
	EnterpriseManager        = "EnterpriseManager"
	EnterpriseManagerVersion = "1.0.0"
	AliYun                   = "Aliyun"
	CNHangzhou               = "CN-HANGZHOU"
	CNBeijing                = "CN-BEIJING"
	CNBeijingName            = "North China(Beijing)"
	CNHangzhouName           = "East China(Hangzhou)"
	CNBeijingG               = "CN-BEIJING-G"
	CNBeijingH               = "CN-BEIJING-H"
	CNHangzhouH              = "CN-HANGZHOU-H"
	CNHangzhouG              = "CN-HANGZHOU-G"
	ZoneNameG                = "ZONE(G)"
	ZoneNameH                = "ZONE(H)"
)

var products = []structs.Product{
	{VendorID: AliYun, RegionID: CNHangzhou, ID: TiDB, Name: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ID: TiDB, Name: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchArm64), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ID: TiDB, Name: TiDB, Version: TiDBVersion50, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNBeijing, ID: TiDB, Name: TiDB, Version: TiDBVersion51, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductNo},
	{VendorID: AliYun, RegionID: CNHangzhou, ID: EnterpriseManager, Name: EnterpriseManager, Version: EnterpriseManagerVersion, Arch: string(constants.ArchX8664), Status: string(constants.ProductStatusOnline), Internal: constants.EMInternalProductYes},
}

var TiDBALIYUNHZX8650Components = []structs.ProductComponentProperty{
	//TiDB v5.0.0
	{ID: "TiDB", Name: "Compute Engine", PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiKV", Name: "Storage Engine", PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiFlash", Name: "Column Storage Engine", PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ID: "PD", Name: "Schedule Engine", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ID: "CDC", Name: "CDC", PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ID: "Grafana", Name: "Monitor GUI", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Prometheus", Name: "Monitor", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "AlertManger", Name: "Alert", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "NodeExporter", Name: "NodeExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "BlackboxExporter", Name: "BlackboxExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNBJX8650Components = []structs.ProductComponentProperty{
	//TiDB v5.0.0
	{ID: "TiDB", Name: "Compute Engine", PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiKV", Name: "Storage Engine", PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiFlash", Name: "Column Storage Engine", PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ID: "PD", Name: "Schedule Engine", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ID: "CDC", Name: "CDC", PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ID: "Grafana", Name: "Monitor GUI", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Prometheus", Name: "Monitor", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "AlertManger", Name: "Alert", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "NodeExporter", Name: "NodeExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "BlackboxExporter", Name: "BlackboxExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNBJX8651Components = []structs.ProductComponentProperty{
	//TiDB v5.1.0
	{ID: "TiDB", Name: "Compute Engine", PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiKV", Name: "Storage Engine", PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiFlash", Name: "Column Storage Engine", PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ID: "PD", Name: "Schedule Engine", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ID: "CDC", Name: "CDC", PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ID: "Grafana", Name: "Monitor GUI", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Prometheus", Name: "Monitor", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "AlertManger", Name: "Alert", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "NodeExporter", Name: "NodeExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "BlackboxExporter", Name: "BlackboxExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var TiDBALIYUNHZARM6451Components = []structs.ProductComponentProperty{
	//TiDB v5.1.0
	{ID: "TiDB", Name: "Compute Engine", PurposeType: "Compute", StartPort: 10000, EndPort: 10020, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiKV", Name: "Storage Engine", PurposeType: "Storage", StartPort: 10020, EndPort: 10040, MaxPort: 2, MinInstance: 1, MaxInstance: 10240},
	{ID: "TiFlash", Name: "Column Storage Engine", PurposeType: "Storage", StartPort: 10120, EndPort: 10180, MaxPort: 6, MinInstance: 1, MaxInstance: 10240},
	{ID: "PD", Name: "Schedule Engine", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 8, MinInstance: 1, MaxInstance: 7},
	{ID: "CDC", Name: "CDC", PurposeType: "Schedule", StartPort: 10180, EndPort: 10200, MaxPort: 2, MinInstance: 1, MaxInstance: 512},
	{ID: "Grafana", Name: "Monitor GUI", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Prometheus", Name: "Monitor", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "AlertManger", Name: "Alert", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "NodeExporter", Name: "NodeExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "BlackboxExporter", Name: "BlackboxExporter", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var EMALIYUNHZX8610Components = []structs.ProductComponentProperty{
	//Enterprise Manager v1.0.0
	{ID: "cluster-server", Name: "cluster-server", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "openapi-server", Name: "openapi-server", PurposeType: "Schedule", StartPort: 11000, EndPort: 12000, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Grafana", Name: "Monitor GUI", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "Prometheus", Name: "Monitor", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
	{ID: "AlertManger", Name: "Alert", PurposeType: "Schedule", StartPort: 10040, EndPort: 10120, MaxPort: 1, MinInstance: 1, MaxInstance: 1},
}

var zones = []structs.ZoneFullInfo{
	{ZoneID: CNBeijingG, ZoneName: ZoneNameG, RegionID: CNBeijing, RegionName: CNBeijingName, VendorID: AliYun, VendorName: AliYun, Comment: ""},
	{ZoneID: CNBeijingH, ZoneName: ZoneNameH, RegionID: CNBeijing, RegionName: CNBeijingName, VendorID: AliYun, VendorName: AliYun, Comment: ""},
	{ZoneID: CNHangzhouG, ZoneName: ZoneNameG, RegionID: CNHangzhou, RegionName: CNHangzhouName, VendorID: AliYun, VendorName: AliYun, Comment: ""},
	{ZoneID: CNHangzhouH, ZoneName: ZoneNameH, RegionID: CNHangzhou, RegionName: CNHangzhouName, VendorID: AliYun, VendorName: AliYun, Comment: ""},
}

var computeSpecs = []structs.SpecInfo{
	{ID: "c1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.large", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.xlarge", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c2.g.2xlarge", Name: "16C32G", CPU: 16, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c3.g.large", Name: "8C32G", CPU: 8, Memory: 32, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "c3.g.xlarge", Name: "16C64G", CPU: 16, Memory: 64, DiskType: "SSD", PurposeType: "Compute", Status: string(constants.ProductSpecStatusOnline)},
}
var scheduleSpecs = []structs.SpecInfo{
	{ID: "sd1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "sd2.g.large", Name: "4C8G", CPU: 4, Memory: 8, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "sd2.g.xlarge", Name: "8C16G", CPU: 8, Memory: 16, DiskType: "SSD", PurposeType: "Schedule", Status: string(constants.ProductSpecStatusOnline)},
}
var storageSpecs = []structs.SpecInfo{
	{ID: "s1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "s2.g.large", Name: "8C64G", CPU: 8, Memory: 64, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "s2.g.xlarge", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "NVMeSSD", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs1.g.large", Name: "2C2G", CPU: 2, Memory: 2, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs2.g.large", Name: "8C16G", CPU: 8, Memory: 64, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
	{ID: "cs2.g.xlarge", Name: "16C128G", CPU: 16, Memory: 128, DiskType: "SATA", PurposeType: "Storage", Status: string(constants.ProductSpecStatusOnline)},
}

func TestProductManager_CreateZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.CreateZonesReq{
		Zones: zones,
	}

	t.Run("CreateZones", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateZones(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.CreateZones(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("CreateZonesWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateZones(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.CreateZones(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("CreateZonesWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateZones(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.CreateZonesError, "create zone failed")).AnyTimes()
		_, err := mgr.CreateZones(context.TODO(), msg)
		assert.Equal(t, errors.CreateZonesError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_DeleteZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.DeleteZoneReq{
		Zones: zones,
	}

	t.Run("DeleteZones", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteZones(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.DeleteZones(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("DeleteZonesWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteZones(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.DeleteZones(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("DeleteZonesWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteZones(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.DeleteZonesError, "delete zone failed")).AnyTimes()
		_, err := mgr.DeleteZones(context.TODO(), msg)
		assert.Equal(t, errors.DeleteZonesError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_QueryZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	t.Run("QueryZones", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryZones(gomock.Any()).Return(zones, nil).AnyTimes()
		resp, err := mgr.QueryZones(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Vendors))
		assert.Equal(t, "Aliyun", resp.Vendors["Aliyun"].VendorInfo.Name)
		assert.Equal(t, 2, len(resp.Vendors["Aliyun"].Regions))
		assert.Equal(t, "East China(Hangzhou)", resp.Vendors["Aliyun"].Regions["CN-HANGZHOU"].Name)
	})

	t.Run("QueryZonesWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryZones(gomock.Any()).Return(make([]structs.ZoneFullInfo, 0), errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.QueryZones(context.TODO())
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QueryZonesWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryZones(gomock.Any()).Return(make([]structs.ZoneFullInfo, 0), errors.NewErrorf(errors.QueryZoneScanRowError, "scan data database error")).AnyTimes()
		_, err := mgr.QueryZones(context.TODO())
		assert.Equal(t, errors.QueryZoneScanRowError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_CreateSpecs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.CreateSpecsReq{}
	msg.Specs = append(scheduleSpecs)
	msg.Specs = append(storageSpecs)
	msg.Specs = append(computeSpecs)

	t.Run("CreateSpecs", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateSpecs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.CreateSpecs(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("CreateSpecsWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateSpecs(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.CreateSpecs(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("CreateSpecsWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateSpecs(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.CreateSpecsError, "create zone failed")).AnyTimes()
		_, err := mgr.CreateSpecs(context.TODO(), msg)
		assert.Equal(t, errors.CreateSpecsError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_DeleteSpecs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.DeleteSpecsReq{}
	for _, value := range scheduleSpecs {
		msg.SpecIDs = append(msg.SpecIDs, value.ID)
	}

	t.Run("DeleteSpecs", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteSpecs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.DeleteSpecs(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("DeleteSpecsWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteSpecs(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.DeleteSpecs(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("DeleteSpecsWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteSpecs(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.DeleteSpecsError, "delete zone failed")).AnyTimes()
		_, err := mgr.DeleteSpecs(context.TODO(), msg)
		assert.Equal(t, errors.DeleteSpecsError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_QuerySpecs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	t.Run("QuerySpecs", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QuerySpecs(gomock.Any()).Return(scheduleSpecs, nil).AnyTimes()
		resp, err := mgr.QuerySpecs(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, true, reflect.DeepEqual(resp.Specs, scheduleSpecs))
	})

	t.Run("QuerySpecsWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QuerySpecs(gomock.Any()).Return([]structs.SpecInfo{{}}, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.QuerySpecs(context.TODO())
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QuerySpecsWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QuerySpecs(gomock.Any()).Return([]structs.SpecInfo{{}}, errors.NewErrorf(errors.QuerySpecScanRowError, "scan data database error")).AnyTimes()
		_, err := mgr.QuerySpecs(context.TODO())
		assert.Equal(t, errors.QuerySpecScanRowError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_CreateProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.CreateProductReq{
		ProductInfo: products[0],
		Components:  TiDBALIYUNHZX8650Components,
	}

	t.Run("CreateProduct", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateProduct(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.CreateProduct(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("CreateProductWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateProduct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.CreateProduct(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("CreateProductWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().CreateProduct(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.CreateProductError, "create zone failed")).AnyTimes()
		_, err := mgr.CreateProduct(context.TODO(), msg)
		assert.Equal(t, errors.CreateProductError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_DeleteProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.DeleteProductReq{
		ProductInfo: products[0],
	}

	t.Run("DeleteProduct", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteProduct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		_, err := mgr.DeleteProduct(context.TODO(), msg)
		assert.NoError(t, err)
	})

	t.Run("DeleteProductWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteProduct(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.DeleteProduct(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("DeleteProductWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().DeleteProduct(gomock.Any(), gomock.Any()).Return(errors.NewErrorf(errors.DeleteProductError, "create zone failed")).AnyTimes()
		_, err := mgr.DeleteProduct(context.TODO(), msg)
		assert.Equal(t, errors.DeleteProductError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_QueryProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.QueryAvailableProductsReq{
		VendorID:        AliYun,
		Status:          string(constants.ProductStatusOnline),
		InternalProduct: constants.EMInternalProductNo,
	}
	t.Run("QueryProducts", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryProducts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(products, nil).AnyTimes()
		resp, err := mgr.QueryProducts(context.TODO(), msg)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.Products))
		assert.Equal(t, 2, len(resp.Products["CN-HANGZHOU"]))
		assert.Equal(t, 2, len(resp.Products["CN-HANGZHOU"]["TiDB"]))
		assert.Equal(t, 1, len(resp.Products["CN-HANGZHOU"]["TiDB"]["X86_64"]))
		assert.Equal(t, "TiDB", resp.Products["CN-HANGZHOU"]["TiDB"]["X86_64"]["5.0.0"].Name)

	})

	t.Run("QueryProductsWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryProducts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Product{{}}, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.QueryProducts(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QueryProductsWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryProducts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]structs.Product{{}}, errors.NewErrorf(errors.QueryProductsScanRowError, "scan data database error")).AnyTimes()
		_, err := mgr.QueryProducts(context.TODO(), msg)
		assert.Equal(t, errors.QueryProductsScanRowError, err.(errors.EMError).GetCode())
	})
}

func TestProductManager_QueryProductDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewProductManager()
	msg := message.QueryProductDetailReq{
		VendorID:        AliYun,
		ProductID:       TiDB,
		RegionID:        CNHangzhou,
		Status:          string(constants.ProductStatusOnline),
		InternalProduct: constants.EMInternalProductNo,
	}
	t.Run("QueryProductDetail", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)

		type Args struct {
			Product    structs.Product
			Components []structs.ProductComponentProperty
		}
		Products := make(map[string]structs.ProductDetail)

		prw.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(Products, nil).AnyTimes()
		resp, err := mgr.QueryProductDetail(context.TODO(), msg)
		assert.NoError(t, err)
		assert.Equal(t, true, reflect.DeepEqual(resp.Products, Products))
	})

	t.Run("QueryProductsWithEmptyParameter", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make(map[string]structs.ProductDetail, 0),
			errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "invalid parameter")).AnyTimes()
		_, err := mgr.QueryProductDetail(context.TODO(), msg)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("QueryProductsWithDBError", func(t *testing.T) {
		prw := mock_product.NewMockProductReadWriterInterface(ctrl)
		models.SetProductReaderWriter(prw)
		prw.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make(map[string]structs.ProductDetail, 0),
			errors.NewErrorf(errors.QueryProductsScanRowError, "scan data database error")).AnyTimes()
		_, err := mgr.QueryProductDetail(context.TODO(), msg)
		assert.Equal(t, errors.QueryProductsScanRowError, err.(errors.EMError).GetCode())
	})
}
