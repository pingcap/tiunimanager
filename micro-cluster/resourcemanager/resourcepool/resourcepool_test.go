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
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/config"
	mock_config "github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	mock_provider "github.com/pingcap-inc/tiem/test/mockresource/mockprovider"
	mock_workflow "github.com/pingcap-inc/tiem/test/mockworkflow"
	workflow "github.com/pingcap-inc/tiem/workflow2"
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
		{"TakeOverFlowWithoutInit", structs.ImportCondition{ReserveHost: true, SkipHostInit: true}, want{rp_consts.FlowImportHostsWithoutInit}},
		{"TakeOverFlow", structs.ImportCondition{ReserveHost: true, SkipHostInit: false}, want{rp_consts.FlowTakeOverHosts}},
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
		host     *structs.HostInfo
		force    bool
		want     want
	}{
		{"DeleteFlow", &structs.HostInfo{IP: "192.168.6.6", Status: string(constants.HostOnline)}, false, want{rp_consts.FlowDeleteHosts}},
		{"DeleteFailHostFlow", &structs.HostInfo{IP: "192.168.6.6", Status: string(constants.HostFailed)}, false, want{rp_consts.FlowDeleteHostsByForce}},
		{"DeleteFlowByForce", &structs.HostInfo{IP: "192.168.6.6", Status: string(constants.HostOnline)}, true, want{rp_consts.FlowDeleteHostsByForce}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			flowName := resourcePool.selectDeleteFlowName(tt.host, tt.force)
			assert.Equal(t, tt.want.flowName, flowName)
		})
	}
}

func Test_ImportHosts(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	ctrl0 := gomock.NewController(t)
	defer ctrl0.Finish()
	rw := mock_config.NewMockReaderWriter(ctrl0)
	rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "9527"}, nil)
	models.SetConfigReaderWriter(rw)

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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).Times(3)
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, flowId string) error {
		return nil
	}).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(9)

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
	mockProvider.EXPECT().QueryHosts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]structs.HostInfo{{ID: "hostId", Status: string(constants.HostOnline), Stat: string(constants.HostLoadLoadLess)}},
		int64(1),
		nil,
	).Times(3)
	resourcePool.SetHostProvider(mockProvider)

	ctrl2 := gomock.NewController(t)
	defer ctrl2.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl2)
	workflow.MockWorkFlowService(workflowService)
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).Times(3)
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, flowId string) error {
		return nil
	}).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(6)

	flowIds, err := resourcePool.DeleteHosts(context.TODO(), []string{"hostId1", "hostId2", "hostId3"}, false)
	assert.Equal(t, 3, len(flowIds))
	assert.Nil(t, err)
}

func Test_DeleteHosts_InvalidHostId(t *testing.T) {
	hostId := "test-fake-host-id"

	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().QueryHosts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil,
		int64(0),
		nil,
	)
	resourcePool.SetHostProvider(mockProvider)

	flowIds, err := resourcePool.DeleteHosts(context.TODO(), []string{hostId}, false)
	assert.Nil(t, flowIds)
	assert.NotNil(t, err)
	emError, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_DELETE_HOST_ERROR, emError.GetCode())
}

func Test_QueryHosts(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().QueryHosts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, location *structs.Location, filter *structs.HostFilter, page *structs.PageRequest) ([]structs.HostInfo, int64, error) {
		return []structs.HostInfo{{ID: "fake_hostId1"}}, 1, nil
	})
	resourcePool.SetHostProvider(mockProvider)

	hosts, total, err := resourcePool.QueryHosts(context.TODO(), &structs.Location{}, &structs.HostFilter{}, &structs.PageRequest{})
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

func Test_getSSHConfigPort_Succeed(t *testing.T) {
	models.MockDB()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mock_config.NewMockReaderWriter(ctrl)
	rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "9527"}, nil)
	models.SetConfigReaderWriter(rw)

	resourcePool := GetResourcePool()

	port := resourcePool.getSSHConfigPort(context.TODO())
	assert.Equal(t, 9527, port)
}

func Test_getSSHConfigPort_InvalidString(t *testing.T) {
	models.MockDB()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mock_config.NewMockReaderWriter(ctrl)
	rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "fault"}, nil)
	models.SetConfigReaderWriter(rw)

	resourcePool := GetResourcePool()

	port := resourcePool.getSSHConfigPort(context.TODO())
	assert.Equal(t, rp_consts.HostSSHPort, port)
}

func Test_getSSHConfigPort_NullString(t *testing.T) {
	models.MockDB()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mock_config.NewMockReaderWriter(ctrl)
	rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: ""}, nil)
	models.SetConfigReaderWriter(rw)

	resourcePool := GetResourcePool()

	port := resourcePool.getSSHConfigPort(context.TODO())
	assert.Equal(t, 22, port)
}

func Test_UpdateHostInfo(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().UpdateHostInfo(gomock.Any(), gomock.Any()).Return(nil)
	resourcePool.SetHostProvider(mockProvider)

	err := resourcePool.UpdateHostInfo(context.TODO(), structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_CUDDisk(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host provider
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl1)
	mockProvider.EXPECT().CreateDisks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	mockProvider.EXPECT().UpdateDisk(gomock.Any(), gomock.Any()).Return(nil)
	mockProvider.EXPECT().DeleteDisks(gomock.Any(), gomock.Any()).Return(nil)
	resourcePool.SetHostProvider(mockProvider)

	_, err := resourcePool.CreateDisks(context.TODO(), "fake-host-id", nil)
	assert.Nil(t, err)
	err = resourcePool.UpdateDisk(context.TODO(), structs.DiskInfo{})
	assert.Nil(t, err)
	err = resourcePool.DeleteDisks(context.TODO(), []string{"fake-disk-id"})
	assert.Nil(t, err)
}