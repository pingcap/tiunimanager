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

package meta

import (
	ctx "context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/platform/config"
	"github.com/pingcap-inc/tiem/models/platform/product"
	mock_deployment "github.com/pingcap-inc/tiem/test/mockdeployment"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestContain(t *testing.T) {
	ports := []int{4000, 4001, 4002}
	got := Contain(ports, 4000)
	assert.Equal(t, got, true)

	got = Contain(ports, 4003)
	assert.Equal(t, got, false)

	hosts := []string{"127.0.0.1", "127.0.0.2"}
	got = Contain(hosts, "127.0.0.1")
	assert.Equal(t, got, true)

	got = Contain(hosts, "127.0.0.3")
	assert.Equal(t, got, false)
}

func TestCompareTiDBVersion(t *testing.T) {
	got, err := CompareTiDBVersion("v5.0.0", "v5.2.2")
	assert.NoError(t, err)
	assert.Equal(t, got, false)

	got, err = CompareTiDBVersion("v5.2.2", "v5.0.0")
	assert.NoError(t, err)
	assert.Equal(t, got, true)

	got, err = CompareTiDBVersion("v5.0.2", "v5.0.0")
	assert.NoError(t, err)
	assert.Equal(t, got, true)

	got, err = CompareTiDBVersion("v5.0.0", "v5.0.0")
	assert.NoError(t, err)
	assert.Equal(t, got, true)

	got, err = CompareTiDBVersion("vx.0.0", "v5.2.2")
	assert.Error(t, err)

	got, err = CompareTiDBVersion("v5.x.0", "v5.2.2")
	assert.Error(t, err)

	got, err = CompareTiDBVersion("v5.0.x", "v5.2.2")
	assert.Error(t, err)

	got, err = CompareTiDBVersion("v5.0.0", "vx.2.2")
	assert.Error(t, err)

	got, err = CompareTiDBVersion("v5.0.0", "v5.x.2")
	assert.Error(t, err)

	got, err = CompareTiDBVersion("v5.0.0", "v5.2.x")
	assert.Error(t, err)
}

func TestScaleOutPreCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mock_deployment.NewMockInterface(ctrl)
	deployment.M = mockTiup

	t.Run("normal", func(t *testing.T) {
		mockTiup.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"true\"}", nil)
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Version: "v5.0.0",
			},
			Instances: map[string][]*management.ClusterInstance{
				"PD": {
					{
						Entity: common.Entity{
							ID:     "instance",
							Status: string(constants.ClusterInstanceRunning),
						},
						Type:   "PD",
						HostIP: []string{"127.0.0.1"},
						Ports:  []int32{4000},
					},
				},
			}}
		computes := []structs.ClusterResourceParameterCompute{
			{
				Type: "TiFlash",
			},
		}

		err := ScaleOutPreCheck(ctx.TODO(), meta, computes)
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"enable-placement-rules\": \"false\"}", nil)
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Version: "v5.0.0",
			},
			Instances: map[string][]*management.ClusterInstance{
				"PD": {
					{
						Entity: common.Entity{
							ID:     "instance",
							Status: string(constants.ClusterInstanceRunning),
						},
						Type:   "PD",
						HostIP: []string{"127.0.0.1"},
						Ports:  []int32{4000},
					},
				},
			}}
		computes := []structs.ClusterResourceParameterCompute{
			{
				Type: "TiFlash",
			},
		}

		err := ScaleOutPreCheck(ctx.TODO(), meta, computes)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {

		computes := make([]structs.ClusterResourceParameterCompute, 0)
		meta := &ClusterMeta{}
		err := ScaleOutPreCheck(ctx.TODO(), meta, computes)
		assert.Error(t, err)
	})
}

func TestScaleInPreCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)

	t.Run("parameter invalid", func(t *testing.T) {
		err := ScaleInPreCheck(ctx.TODO(), nil, nil)
		assert.Error(t, err)
	})

	t.Run("connect fail", func(t *testing.T) {
		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.0.0"}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDTiDB): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
			},
		}, DBUsers: map[string]*management.DBUser{
			string(constants.Root): {
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
		}}
		mockQueryTiDBFromDB(productRW.EXPECT())
		instance := &management.ClusterInstance{Type: string(constants.ComponentIDTiFlash)}
		err := ScaleInPreCheck(ctx.TODO(), meta, instance)
		assert.Error(t, err)
	})

	t.Run("component required error", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", CpuArchitecture: "x86_64"}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDTiDB): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
			},
		}, DBUsers: map[string]*management.DBUser{
			string(constants.Root): {
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
		}}
		instance := &management.ClusterInstance{Type: string(constants.ComponentIDPD)}
		err := ScaleInPreCheck(ctx.TODO(), meta, instance)
		assert.Error(t, err)
	})

	t.Run("check copies error", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())

		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.0.0", Copies: 3}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		}, DBUsers: map[string]*management.DBUser{
			string(constants.Root): {
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  "12345678",
				RoleType:  string(constants.Root),
			},
		}}
		instance := &management.ClusterInstance{Type: string(constants.ComponentIDTiKV)}
		err := ScaleInPreCheck(ctx.TODO(), meta, instance)
		assert.Error(t, err)
	})

}

func TestClonePreCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)
	rw.EXPECT().GetMasters(gomock.Any(), gomock.Any()).Return(make([]*management.ClusterRelation, 0), nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		sourceMeta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 1}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDCDC): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 1}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDCDC): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		err := ClonePreCheck(ctx.TODO(), sourceMeta, meta, string(constants.CDCSyncClone))
		assert.NoError(t, err)
	})

	t.Run("cdc fail", func(t *testing.T) {
		sourceMeta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 1}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDTiDB): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 1}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDTiDB): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		err := ClonePreCheck(ctx.TODO(), sourceMeta, meta, string(constants.CDCSyncClone))
		assert.Error(t, err)
	})

	t.Run("copies fail", func(t *testing.T) {
		sourceMeta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 3}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDCDC): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		meta := &ClusterMeta{Cluster: &management.Cluster{Type: "TiDB", Version: "v5.2.2", Copies: 3}, Instances: map[string][]*management.ClusterInstance{
			string(constants.ComponentIDCDC): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
			string(constants.ComponentIDTiKV): {
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8001},
				},
				{
					Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{8002},
				},
			},
		},
		}
		err := ClonePreCheck(ctx.TODO(), sourceMeta, meta, string(constants.CDCSyncClone))
		assert.Error(t, err)
	})
}

func TestWaitWorkflow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs.WorkFlowInfo{
					Status: constants.WorkFlowStatusFinished}}, nil).AnyTimes()
		err := WaitWorkflow(ctx.TODO(), "111", 1*time.Second, 2*time.Minute)
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs.WorkFlowInfo{
					Status: constants.WorkFlowStatusError}}, nil).AnyTimes()
		err := WaitWorkflow(ctx.TODO(), "111", 1*time.Second, 2*time.Minute)
		assert.Error(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().DetailWorkFlow(gomock.Any(), gomock.Any()).Return(
			message.QueryWorkFlowDetailResp{
				Info: &structs.WorkFlowInfo{
					Status: constants.WorkFlowStatusProcessing}}, nil).AnyTimes()
		err := WaitWorkflow(ctx.TODO(), "111", 1*time.Second, 2*time.Second)
		assert.Error(t, err)
	})
}

func Test_getRetainedPortRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rw := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(rw)

	t.Run("normal", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{
			ConfigValue: "[10,11]",
		}, nil).Times(1)
		portRange, err := getRetainedPortRange(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, []int{10,11}, portRange)
	})
	t.Run("error", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{}, errors.Error(errors.TIEM_SYSTEM_MISSING_CONFIG)).Times(1)
		_, err := getRetainedPortRange(context.TODO())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing system config")
	})
	t.Run("empty", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{
			ConfigValue: "",
		}, nil).Times(1)
		_, err := getRetainedPortRange(context.TODO())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing config config_retained_port_range")
	})
	t.Run("unmarshal error", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{
			ConfigValue: "ssss",
		}, nil).Times(1)
		_, err := getRetainedPortRange(context.TODO())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value for config config_retained_port_range")
	})
}

