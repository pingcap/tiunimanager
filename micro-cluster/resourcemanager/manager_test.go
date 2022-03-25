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
 *                                                                            *
 ******************************************************************************/

package resourcemanager

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	allocrecycle "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/allocator_recycler"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	host_provider "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	config "github.com/pingcap-inc/tiem/models/platform/config"
	resource_models "github.com/pingcap-inc/tiem/models/resource"
	resourcepool "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_cluster "github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_config "github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	mock_resource "github.com/pingcap-inc/tiem/test/mockmodels/mockresource"
	mock_initiator "github.com/pingcap-inc/tiem/test/mockresource/mockinitiator"
	mock_workflow "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
)

var emptyNode = func(task *wfModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func getEmptyFlow(name string) *workflow.WorkFlowDefine {
	return &workflow.WorkFlowDefine{
		FlowName: name,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start": {Name: "start", SuccessEvent: "done", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: emptyNode},
			"done":  {Name: "end", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: emptyNode},
			"fail":  {Name: "end", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: emptyNode},
		},
	}
}

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

func genHostRspFromDB(hostId, hostName string) *resourcepool.Host {
	host := resourcepool.Host{
		ID:       hostId,
		HostName: hostName,
		IP:       "192.168.56.11",
		OS:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		AZ:       "TEST_REGION,TEST_AZ",
		Rack:     "TEST_REGION,TEST_AZ,TEST_RACK",
		Status:   string(constants.HostOnline),
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, resourcepool.Disk{
		Name:     "sda",
		Path:     "/",
		Status:   string(constants.DiskAvailable),
		Capacity: 512,
	})
	host.Disks = append(host.Disks, resourcepool.Disk{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	return &host
}

func doMockInitiator(mockInitiator *mock_initiator.MockHostInitiator) {
	mockInitiator.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, h *structs.HostInfo) error {
		return nil
	}).AnyTimes()
	mockInitiator.EXPECT().InstallSoftware(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []structs.HostInfo) error {
		return nil
	}).AnyTimes()
}
func Test_ImportHosts_Succeed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	// Mock models readerwriter
	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()
	mockModels := mock_resource.NewMockReaderWriter(ctrl1)
	mockModels.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []resourcepool.Host) ([]string, error) {
		if hosts[0].HostName == "TEST_HOST1" {
			var hostIds []string
			hostIds = append(hostIds, fake_hostId)
			return hostIds, nil
		} else {
			return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	mockConfigModels := mock_config.NewMockReaderWriter(ctrl1)
	mockConfigModels.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigValue: "9527"}, nil)
	models.MockDB()
	models.SetConfigReaderWriter(mockConfigModels)

	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockModels)

	// Mock host initiator
	ctrl2 := gomock.NewController(t)
	defer ctrl2.Finish()
	mockInitiator := mock_initiator.NewMockHostInitiator(ctrl2)
	doMockInitiator(mockInitiator)

	resourceManager.GetResourcePool().SetHostInitiator(mockInitiator)

	ctrl3 := gomock.NewController(t)
	defer ctrl3.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl3)
	workflow.MockWorkFlowService(workflowService)
	//defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{}, 0)},
	}, nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(3)

	var hosts []structs.HostInfo
	host := genHostInfo("TEST_HOST1")
	hosts = append(hosts, *host)

	flowIds, hostIds, err := resourceManager.ImportHosts(context.TODO(), hosts, &structs.ImportCondition{})
	assert.Nil(t, err)

	assert.Equal(t, fake_hostId, hostIds[0])
	assert.Equal(t, "flow01", flowIds[0])
	// UnsetInEmTiup in case the background import routine is not finished
	framework.UnsetInEmTiupProcess()
}

