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

package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/micro-metadb/models"

	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/stretchr/testify/assert"
)

type UsedComputeStatistic struct {
	HostId        string
	TotalCpuCores int
	TotalMemory   int
}

type UsedDiskStatistic struct {
	HostId    string
	DiskCount int
}

type usedPortStatistic struct {
	HostId    string
	PortCount int
}

func TestDBServiceHandler_Alloc_Recycle_Resources(t *testing.T) {

	err := Dao.InitResourceDataForDev("Region1", "Zone1", "Rack1", "168.168.168.1", "168.168.168.2", "168.168.168.3")
	if err != nil {
		t.Errorf("InitResourceDataForDev failed in RecycleResources() error = %v, ", err)
	}

	type args struct {
		ctx context.Context
		req *dbpb.DBBatchAllocRequest
		rsp *dbpb.DBBatchAllocResponse
	}

	loc := new(dbpb.DBLocation)
	loc.Region = "Region1"
	loc.Zone = "Zone1"
	filter1 := new(dbpb.DBFilter)
	filter1.Arch = string(resource.X86)
	filter1.DiskType = string(resource.Sata)
	filter1.Purpose = string(resource.Compute)
	require := new(dbpb.DBRequirement)
	require.Exclusive = false
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.Sata)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	clusterId := "TestClusterId1"
	requestId := "TestRequestId1"
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = clusterId
	test_req.Applicant.RequestId = requestId
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	var batchRsp dbpb.DBBatchAllocResponse

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal",
			args: args{
				ctx: context.TODO(),
				req: &batchReq,
				rsp: &batchRsp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handler.AllocResourcesInBatch(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("AllocResourcesInBatch() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, int32(0), tt.args.rsp.Rs.Code)
				t.Log(tt.args.rsp.Rs.Message)
			}
			assert.Equal(t, 3, len(batchRsp.BatchResults))
			assert.Equal(t, 3, len(batchRsp.BatchResults[0].Results))
			assert.True(t, batchRsp.BatchResults[0].Results[0].HostIp != batchRsp.BatchResults[0].Results[1].HostIp && batchRsp.BatchResults[0].Results[0].HostIp != batchRsp.BatchResults[0].Results[2].HostIp &&
				batchRsp.BatchResults[0].Results[1].HostIp != batchRsp.BatchResults[0].Results[2].HostIp)
			assert.True(t, batchRsp.BatchResults[0].Results[0].HostIp == "168.168.168.1" || batchRsp.BatchResults[0].Results[0].HostIp == "168.168.168.2" ||
				batchRsp.BatchResults[0].Results[0].HostIp == "168.168.168.3")
			assert.Equal(t, 3, len(batchRsp.BatchResults[1].Results))
			assert.Equal(t, 3, len(batchRsp.BatchResults[2].Results))
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "168.168.168.1")
			assert.Equal(t, int32(16-4-4-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "168.168.168.2")
			assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			stat, isExaust := host2.IsExhaust()
			assert.True(t, stat == resource.HOST_DISK_EXHAUST && isExaust == true)
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "168.168.168.3")
			assert.Equal(t, int32(12-4-4-4), host3.FreeCpuCores)
			assert.Equal(t, int32(24-8-8-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
			stat, isExaust = host3.IsExhaust()
			assert.True(t, stat == resource.HOST_EXHAUST && isExaust == true)

			var usedComputes []UsedComputeStatistic
			MetaDB.Model(&resource.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedComputes)
			assert.Equal(t, 3, len(usedComputes))
			assert.True(t, usedComputes[0].TotalCpuCores == 12 && usedComputes[1].TotalCpuCores == 12 && usedComputes[2].TotalCpuCores == 12)
			assert.True(t, usedComputes[0].TotalMemory == 24 && usedComputes[1].TotalMemory == 24 && usedComputes[2].TotalMemory == 24)
			var usedDisks []UsedDiskStatistic
			MetaDB.Model(&resource.UsedDisk{}).Select("host_id, count(disk_id) as disk_count").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedDisks)
			assert.True(t, usedDisks[0].DiskCount == 3 && usedDisks[1].DiskCount == 3 && usedDisks[2].DiskCount == 3)
			var usedPorts []usedPortStatistic
			MetaDB.Model(&resource.UsedPort{}).Select("host_id, count(port) as port_count").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedPorts)
			assert.True(t, usedPorts[0].PortCount == 15 && usedPorts[1].PortCount == 15 && usedPorts[2].PortCount == 15)

			recycleRequest := new(dbpb.DBRecycleRequest)
			recycleRequest.RecycleReqs = append(recycleRequest.RecycleReqs, &dbpb.DBRecycleRequire{
				RecycleType: int32(resource.RecycleHolder),
				HolderId:    clusterId,
			})
			recycleResponse := new(dbpb.DBRecycleResponse)
			if err := handler.RecycleResources(tt.args.ctx, recycleRequest, recycleResponse); (err != nil) != tt.wantErr {
				t.Errorf("RecycleResources() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, int32(0), tt.args.rsp.Rs.Code)
				t.Log(tt.args.rsp.Rs.Message)
			}

			MetaDB.First(&host, "IP = ?", "168.168.168.1")
			assert.Equal(t, int32(16), host.FreeCpuCores)
			assert.Equal(t, int32(64), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_LOADLESS))

			var usedComputes2 []UsedComputeStatistic
			MetaDB.Model(&resource.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedComputes2)
			assert.Equal(t, 0, len(usedComputes2))
			var usedDisks2 []UsedDiskStatistic
			MetaDB.Model(&resource.UsedDisk{}).Select("host_id, count(disk_id) as disk_count").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedDisks2)
			assert.Equal(t, 0, len(usedDisks2))
			var usedPorts2 []usedPortStatistic
			MetaDB.Model(&resource.UsedPort{}).Select("host_id, count(port) as port_count").Where("holder_id = ?", clusterId).Group("host_id").Scan(&usedPorts2)
			assert.Equal(t, 0, len(usedPorts2))

		})
	}
}

