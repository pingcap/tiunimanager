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

package resource

import (
	"context"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/golang/mock/gomock"
	rpc_client "github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	mock "github.com/pingcap-inc/tiem/test/mockdb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_AllocResourcesInBatch_Succeed(t *testing.T) {

	fake_host_id1 := "TEST_host_id1"
	fake_host_ip1 := "199.199.199.1"
	fake_holder_id := "TEST_holder1"
	fake_request_id := "TEST_reqeust1"
	fake_disk_id1 := "TEST_disk_id1"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().AllocResourcesInBatch(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error) {
		rsp := new(dbpb.DBBatchAllocResponse)
		rsp.Rs = new(dbpb.DBAllocResponseStatus)
		if in.BatchRequests[0].Applicant.HolderId == fake_holder_id && in.BatchRequests[1].Applicant.RequestId == fake_request_id &&
			in.BatchRequests[0].Requires[0].Count == 1 && in.BatchRequests[1].Requires[1].Require.PortReq[1].PortCnt == 2 {
			rsp.Rs.Code = int32(codes.OK)
		} else {
			return nil, status.Error(codes.Internal, "BAD alloc resource request")
		}
		var r dbpb.DBHostResource
		r.Reqseq = 0
		r.HostId = fake_host_id1
		r.HostIp = fake_host_ip1
		r.ComputeRes = new(dbpb.DBComputeRequirement)
		r.ComputeRes.CpuCores = in.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.ComputeRes.Memory = in.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.Location = new(dbpb.DBLocation)
		r.Location.Region = in.BatchRequests[0].Requires[0].Location.Region
		r.Location.Zone = in.BatchRequests[0].Requires[0].Location.Zone
		r.DiskRes = new(dbpb.DBDiskResource)
		r.DiskRes.DiskId = fake_disk_id1
		r.DiskRes.Type = in.BatchRequests[0].Requires[0].Require.DiskReq.DiskType
		r.DiskRes.Capacity = in.BatchRequests[0].Requires[0].Require.DiskReq.Capacity
		for _, portRes := range in.BatchRequests[0].Requires[0].Require.PortReq {
			var portResource dbpb.DBPortResource
			portResource.Start = portRes.Start
			portResource.End = portRes.End
			portResource.Ports = append(portResource.Ports, portRes.Start+1)
			portResource.Ports = append(portResource.Ports, portRes.Start+2)
			r.PortRes = append(r.PortRes, &portResource)
		}

		var one_rsp dbpb.DBAllocResponse
		one_rsp.Rs = new(dbpb.DBAllocResponseStatus)
		one_rsp.Rs.Code = int32(codes.OK)
		one_rsp.Results = append(one_rsp.Results, &r)

		rsp.BatchResults = append(rsp.BatchResults, &one_rsp)

		var two_rsp dbpb.DBAllocResponse
		two_rsp.Rs = new(dbpb.DBAllocResponseStatus)
		two_rsp.Rs.Code = int32(codes.OK)
		two_rsp.Results = append(two_rsp.Results, &r)
		two_rsp.Results = append(two_rsp.Results, &r)
		rsp.BatchResults = append(rsp.BatchResults, &two_rsp)
		return rsp, nil
	})
	rpc_client.DBClient = mockClient

	in := new(clusterpb.BatchAllocRequest)

	var require clusterpb.AllocRequirement
	require.Location = new(clusterpb.Location)
	require.Location.Region = "TesT_Region1"
	require.Location.Zone = "TEST_Zone1"
	require.HostFilter = new(clusterpb.Filter)
	require.HostFilter.Arch = "X86"
	require.HostFilter.DiskType = "sata"
	require.HostFilter.Purpose = "General"
	require.Require = new(clusterpb.Requirement)
	require.Require.ComputeReq = new(clusterpb.ComputeRequirement)
	require.Require.ComputeReq.CpuCores = 4
	require.Require.ComputeReq.Memory = 8
	require.Require.DiskReq = new(clusterpb.DiskRequirement)
	require.Require.DiskReq.NeedDisk = true
	require.Require.DiskReq.DiskType = "sata"
	require.Require.DiskReq.Capacity = 256
	require.Require.PortReq = append(require.Require.PortReq, &clusterpb.PortRequirement{
		Start:   10000,
		End:     10010,
		PortCnt: 2,
	})
	require.Require.PortReq = append(require.Require.PortReq, &clusterpb.PortRequirement{
		Start:   10010,
		End:     10020,
		PortCnt: 2,
	})
	require.Count = 1

	var req1 clusterpb.AllocRequest
	req1.Applicant = new(clusterpb.Applicant)
	req1.Applicant.HolderId = fake_holder_id
	req1.Applicant.RequestId = fake_request_id
	req1.Requires = append(req1.Requires, &require)

	var req2 clusterpb.AllocRequest
	req2.Applicant = new(clusterpb.Applicant)
	req2.Applicant.HolderId = fake_holder_id
	req2.Applicant.RequestId = fake_request_id
	req2.Requires = append(req2.Requires, &require)
	req2.Requires = append(req2.Requires, &require)

	in.BatchRequests = append(in.BatchRequests, &req1)
	in.BatchRequests = append(in.BatchRequests, &req2)

	out := new(clusterpb.BatchAllocResponse)
	if err := resourceManager.AllocResourcesInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("alloc resource failed, err: %v\n", err)
	}

	assert.True(t, out.Rs.Code == int32(codes.OK))
	assert.Equal(t, 2, len(out.BatchResults))
	assert.Equal(t, 1, len(out.BatchResults[0].Results))
	assert.Equal(t, 2, len(out.BatchResults[1].Results))
	assert.True(t, out.BatchResults[0].Results[0].DiskRes.DiskId == fake_disk_id1 && out.BatchResults[0].Results[0].DiskRes.Capacity == 256 && out.BatchResults[1].Results[0].DiskRes.Type == "sata")
	assert.True(t, out.BatchResults[1].Results[1].PortRes[0].Ports[0] == 10001 && out.BatchResults[1].Results[1].PortRes[0].Ports[1] == 10002)
	assert.True(t, out.BatchResults[1].Results[1].PortRes[1].Ports[0] == 10011 && out.BatchResults[1].Results[1].PortRes[1].Ports[1] == 10012)
}

