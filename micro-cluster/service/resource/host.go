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

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"

	"google.golang.org/grpc/codes"
)

type ResourceManager struct{}

func NewResourceManager() *ResourceManager {
	m := new(ResourceManager)
	return m
}

func copyAllocRequirement(src *clusterpb.AllocRequirement, dst *dbpb.DBAllocRequirement) {
	dst.Location = new(dbpb.DBLocation)
	dst.Location.Region = src.Location.Region
	dst.Location.Zone = src.Location.Zone
	dst.Location.Rack = src.Location.Rack
	dst.Location.Host = src.Location.Host

	dst.HostExcluded = new(dbpb.DBExcluded)
	if src.HostExcluded != nil {
		dst.HostExcluded.Hosts = append(dst.HostExcluded.Hosts, src.HostExcluded.Hosts...)
	}

	dst.HostFilter = new(dbpb.DBFilter)
	dst.HostFilter.Arch = src.HostFilter.Arch
	dst.HostFilter.Purpose = src.HostFilter.Purpose
	dst.HostFilter.DiskType = src.HostFilter.DiskType
	dst.HostFilter.HostTraits = src.HostFilter.HostTraits

	// copy requirement
	dst.Require = new(dbpb.DBRequirement)
	dst.Require.Exclusive = src.Require.Exclusive
	dst.Require.ComputeReq = new(dbpb.DBComputeRequirement)
	dst.Require.ComputeReq.CpuCores = src.Require.ComputeReq.CpuCores
	dst.Require.ComputeReq.Memory = src.Require.ComputeReq.Memory

	dst.Require.DiskReq = new(dbpb.DBDiskRequirement)
	dst.Require.DiskReq.NeedDisk = src.Require.DiskReq.NeedDisk
	dst.Require.DiskReq.DiskType = src.Require.DiskReq.DiskType
	dst.Require.DiskReq.Capacity = src.Require.DiskReq.Capacity

	for _, portReq := range src.Require.PortReq {
		dst.Require.PortReq = append(dst.Require.PortReq, &dbpb.DBPortRequirement{
			Start:   portReq.Start,
			End:     portReq.End,
			PortCnt: portReq.PortCnt,
		})
	}

	dst.Strategy = src.Strategy
	dst.Count = src.Count
}

func buildDBAllocRequest(src *clusterpb.BatchAllocRequest, dst *dbpb.DBBatchAllocRequest) {
	for _, request := range src.BatchRequests {
		var dbReq dbpb.DBAllocRequest
		dbReq.Applicant = new(dbpb.DBApplicant)
		dbReq.Applicant.HolderId = request.Applicant.HolderId
		dbReq.Applicant.RequestId = request.Applicant.RequestId
		for _, require := range request.Requires {
			var dbRequire dbpb.DBAllocRequirement
			copyAllocRequirement(require, &dbRequire)
			dbReq.Requires = append(dbReq.Requires, &dbRequire)
		}
		dst.BatchRequests = append(dst.BatchRequests, &dbReq)
	}
}

func copyResourcesInfoFromRsp(src *dbpb.DBHostResource, dst *clusterpb.HostResource) {
	dst.Reqseq = src.Reqseq
	dst.HostId = src.HostId
	dst.HostIp = src.HostIp
	dst.HostName = src.HostName
	dst.Passwd = src.Passwd
	dst.UserName = src.UserName
	dst.ComputeRes = new(clusterpb.ComputeRequirement)
	dst.ComputeRes.CpuCores = src.ComputeRes.CpuCores
	dst.ComputeRes.Memory = src.ComputeRes.Memory
	dst.Location = new(clusterpb.Location)
	dst.Location.Region = src.Location.Region
	dst.Location.Zone = src.Location.Zone
	dst.Location.Rack = src.Location.Rack
	dst.Location.Host = src.Location.Host
	dst.DiskRes = new(clusterpb.DiskResource)
	dst.DiskRes.DiskId = src.DiskRes.DiskId
	dst.DiskRes.DiskName = src.DiskRes.DiskName
	dst.DiskRes.Path = src.DiskRes.Path
	dst.DiskRes.Type = src.DiskRes.Type
	dst.DiskRes.Capacity = src.DiskRes.Capacity
	for _, portRes := range src.PortRes {
		var portResource clusterpb.PortResource
		portResource.Start = portRes.Start
		portResource.End = portRes.End
		portResource.Ports = append(portResource.Ports, portRes.Ports...)
		dst.PortRes = append(dst.PortRes, &portResource)
	}
}

