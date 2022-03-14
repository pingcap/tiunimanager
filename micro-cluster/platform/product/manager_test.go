/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 * @File: manager_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/product"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocksystem"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManager_QueryVendors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("query all vendors failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).
			Return(nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).
			Times(1)

		_, err := NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{VendorIDs: []string{}})
		assert.Error(t, err)
	})
	t.Run("get vendor failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).Return([]*product.Vendor{
			{
				VendorID:   "AWS",
				VendorName: "aws",
			},
		}, nil).Times(1)
		productRW.EXPECT().GetVendor(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)
		_, err := NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{VendorIDs: []string{}})
		assert.Error(t, err)
	})
	t.Run("query all", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).Return([]*product.Vendor{
			{
				VendorID:   "Local",
				VendorName: "local",
			},
		}, nil).Times(1)
		mockQueryLocalFromDB(productRW.EXPECT())
		_, err := NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{VendorIDs: []string{}})
		assert.NoError(t, err)
	})
	t.Run("query exact", func(t *testing.T) {
		mockQueryLocalFromDB(productRW.EXPECT())
		_, err := NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{VendorIDs: []string{"Local"}})
		assert.NoError(t, err)
	})
}

func TestManager_UpdateVendors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	systemRW := mocksystem.NewMockReaderWriter(ctrl)
	models.SetSystemReaderWriter(systemRW)

	t.Run("update failed", func(t *testing.T) {
		productRW.EXPECT().SaveVendor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)
		_, err := NewManager().UpdateVendors(context.TODO(), message.UpdateVendorInfoReq{
			Vendors: vendors(),
		})
		assert.Error(t, err)
	})

	t.Run("flag failed", func(t *testing.T) {
		productRW.EXPECT().SaveVendor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		systemRW.EXPECT().VendorInitialized(gomock.Any()).Return(errors.Error(errors.TIEM_PARAMETER_INVALID)).Times(1)
		_, err := NewManager().UpdateVendors(context.TODO(), message.UpdateVendorInfoReq{
			Vendors: vendors(),
		})
		assert.Error(t, err)
	})

	t.Run("flag failed", func(t *testing.T) {
		productRW.EXPECT().SaveVendor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		systemRW.EXPECT().VendorInitialized(gomock.Any()).Return(nil).Times(1)
		_, err := NewManager().UpdateVendors(context.TODO(), message.UpdateVendorInfoReq{
			Vendors: vendors(),
		})
		assert.NoError(t, err)
	})
}

func TestManager_QueryAvailableVendors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("normal", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).Return([]*product.Vendor{
			{
				VendorID:   "Local",
				VendorName: "local",
			},
		}, nil).Times(1)
		mockQueryLocalFromDB(productRW.EXPECT())
		resp, err := NewManager().QueryAvailableVendors(context.TODO(), message.QueryAvailableVendorsReq{})
		assert.NoError(t, err)
		assert.Equal(t, "Region2", resp.Vendors["Local"].Regions["Region2"].ID)
	})
	t.Run("query all vendors failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).
			Return(nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).
			Times(1)

		_, err := NewManager().QueryAvailableVendors(context.TODO(), message.QueryAvailableVendorsReq{})
		assert.Error(t, err)
	})
	t.Run("get vendor failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllVendors(gomock.Any()).Return([]*product.Vendor{
			{
				VendorID:   "AWS",
				VendorName: "aws",
			},
		}, nil).Times(1)
		productRW.EXPECT().GetVendor(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)
		_, err := NewManager().QueryAvailableVendors(context.TODO(), message.QueryAvailableVendorsReq{})
		assert.Error(t, err)
	})

}