func Test_ImportHosts_Failed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []resourcepool.Host) ([]string, error) {
		if hosts[0].HostName == "TEST_HOST1" {
			var hostIds []string
			hostIds = append(hostIds, fake_hostId)
			return hostIds, nil
		} else {
			return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})

	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	// Mock host initiator
	ctrl2 := gomock.NewController(t)
	defer ctrl2.Finish()
	mockInitiator := mock_initiator.NewMockHostInitiator(ctrl2)
	doMockInitiator(mockInitiator)

	resourceManager.GetResourcePool().SetHostInitiator(mockInitiator)

	ctrl3 := gomock.NewController(t)
	defer ctrl3.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl3)
	workflow.MockWorkFlowService(workflowService)
	//defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{}, 0)},
	}, nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	var hosts []structs.HostInfo
	host := genHostInfo("TEST_HOST2")
	hosts = append(hosts, *host)
	flowIds, _, err := resourceManager.ImportHosts(context.TODO(), hosts, &structs.ImportCondition{})
	assert.NotNil(t, err)
	tiemErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_PARAMETER_INVALID, tiemErr.GetCode())
	assert.Nil(t, flowIds)

}

func Test_QueryHosts_Succeed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_hostname := "fake_host_name"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	models.MockDB()
	clusterRW := mock_cluster.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().QueryHostInstances(gomock.Any(), gomock.Any()).Return(nil, nil)

	mockClient.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, location *structs.Location, filter *structs.HostFilter, offset, limit int) (hosts []resourcepool.Host, total int64, err error) {
		assert.Equal(t, 20, offset)
		assert.Equal(t, 10, limit)
		if filter.HostID == fake_hostId {
			dbhost := genHostRspFromDB(fake_hostId, fake_hostname)
			hosts = append(hosts, *dbhost)
			return hosts, 1, nil
		} else {
			return nil, 0, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	filter := &structs.HostFilter{
		HostID: fake_hostId,
	}
	page := &structs.PageRequest{
		Page:     3,
		PageSize: 10,
	}

	hosts, total, err := resourceManager.QueryHosts(context.TODO(), &structs.Location{}, filter, page)
	assert.Nil(t, err)
	assert.Equal(t, 1, int(total))

	assert.Equal(t, fake_hostname, hosts[0].HostName)
	assert.Equal(t, "TEST_REGION", hosts[0].Region)
	assert.Equal(t, "TEST_AZ", hosts[0].AZ)
	assert.Equal(t, "TEST_RACK", hosts[0].Rack)
}

func Test_DeleteHosts_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockClient.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]resourcepool.Host{{IP: "192.168.168.999", Status: string(constants.HostOnline), Stat: string(constants.HostLoadLoadLess)}},
		int64(1),
		nil,
	).Times(2)
	models.MockDB()
	clusterRW := mock_cluster.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().QueryHostInstances(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)

	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	ctrl3 := gomock.NewController(t)
	defer ctrl3.Finish()
	workflowService := mock_workflow.NewMockWorkFlowService(ctrl3)
	workflow.MockWorkFlowService(workflowService)
	//defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{}, 0)},
	}, nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(4)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	flowIds, err := resourceManager.DeleteHosts(context.TODO(), hostIds, false)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(flowIds))
}

