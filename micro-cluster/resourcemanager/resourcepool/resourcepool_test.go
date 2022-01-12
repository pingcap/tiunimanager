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

package resourcepool

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_provider "github.com/pingcap-inc/tiem/test/mockresource/mockprovider"
	mock_workflow "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
)

func genHostInfo(hostName string) *structs.HostInfo {
	host := structs.HostInfo{
		IP:       "192.168.56.11",
		HostName: hostName,
		OS:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		AZ:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   string(constants.HostOnline),
		Stat:     string(constants.HostLoadLoadLess),
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sda",
		Path:     "/",
		Status:   string(constants.DiskAvailable),
		Capacity: 512,
	})
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	return &host
}

func Test_SelectImportFlowName(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()
	type want struct {
		flowName string
	}
	tests := []struct {
		testName string
		cond     structs.ImportCondition
		want     want
	}{
		{"ImportFlow", structs.ImportCondition{ReserveHost: false, SkipHostInit: false}, want{rp_consts.FlowImportHosts}},
		{"ImportFlowWithoutInit", structs.ImportCondition{ReserveHost: false, SkipHostInit: true}, want{rp_consts.FlowImportHostsWithoutInit}},
		{"TakeOverFlow", structs.ImportCondition{ReserveHost: true}, want{rp_consts.FlowTakeOverHosts}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			flowName := resourcePool.selectImportFlowName(&tt.cond)
			assert.Equal(t, tt.want.flowName, flowName)
		})
	}
}

func Test_SelectDeleteFlowName(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()
	type want struct {
		flowName string
	}
	tests := []struct {
		testName string
		force    bool
		want     want
	}{
		{"DeleteFlow", false, want{rp_consts.FlowDeleteHosts}},
		{"DeleteFlowByForce", true, want{rp_consts.FlowDeleteHostsByForce}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			flowName := resourcePool.selectDeleteFlowName(tt.force)
			assert.Equal(t, tt.want.flowName, flowName)
		})
	}
}

func Test_ImportHosts(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().ImportHosts(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []structs.HostInfo) ([]string, error) {
		return []string{"hostId1", "hostId2", "hostId3"}, nil
	})
	resourcePool.SetHostProvider(mockProvider)

	ctrl2 := gomock.NewController(t)
	defer ctrl2.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl2)
	workflow.MockWorkFlowService(workflowService)
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).Times(3)
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, flow *workflow.WorkFlowAggregation) error {
		return nil
	}).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(9)

	host1 := genHostInfo("Test_Host1")
	host2 := genHostInfo("Test_Host2")
	host3 := genHostInfo("Test_Host3")
	flowIds, hostIds, err := resourcePool.ImportHosts(context.TODO(), []structs.HostInfo{*host1, *host2, *host3}, &structs.ImportCondition{ReserveHost: false, SkipHostInit: false, IgnoreWarings: true})
	assert.Equal(t, 3, len(flowIds))
	assert.Equal(t, 3, len(hostIds))
	assert.Nil(t, err)
}

func Test_DeleteHosts(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	resourcePool.SetHostProvider(mockProvider)

	ctrl2 := gomock.NewController(t)
	defer ctrl2.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl2)
	workflow.MockWorkFlowService(workflowService)
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).Times(3)
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, flow *workflow.WorkFlowAggregation) error {
		return nil
	}).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(6)

	flowIds, err := resourcePool.DeleteHosts(context.TODO(), []string{"hostId1", "hostId2", "hostId3"}, false)
	assert.Equal(t, 3, len(flowIds))
	assert.Nil(t, err)
}

func Test_QueryHosts(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().QueryHosts(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) ([]structs.HostInfo, int64, error) {
		return []structs.HostInfo{{ID: "fake_hostId1"}}, 1, nil
	})
	resourcePool.SetHostProvider(mockProvider)

	hosts, total, err := resourcePool.QueryHosts(context.TODO(), &structs.HostFilter{}, &structs.PageRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 1, int(total))
	assert.Equal(t, 1, len(hosts))
}

func Test_UpdateHostStatus(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	resourcePool.SetHostProvider(mockProvider)

	err := resourcePool.UpdateHostStatus(context.TODO(), []string{"hostId1", "hostId2"}, string(constants.HostOffline))
	assert.Nil(t, err)
}

func Test_UpdateHostReserved(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().UpdateHostReserved(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	resourcePool.SetHostProvider(mockProvider)

	err := resourcePool.UpdateHostReserved(context.TODO(), []string{"hostId1", "hostId2"}, true)
	assert.Nil(t, err)
}
