package service

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/library/common/resource-type"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
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

	Dao.InitResourceDataForDev()

	type args struct {
		ctx context.Context
		req *dbPb.DBBatchAllocRequest
		rsp *dbPb.DBBatchAllocResponse
	}

	loc := new(dbPb.DBLocation)
	loc.Region = "Region1"
	loc.Zone = "Zone1"
	filter1 := new(dbPb.DBFilter)
	filter1.Arch = string(resource.X86)
	filter1.DiskType = string(resource.Sata)
	filter1.Purpose = string(resource.General)
	require := new(dbPb.DBRequirement)
	require.Exclusive = false
	require.ComputeReq = new(dbPb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbPb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.Sata)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbPb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbPb.DBAllocRequest
	clusterId := "TestClusterId1"
	requestId := "TestRequestId1"
	test_req.Applicant = new(dbPb.DBApplicant)
	test_req.Applicant.HolderId = clusterId
	test_req.Applicant.RequestId = requestId
	test_req.Requires = append(test_req.Requires, &dbPb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbPb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	var batchRsp dbPb.DBBatchAllocResponse

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
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "168.168.168.2")
			assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_EXHAUST))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "168.168.168.3")
			assert.Equal(t, int32(16-4-4-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_EXHAUST))

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

			recycleRequest := new(dbPb.DBRecycleRequest)
			recycleRequest.RecycleReqs = append(recycleRequest.RecycleReqs, &dbPb.DBRecycleRequire{
				RecycleType: int32(resource.RecycleHolder),
				HolderId:    clusterId,
			})
			recycleResponse := new(dbPb.DBRecycleResponse)
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