func Test_UpdateHostReserved_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host1 := genHostInfo("TEST_HOST1")
	host2 := genHostInfo("TEST_HOST2")
	mockClient.EXPECT().UpdateHostReserved(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string, reserved bool) error {
		if hostIds[0] == fake_hostId1 && hostIds[1] == fake_hostId2 {
			host1.Reserved = reserved
			host2.Reserved = reserved
			return nil
		} else {
			return errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	err := resourceManager.UpdateHostReserved(context.TODO(), hostIds, true)
	assert.Nil(t, err)
	assert.Equal(t, true, host1.Reserved)
	assert.Equal(t, true, host2.Reserved)
}

func Test_UpdateHostStatus_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host1 := genHostInfo("TEST_HOST1")
	host2 := genHostInfo("TEST_HOST2")
	mockClient.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string, status string) error {
		if hostIds[0] == fake_hostId1 && hostIds[1] == fake_hostId2 {
			host1.Status = status
			host2.Status = status
			return nil
		} else {
			return errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	status := "Offline"
	err := resourceManager.UpdateHostStatus(context.TODO(), hostIds, status)
	assert.Nil(t, err)
	assert.Equal(t, status, host1.Status)
	assert.Equal(t, status, host2.Status)
}

func Test_GetHierarchy_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetHostItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter *structs.HostFilter, level, depth int32) (items []resource_models.HostItem, err error) {
		if filter.Arch == string(constants.ArchX8664) && level == 1 && depth == 3 {
			item1 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone1",
				Rack:   "TEST_Region1,TEST_Zone1,TEST_Rack1",
				Ip:     "192.168.9.111",
				Name:   "HostName1",
			}
			item2 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone1",
				Rack:   "TEST_Region1,TEST_Zone1,TEST_Rack2",
				Ip:     "192.168.9.112",
				Name:   "HostName2",
			}
			item3 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone2",
				Rack:   "TEST_Region1,TEST_Zone2,TEST_Rack1",
				Ip:     "192.168.9.113",
				Name:   "HostName3",
			}
			items = append(items, item1)
			items = append(items, item2)
			items = append(items, item3)
			return items, nil
		} else {
			return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	filter := structs.HostFilter{
		Arch: string(constants.ArchX8664),
	}

	root, err := resourceManager.GetHierarchy(context.TODO(), &filter, 1, 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(root.SubNodes))
	assert.Equal(t, "TEST_Region1", root.SubNodes[0].Name)
	assert.Equal(t, 2, len(root.SubNodes[0].SubNodes[0].SubNodes))
	assert.Equal(t, 1, len(root.SubNodes[0].SubNodes[1].SubNodes))
}