func (m *ResourceManager) AllocResourcesInBatch(ctx context.Context, in *clusterpb.BatchAllocRequest, out *clusterpb.BatchAllocResponse) error {
	var req dbpb.DBBatchAllocRequest
	buildDBAllocRequest(in, &req)
	rsp, err := client.DBClient.AllocResourcesInBatch(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("alloc resources error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.AllocResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("alloc resources from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("alloc resources from db service succeed, holde id: %s, repest id: %s", in.BatchRequests[0].Applicant.HolderId, in.BatchRequests[0].Applicant.RequestId)
	for _, result := range rsp.BatchResults {
		var res clusterpb.AllocResponse
		res.Rs = new(clusterpb.AllocResponseStatus)
		res.Rs.Code = result.Rs.Code
		res.Rs.Message = result.Rs.Message
		for _, r := range result.Results {
			var hostResource clusterpb.HostResource
			copyResourcesInfoFromRsp(r, &hostResource)
			res.Results = append(res.Results, &hostResource)
		}
		out.BatchResults = append(out.BatchResults, &res)
	}
	return nil
}

func buildDBRecycleRequire(src *clusterpb.RecycleRequire, dst *dbpb.DBRecycleRequire) {
	dst.RecycleType = src.RecycleType
	dst.HolderId = src.HolderId
	dst.RequestId = src.RequestId
	dst.HostId = src.HostId
	dst.HostIp = src.HostIp
	dst.ComputeReq = new(dbpb.DBComputeRequirement)
	if src.ComputeReq != nil {
		dst.ComputeReq.CpuCores = src.ComputeReq.CpuCores
		dst.ComputeReq.Memory = src.ComputeReq.Memory
	}

	if src.DiskReq != nil {
		for _, req := range src.DiskReq {
			dst.DiskReq = append(dst.DiskReq, &dbpb.DBDiskResource{
				DiskId:   req.DiskId,
				DiskName: req.DiskName,
				Path:     req.Path,
				Type:     req.Type,
				Capacity: req.Capacity,
			})
		}
	}

	for _, portReq := range src.PortReq {
		var portResource dbpb.DBPortResource
		portResource.Start = portReq.Start
		portResource.End = portReq.End
		portResource.Ports = append(portResource.Ports, portReq.Ports...)
		dst.PortReq = append(dst.PortReq, &portResource)
	}

}

func (m *ResourceManager) RecycleResources(ctx context.Context, in *clusterpb.RecycleRequest, out *clusterpb.RecycleResponse) error {
	var req dbpb.DBRecycleRequest
	for _, request := range in.RecycleReqs {
		var r dbpb.DBRecycleRequire
		buildDBRecycleRequire(request, &r)
		req.RecycleReqs = append(req.RecycleReqs, &r)
	}
	rsp, err := client.DBClient.RecycleResources(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("recycle resources for type %d error, %v", err, in.RecycleReqs[0].RecycleType)
		return err
	}
	out.Rs = new(clusterpb.AllocResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("recycle resources from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("recycle resources from db service succeed, recycle type %d, holderId %s, requestId %s", in.RecycleReqs[0].RecycleType, in.RecycleReqs[0].HolderId, in.RecycleReqs[0].RequestId)

	return nil
}
