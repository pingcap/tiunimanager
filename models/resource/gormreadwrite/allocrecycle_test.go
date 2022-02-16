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

package gormreadwrite

import (
	"context"
	"fmt"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models/resource/management"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/stretchr/testify/assert"
)

func createTestHost(region, zone, rack, hostName, ip, clusterType, purpose, diskType string, freeCpuCores, freeMemory, availableDiskCount int32) (id []string, err error) {
	h := resourcepool.Host{
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       "admin2",
		Status:       string(constants.HostOnline),
		Stat:         string(constants.HostLoadLoadLess),
		Arch:         string(constants.ArchX8664),
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     freeCpuCores,
		Memory:       freeMemory,
		FreeCpuCores: freeCpuCores,
		FreeMemory:   freeMemory,
		Nic:          "1GE",
		Region:       region,
		AZ:           zone,
		Rack:         rack,
		ClusterType:  clusterType,
		Purpose:      purpose,
		DiskType:     diskType,
		Disks: []resourcepool.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: string(constants.DiskReserved), Type: diskType},
		},
	}
	h.BuildDefaultTraits()
	for i := 0; i < int(availableDiskCount); i++ {
		deviceName := fmt.Sprintf("sd%c", 'b'+i)
		path := fmt.Sprintf("/mnt%d", i+1)
		h.Disks = append(h.Disks, resourcepool.Disk{
			Name:     deviceName,
			Path:     path,
			Capacity: 256,
			Status:   string(constants.DiskAvailable),
			Type:     diskType,
		})
	}
	return GormRW.Create(context.TODO(), []resourcepool.Host{h})
}

func recycleClusterResources(holderId string) error {
	var request resource_structs.RecycleRequest
	var recycleRequire resource_structs.RecycleRequire
	recycleRequire.HolderID = holderId

	recycleRequire.RecycleType = resource_structs.RecycleHolder
	request.RecycleReqs = append(request.RecycleReqs, recycleRequire)
	return GormRW.RecycleResources(context.TODO(), &request)
}

func recycleRequestResources(requestId string) error {
	var request resource_structs.RecycleRequest
	var recycleRequire resource_structs.RecycleRequire
	recycleRequire.RequestID = requestId

	recycleRequire.RecycleType = resource_structs.RecycleOperate
	request.RecycleReqs = append(request.RecycleReqs, recycleRequire)
	return GormRW.RecycleResources(context.TODO(), &request)
}

func Test_Create_Delete_Host_Succeed(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
}

func Test_Create_Dup_Host(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	assert.NotNil(t, err)
	assert.Nil(t, id2)
}

func Test_Create_Query_Host_Succeed(t *testing.T) {
	hostIp1 := "192.168.999.998"
	hostName1 := "Test_Host1"
	hostIp2 := "192.168.999.999"
	hostName2 := "Test_Host2"
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", hostName1, hostIp1,
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", hostName2, hostIp2,
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	hosts, total, err := GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{HostID: id1[0]}, 0, 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hosts))
	assert.Equal(t, 1, int(total))
	assert.Equal(t, id1[0], hosts[0].ID)
	assert.Equal(t, hostName1, hosts[0].HostName)
	assert.Equal(t, hostIp1, hosts[0].IP)
	assert.Equal(t, 3, len(hosts[0].Disks))
	hosts, total, err = GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{Purpose: string(constants.PurposeCompute)}, 0, 3)
	assert.Nil(t, err)
	assert.Equal(t, 2, int(total))
	assert.Equal(t, 2, len(hosts))
}

func Test_UpdateHostReserved_Succeed(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, total, err := GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, 1, int(total))
	assert.False(t, hosts[0].Reserved)
	GormRW.UpdateHostReserved(context.TODO(), id1, true)
	hosts, total, err = GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{Stat: string(constants.HostLoadLoadLess)}, 0, 3)
	assert.Equal(t, 1, int(total))
	assert.True(t, err == nil && len(hosts) == 1)
	assert.True(t, hosts[0].Reserved)
}

