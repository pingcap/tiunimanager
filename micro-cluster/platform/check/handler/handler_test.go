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
	"github.com/pingcap-inc/tiem/common/structs"
	hostInspector "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/inspect"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	mock_check "github.com/pingcap-inc/tiem/test/mockcheck"
	mock_hosts_inspect "github.com/pingcap-inc/tiem/test/mockhostsinspect"
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


