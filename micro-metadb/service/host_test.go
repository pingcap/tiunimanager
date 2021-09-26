package service

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/library/common/resource-type"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/stretchr/testify/assert"
)

func TestDBServiceHandler_AllocResourcesInBatch(t *testing.T) {

	Dao.InitResourceDataForDev()

	type args struct {
		ctx context.Context
		req *dbPb.BatchAllocRequest
		rsp *dbPb.BatchAllocResponse
	}

	loc := new(dbPb.Location)
	loc.Region = "Region1"
	loc.Zone = "Zone1"
	filter1 := new(dbPb.Filter)
	filter1.DiskType = string(resource.Sata)
	filter1.Purpose = string(resource.General)
	require := new(dbPb.Requirement)
	require.ComputeReq = new(dbPb.ComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbPb.DiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.Sata)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbPb.PortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbPb.AllocRequest
	test_req.Applicant = new(dbPb.Applicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbPb.AllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbPb.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	var batchRsp dbPb.BatchAllocResponse

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
		})
	}
}