func Test_UpdateHostStatus_Succeed(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, total, err := GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, 1, int(total))
	assert.Equal(t, string(constants.HostOnline), hosts[0].Status)
	GormRW.UpdateHostStatus(context.TODO(), id1, string(constants.HostOffline))
	hosts, total, err = GormRW.Query(context.TODO(), &structs.Location{}, &structs.HostFilter{Purpose: string(constants.PurposeCompute)}, 0, 3)
	assert.Equal(t, 1, int(total))
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, string(constants.HostOffline), hosts[0].Status)
}

func Test_GetHostItems(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	items, err := GormRW.GetHostItems(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664)}, 1, 3)
	assert.True(t, err == nil && len(items) == 3)
	assert.Equal(t, "192.168.192.168", items[0].Ip)
	assert.Equal(t, "192.168.192.169", items[1].Ip)
	assert.Equal(t, "192.168.192.170", items[2].Ip)
}

func Test_GetHostStocks(t *testing.T) {
	id1, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := createTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 32, 64, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	stocks, err := GormRW.GetHostStocks(context.TODO(), &structs.Location{Region: "Test_Region1", Zone: "Test_Zone2"}, &structs.HostFilter{}, &structs.DiskFilter{})
	t.Log(stocks)
	assert.True(t, err == nil)
	assert.Equal(t, 1, len(stocks))
	assert.Equal(t, "Test_Region1,Test_Zone2", stocks[0].Zone)
	assert.Equal(t, int32(32), stocks[0].FreeCpuCores)
	assert.Equal(t, int32(64), stocks[0].FreeMemory)
	assert.Equal(t, int32(3), stocks[0].FreeDiskCount)
	assert.Equal(t, int32(3*256), stocks[0].FreeDiskCapacity)
}

func TestAllocResources_3RequestsInBatch_3Hosts(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.117",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.118",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.119",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc := structs.Location{}
	loc.Region = "Test_Region1"
	loc.Zone = "Test_Zone1"
	filter1 := resource_structs.Filter{}
	filter1.DiskType = string(constants.SSD)
	filter1.Purpose = string(constants.PurposeCompute)
	filter1.Arch = string(constants.ArchX8664)
	require := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   resource_structs.RandomRack,
		Require:    *require,
		Count:      3,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rsp.BatchResults))
	assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
	assert.Equal(t, int32(0), rsp.BatchResults[0].Results[2].Reqseq)
	assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[0].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[0].HostIp == "474.111.111.119")
	assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[1].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[1].HostIp == "474.111.111.119")
	assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[2].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[2].HostIp == "474.111.111.119")
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.117")
	assert.Equal(t, int32(17-4-4-4), host.FreeCpuCores)
	assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))
	//stat, isExaust := host.IsExhaust()
	//assert.True(t, stat == resource.HOST_DISK_EXHAUST && isExaust == true)
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "474.111.111.118")
	assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
	assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
	assert.True(t, host2.Stat == string(constants.HostLoadInUsed))
	var host3 resourcepool.Host
	MetaDB.First(&host3, "IP = ?", "474.111.111.119")
	assert.Equal(t, int32(15-4-4-4), host3.FreeCpuCores)
	assert.Equal(t, int32(64-8-8-8), host3.FreeMemory)
	assert.True(t, host3.Stat == string(constants.HostLoadInUsed))
	//var usedPorts []int32
	var usedPorts []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
	assert.Equal(t, 15, len(usedPorts))
	assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
	assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
	for i := 0; i < 15; i++ {
		assert.Equal(t, int32(10000+i), usedPorts[i].Port)
	}
	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}