func TestDBServiceHandler_Recycle_Host_Resources(t *testing.T) {

	err := Dao.InitResourceDataForDev("Region2", "Zone2", "Rack2", "168.168.168.4", "168.168.168.5", "168.168.168.6")
	if err != nil {
		t.Errorf("InitResourceDataForDev failed in RecycleResources() error = %v, ", err)
	}

	type args struct {
		ctx context.Context
		req *dbpb.DBBatchAllocRequest
		rsp *dbpb.DBBatchAllocResponse
	}

	loc := new(dbpb.DBLocation)
	loc.Region = "Region2"
	loc.Zone = "Zone2"
	filter1 := new(dbpb.DBFilter)
	filter1.Arch = string(resource.X86)
	filter1.DiskType = string(resource.Sata)
	filter1.Purpose = string(resource.Compute)
	require := new(dbpb.DBRequirement)
	require.Exclusive = false
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.Sata)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	clusterId := "TestClusterId2"
	requestId := "TestRequestId2"
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = clusterId
	test_req.Applicant.RequestId = requestId
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	var batchRsp dbpb.DBBatchAllocResponse

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal",
			args: args{
				ctx: context.TODO(),
				req: &batchReq,
				rsp: &batchRsp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handler.AllocResourcesInBatch(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("AllocResourcesInBatch() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, int32(0), tt.args.rsp.Rs.Code)
				t.Log(tt.args.rsp.Rs.Message)
			}
			assert.Equal(t, 1, len(batchRsp.BatchResults))
			assert.Equal(t, 3, len(batchRsp.BatchResults[0].Results))

			var host resource.Host
			MetaDB.First(&host, "IP = ?", "168.168.168.4")
			assert.Equal(t, int32(16-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var diskId string
			for _, disk := range host.Disks {
				if disk.Status == int32(resource.DISK_EXHAUST) {
					diskId = disk.ID
					break
				}
			}
			assert.True(t, diskId != "")

			var usedDisks []UsedDiskStatistic
			MetaDB.Model(&resource.UsedDisk{}).Select("host_id, count(disk_id) as disk_count").Where("holder_id = ?", clusterId).Group("host_id").Having("host_id = ?", host.ID).Scan(&usedDisks)
			assert.True(t, len(usedDisks) == 1 && usedDisks[0].DiskCount == 1)
			var usedPorts []usedPortStatistic
			MetaDB.Model(&resource.UsedPort{}).Select("host_id, count(port) as port_count").Where("holder_id = ?", clusterId).Group("host_id").Having("host_id = ?", host.ID).Scan(&usedPorts)
			assert.True(t, len(usedPorts) == 1 && usedPorts[0].PortCount == 5)

			recycleRequest := new(dbpb.DBRecycleRequest)
			recycleRequest.RecycleReqs = append(recycleRequest.RecycleReqs, &dbpb.DBRecycleRequire{
				RecycleType: int32(resource.RecycleHost),
				HostId:      host.ID,
				HolderId:    clusterId,
				ComputeReq: &dbpb.DBComputeRequirement{
					CpuCores: 4,
					Memory:   8,
				},
			})
			recycleRequest.RecycleReqs[0].DiskReq = append(recycleRequest.RecycleReqs[0].DiskReq, &dbpb.DBDiskResource{
				DiskId: diskId,
			})
			recycleRequest.RecycleReqs[0].PortReq = append(recycleRequest.RecycleReqs[0].PortReq, &dbpb.DBPortResource{
				Ports: []int32{10000, 10001, 10002, 10003, 10004},
			})
			recycleResponse := new(dbpb.DBRecycleResponse)
			if err := handler.RecycleResources(tt.args.ctx, recycleRequest, recycleResponse); (err != nil) != tt.wantErr {
				t.Errorf("RecycleResources() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, int32(0), tt.args.rsp.Rs.Code)
				t.Log(tt.args.rsp.Rs.Message)
			}

			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "168.168.168.4")
			assert.Equal(t, int32(16), host2.FreeCpuCores)
			assert.Equal(t, int32(64), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_LOADLESS))

			var usedDisks2 []UsedDiskStatistic
			MetaDB.Model(&resource.UsedDisk{}).Select("host_id, count(disk_id) as disk_count").Where("holder_id = ?", clusterId).Group("host_id").Having("host_id = ?", host.ID).Scan(&usedDisks2)
			assert.True(t, len(usedDisks2) == 0)
			var usedPorts2 []usedPortStatistic
			MetaDB.Model(&resource.UsedPort{}).Select("host_id, count(port) as port_count").Where("holder_id = ?", clusterId).Group("host_id").Having("host_id = ?", host.ID).Scan(&usedPorts2)
			assert.True(t, len(usedPorts2) == 0)

		})
	}

}