func TestManager_QueryProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("query all products failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).
			Return(nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).
			Times(1)

		_, err := NewManager().QueryProducts(context.TODO(), message.QueryProductsInfoReq{ProductIDs: []string{}})
		assert.Error(t, err)
	})
	t.Run("get product failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).Return([]*product.ProductInfo{
			{
				ProductID:   "TiDB",
				ProductName: "tidb",
			},
		}, nil).Times(1)
		productRW.EXPECT().GetProduct(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)
		_, err := NewManager().QueryProducts(context.TODO(), message.QueryProductsInfoReq{ProductIDs: []string{}})
		assert.Error(t, err)
	})
	t.Run("query all", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).Return([]*product.ProductInfo{
			{
				ProductID:   "TiDB",
				ProductName: "tidb",
			},
		}, nil).Times(1)
		mockQueryTiDBFromDB(productRW.EXPECT())
		_, err := NewManager().QueryProducts(context.TODO(), message.QueryProductsInfoReq{ProductIDs: []string{}})
		assert.NoError(t, err)
	})
	t.Run("query exact", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		_, err := NewManager().QueryProducts(context.TODO(), message.QueryProductsInfoReq{ProductIDs: []string{"TiDB"}})
		assert.NoError(t, err)
	})
}

func TestManager_UpdateProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	systemRW := mocksystem.NewMockReaderWriter(ctrl)
	models.SetSystemReaderWriter(systemRW)

	t.Run("update failed", func(t *testing.T) {
		productRW.EXPECT().SaveProduct(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)
		_, err := NewManager().UpdateProducts(context.TODO(), message.UpdateProductsInfoReq{
			Products: products(),
		})
		assert.Error(t, err)
	})

	t.Run("flag failed", func(t *testing.T) {
		productRW.EXPECT().SaveProduct(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		systemRW.EXPECT().ProductInitialized(gomock.Any()).Return(errors.Error(errors.TIEM_PARAMETER_INVALID)).Times(1)
		_, err := NewManager().UpdateProducts(context.TODO(), message.UpdateProductsInfoReq{
			Products: products(),
		})
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		productRW.EXPECT().SaveProduct(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		systemRW.EXPECT().ProductInitialized(gomock.Any()).Return(nil).Times(1)
		_, err := NewManager().UpdateProducts(context.TODO(), message.UpdateProductsInfoReq{
			Products: products(),
		})
		assert.NoError(t, err)
	})
}

func TestManager_QueryAvailableProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("QueryProducts failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).
			Return(nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).
			Times(1)

		_, err := NewManager().QueryAvailableProducts(context.TODO(), message.QueryAvailableProductsReq{
			VendorID: "Local",
		})
		assert.Error(t, err)
	})
	t.Run("QueryVendors failed", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).Return([]*product.ProductInfo{
			{
				ProductID:   "TiDB",
				ProductName: "tidb",
			},
		}, nil).Times(1)
		mockQueryTiDBFromDB(productRW.EXPECT())
		productRW.EXPECT().GetVendor(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)

		_, err := NewManager().QueryAvailableProducts(context.TODO(), message.QueryAvailableProductsReq{
			VendorID: "Local",
		})
		assert.Error(t, err)
	})
	t.Run("normal", func(t *testing.T) {
		productRW.EXPECT().QueryAllProducts(gomock.Any()).Return([]*product.ProductInfo{
			{
				ProductID:   "TiDB",
				ProductName: "tidb",
			},
		}, nil).Times(1)
		mockQueryTiDBFromDB(productRW.EXPECT())

		mockQueryLocalFromDB(productRW.EXPECT())
		_, err := NewManager().QueryAvailableProducts(context.TODO(), message.QueryAvailableProductsReq{
			VendorID: "Local",
		})
		assert.NoError(t, err)
	})

}