func TestAllocResources_1Requirement_3Hosts_Filted_by_Label(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host1", "474.111.111.158",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute)+","+string(constants.PurposeStorage), string(constants.SSD), 17, 64, 1)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host2", "474.111.111.159",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 2)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host3", "474.111.111.160",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 1)

	loc := structs.Location{}
	loc.Region = "Test_Region1"
	loc.Zone = "Test_Zone3"
	filter := resource_structs.Filter{}
	filter.Arch = string(constants.ArchX8664)
	//filter.DiskType = string(resource.SSD)
	//filter.Purpose = string(resource.Compute)
	filter.HostTraits = 136
	require := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10010, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster59"
	test_req.Applicant.RequestId = "TestRequestID59"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc,
		HostFilter: filter,
		Strategy:   resource_structs.RandomRack,
		Require:    *require,
		Count:      3,
	})
	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	res := rsp.BatchResults[0]
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res.Results))
	assert.Equal(t, int32(0), res.Results[0].Reqseq)
	assert.Equal(t, int32(0), res.Results[1].Reqseq)
	assert.Equal(t, int32(0), res.Results[2].Reqseq)
	assert.Equal(t, int32(10003), res.Results[0].PortRes[0].Ports[3])
	assert.True(t, res.Results[0].HostIp != res.Results[1].HostIp)
	assert.True(t, res.Results[0].HostIp != res.Results[2].HostIp)
	assert.True(t, res.Results[2].HostIp != res.Results[1].HostIp)
	assert.True(t, res.Results[0].HostIp == "474.111.111.158" || res.Results[0].HostIp == "474.111.111.159" || res.Results[0].HostIp == "474.111.111.160")
	assert.True(t, res.Results[1].HostIp == "474.111.111.158" || res.Results[1].HostIp == "474.111.111.159" || res.Results[1].HostIp == "474.111.111.160")
	assert.True(t, res.Results[2].HostIp == "474.111.111.158" || res.Results[2].HostIp == "474.111.111.159" || res.Results[2].HostIp == "474.111.111.160")
	assert.Equal(t, int32(4), res.Results[0].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), res.Results[0].ComputeRes.Memory)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.158")
	assert.Equal(t, int32(17-4), host.FreeCpuCores)
	assert.Equal(t, int32(64-8), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "474.111.111.159")
	assert.Equal(t, int32(16-4), host2.FreeCpuCores)
	assert.Equal(t, int32(64-8), host2.FreeMemory)
	assert.True(t, host2.Stat == string(constants.HostLoadInUsed))
	var host3 resourcepool.Host
	MetaDB.First(&host3, "IP = ?", "474.111.111.160")
	assert.Equal(t, int32(15-4), host3.FreeCpuCores)
	assert.Equal(t, int32(64-8), host3.FreeMemory)
	assert.True(t, host3.Stat == string(constants.HostLoadInUsed))

	err = recycleClusterResources("TestCluster59")
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}

func newRequirementForRequest(cpuCores, memory int32, needDisk bool, diskcap int32, disktype string, portStart, portEnd, portCount int32) *resource_structs.Requirement {
	require := resource_structs.Requirement{}
	require.ComputeReq.CpuCores = cpuCores
	require.ComputeReq.Memory = memory
	require.DiskReq.NeedDisk = needDisk
	if require.DiskReq.NeedDisk {
		require.DiskReq.Capacity = diskcap
		require.DiskReq.DiskType = disktype
	}
	require.PortReq = append(require.PortReq, resource_structs.PortRequirement{
		Start:   portStart,
		End:     portEnd,
		PortCnt: portCount,
	})
	return &require
}

func newLocation(regionName, zoneName, rackName, hostIp string) *structs.Location {
	var location structs.Location
	location.Region = regionName
	location.Zone = zoneName
	location.Rack = rackName
	location.HostIp = hostIp

	return &location
}

