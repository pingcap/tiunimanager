/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: readwriteimpl_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProductReadWrite_Product(t *testing.T) {
	t.Run("delete empty", func(t *testing.T) {
		err := testRW.DeleteProduct(context.TODO(), "")
		assert.Error(t, err)
		err = testRW.DeleteProduct(context.TODO(), "111")
		assert.Error(t, err)
	})
	t.Run("get empty", func(t *testing.T) {
		_, _, _, err := testRW.GetProduct(context.TODO(), "")
		assert.Error(t, err)
		_, _, _, err = testRW.GetProduct(context.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("save empty", func(t *testing.T) {
		err := testRW.SaveProduct(context.TODO(), nil, []*ProductVersion{}, []*ProductComponentInfo{})
		assert.Error(t, err)
		err = testRW.SaveProduct(context.TODO(), &ProductInfo{}, []*ProductVersion{}, []*ProductComponentInfo{})
		assert.Error(t, err)
	})

	t.Run("versions conflict", func(t *testing.T) {
		defer testRW.DeleteProduct(context.TODO(), "TiDB")
		err := testRW.SaveProduct(context.TODO(), &ProductInfo{
			ProductID:   "TiDB",
			ProductName: "my-tidb",
		}, []*ProductVersion{
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "arm64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.4.0",
				Desc:      "default",
			},
		}, []*ProductComponentInfo{
			{
				ProductID:              "TiDB",
				ComponentID:            "TiKV",
				PurposeType:            "Storage",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[1,3,5,7]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "PD",
				PurposeType:            "Schedule",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "TiDB",
				PurposeType:            "Compute",
				StartPort:              2,
				EndPort:                5,
				MaxPort:                3,
				MinInstance:            3,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
		})
		assert.Error(t, err)
	})
	t.Run("components conflict", func(t *testing.T) {
		defer testRW.DeleteProduct(context.TODO(), "TiDB")
		err := testRW.SaveProduct(context.TODO(),
			&ProductInfo{
				ProductID:   "TiDB",
				ProductName: "my-tidb",
			},
			[]*ProductVersion{
				{
					ProductID: "TiDB",
					Arch:      "x86_64",
					Version:   "v5.2.2",
					Desc:      "default",
				},
				{
					ProductID: "TiDB",
					Arch:      "arm64",
					Version:   "v5.2.2",
					Desc:      "default",
				},
				{
					ProductID: "TiDB",
					Arch:      "x86_64",
					Version:   "v5.4.0",
					Desc:      "default",
				},
			},
			[]*ProductComponentInfo{
				{
					ProductID:              "TiDB",
					ComponentID:            "TiKV",
					PurposeType:            "Storage",
					StartPort:              12,
					EndPort:                17,
					MaxPort:                3,
					MinInstance:            9,
					MaxInstance:            18,
					SuggestedInstancesInfo: "[1,3,5,7]",
				},
				{
					ProductID:              "TiDB",
					ComponentID:            "TiKV",
					PurposeType:            "Storage",
					StartPort:              12,
					EndPort:                17,
					MaxPort:                3,
					MinInstance:            9,
					MaxInstance:            18,
					SuggestedInstancesInfo: "[1,3,5,7]",
				},
				{
					ProductID:              "TiDB",
					ComponentID:            "PD",
					PurposeType:            "Schedule",
					StartPort:              12,
					EndPort:                17,
					MaxPort:                3,
					MinInstance:            9,
					MaxInstance:            18,
					SuggestedInstancesInfo: "[]",
				},
				{
					ProductID:              "TiDB",
					ComponentID:            "TiDB",
					PurposeType:            "Compute",
					StartPort:              2,
					EndPort:                5,
					MaxPort:                3,
					MinInstance:            3,
					MaxInstance:            18,
					SuggestedInstancesInfo: "[]",
				},
			})
		assert.Error(t, err)
	})
	t.Run("save", func(t *testing.T) {
		defer testRW.DeleteProduct(context.TODO(), "TiDB")
		defer testRW.DeleteProduct(context.TODO(), "DM")

		all, err := testRW.QueryAllProducts(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(all))

		err = testRW.SaveProduct(context.TODO(), &ProductInfo{
			ProductID:   "DM",
			ProductName: "DM",
		}, []*ProductVersion{
			{
				ProductID: "DM",
				Arch:      "x86_64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "DM",
				Arch:      "arm64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "DM",
				Arch:      "x86_64",
				Version:   "v5.4.0",
				Desc:      "default",
			},
		}, []*ProductComponentInfo{
			{
				ProductID:              "DM",
				ComponentID:            "TiKV",
				PurposeType:            "Storage",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[1,3,5,7]",
			},
			{
				ProductID:              "DM",
				ComponentID:            "PD",
				PurposeType:            "Schedule",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
			{
				ProductID:              "DM",
				ComponentID:            "TiDB",
				PurposeType:            "Compute",
				StartPort:              2,
				EndPort:                5,
				MaxPort:                3,
				MinInstance:            3,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
		})
		assert.NoError(t, err)

		all, err = testRW.QueryAllProducts(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(all))

		product := &ProductInfo{
			ProductID:   "TiDB",
			ProductName: "my-tidb",
		}
		versions := []*ProductVersion{
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "arm64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.4.0",
				Desc:      "default",
			},
		}
		components := []*ProductComponentInfo{
			{
				ProductID:              "TiDB",
				ComponentID:            "TiKV",
				PurposeType:            "Storage",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[1,3,5,7]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "PD",
				PurposeType:            "Schedule",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "TiDB",
				PurposeType:            "Compute",
				StartPort:              2,
				EndPort:                5,
				MaxPort:                3,
				MinInstance:            3,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
		}
		err = testRW.SaveProduct(context.TODO(), product, versions, components)
		assert.NoError(t, err)

		all, err = testRW.QueryAllProducts(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(all))

		gotProduct, gotVersions, gotComponents, err := testRW.GetProduct(context.TODO(), "TiDB")
		assert.NoError(t, err)
		assert.Equal(t, *product, *gotProduct)
		assert.Equal(t, len(versions), len(gotVersions))
		assert.Equal(t, "TiDB", gotVersions[0].ProductID)
		assert.Equal(t, len(components), len(gotComponents))
		assert.Equal(t, "TiDB", gotComponents[1].ProductID)
	})
	t.Run("save again", func(t *testing.T) {
		defer testRW.DeleteProduct(context.TODO(), "TiDB")
		err := testRW.SaveProduct(context.TODO(), &ProductInfo{
			ProductID:   "TiDB",
			ProductName: "existed",
		}, []*ProductVersion{
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v9.9.9",
				Desc:      "default",
			},
		}, []*ProductComponentInfo{
			{
				ProductID:              "TiDB",
				ComponentID:            "TiKV",
				PurposeType:            "Storage",
				StartPort:              999,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[1,3,5,7]",
			},
		})
		assert.NoError(t, err)

		product := &ProductInfo{
			ProductID:   "TiDB",
			ProductName: "my-tidb",
		}
		versions := []*ProductVersion{
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "arm64",
				Version:   "v5.2.2",
				Desc:      "default",
			},
			{
				ProductID: "TiDB",
				Arch:      "x86_64",
				Version:   "v5.4.0",
				Desc:      "default",
			},
		}
		components := []*ProductComponentInfo{
			{
				ProductID:              "TiDB",
				ComponentID:            "TiKV",
				PurposeType:            "Storage",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[1,3,5,7]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "PD",
				PurposeType:            "Schedule",
				StartPort:              12,
				EndPort:                17,
				MaxPort:                3,
				MinInstance:            9,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
			{
				ProductID:              "TiDB",
				ComponentID:            "TiDB",
				PurposeType:            "Compute",
				StartPort:              2,
				EndPort:                5,
				MaxPort:                3,
				MinInstance:            3,
				MaxInstance:            18,
				SuggestedInstancesInfo: "[]",
			},
		}
		err = testRW.SaveProduct(context.TODO(), product, versions, components)
		assert.NoError(t, err)

		gotProduct, gotVersions, gotComponents, err := testRW.GetProduct(context.TODO(), "TiDB")
		assert.NoError(t, err)
		assert.Equal(t, *product, *gotProduct)
		assert.Equal(t, len(versions), len(gotVersions))
		assert.Equal(t, "TiDB", gotVersions[0].ProductID)
		assert.NotEqual(t, "v9.9.9", gotVersions[0].Version)

		assert.Equal(t, len(components), len(gotComponents))
		assert.NotEqual(t, int32(999), gotComponents[0].StartPort)
		assert.NotEqual(t, int32(999), gotComponents[1].StartPort)
		assert.NotEqual(t, int32(999), gotComponents[2].StartPort)
	})
}

func TestProductReadWrite_Vendor(t *testing.T) {
	t.Run("delete empty", func(t *testing.T) {
		err := testRW.DeleteVendor(context.TODO(), "")
		assert.Error(t, err)
		err = testRW.DeleteVendor(context.TODO(), "111")
		assert.Error(t, err)
	})
	t.Run("get empty", func(t *testing.T) {
		_, _, _, err := testRW.GetVendor(context.TODO(), "")
		assert.Error(t, err)
		_, _, _, err = testRW.GetVendor(context.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("save empty", func(t *testing.T) {
		err := testRW.SaveVendor(context.TODO(), nil, []*VendorZone{}, []*VendorSpec{})
		assert.Error(t, err)
		err = testRW.SaveVendor(context.TODO(), &Vendor{}, []*VendorZone{}, []*VendorSpec{})
		assert.Error(t, err)
	})

	t.Run("zone conflict", func(t *testing.T) {
		defer testRW.DeleteVendor(context.TODO(), "Local")
		err := testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "Local",
			VendorName: "datacenter",
		}, []*VendorZone{
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
		}, []*VendorSpec{})
		assert.Error(t, err)
	})
	t.Run("spec conflict", func(t *testing.T) {
		defer testRW.DeleteVendor(context.TODO(), "Local")
		err := testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "Local",
			VendorName: "datacenter",
		}, []*VendorZone{
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
		}, []*VendorSpec{
			{
				VendorID: "Local",
				SpecID:   "x.large",
			},
			{
				VendorID: "Local",
				SpecID:   "x.large",
			},
		})
		assert.Error(t, err)
	})
	t.Run("save", func(t *testing.T) {
		defer testRW.DeleteVendor(context.TODO(), "Local")
		defer testRW.DeleteVendor(context.TODO(), "AWS")

		all, err := testRW.QueryAllVendors(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(all))

		// save local
		err = testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "Local",
			VendorName: "datacenter",
		}, []*VendorZone{
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone2",
			},
			{
				VendorID: "Local",
				RegionID: "region2",
				ZoneID:   "zone1",
			},
		}, []*VendorSpec{
			{
				VendorID: "Local",
				SpecID:   "x.large",
			},
			{
				VendorID: "Local",
				SpecID:   "xxx.large",
			},
		})
		assert.NoError(t, err)

		all, err = testRW.QueryAllVendors(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(all))

		// save aws
		err = testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "AWS",
			VendorName: "AWS",
		}, []*VendorZone{
			{
				VendorID: "AWS",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
			{
				VendorID: "AWS",
				RegionID: "region1",
				ZoneID:   "zone2",
			},
			{
				VendorID: "AWS",
				RegionID: "region2",
				ZoneID:   "zone1",
			},
		}, []*VendorSpec{
			{
				VendorID: "AWS",
				SpecID:   "x.large",
			},
			{
				VendorID: "AWS",
				SpecID:   "xxx.large",
			},
		})

		all, err = testRW.QueryAllVendors(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(all))

		// get local
		vendor, zones, specs, err := testRW.GetVendor(context.TODO(), "Local")
		assert.NoError(t, err)
		assert.Equal(t, "datacenter", vendor.VendorName)
		assert.Equal(t, "Local", zones[0].VendorID)
		assert.Equal(t, "Local", zones[1].VendorID)
		assert.Equal(t, "Local", zones[2].VendorID)
		assert.Equal(t, "Local", specs[0].VendorID)
		assert.Equal(t, "Local", specs[1].VendorID)

		// get aws
		vendor, zones, specs, err = testRW.GetVendor(context.TODO(), "AWS")
		assert.NoError(t, err)
		assert.Equal(t, "AWS", vendor.VendorName)
		assert.Equal(t, "AWS", zones[0].VendorID)
		assert.Equal(t, "AWS", zones[1].VendorID)
		assert.Equal(t, "AWS", zones[2].VendorID)
		assert.Equal(t, "AWS", specs[0].VendorID)
		assert.Equal(t, "AWS", specs[1].VendorID)
	})
	t.Run("save again", func(t *testing.T) {
		defer testRW.DeleteVendor(context.TODO(), "Local")
		err := testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "Local",
			VendorName: "datacenter",
		}, []*VendorZone{
			{
				VendorID: "Local",
				RegionID: "aaa",
				ZoneID:   "aaa",
			},
			{
				VendorID: "Local",
				RegionID: "aaa",
				ZoneID:   "bbb",
			},
		}, []*VendorSpec{
			{
				VendorID: "Local",
				SpecID:   "aaa",
			},
			{
				VendorID: "Local",
				SpecID:   "aaa",
			},
		})

		// save local
		err = testRW.SaveVendor(context.TODO(), &Vendor{
			VendorID:   "Local",
			VendorName: "datacenter",
		}, []*VendorZone{
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone1",
			},
			{
				VendorID: "Local",
				RegionID: "region1",
				ZoneID:   "zone2",
			},
			{
				VendorID: "Local",
				RegionID: "region2",
				ZoneID:   "zone1",
			},
		}, []*VendorSpec{
			{
				VendorID: "Local",
				SpecID:   "x.large",
			},
			{
				VendorID: "Local",
				SpecID:   "xxx.large",
			},
		})
		assert.NoError(t, err)

		// get local
		vendor, zones, specs, err := testRW.GetVendor(context.TODO(), "Local")
		assert.NoError(t, err)
		assert.Equal(t, "datacenter", vendor.VendorName)
		assert.Equal(t, "Local", zones[0].VendorID)
		assert.Equal(t, "Local", zones[1].VendorID)
		assert.Equal(t, "Local", zones[2].VendorID)
		assert.Equal(t, "Local", specs[0].VendorID)
		assert.Equal(t, "Local", specs[1].VendorID)

		assert.NotEqual(t, "aaa", zones[0].ZoneID)
		assert.NotEqual(t, "bbb", zones[0].ZoneID)
		assert.NotEqual(t, "aaa", zones[0].RegionID)

		assert.NotEqual(t, "aaa", specs[0].SpecID)
		assert.NotEqual(t, "bbb", specs[0].SpecID)

		assert.NotEqual(t, "aaa", specs[1].SpecID)
	})
}