func Test_GetStocks_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error) {
		if location.Region == "TEST_Region1" {
			stocks1 := structs.Stocks{
				Zone:             "TEST_Region1,TEST_Zone1",
				FreeCpuCores:     2,
				FreeMemory:       4,
				FreeDiskCount:    2,
				FreeDiskCapacity: 256,
			}
			stocks = append(stocks, stocks1)
			stocks2 := structs.Stocks{
				Zone:             "TEST_Region1,TEST_Zone1",
				FreeCpuCores:     1,
				FreeMemory:       1,
				FreeDiskCount:    1,
				FreeDiskCapacity: 256,
			}
			stocks = append(stocks, stocks2)
			return stocks, nil
		} else {
			return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	location := structs.Location{Region: "TEST_Region1"}

	stocks, err := resourceManager.GetStocks(context.TODO(), &location, &structs.HostFilter{}, &structs.DiskFilter{})
	assert.Nil(t, err)
	assert.Equal(t, int32(2), stocks["TEST_Region1,TEST_Zone1"].FreeHostCount)
	assert.Equal(t, int32(3), stocks["TEST_Region1,TEST_Zone1"].FreeCpuCores)
	assert.Equal(t, int32(5), stocks["TEST_Region1,TEST_Zone1"].FreeMemory)
	assert.Equal(t, int32(3), stocks["TEST_Region1,TEST_Zone1"].FreeDiskCount)
	assert.Equal(t, int32(512), stocks["TEST_Region1,TEST_Zone1"].FreeDiskCapacity)
}

func Test_AllocResources_Succeed(t *testing.T) {

	fake_host_id1 := "TEST_host_id1"
	fake_host_ip1 := "199.199.199.1"
	fake_holder_id := "TEST_holder1"
	fake_request_id := "TEST_reqeust1"
	fake_disk_id1 := "TEST_disk_id1"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().AllocResources(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, batchReq *resource_structs.BatchAllocRequest) (*resource_structs.BatchAllocResponse, error) {
		rsp := new(resource_structs.BatchAllocResponse)
		if batchReq.BatchRequests[0].Applicant.HolderId == fake_holder_id && batchReq.BatchRequests[1].Applicant.RequestId == fake_request_id &&
			batchReq.BatchRequests[0].Requires[0].Count == 1 && batchReq.BatchRequests[1].Requires[1].Require.PortReq[1].PortCnt == 2 {
		} else {
			return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
		var r resource_structs.Compute
		r.Reqseq = 0
		r.HostId = fake_host_id1
		r.HostIp = fake_host_ip1
		r.ComputeRes.CpuCores = batchReq.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.ComputeRes.Memory = batchReq.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.Location.Region = batchReq.BatchRequests[0].Requires[0].Location.Region
		r.Location.Zone = batchReq.BatchRequests[0].Requires[0].Location.Zone
		r.DiskRes.DiskId = fake_disk_id1
		r.DiskRes.Type = batchReq.BatchRequests[0].Requires[0].Require.DiskReq.DiskType
		r.DiskRes.Capacity = batchReq.BatchRequests[0].Requires[0].Require.DiskReq.Capacity
		for _, portRes := range batchReq.BatchRequests[0].Requires[0].Require.PortReq {
			var portResource resource_structs.PortResource
			portResource.Start = portRes.Start
			portResource.End = portRes.End
			portResource.Ports = append(portResource.Ports, portRes.Start+1)
			portResource.Ports = append(portResource.Ports, portRes.Start+2)
			r.PortRes = append(r.PortRes, portResource)
		}

		var one_rsp resource_structs.AllocRsp
		one_rsp.Results = append(one_rsp.Results, r)

		rsp.BatchResults = append(rsp.BatchResults, &one_rsp)

		var two_rsp resource_structs.AllocRsp
		two_rsp.Results = append(two_rsp.Results, r)
		two_rsp.Results = append(two_rsp.Results, r)
		rsp.BatchResults = append(rsp.BatchResults, &two_rsp)
		return rsp, nil
	})
	allocRecycle := resourceManager.GetManagement().GetAllocatorRecycler()
	localHostManage, ok := (allocRecycle).(*(allocrecycle.LocalHostManagement))
	assert.True(t, ok)
	localHostManage.SetResourceReaderWriter(mockClient)

	var batchReq resource_structs.BatchAllocRequest

	var require resource_structs.AllocRequirement
	require.Location.Region = "TesT_Region1"
	require.Location.Zone = "TEST_Zone1"
	require.HostFilter.Arch = string(constants.ArchX8664)
	require.HostFilter.DiskType = string(constants.SSD)
	require.HostFilter.Purpose = string(constants.PurposeSchedule)
	require.Require.ComputeReq.CpuCores = 4
	require.Require.ComputeReq.Memory = 8
	require.Require.DiskReq.NeedDisk = true
	require.Require.DiskReq.DiskType = string(constants.SATA)
	require.Require.DiskReq.Capacity = 256
	require.Require.PortReq = append(require.Require.PortReq, resource_structs.PortRequirement{
		Start:   10000,
		End:     10010,
		PortCnt: 2,
	})
	require.Require.PortReq = append(require.Require.PortReq, resource_structs.PortRequirement{
		Start:   10010,
		End:     10020,
		PortCnt: 2,
	})
	require.Count = 1

	var req1 resource_structs.AllocReq
	req1.Applicant.HolderId = fake_holder_id
	req1.Applicant.RequestId = fake_request_id
	req1.Requires = append(req1.Requires, require)

	var req2 resource_structs.AllocReq
	req2.Applicant.HolderId = fake_holder_id
	req2.Applicant.RequestId = fake_request_id
	req2.Requires = append(req2.Requires, require)
	req2.Requires = append(req2.Requires, require)

	batchReq.BatchRequests = append(batchReq.BatchRequests, req1)
	batchReq.BatchRequests = append(batchReq.BatchRequests, req2)

	resp, err := resourceManager.AllocResources(context.TODO(), &batchReq)
	if err != nil {
		t.Errorf("alloc resource failed, err: %v\n", err)
	}

	assert.Equal(t, 2, len(resp.BatchResults))
	assert.Equal(t, 1, len(resp.BatchResults[0].Results))
	assert.Equal(t, 2, len(resp.BatchResults[1].Results))
	assert.True(t, resp.BatchResults[0].Results[0].DiskRes.DiskId == fake_disk_id1 && resp.BatchResults[0].Results[0].DiskRes.Capacity == 256 && resp.BatchResults[1].Results[0].DiskRes.Type == string(constants.SATA))
	assert.True(t, resp.BatchResults[1].Results[1].PortRes[0].Ports[0] == 10001 && resp.BatchResults[1].Results[1].PortRes[0].Ports[1] == 10002)
	assert.True(t, resp.BatchResults[1].Results[1].PortRes[1].Ports[0] == 10011 && resp.BatchResults[1].Results[1].PortRes[1].Ports[1] == 10012)
}

func Test_RecycleResources_Succeed(t *testing.T) {
	fake_cluster_id := "TEST_Fake_CLUSTER_ID"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, request *resource_structs.RecycleRequest) error {
		if request.RecycleReqs[0].RecycleType == 2 && request.RecycleReqs[0].HolderID == fake_cluster_id {

		} else {
			return errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
		return nil
	})
	allocRecycle := resourceManager.GetManagement().GetAllocatorRecycler()
	localHostManage, ok := (allocRecycle).(*(allocrecycle.LocalHostManagement))
	assert.True(t, ok)
	localHostManage.SetResourceReaderWriter(mockClient)

	var req resource_structs.RecycleRequest
	var require resource_structs.RecycleRequire
	require.HolderID = fake_cluster_id
	require.RecycleType = 2

	req.RecycleReqs = append(req.RecycleReqs, require)
	err := resourceManager.RecycleResources(context.TODO(), &req)
	assert.Nil(t, err)
}