func Test_RecycleResources_Succeed(t *testing.T) {
	fake_cluster_id := "TEST_Fake_CLUSTER_ID"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, in *dbpb.DBRecycleRequest, opts ...client.CallOption) (*dbpb.DBRecycleResponse, error) {
		rsp := new(dbpb.DBRecycleResponse)
		rsp.Rs = new(dbpb.DBAllocResponseStatus)
		if in.RecycleReqs[0].RecycleType == 2 && in.RecycleReqs[0].HolderId == fake_cluster_id {
			rsp.Rs.Code = int32(codes.OK)
		} else {
			return nil, status.Error(codes.Internal, "BAD recycle resource request")
		}
		return rsp, nil
	})
	rpc_client.DBClient = mockClient

	var req clusterpb.RecycleRequest
	var require clusterpb.RecycleRequire
	require.HolderId = fake_cluster_id
	require.RecycleType = 2

	req.RecycleReqs = append(req.RecycleReqs, &require)
	out := new(clusterpb.RecycleResponse)
	if err := resourceManager.RecycleResources(context.TODO(), &req, out); err != nil {
		t.Errorf("recycle resource failed, err: %v\n", err)
	}

	assert.True(t, out.Rs.Code == int32(codes.OK))
}

func Test_buildDBRecycleRequire(t *testing.T) {
	var recycleType int32 = 3
	holder_id := "TEST_CLUSTER3"
	request_id := "TEST_REQ3"
	host_id := "TEST_HOST3"
	host_ip := "111.111.111.111"
	diskId1 := "TEST_DISK1"
	diskId2 := "TEST_DISK2"

	var src clusterpb.RecycleRequire
	src.RecycleType = recycleType
	src.HolderId = holder_id
	src.RequestId = request_id
	src.HostId = host_id
	src.HostIp = host_ip
	src.ComputeReq = new(clusterpb.ComputeRequirement)
	src.ComputeReq.CpuCores = 32
	src.ComputeReq.Memory = 64
	src.DiskReq = append(src.DiskReq, &clusterpb.DiskResource{
		DiskId: diskId1,
	})
	src.DiskReq = append(src.DiskReq, &clusterpb.DiskResource{
		DiskId: diskId2,
	})
	src.PortReq = append(src.PortReq, &clusterpb.PortResource{
		Ports: []int32{1000, 1001, 1002},
	})
	src.PortReq = append(src.PortReq, &clusterpb.PortResource{
		Ports: []int32{1003, 1004, 1005},
	})

	var dst dbpb.DBRecycleRequire
	buildDBRecycleRequire(&src, &dst)
	assert.Equal(t, holder_id, dst.HolderId)
	assert.Equal(t, host_id, dst.HostId)
	assert.Equal(t, int32(32), dst.ComputeReq.CpuCores)
	assert.Equal(t, int32(64), dst.ComputeReq.Memory)
	assert.Equal(t, 2, len(dst.DiskReq))
	assert.Equal(t, 2, len(dst.PortReq))
	assert.Equal(t, diskId1, dst.DiskReq[0].DiskId)
	assert.Equal(t, diskId2, dst.DiskReq[1].DiskId)
	assert.True(t, dst.PortReq[0].Ports[0] == 1000 && dst.PortReq[0].Ports[1] == 1001 && dst.PortReq[0].Ports[2] == 1002)
	assert.True(t, dst.PortReq[1].Ports[0] == 1003 && dst.PortReq[1].Ports[1] == 1004 && dst.PortReq[1].Ports[2] == 1005)
}