func TestManager_QueryProductDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	t.Run("QueryProducts failed", func(t *testing.T) {
		productRW.EXPECT().GetProduct(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)

		_, err := NewManager().QueryProductDetail(context.TODO(), message.QueryProductDetailReq{
			VendorID:  "Local",
			ProductID: "TiDB",
			RegionID:  "Region1",
		})
		assert.Error(t, err)
	})
	t.Run("QueryVendors failed", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		productRW.EXPECT().GetVendor(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT)).Times(1)

		_, err := NewManager().QueryProductDetail(context.TODO(), message.QueryProductDetailReq{
			VendorID:  "Local",
			ProductID: "TiDB",
			RegionID:  "Region1",
		})
		assert.Error(t, err)
	})

	t.Run("region error", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		mockQueryLocalFromDB(productRW.EXPECT())

		_, err := NewManager().QueryProductDetail(context.TODO(), message.QueryProductDetailReq{
			VendorID:  "Local",
			ProductID: "TiDB",
			RegionID:  "Region999",
		})
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		mockQueryLocalFromDB(productRW.EXPECT())

		_, err := NewManager().QueryProductDetail(context.TODO(), message.QueryProductDetailReq{
			VendorID:  "Local",
			ProductID: "TiDB",
			RegionID:  "Region1",
		})
		assert.NoError(t, err)
	})

}

func vendors() []structs.VendorConfigInfo {
	return []structs.VendorConfigInfo{
		{
			VendorInfo: structs.VendorInfo{
				ID:   "Local",
				Name: "local",
			},
			Regions: []structs.RegionConfigInfo{
				{
					RegionInfo: structs.RegionInfo{ID: "Region1", Name: "region1"},
					Zones: []structs.ZoneInfo{
						{ZoneID: "Zone1_1", ZoneName: "zone1_1"},
						{ZoneID: "Zone1_2", ZoneName: "zone1_2"},
					},
				},
				{
					RegionInfo: structs.RegionInfo{ID: "Region2", Name: "region2"},
					Zones: []structs.ZoneInfo{
						{ZoneID: "Zone2_1", ZoneName: "zone2_1"},
					},
				},
			},
			Specs: []structs.SpecInfo{
				{
					ID:          "c.large",
					Name:        "c.large",
					CPU:         4,
					Memory:      8,
					DiskType:    "SATA",
					PurposeType: "Compute",
				},
				{
					ID:          "s.large",
					Name:        "s.large",
					CPU:         4,
					Memory:      8,
					DiskType:    "SATA",
					PurposeType: "Storage",
				},
				{
					ID:          "s.xlarge",
					Name:        "s.xlarge",
					CPU:         8,
					Memory:      16,
					DiskType:    "SATA",
					PurposeType: "Storage",
				},
				{
					ID:          "sc.large",
					Name:        "sc.large",
					CPU:         4,
					Memory:      8,
					DiskType:    "SATA",
					PurposeType: "Schedule",
				},
			},
		},
	}
}

func mockQueryLocalFromDB(expect *mock_product.MockReaderWriterMockRecorder) {
	expect.GetVendor(gomock.Any(), gomock.Any()).Return(&product.Vendor{
		VendorID:   "Local",
		VendorName: "local",
	}, []*product.VendorZone{
		{
			VendorID:   "Local",
			RegionID:   "Region1",
			RegionName: "region1",
			ZoneID:     "Zone1_1",
			ZoneName:   "zone1_1",
			Comment:    "aa",
		},
		{
			VendorID:   "Local",
			RegionID:   "Region1",
			RegionName: "region1",
			ZoneID:     "Zone1_2",
			ZoneName:   "zone1_2",
			Comment:    "aa",
		},
		{
			VendorID:   "Local",
			RegionID:   "Region2",
			RegionName: "region2",
			ZoneID:     "Zone2_1",
			ZoneName:   "zone2_1",
			Comment:    "aa",
		},
	}, []*product.VendorSpec{
		{
			VendorID:    "Local",
			SpecID:      "c.large",
			SpecName:    "c.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Compute",
		},
		{
			VendorID:    "Local",
			SpecID:      "s.large",
			SpecName:    "s.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Storage",
		},
		{
			VendorID:    "Local",
			SpecID:      "s.xlarge",
			SpecName:    "s.xlarge",
			CPU:         8,
			Memory:      16,
			DiskType:    "SATA",
			PurposeType: "Storage",
		},
		{
			VendorID:    "Local",
			SpecID:      "sc.large",
			SpecName:    "sc.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Schedule",
		},
	}, nil).Times(1)
}