func TestAllocResources_3RequestsInBatch_SpecifyHost_Strategy(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.127",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.128",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.129",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc1 := structs.Location{}
	loc1.Region = "Test_Region1"
	loc1.Zone = "Test_Zone2"
	loc1.HostIp = "474.111.111.127"

	require1 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc1,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require1,
		Count:    1,
	})

	loc2 := structs.Location{}
	loc2.Region = "Test_Region1"
	loc2.Zone = "Test_Zone2"
	loc2.HostIp = "474.111.111.128"

	require2 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc2,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require2,
		Count:    1,
	})

	loc3 := structs.Location{}
	loc3.Region = "Test_Region1"
	loc3.Zone = "Test_Zone2"
	loc3.HostIp = "474.111.111.129"

	require3 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc3,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require3,
		Count:    1,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rsp.BatchResults))
	assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
	assert.Equal(t, int32(2), rsp.BatchResults[0].Results[2].Reqseq)
	assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.111.111.127")
	assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.111.111.128")
	assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.111.111.129")
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
	assert.Equal(t, "admin2", rsp.BatchResults[0].Results[0].Passwd)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.127")
	assert.Equal(t, int32(17-4), host.FreeCpuCores)
	assert.Equal(t, int32(64-8), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "474.111.111.128")
	assert.Equal(t, int32(16-4), host2.FreeCpuCores)
	assert.Equal(t, int32(64-8), host2.FreeMemory)
	assert.True(t, host2.Stat == string(constants.HostLoadInUsed))
	var host3 resourcepool.Host
	MetaDB.First(&host3, "IP = ?", "474.111.111.129")
	assert.Equal(t, int32(15-4), host3.FreeCpuCores)
	assert.Equal(t, int32(64-8), host3.FreeMemory)
	assert.True(t, host3.Stat == string(constants.HostLoadInUsed))
	//var usedPorts []int32
	var usedPorts []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
	assert.Equal(t, 5, len(usedPorts))
	assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
	assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
	for i := 0; i < 5; i++ {
		assert.Equal(t, int32(10000+i), usedPorts[i].Port)
	}

	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}

func TestAllocResources_SpecifyHost_Strategy_TakeOver(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone4", "Test_Region1,Test_Zon4,Test_Rack1", "Test_Host1", "474.111.111.147",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)

	err := GormRW.UpdateHostReserved(context.TODO(), id1, true)
	assert.Equal(t, nil, err)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.147")
	assert.Equal(t, true, host.Reserved)

	loc1 := structs.Location{}
	loc1.Region = "Test_Region1"
	loc1.Zone = "Test_Zone4"
	loc1.HostIp = "474.111.111.147"

	require1 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc1,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require1,
		Count:    1,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.True(t, nil == rsp && err != nil)
	t.Log(err)
	te, ok := err.(errors.EMError)
	assert.Equal(t, true, ok)
	assert.True(t, errors.TIEM_RESOURCE_ALLOCATE_ERROR.Equal(int32(te.GetCode())))

	batchReq.BatchRequests[0].Applicant.TakeoverOperation = true
	rsp, err2 := GormRW.AllocResources(context.TODO(), &batchReq)
	t.Log(err2)
	assert.Equal(t, nil, err2)
	assert.True(t, rsp.BatchResults[0].Results[0].HostId == id1[0])

	err = recycleRequestResources("TestRequestID1")
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
}

func Test_AllocResources_ClusterPorts_Strategy(t *testing.T) {
	var ids []string
	id1, _ := createTestHost("Test_Region29", "Test_Region29,Zone5", "Test_Region29,Zone5,Rack1", "HostName1", "429.111.111.137", string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region29", "Test_Region29,Zone5", "Test_Region29,Zone5,Rack2", "HostName2", "429.111.111.138", string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region29", "Test_Region29,Zone5", "Test_Region29,Zone5,Rack1", "HostName3", "429.111.111.139", string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)
	ids = append(ids, id1...)
	ids = append(ids, id2...)
	ids = append(ids, id3...)
	loc1 := structs.Location{}
	loc1.Region = "Test_Region29"
	loc1.Zone = "Zone5"

	filter1 := resource_structs.Filter{}
	filter1.Arch = string(constants.ArchX8664)
	require1 := newRequirementForRequest(4, 8, false, 0, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant = resource_structs.Applicant{}
	test_req.Applicant.HolderId = "TestCluster29"
	test_req.Applicant.RequestId = "TestRequestID29"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc1,
		HostFilter: filter1,
		Strategy:   resource_structs.RandomRack,
		Require:    *require1,
		Count:      3,
	})

	loc2 := structs.Location{}
	loc2.Region = "Test_Region29"

	require2 := resource_structs.Requirement{}
	require2.PortReq = append(require2.PortReq, resource_structs.PortRequirement{
		Start:   11000,
		End:     11005,
		PortCnt: 5,
	})

	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc2,
		HostFilter: filter1,
		Strategy:   resource_structs.ClusterPorts,
		Require:    require2,
		Count:      1,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rsp.BatchResults))
	assert.Equal(t, 4, len(rsp.BatchResults[0].Results))
	assert.Equal(t, int32(1), rsp.BatchResults[0].Results[3].Reqseq)
	assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "429.111.111.137")
	assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "429.111.111.138")
	assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "429.111.111.139")
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
	assert.Equal(t, "", rsp.BatchResults[0].Results[0].DiskRes.DiskId)
	assert.Equal(t, "", rsp.BatchResults[0].Results[1].DiskRes.DiskId)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "429.111.111.137")
	assert.Equal(t, int32(17-4), host.FreeCpuCores)
	assert.Equal(t, int32(64-8), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))
	//var usedPorts []int32
	for _, id := range ids {
		var usedPorts []management.UsedPort
		MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", id).Scan(&usedPorts)
		assert.Equal(t, 10, len(usedPorts))

		for i := 0; i < 5; i++ {
			assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[i].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[i].RequestId)
		}
		for i := 0; i < 5; i++ {
			assert.Equal(t, int32(11000+i), usedPorts[i+5].Port)
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[i+5].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[i+5].RequestId)
		}
	}

	err = recycleClusterResources("TestCluster29")
	assert.Nil(t, err)
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "429.111.111.138")
	assert.Equal(t, int32(16), host2.FreeCpuCores)
	assert.Equal(t, int32(64), host2.FreeMemory)
	assert.True(t, host2.Stat == string(constants.HostLoadLoadLess))
	var usedPorts2 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("holder_id = ?", "TestCluster29").Scan(&usedPorts2)
	assert.Equal(t, 0, len(usedPorts2))

	err = GormRW.Delete(context.TODO(), ids)
	assert.Nil(t, err)
}

