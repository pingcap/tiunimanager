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

package handler

import (
	ctx "context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/deployment"
	hostInspector "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/inspect"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	mock_check "github.com/pingcap-inc/tiem/test/mockcheck"
	mock_deployment "github.com/pingcap-inc/tiem/test/mockdeployment"
	mock_hosts_inspect "github.com/pingcap-inc/tiem/test/mockhostsinspect"
	mock_account "github.com/pingcap-inc/tiem/test/mockmodels/mockaccount"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockresource"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestReport_ParseFrom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().GetReport(gomock.Any(), "111").Return(structs.CheckReportInfo{}, nil)

		report := &Report{}
		err := report.ParseFrom(ctx.TODO(), "111")
		assert.NoError(t, err)
		assert.NotNil(t, report.Info)
	})

	t.Run("error", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().GetReport(gomock.Any(), "111").Return(structs.CheckReportInfo{}, errors.New("parse from error"))

		report := &Report{}
		err := report.ParseFrom(ctx.TODO(), "111")
		assert.Error(t, err)
		assert.Nil(t, report.Info)
	})
}

func TestReport_Serialize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		report := &Report{Info: &structs.CheckReportInfo{
			Tenants: map[string]structs.TenantCheck{
				"tenant01": {
					CPURatio: 0.8,
				},
			},
			Hosts: structs.HostsCheck{
				Hosts: map[string]structs.HostCheck{
					"host01": {
						StorageRatio: 0.8,
					},
				},
			},
		}}

		got, err := report.Serialize(ctx.TODO())
		assert.NoError(t, err)
		assert.NotEmpty(t, got)
	})

	t.Run("error", func(t *testing.T) {
		report := &Report{Info: &structs.CheckReportInfo{
			Tenants: map[string]structs.TenantCheck{
				"tenant01": {
					CPURatio: float32(math.Inf(1)),
				},
			},
		}}

		_, err := report.Serialize(ctx.TODO())
		assert.Error(t, err)
	})
}

