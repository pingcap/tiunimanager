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

package management

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_validateCreating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productRW := mock_product.NewMockProductReadWriterInterface(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("product error", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.Error(errors.TIEM_PARAMETER_INVALID)).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq{})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})

	t.Run("unsupported product", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"DM": {
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			 CreateClusterParameter: structs.CreateClusterParameter{
			 	Type: "TiDB",
			 },
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_UNSUPPORT_PRODUCT, err.(errors.EMError).GetCode())
	})

	t.Run("unsupported version", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.0.0": {},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_UNSUPPORT_PRODUCT, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "version v5.2.2 is not supported")
	})
	t.Run("unsupported arch", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"amd64": {

							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_UNSUPPORT_PRODUCT, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "arch x86_64 is not supported")
	})
	t.Run("max instance", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"x86_64": {
								{
									ID: "TiDB",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiKV",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "PD",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiFlash",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
			},
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 10},
				},
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_INVALID_TOPOLOGY, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "should be less than")
	})
	t.Run("min instance", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"x86_64": {
								{
									ID: "TiDB",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiKV",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "PD",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{1,3,5,7},
								},
								{
									ID: "TiFlash",
									MinInstance: 0,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
			},
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4},
				},
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_INVALID_TOPOLOGY, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "should be more than")
	})
	t.Run("TiKV and copies", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"x86_64": {
								{
									ID: "TiDB",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiKV",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "PD",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{1,3,5,7},
								},
								{
									ID: "TiFlash",
									MinInstance: 0,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
				Copies: 5,
			},
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4},
					{Type: "TiKV", Count: 4},
					{Type: "PD", Count: 4},
				},
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_INVALID_TOPOLOGY, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "is less than copies ")
	})
	t.Run("suggested count", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"x86_64": {
								{
									ID: "TiDB",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiKV",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "PD",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{1,3,5,7},
								},
								{
									ID: "TiFlash",
									MinInstance: 0,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
				Copies: 3,

			},
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4},
					{Type: "TiKV", Count: 4},
					{Type: "PD", Count: 4},
				},
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_INVALID_TOPOLOGY, err.(errors.EMError).GetCode())
		assert.Contains(t, err.Error(), "total number of PD should be in ")
	})
	t.Run("OK", func(t *testing.T) {
		productRW.EXPECT().QueryProductDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]structs.ProductDetail{
			"TiDB": {
				Versions: map[string]structs.ProductVersion{
					"v5.2.2": {
						Version: "v5.2.2",
						Arch: map[string][]structs.ProductComponentProperty {
							"x86_64": {
								{
									ID: "TiDB",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "TiKV",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
								{
									ID: "PD",
									MinInstance: 1,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{1,3,5,7},
								},
								{
									ID: "TiFlash",
									MinInstance: 0,
									MaxInstance: 8,
									SuggestedInstancesCount: []int32{},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		err := validateCreating(context.TODO(), &cluster.CreateClusterReq {
			CreateClusterParameter: structs.CreateClusterParameter{
				Type: "TiDB",
				Version: "v5.2.2",
				CpuArchitecture: "x86_64",
				Copies: 5,
			},
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 4},
					{Type: "TiKV", Count: 5},
					{Type: "PD", Count: 5},
				},
			},
		})
		assert.NoError(t, err)
	})
}