func products() []structs.ProductConfigInfo {
	return []structs.ProductConfigInfo{
		{
			ProductID:   "TiDB",
			ProductName: "whatever",
			Components: []structs.ProductComponentPropertyWithZones{
				{
					ID:                      "TiDB",
					Name:                    "TiDB",
					PurposeType:             "Compute",
					StartPort:               8,
					EndPort:                 16,
					MaxPort:                 4,
					MinInstance:             1,
					MaxInstance:             128,
					SuggestedInstancesCount: []int32{},
				},
				{
					ID:                      "TiKV",
					Name:                    "TiKV",
					PurposeType:             "Storage",
					StartPort:               1,
					EndPort:                 2,
					MaxPort:                 2,
					MinInstance:             1,
					MaxInstance:             128,
					SuggestedInstancesCount: []int32{},
				}, {
					ID:                      "PD",
					Name:                    "PD",
					PurposeType:             "Schedule",
					StartPort:               23,
					EndPort:                 413,
					MaxPort:                 22,
					MinInstance:             1,
					MaxInstance:             128,
					SuggestedInstancesCount: []int32{1, 3, 5, 7},
				}, {
					ID:                      "CDC",
					Name:                    "CDC",
					PurposeType:             "Compute",
					StartPort:               23,
					EndPort:                 43,
					MaxPort:                 2,
					MinInstance:             0,
					MaxInstance:             128,
					SuggestedInstancesCount: []int32{},
				},
			},
			Versions: []structs.SpecificVersionProduct{
				{
					ProductID: "TiDB",
					Arch:      "x86_64",
					Version:   "v5.2.2",
				}, {
					ProductID: "TiDB",
					Arch:      "ARM64",
					Version:   "v5.2.2",
				}, {
					ProductID: "TiDB",
					Arch:      "x86_64",
					Version:   "v5.3.0",
				},
			},
		},
	}
}

func mockQueryTiDBFromDB(expect *mock_product.MockReaderWriterMockRecorder) {
	expect.GetProduct(gomock.Any(), gomock.Any()).Return(&product.ProductInfo{
		ProductID:   "TiDB",
		ProductName: "tidb",
	}, []*product.ProductVersion{
		{
			ProductID: "TiDB",
			Arch:      "x86_64",
			Version:   "v5.2.2",
		}, {
			ProductID: "TiDB",
			Arch:      "ARM64",
			Version:   "v5.2.2",
		}, {
			ProductID: "TiDB",
			Arch:      "x86_64",
			Version:   "v5.3.0",
		},
	}, []*product.ProductComponentInfo{
		{
			ProductID:               "TiDB",
			ComponentID:             "TiDB",
			ComponentName:           "TiDB",
			PurposeType:             "Compute",
			StartPort:               8,
			EndPort:                 16,
			MaxPort:                 4,
			MinInstance:             1,
			MaxInstance:             128,
			SuggestedInstancesCount: []int32{},
		},
		{
			ProductID:               "TiDB",
			ComponentID:             "TiKV",
			ComponentName:           "TiKV",
			PurposeType:             "Storage",
			StartPort:               1,
			EndPort:                 2,
			MaxPort:                 2,
			MinInstance:             1,
			MaxInstance:             128,
			SuggestedInstancesCount: []int32{},
		}, {
			ProductID:               "TiDB",
			ComponentID:             "PD",
			ComponentName:           "PD",
			PurposeType:             "Schedule",
			StartPort:               23,
			EndPort:                 413,
			MaxPort:                 22,
			MinInstance:             1,
			MaxInstance:             128,
			SuggestedInstancesCount: []int32{1, 3, 5, 7},
		}, {
			ProductID:               "TiDB",
			ComponentID:             "CDC",
			ComponentName:           "CDC",
			PurposeType:             "Compute",
			StartPort:               23,
			EndPort:                 43,
			MaxPort:                 2,
			MinInstance:             0,
			MaxInstance:             128,
			SuggestedInstancesCount: []int32{},
		},
	}, nil).Times(1)
}