func printTree(root *node, pre string) {
	if root == nil {
		return
	}
	fmt.Println(pre, "Code: ", root.Code, "Prefix: ", root.Prefix, "Name: ", root.Name)
	for _, subNode := range root.subNodes {
		printTree(subNode, "--"+pre)
	}
}

func printDBTree(root *dbpb.DBNode, pre string) {
	if root == nil {
		return
	}
	fmt.Println(pre, "Code: ", root.Code, "Prefix: ", root.Prefix, "Name: ", root.Name)
	for _, subNode := range root.SubNodes {
		printDBTree(subNode, "--"+pre)
	}
}

func generateItems() []models.Item {
	Items := []models.Item{
		{Region: "Region1", Az: "Region1,Zone1", Rack: "Region1,Zone1,Rack1", Ip: "111.111.111.111", Name: "Host1"},
		{Region: "Region1", Az: "Region1,Zone2", Rack: "Region1,Zone2,Rack1", Ip: "111.111.111.112", Name: "Host2"},
		{Region: "Region2", Az: "Region2,Zone3", Rack: "Region2,Zone3,Rack2", Ip: "111.111.111.113", Name: "Host3"},
		{Region: "Region2", Az: "Region2,Zone3", Rack: "Region2,Zone3,Rack3", Ip: "111.111.111.114", Name: "Host4"},
		{Region: "Region3", Az: "Region3,Zone4", Rack: "Region3,Zone4,Rack4", Ip: "111.111.111.115", Name: "Host5"},
		{Region: "Region3", Az: "Region3,Zone5", Rack: "Region3,Zone5,Rack1", Ip: "111.111.111.116", Name: "Host6"},
		{Region: "Region3", Az: "Region3,Zone5", Rack: "Region3,Zone5,Rack1", Ip: "111.111.111.117", Name: "Host7"},
	}
	return Items
}