func TestReport_CheckHosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("normal", func(t *testing.T) {
		resourceRW := mockresource.NewMockReaderWriter(ctrl)
		models.SetResourceReaderWriter(resourceRW)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return([]resourcepool.Host{{ID: "01"}, {ID: "02"}}, int64(2), nil)

		inspectHosts := mock_hosts_inspect.NewMockHostInspector(ctrl)
		hostInspector.MockHostInspector(inspectHosts)
		inspectHosts.EXPECT().CheckCpuAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         false,
				RealValue:     12,
				ExpectedValue: 13,
			},
			"02": {
				Valid:         true,
				RealValue:     10,
				ExpectedValue: 10,
			},
		}, nil)
		inspectHosts.EXPECT().CheckMemAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         true,
				RealValue:     100,
				ExpectedValue: 100,
			},
			"02": {
				Valid:         true,
				RealValue:     120,
				ExpectedValue: 120,
			},
		}, nil)
		inspectHosts.EXPECT().CheckDiskAllocated(gomock.Any(), gomock.Any()).Return(map[string]map[string]*structs.CheckString{
			"01": {
				"sda": {
					Valid:         true,
					RealValue:     "/sda",
					ExpectedValue: "/sda",
				},
			},
			"02": {
				"sdb": {
					Valid:         true,
					RealValue:     "/sdb",
					ExpectedValue: "/sdb",
				},
			},
		}, nil)
		report := &Report{}
		err := report.CheckHosts(ctx.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(report.Info.Hosts.Hosts))
		assert.Equal(t, false, report.Info.Hosts.Hosts["01"].CPUAllocated.Valid)
	})

	t.Run("query hosts error", func(t *testing.T) {
		resourceRW := mockresource.NewMockReaderWriter(ctrl)
		models.SetResourceReaderWriter(resourceRW)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return([]resourcepool.Host{{ID: "01"}, {ID: "02"}}, int64(2), errors.New("query hosts error"))
		report := &Report{}
		err := report.CheckHosts(ctx.TODO())
		assert.Error(t, err)
	})

	t.Run("inspect cpu allocated error", func(t *testing.T) {
		resourceRW := mockresource.NewMockReaderWriter(ctrl)
		models.SetResourceReaderWriter(resourceRW)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return([]resourcepool.Host{{ID: "01"}, {ID: "02"}}, int64(2), nil)

		inspectHosts := mock_hosts_inspect.NewMockHostInspector(ctrl)
		hostInspector.MockHostInspector(inspectHosts)
		inspectHosts.EXPECT().CheckCpuAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         false,
				RealValue:     12,
				ExpectedValue: 13,
			},
			"02": {
				Valid:         true,
				RealValue:     10,
				ExpectedValue: 10,
			},
		}, errors.New("inspect cpu allocated error"))
		report := &Report{}
		err := report.CheckHosts(ctx.TODO())
		assert.Error(t, err)
	})

	t.Run("inspect memory allocated error", func(t *testing.T) {
		resourceRW := mockresource.NewMockReaderWriter(ctrl)
		models.SetResourceReaderWriter(resourceRW)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return([]resourcepool.Host{{ID: "01"}, {ID: "02"}}, int64(2), nil)

		inspectHosts := mock_hosts_inspect.NewMockHostInspector(ctrl)
		hostInspector.MockHostInspector(inspectHosts)
		inspectHosts.EXPECT().CheckCpuAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         false,
				RealValue:     12,
				ExpectedValue: 13,
			},
			"02": {
				Valid:         true,
				RealValue:     10,
				ExpectedValue: 10,
			},
		}, nil)
		inspectHosts.EXPECT().CheckMemAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         true,
				RealValue:     100,
				ExpectedValue: 100,
			},
			"02": {
				Valid:         true,
				RealValue:     120,
				ExpectedValue: 120,
			},
		}, errors.New("inspect memory allocated error"))
		report := &Report{}
		err := report.CheckHosts(ctx.TODO())
		assert.Error(t, err)
	})

	t.Run("inspect disk allocated error", func(t *testing.T) {
		resourceRW := mockresource.NewMockReaderWriter(ctrl)
		models.SetResourceReaderWriter(resourceRW)

		resourceRW.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return([]resourcepool.Host{{ID: "01"}, {ID: "02"}}, int64(2), nil)

		inspectHosts := mock_hosts_inspect.NewMockHostInspector(ctrl)
		hostInspector.MockHostInspector(inspectHosts)
		inspectHosts.EXPECT().CheckCpuAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         false,
				RealValue:     12,
				ExpectedValue: 13,
			},
			"02": {
				Valid:         true,
				RealValue:     10,
				ExpectedValue: 10,
			},
		}, nil)
		inspectHosts.EXPECT().CheckMemAllocated(gomock.Any(), gomock.Any()).Return(map[string]*structs.CheckInt32{
			"01": {
				Valid:         true,
				RealValue:     100,
				ExpectedValue: 100,
			},
			"02": {
				Valid:         true,
				RealValue:     120,
				ExpectedValue: 120,
			},
		}, nil)
		inspectHosts.EXPECT().CheckDiskAllocated(gomock.Any(), gomock.Any()).Return(map[string]map[string]*structs.CheckString{
			"01": {
				"sda": {
					Valid:         true,
					RealValue:     "/sda",
					ExpectedValue: "/sda",
				},
			},
			"02": {
				"sdb": {
					Valid:         true,
					RealValue:     "/sdb",
					ExpectedValue: "/sdb",
				},
			},
		}, errors.New("inspect disk allocated error"))
		report := &Report{}
		err := report.CheckHosts(ctx.TODO())
		assert.Error(t, err)
	})
}

func TestReport_CheckTenants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("normal", func(t *testing.T) {
		accountRW := mock_account.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().QueryTenants(gomock.Any()).Return(make(map[string]structs.TenantInfo), nil)
		report := &Report{}
		err := report.CheckTenants(ctx.TODO())
		assert.NoError(t, err)
	})

	t.Run("query tenants error", func(t *testing.T) {
		accountRW := mock_account.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().QueryTenants(gomock.Any()).Return(make(map[string]structs.TenantInfo), errors.New("query tenants error"))
		report := &Report{}
		err := report.CheckTenants(ctx.TODO())
		assert.Error(t, err)
	})

	t.Run("check tenant error", func(t *testing.T) {
		accountRW := mock_account.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().QueryTenants(gomock.Any()).Return(map[string]structs.TenantInfo{"tenant": {ID: "id"}}, nil)

		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().QueryClusters(gomock.Any(), gomock.Any()).Return(nil, errors.New("check tenant error"))
		report := &Report{}
		err := report.CheckTenants(ctx.TODO())
		assert.Error(t, err)
	})
}

