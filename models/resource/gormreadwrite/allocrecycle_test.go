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
	"github.com/pingcap-inc/tiem/common/structs"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/structs"
	"github.com/pingcap-inc/tiem/models/resource/management"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/stretchr/testify/assert"
)

func CreateTestHost(region, zone, rack, hostName, ip, clusterType, purpose, diskType string, freeCpuCores, freeMemory, availableDiskCount int32) (id []string, err error) {
	h := resourcepool.Host{
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
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

func Test_Create_Delete_Host_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
}

func Test_Create_Dup_Host(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	assert.NotNil(t, err)
	assert.Nil(t, id2)
}

func Test_Create_Query_Host_Succeed(t *testing.T) {
	hostIp := "192.168.999.999"
	hostName := "Test_Host2"
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", hostName, hostIp,
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{HostID: id1[0]}, 0, 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hosts))
	assert.Equal(t, id1[0], hosts[0].ID)
	assert.Equal(t, hostName, hosts[0].HostName)
	assert.Equal(t, hostIp, hosts[0].IP)
	assert.Equal(t, 3, len(hosts[0].Disks))
}

func Test_UpdateHostReserved_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.False(t, hosts[0].Reserved)
	GormRW.UpdateHostReserved(context.TODO(), id1, true)
	hosts, err = GormRW.Query(context.TODO(), &structs.HostFilter{Stat: string(constants.HostLoadLoadLess)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.True(t, hosts[0].Reserved)
}

func Test_UpdateHostStatus_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, string(constants.HostOnline), hosts[0].Status)
	GormRW.UpdateHostStatus(context.TODO(), id1, string(constants.HostOffline))
	hosts, err = GormRW.Query(context.TODO(), &structs.HostFilter{Purpose: string(constants.PurposeCompute)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, string(constants.HostOffline), hosts[0].Status)
}

func Test_GetHostItems(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	items, err := GormRW.GetHostItems(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664)}, 1, 3)
	assert.True(t, err == nil && len(items) == 3)
	assert.Equal(t, "192.168.192.168", items[0].Ip)
	assert.Equal(t, "192.168.192.169", items[1].Ip)
	assert.Equal(t, "192.168.192.170", items[2].Ip)
}

func Test_GetHostStocks(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 32, 64, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	stocks, err := GormRW.GetHostStocks(context.TODO(), &structs.Location{Region: "Test_Region1", Zone: "Test_Zone2"}, &structs.HostFilter{}, &structs.DiskFilter{})
	t.Log(stocks)
	assert.True(t, err == nil)
	assert.Equal(t, 1, len(stocks))
	assert.Equal(t, int32(32), stocks[0].FreeCpuCores)
	assert.Equal(t, int32(64), stocks[0].FreeMemory)
	assert.Equal(t, int32(3), stocks[0].FreeDiskCount)
	assert.Equal(t, int32(3*256), stocks[0].FreeDiskCapacity)
}

func TestAllocResources_3RequestsInBatch_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.117",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.118",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.119",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

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

}

func TestAllocResources_1Requirement_3Hosts_Filted_by_Label(t *testing.T) {
	id1, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host1", "474.111.111.158",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute)+","+string(constants.PurposeStorage), string(constants.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host2", "474.111.111.159",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone3", "Test_Region1,Test_Zone3,Test_Rack1", "Test_Host3", "474.111.111.160",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 1)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

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

func TestAllocResources_3RequestsInBatch_SpecifyHost_Strategy(t *testing.T) {
	id1, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.127",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.128",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.129",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

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
}