func TestAllocResources_Host_Recycle_Strategy(t *testing.T) {
	id1, _ := createTestHost("Test_Region37", "Test_Region37,Zone41", "Test_Region37,Zone41,3-1", "HostName1", "437.111.111.127",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 32, 64, 3)

	loc1 := newLocation("Test_Region37", "Zone41", "", "437.111.111.127")
	require1 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster11"
	test_req.Applicant.RequestId = "TestRequestID11"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: *loc1,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require1,
		Count:    1,
	})

	loc2 := newLocation("Test_Region37", "Zone41", "", "437.111.111.127")
	require2 := newRequirementForRequest(16, 16, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req2 resource_structs.AllocReq
	test_req2.Applicant.HolderId = "TestCluster12"
	test_req2.Applicant.RequestId = "TestRequestID12"
	test_req2.Requires = append(test_req2.Requires, resource_structs.AllocRequirement{
		Location: *loc2,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require2,
		Count:    1,
	})

	loc3 := newLocation("Test_Region37", "Zone41", "", "437.111.111.127")
	require3 := newRequirementForRequest(8, 32, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req3 resource_structs.AllocReq
	test_req3.Applicant.HolderId = "TestCluster11"
	test_req3.Applicant.RequestId = "TestRequestID13"
	test_req3.Requires = append(test_req3.Requires, resource_structs.AllocRequirement{
		Location: *loc3,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require3,
		Count:    1,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req2)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req3)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rsp.BatchResults))
	assert.True(t, len(rsp.BatchResults[0].Results) == 1 && len(rsp.BatchResults[1].Results) == 1 && len(rsp.BatchResults[2].Results) == 1)
	assert.Equal(t, int32(10008), rsp.BatchResults[1].Results[0].PortRes[0].Ports[3])

	assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "437.111.111.127" && rsp.BatchResults[1].Results[0].HostIp == "437.111.111.127" && rsp.BatchResults[2].Results[0].HostIp == "437.111.111.127")
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "437.111.111.127")
	assert.Equal(t, int32(32-4-16-8), host.FreeCpuCores)
	assert.Equal(t, int32(64-8-16-32), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))

	//var usedPorts []int32
	var usedPorts []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts)
	assert.Equal(t, 15, len(usedPorts))
	for i := 0; i < 15; i++ {
		assert.Equal(t, int32(10000+i), usedPorts[i].Port)
	}

	type UsedCompute struct {
		TotalCpuCores int
		TotalMemory   int
		HolderId      string
	}
	var usedCompute []UsedCompute
	MetaDB.Model(&management.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute)
	assert.Equal(t, 2, len(usedCompute))
	assert.True(t, usedCompute[0].TotalCpuCores == 12 && usedCompute[0].TotalMemory == 40 && usedCompute[0].HolderId == "TestCluster11")
	assert.True(t, usedCompute[1].TotalCpuCores == 16 && usedCompute[1].TotalMemory == 16 && usedCompute[1].HolderId == "TestCluster12")

	diskId1 := rsp.BatchResults[0].Results[0].DiskRes.DiskId // used by cluster1
	diskId2 := rsp.BatchResults[1].Results[0].DiskRes.DiskId // used by cluster2
	diskId3 := rsp.BatchResults[2].Results[0].DiskRes.DiskId // used by cluster3
	var disks []resourcepool.Disk
	MetaDB.Model(&resourcepool.Disk{}).Where("id = ? or id = ? or id = ?", diskId1, diskId2, diskId3).Scan(&disks)

	assert.Equal(t, string(constants.DiskExhaust), disks[0].Status)
	assert.Equal(t, string(constants.DiskExhaust), disks[1].Status)
	assert.Equal(t, string(constants.DiskExhaust), disks[2].Status)

	// recycle part 1
	var request resource_structs.RecycleRequest
	var recycleRequire resource_structs.RecycleRequire
	recycleRequire.HostID = id1[0]
	recycleRequire.HolderID = "TestCluster11"
	recycleRequire.ComputeReq.CpuCores = 4
	recycleRequire.ComputeReq.Memory = 8
	recycleRequire.DiskReq = append(recycleRequire.DiskReq, resource_structs.DiskResource{
		DiskId: diskId1,
	})
	recycleRequire.PortReq = append(recycleRequire.PortReq, resource_structs.PortResource{
		Ports: []int32{10000, 10001, 10002},
	})
	recycleRequire.RecycleType = resource_structs.RecycleHost
	request.RecycleReqs = append(request.RecycleReqs, recycleRequire)
	err = GormRW.RecycleResources(context.TODO(), &request)
	assert.Nil(t, err)
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "437.111.111.127")
	assert.Equal(t, int32(32-16-8), host2.FreeCpuCores)
	assert.Equal(t, int32(64-16-32), host2.FreeMemory)
	var usedCompute2 []UsedCompute
	MetaDB.Model(&management.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute2)
	assert.Equal(t, 2, len(usedCompute2))
	assert.True(t, usedCompute2[0].TotalCpuCores == 8 && usedCompute2[0].TotalMemory == 32 && usedCompute2[0].HolderId == "TestCluster11")
	assert.True(t, usedCompute2[1].TotalCpuCores == 16 && usedCompute2[1].TotalMemory == 16 && usedCompute2[1].HolderId == "TestCluster12")
	var usedPorts2 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts2)
	assert.Equal(t, 12, len(usedPorts2))
	for i := 3; i < 15; i++ {
		assert.Equal(t, int32(10000+i), usedPorts2[i-3].Port)
	}
	var disk1 resourcepool.Disk
	MetaDB.Model(&resourcepool.Disk{}).Where("id = ?", diskId1).Scan(&disk1)
	assert.Equal(t, string(constants.DiskAvailable), disk1.Status)

	// recycle part 2
	var request2 resource_structs.RecycleRequest
	var recycleRequire2 resource_structs.RecycleRequire
	recycleRequire2.HostID = id1[0]
	recycleRequire2.HolderID = "TestCluster11"
	recycleRequire2.ComputeReq.CpuCores = 8
	recycleRequire2.ComputeReq.Memory = 32
	recycleRequire2.DiskReq = append(recycleRequire2.DiskReq, resource_structs.DiskResource{
		DiskId: diskId3,
	})

	recycleRequire2.PortReq = append(recycleRequire2.PortReq, resource_structs.PortResource{
		Ports: []int32{10003, 10004, 10010, 10011, 10012, 10013, 10014},
	})
	recycleRequire2.RecycleType = resource_structs.RecycleHost
	request2.RecycleReqs = append(request2.RecycleReqs, recycleRequire2)
	err = GormRW.RecycleResources(context.TODO(), &request2)
	assert.Nil(t, err)
	var host3 resourcepool.Host
	MetaDB.First(&host3, "IP = ?", "437.111.111.127")
	assert.Equal(t, int32(32-16), host3.FreeCpuCores)
	assert.Equal(t, int32(64-16), host3.FreeMemory)
	var usedCompute3 []UsedCompute
	MetaDB.Model(&management.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute3)
	assert.Equal(t, 1, len(usedCompute3))
	assert.True(t, usedCompute3[0].TotalCpuCores == 16 && usedCompute3[0].TotalMemory == 16 && usedCompute3[0].HolderId == "TestCluster12")
	var usedPorts3 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts3)
	assert.Equal(t, 5, len(usedPorts3))
	for i := 5; i < 10; i++ {
		assert.Equal(t, int32(10000+i), usedPorts3[i-5].Port)
	}
	var disk2 resourcepool.Disk
	MetaDB.Model(&resourcepool.Disk{}).Where("id = ?", diskId3).Scan(&disk2)
	assert.Equal(t, string(constants.DiskAvailable), disk2.Status)

	// recycle all
	var request3 resource_structs.RecycleRequest
	var recycleRequire3 resource_structs.RecycleRequire
	recycleRequire3.HostID = id1[0]
	recycleRequire3.HolderID = "TestCluster12"
	recycleRequire3.ComputeReq.CpuCores = 16
	recycleRequire3.ComputeReq.Memory = 16
	recycleRequire3.DiskReq = append(recycleRequire3.DiskReq, resource_structs.DiskResource{
		DiskId: diskId2,
	})
	recycleRequire3.PortReq = append(recycleRequire3.PortReq, resource_structs.PortResource{
		Ports: []int32{10005, 10006, 10007, 10008, 10009},
	})
	recycleRequire3.RecycleType = resource_structs.RecycleHost
	request3.RecycleReqs = append(request3.RecycleReqs, recycleRequire3)
	err = GormRW.RecycleResources(context.TODO(), &request3)
	assert.Nil(t, err)
	var host4 resourcepool.Host
	MetaDB.First(&host4, "IP = ?", "437.111.111.127")
	assert.Equal(t, int32(32), host4.FreeCpuCores)
	assert.Equal(t, int32(64), host4.FreeMemory)
	//assert.Equal(t, int32(resource.HOST_LOADLESS), host4.Stat)
	var usedCompute4 []UsedCompute
	MetaDB.Model(&management.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute4)
	assert.Equal(t, 0, len(usedCompute4))
	var usedPorts4 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts4)
	assert.Equal(t, 0, len(usedPorts4))
	var disk3 resourcepool.Disk
	MetaDB.Model(&resourcepool.Disk{}).Where("id = ?", diskId2).Scan(&disk3)
	assert.Equal(t, string(constants.DiskAvailable), disk3.Status)

	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
}

