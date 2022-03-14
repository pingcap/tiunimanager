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

package management

import (
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/product"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			models.MockDB()
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)
			models.MockDB()
			return models.Open(d)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func mockQueryTiDBFromDBAnyTimes(expect *mock_product.MockReaderWriterMockRecorder) {
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
	}, nil).AnyTimes()
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