func Test_UpdateHostInfo(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host1 := genHostInfo("TEST_HOST1")
	mockClient.EXPECT().UpdateHostInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, host resourcepool.Host) error {
		if host.ID == fake_hostId1 {
			return nil
		} else {
			return errors.NewError(errors.TIEM_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	err := resourceManager.UpdateHostInfo(context.TODO(), *host1)
	assert.NotNil(t, err)
	assert.Equal(t, errors.NewError(errors.TIEM_PARAMETER_INVALID, "update host failed without host id").GetMsg(), err.(errors.EMError).GetMsg())

	host1.ID = fake_hostId1
	err = resourceManager.UpdateHostInfo(context.TODO(), *host1)
	assert.Nil(t, err)
}

func Test_CUD_Disk(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)

	mockClient.EXPECT().CreateDisks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	mockClient.EXPECT().DeleteDisks(gomock.Any(), gomock.Any()).Return(nil)
	mockClient.EXPECT().UpdateDisk(gomock.Any(), gomock.Any()).Return(nil)
	hostprovider := resourceManager.GetResourcePool().GetHostProvider()
	file_hostprovider, ok := (hostprovider).(*(host_provider.FileHostProvider))
	assert.True(t, ok)
	file_hostprovider.SetResourceReaderWriter(mockClient)

	var disk structs.DiskInfo
	_, err := resourceManager.CreateDisks(context.TODO(), fake_hostId1, []structs.DiskInfo{disk})
	assert.NotNil(t, err)
	assert.Equal(t, errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
		fake_hostId1, disk.Name, disk.Path, disk.Capacity).GetMsg(), err.(errors.EMError).GetMsg())

	_, err = resourceManager.CreateDisks(context.TODO(), fake_hostId1, nil)
	assert.Nil(t, err)

	err = resourceManager.DeleteDisks(context.TODO(), nil)
	assert.Nil(t, err)

	err = resourceManager.UpdateDisk(context.TODO(), disk)
	assert.NotNil(t, err)
	assert.Equal(t, errors.NewError(errors.TIEM_PARAMETER_INVALID, "update disk failed without disk id").GetMsg(), err.(errors.EMError).GetMsg())
	disk.ID = "fake-disk-id"
	err = resourceManager.UpdateDisk(context.TODO(), disk)
	assert.Nil(t, err)
}