func TestAllocResources_Request3Count_SpecifyHost_Strategy(t *testing.T) {
	id1, _ := createTestHost("Test_Region12", "Test_Region12,Test_Zone2", "Test_Region12,Test_Zone2,Test_Rack1", "Test_Host1", "494.111.111.127",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region12", "Test_Region12,Test_Zone2", "Test_Region12,Test_Zone2,Test_Rack1", "Test_Host2", "494.111.111.128",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region12", "Test_Region12,Test_Zone2", "Test_Region12,Test_Zone2,Test_Rack1", "Test_Host3", "494.111.111.129",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc1 := structs.Location{}
	loc1.Region = "Test_Region12"
	loc1.Zone = "Test_Zone2"
	loc1.HostIp = "494.111.111.127"

	require1 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc1,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require1,
		Count:    3,
	})

	loc2 := structs.Location{}
	loc2.Region = "Test_Region12"
	loc2.Zone = "Test_Zone2"
	loc2.HostIp = "494.111.111.128"

	require2 := newRequirementForRequest(4, 8, false, 256, string(constants.SSD), 10000, 10015, 5)

	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc2,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require2,
		Count:    3,
	})

	loc3 := structs.Location{}
	loc3.Region = "Test_Region12"
	loc3.Zone = "Test_Zone2"
	loc3.HostIp = "494.111.111.129"

	require3 := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location: loc3,
		Strategy: resource_structs.UserSpecifyHost,
		Require:  *require3,
		Count:    1,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	rsp, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rsp.BatchResults))
	assert.Equal(t, 7, len(rsp.BatchResults[0].Results))
	assert.Equal(t, int32(0), rsp.BatchResults[0].Results[2].Reqseq)
	assert.Equal(t, int32(1), rsp.BatchResults[0].Results[3].Reqseq)
	assert.Equal(t, int32(2), rsp.BatchResults[0].Results[6].Reqseq)
	assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
	assert.Equal(t, int32(10008), rsp.BatchResults[0].Results[1].PortRes[0].Ports[3])
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[3].HostIp)
	assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "494.111.111.127")
	assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "494.111.111.127")
	assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "494.111.111.127")
	assert.True(t, rsp.BatchResults[0].Results[3].HostIp == "494.111.111.128")
	assert.True(t, rsp.BatchResults[0].Results[4].HostIp == "494.111.111.128")
	assert.True(t, rsp.BatchResults[0].Results[5].HostIp == "494.111.111.128")
	assert.True(t, rsp.BatchResults[0].Results[6].HostIp == "494.111.111.129")
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[1].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[2].ComputeRes.Memory)
	assert.Equal(t, int32(4), rsp.BatchResults[0].Results[4].ComputeRes.CpuCores)
	assert.Equal(t, int32(8), rsp.BatchResults[0].Results[5].ComputeRes.Memory)
	assert.Equal(t, "admin2", rsp.BatchResults[0].Results[0].Passwd)
	var host resourcepool.Host
	MetaDB.First(&host, "IP = ?", "494.111.111.127")
	assert.Equal(t, int32(17-4-4-4), host.FreeCpuCores)
	assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
	assert.True(t, host.Stat == string(constants.HostLoadInUsed))
	var host2 resourcepool.Host
	MetaDB.First(&host2, "IP = ?", "494.111.111.128")
	assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
	assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
	assert.True(t, host2.Stat == string(constants.HostLoadInUsed))
	var host3 resourcepool.Host
	MetaDB.First(&host3, "IP = ?", "494.111.111.129")
	assert.Equal(t, int32(15-4), host3.FreeCpuCores)
	assert.Equal(t, int32(64-8), host3.FreeMemory)
	assert.True(t, host3.Stat == string(constants.HostLoadInUsed))
	//var usedPorts []int32
	var usedPorts []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
	assert.Equal(t, 5, len(usedPorts))
	assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
	assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
	for i := 0; i < 5; i++ {
		assert.Equal(t, int32(10000+i), usedPorts[i].Port)
	}

	var usedPorts2 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host2.ID).Scan(&usedPorts2)
	assert.Equal(t, 15, len(usedPorts2))
	assert.Equal(t, test_req.Applicant.HolderId, usedPorts2[1].HolderId)
	assert.Equal(t, test_req.Applicant.RequestId, usedPorts2[2].RequestId)
	for i := 0; i < 15; i++ {
		assert.Equal(t, int32(10000+i), usedPorts2[i].Port)
	}

	var usedPorts3 []management.UsedPort
	MetaDB.Order("port").Model(&management.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts3)
	assert.Equal(t, 15, len(usedPorts3))
	assert.Equal(t, test_req.Applicant.HolderId, usedPorts3[1].HolderId)
	assert.Equal(t, test_req.Applicant.RequestId, usedPorts3[2].RequestId)
	for i := 0; i < 15; i++ {
		assert.Equal(t, int32(10000+i), usedPorts3[i].Port)
	}

	var usedComputes []management.UsedCompute
	MetaDB.Order("cpu_cores").Model(&management.UsedCompute{}).Where("host_id = ?", host.ID).Scan(&usedComputes)
	var totalUsedCpu, totalUsedMem int32
	for i := range usedComputes {
		totalUsedCpu += usedComputes[i].CpuCores
		totalUsedMem += usedComputes[i].Memory
	}
	assert.Equal(t, int32(4+4+4), totalUsedCpu)
	assert.Equal(t, int32(8+8+8), totalUsedMem)

	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}