func TestGetRandomString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"normal", args{10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRandomString(tt.args.n)
			fmt.Println(got)
		})
	}
}
func mockQueryTiDBFromDBAnyTimes(expect *mock_product.MockReaderWriterMockRecorder) {
	expect.GetProduct(gomock.Any(), gomock.Any()).Return(&product.ProductInfo{
		ProductID: "TiDB",
		ProductName: "tidb",
	},[]*product.ProductVersion{
		{
			ProductID: "TiDB",
			Arch: "x86_64",
			Version: "v5.2.2",
		},{
			ProductID: "TiDB",
			Arch: "ARM64",
			Version: "v5.2.2",
		},{
			ProductID: "TiDB",
			Arch: "x86_64",
			Version: "v5.3.0",
		},
	},[]*product.ProductComponentInfo {
		{
			ProductID: "TiDB",
			ComponentID: "TiDB",
			ComponentName: "TiDB",
			PurposeType: "Compute",
			StartPort: 8,
			EndPort: 16,
			MaxPort: 4,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},
		{
			ProductID: "TiDB",
			ComponentID: "TiKV",
			ComponentName: "TiKV",
			PurposeType: "Storage",
			StartPort: 1,
			EndPort: 2,
			MaxPort: 2,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},{
			ProductID: "TiDB",
			ComponentID: "PD",
			ComponentName: "PD",
			PurposeType: "Schedule",
			StartPort: 23,
			EndPort: 413,
			MaxPort: 22,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{1,3,5,7},
		},{
			ProductID: "TiDB",
			ComponentID: "CDC",
			ComponentName: "CDC",
			PurposeType: "Compute",
			StartPort: 23,
			EndPort: 43,
			MaxPort: 2,
			MinInstance: 0,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},
	}, nil).AnyTimes()
}

func mockQueryTiDBFromDB(expect *mock_product.MockReaderWriterMockRecorder) {
	expect.GetProduct(gomock.Any(), gomock.Any()).Return(&product.ProductInfo{
		ProductID: "TiDB",
		ProductName: "tidb",
	},[]*product.ProductVersion{
		{
			ProductID: "TiDB",
			Arch: "x86_64",
			Version: "v5.2.2",
		},{
			ProductID: "TiDB",
			Arch: "ARM64",
			Version: "v5.2.2",
		},{
			ProductID: "TiDB",
			Arch: "x86_64",
			Version: "v5.3.0",
		},
	},[]*product.ProductComponentInfo {
		{
			ProductID: "TiDB",
			ComponentID: "TiDB",
			ComponentName: "TiDB",
			PurposeType: "Compute",
			StartPort: 8,
			EndPort: 16,
			MaxPort: 4,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},
		{
			ProductID: "TiDB",
			ComponentID: "TiKV",
			ComponentName: "TiKV",
			PurposeType: "Storage",
			StartPort: 1,
			EndPort: 2,
			MaxPort: 2,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},{
			ProductID: "TiDB",
			ComponentID: "PD",
			ComponentName: "PD",
			PurposeType: "Schedule",
			StartPort: 23,
			EndPort: 413,
			MaxPort: 22,
			MinInstance: 1,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{1,3,5,7},
		},{
			ProductID: "TiDB",
			ComponentID: "CDC",
			ComponentName: "CDC",
			PurposeType: "Compute",
			StartPort: 23,
			EndPort: 43,
			MaxPort: 2,
			MinInstance: 0,
			MaxInstance: 128,
			SuggestedInstancesCount: []int32{},
		},
	}, nil).Times(1)
}