func TestReport_CheckTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().QueryClusters(gomock.Any(), gomock.Any()).Return(make([]*management.Result, 0), nil)

		accountRW := mock_account.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{
			ID: "tenant01", MaxCluster: 2, MaxCPU: 3, MaxMemory: 100, MaxStorage: 500}, nil)
		report := &Report{}
		err := report.CheckTenant(ctx.TODO(), "tenant01")
		assert.NoError(t, err)
		assert.Equal(t, report.Info.Tenants["tenant01"].ClusterCount.Valid, true)
	})

	t.Run("query clusters error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().QueryClusters(gomock.Any(), gomock.Any()).Return(make([]*management.Result, 0),
			errors.New("query clusters error"))
		report := &Report{}
		err := report.CheckTenant(ctx.TODO(), "tenant01")
		assert.Error(t, err)
	})

	t.Run("get tenant error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().QueryClusters(gomock.Any(), gomock.Any()).Return(make([]*management.Result, 0), nil)

		accountRW := mock_account.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{
			ID: "tenant01", MaxCluster: 2, MaxCPU: 3, MaxMemory: 100, MaxStorage: 500}, errors.New("get tenant error"))
		report := &Report{}
		err := report.CheckTenant(ctx.TODO(), "tenant01")
		assert.Error(t, err)
	})
}

func TestReport_GetClusterAllocatedResource(t *testing.T) {
	meta := &management.Result{
		Instances: []*management.ClusterInstance{
			{
				CpuCores:     8,
				Memory:       16,
				DiskCapacity: 120,
			},
			{
				CpuCores:     8,
				Memory:       32,
				DiskCapacity: 200,
			},
		},
	}

	report := &Report{}
	cpu, memory, storage := report.GetClusterAllocatedResource(ctx.TODO(), meta)
	assert.Equal(t, cpu, int32(16))
	assert.Equal(t, memory, int32(48))
	assert.Equal(t, storage, int32(320))
}

func TestReport_GetClusterCopies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), nil)
		mockTiupManager := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiupManager
		mockTiupManager.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"max-replicas\": 3}", nil)
		report := &Report{}
		got, err := report.GetClusterCopies(ctx.TODO(), "111")
		assert.NoError(t, err)
		assert.Equal(t, int32(3), got)
	})

	t.Run("get cluster meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), errors.New("get cluster meta error"))
		report := &Report{}
		_, err := report.GetClusterCopies(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("pd address is empty", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{},
			[]*management.ClusterInstance{}, make([]*management.DBUser, 0), nil)
		report := &Report{}
		_, err := report.GetClusterCopies(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("tiup error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), nil)
		mockTiupManager := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiupManager
		mockTiupManager.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"max-replicas\": 3}", errors.New("tiup error"))
		report := &Report{}
		_, err := report.GetClusterCopies(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), nil)
		mockTiupManager := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mockTiupManager
		mockTiupManager.EXPECT().Ctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any()).Return("{\"max-replicas\": 3", nil)
		report := &Report{}
		_, err := report.GetClusterCopies(ctx.TODO(), "111")
		assert.Error(t, err)
	})
}

func TestReport_GetClusterAccountStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("create sql link error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{},
			[]*management.ClusterInstance{}, make([]*management.DBUser, 0), nil)

		report := &Report{}
		got, err := report.GetClusterAccountStatus(ctx.TODO(), "111")
		assert.NoError(t, err)
		assert.Equal(t, false, got.Health)
	})

	t.Run("get cluster meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{},
			[]*management.ClusterInstance{}, make([]*management.DBUser, 0), errors.New("get cluster meta error"))
		report := &Report{}
		_, err := report.GetClusterAccountStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})
}

func TestReport_GetClusterRegionStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("get region status error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), nil)

		report := &Report{}
		_, err := report.GetClusterRegionStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("get cluster meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), errors.New("get cluster meta error"))
		report := &Report{}
		_, err := report.GetClusterRegionStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("pd address is empty", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{},
			[]*management.ClusterInstance{}, make([]*management.DBUser, 0), nil)
		report := &Report{}
		_, err := report.GetClusterRegionStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})
}

func TestReport_GetClusterHealthStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	models.MockDB()

	t.Run("get cluster meta error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), errors.New("get cluster meta error"))
		report := &Report{}
		_, err := report.GetClusterHealthStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("pd address is empty", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{},
			[]*management.ClusterInstance{}, make([]*management.DBUser, 0), nil)
		report := &Report{}
		_, err := report.GetClusterHealthStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})

	t.Run("get health status error", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "instance",
					Status: string(constants.ClusterInstanceRunning),
				},
				Type:   "PD",
				HostIP: []string{"127.0.0.3"},
				Ports:  []int32{8001},
			},
		}, make([]*management.DBUser, 0), nil)

		report := &Report{}
		_, err := report.GetClusterHealthStatus(ctx.TODO(), "111")
		assert.Error(t, err)
	})
}