func TestDBServiceHandler_Build_Hierarchy(t *testing.T) {
	Items := generateItems()
	root := handler.buildHierarchy(Items)
	assert.Equal(t, 3, len(root.subNodes))
	assert.Equal(t, 2, len(root.subNodes[0].subNodes))
	assert.Equal(t, 1, len(root.subNodes[1].subNodes))
	assert.Equal(t, 2, len(root.subNodes[2].subNodes))

	printTree(root, "-->")
}

func TestDBServiceHandler_TrimRegion_Hierarchy(t *testing.T) {
	Items := generateItems()
	root := handler.buildHierarchy(Items)

	newRoot := handler.trimTree(root, resource.REGION, 1)
	printTree(newRoot, "-->")
	assert.Equal(t, 3, len(newRoot.subNodes)) // Regions
	assert.Equal(t, 2, len(newRoot.subNodes[0].subNodes))
	assert.Equal(t, 1, len(newRoot.subNodes[1].subNodes))
	assert.Equal(t, 2, len(newRoot.subNodes[2].subNodes))
	assert.True(t, newRoot.subNodes[0].subNodes[1].subNodes == nil)
	assert.True(t, newRoot.subNodes[1].subNodes[0].subNodes == nil)
	assert.True(t, newRoot.subNodes[2].subNodes[0].subNodes == nil)
}

func TestDBServiceHandler_TrimZone_Hierarchy(t *testing.T) {
	Items := generateItems()
	root := handler.buildHierarchy(Items)

	newRoot := handler.trimTree(root, resource.ZONE, 2)
	printTree(newRoot, "-->")
	assert.Equal(t, 5, len(newRoot.subNodes)) // Zones
	assert.Equal(t, 1, len(newRoot.subNodes[0].subNodes))
	assert.Equal(t, 1, len(newRoot.subNodes[1].subNodes))
	assert.Equal(t, 2, len(newRoot.subNodes[2].subNodes))
	assert.Equal(t, 1, len(newRoot.subNodes[3].subNodes))
	assert.Equal(t, 1, len(newRoot.subNodes[4].subNodes))
}

func TestDBServiceHandler_TrimRack_Hierarchy(t *testing.T) {
	Items := generateItems()
	root := handler.buildHierarchy(Items)

	newRoot := handler.trimTree(root, resource.RACK, 0)
	printTree(newRoot, "-->")
	assert.Equal(t, 6, len(newRoot.subNodes)) // Racks
	assert.True(t, newRoot.subNodes[2].subNodes == nil)
}

func TestDBServiceHandler_Copy_Hierarchy(t *testing.T) {
	Items := generateItems()
	root := handler.buildHierarchy(Items)
	var newRoot *dbpb.DBNode
	copyHierarchyToRsp(root, &newRoot)
	assert.Equal(t, 3, len(newRoot.SubNodes))
	assert.Equal(t, 2, len(newRoot.SubNodes[0].SubNodes))
	assert.Equal(t, 1, len(newRoot.SubNodes[1].SubNodes))
	assert.Equal(t, 2, len(newRoot.SubNodes[2].SubNodes))

	printDBTree(newRoot, "-->")
}
