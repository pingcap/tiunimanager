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

func copyHostToDBReq(src *clusterpb.HostInfo, dst *dbpb.DBHostInfoDTO) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.UserName = src.UserName
	dst.Passwd = src.Passwd
	dst.Arch = src.Arch
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.Reserved = src.Reserved
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbpb.DBDiskDTO{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
			Type:     disk.Type,
		})
	}
}

func copyHostFromDBRsp(src *dbpb.DBHostInfoDTO, dst *clusterpb.HostInfo) {
	dst.HostId = src.HostId
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Arch = src.Arch
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.CreateAt = src.CreateAt
	dst.Reserved = src.Reserved
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &clusterpb.Disk{
			DiskId:   disk.DiskId,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
			Type:     disk.Type,
			UsedBy:   disk.UsedBy,
		})
	}
}

func (m *ResourceManager) ImportHost(ctx context.Context, in *clusterpb.ImportHostRequest, out *clusterpb.ImportHostResponse) error {
	var req dbpb.DBAddHostRequest
	req.Host = new(dbpb.DBHostInfoDTO)
	copyHostToDBReq(in.Host, req.Host)
	var err error
	rsp, err := client.DBClient.AddHost(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("import host %s error, %v", req.Host.Ip, err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("import host failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	framework.LogWithContext(ctx).Infof("import host(%s) succeed from db service: %s", in.Host.Ip, rsp.HostId)
	out.HostId = rsp.HostId

	return nil
}

func (m *ResourceManager) ImportHostsInBatch(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, out *clusterpb.ImportHostsInBatchResponse) error {
	var req dbpb.DBAddHostsInBatchRequest
	for _, v := range in.Hosts {
		var host dbpb.DBHostInfoDTO
		copyHostToDBReq(v, &host)
		req.Hosts = append(req.Hosts, &host)
	}
	var err error
	rsp, err := client.DBClient.AddHostsInBatch(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("import hosts in batch error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("import hosts in batch failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	framework.LogWithContext(ctx).Infof("import %d hosts in batch succeed from db service.", len(rsp.HostIds))
	out.HostIds = rsp.HostIds

	return nil
}

func (m *ResourceManager) RemoveHost(ctx context.Context, in *clusterpb.RemoveHostRequest, out *clusterpb.RemoveHostResponse) error {
	var req dbpb.DBRemoveHostRequest
	req.HostId = in.HostId
	rsp, err := client.DBClient.RemoveHost(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("remove host %s error, %v", req.HostId, err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("remove host %s failed from db service: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("remove host %s succeed from db service", req.HostId)
	return nil
}

func (m *ResourceManager) RemoveHostsInBatch(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, out *clusterpb.RemoveHostsInBatchResponse) error {
	var req dbpb.DBRemoveHostsInBatchRequest
	req.HostIds = in.HostIds
	rsp, err := client.DBClient.RemoveHostsInBatch(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("remove hosts in batch error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("remove hosts in batch failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("remove %d hosts succeed from db service", len(req.HostIds))
	return nil
}

func (m *ResourceManager) ListHost(ctx context.Context, in *clusterpb.ListHostsRequest, out *clusterpb.ListHostsResponse) error {
	var req dbpb.DBListHostsRequest
	req.Purpose = in.Purpose
	req.Status = in.Status
	req.Stat = in.Stat
	req.Page = new(dbpb.DBHostPageDTO)
	req.Page.Page = in.PageReq.Page
	req.Page.PageSize = in.PageReq.PageSize
	rsp, err := client.DBClient.ListHost(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("list hosts error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.PageReq = new(clusterpb.PageDTO)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("list hosts info from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("list %d hosts info from db service succeed", len(rsp.HostList))
	for _, v := range rsp.HostList {
		var host clusterpb.HostInfo
		copyHostFromDBRsp(v, &host)
		out.HostList = append(out.HostList, &host)
	}
	out.PageReq.Page = rsp.Page.Page
	out.PageReq.PageSize = rsp.Page.PageSize
	out.PageReq.Total = rsp.Page.Total
	return nil
}

func (m *ResourceManager) CheckDetails(ctx context.Context, in *clusterpb.CheckDetailsRequest, out *clusterpb.CheckDetailsResponse) error {
	var req dbpb.DBCheckDetailsRequest
	req.HostId = in.HostId
	rsp, err := client.DBClient.CheckDetails(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check host %s details failed, %v", req.HostId, err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("check host %s details from db service failed: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("check host %s details from db service succeed", req.HostId)
	out.Details = new(clusterpb.HostInfo)
	copyHostFromDBRsp(rsp.Details, out.Details)

	return nil
}

func makeAllocReq(src []*clusterpb.AllocationReq) (dst []*dbpb.DBAllocationReq) {
	for _, req := range src {
		dst = append(dst, &dbpb.DBAllocationReq{
			FailureDomain: req.FailureDomain,
			CpuCores:      req.CpuCores,
			Memory:        req.Memory,
			Count:         req.Count,
			Purpose:       req.Purpose,
		})
	}
	return dst
}

func buildDiskFromDB(disk *dbpb.DBDiskDTO) *clusterpb.Disk {
	var hd clusterpb.Disk
	hd.DiskId = disk.DiskId
	hd.Name = disk.Name
	hd.Path = disk.Path
	hd.Capacity = disk.Capacity
	hd.Status = disk.Status

	return &hd
}

func getAllocRsp(src []*dbpb.DBAllocHostDTO) (dst []*clusterpb.AllocHost) {
	for _, rsp := range src {
		dst = append(dst, &clusterpb.AllocHost{
			HostName: rsp.HostName,
			Ip:       rsp.Ip,
			UserName: rsp.UserName,
			Passwd:   rsp.Passwd,
			CpuCores: rsp.CpuCores,
			Memory:   rsp.Memory,
			Disk:     buildDiskFromDB(rsp.Disk),
		})
	}
	return dst
}

func (m *ResourceManager) AllocHosts(ctx context.Context, in *clusterpb.AllocHostsRequest, out *clusterpb.AllocHostResponse) error {
	req := new(dbpb.DBAllocHostsRequest)
	req.PdReq = makeAllocReq(in.PdReq)
	req.TidbReq = makeAllocReq(in.TidbReq)
	req.TikvReq = makeAllocReq(in.TikvReq)

	rsp, err := client.DBClient.AllocHosts(ctx, req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("alloc hosts error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("alloc hosts from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	out.PdHosts = getAllocRsp(rsp.PdHosts)
	out.TidbHosts = getAllocRsp(rsp.TidbHosts)
	out.TikvHosts = getAllocRsp(rsp.TikvHosts)
	return nil
}

func (m *ResourceManager) GetFailureDomain(ctx context.Context, in *clusterpb.GetFailureDomainRequest, out *clusterpb.GetFailureDomainResponse) error {
	var req dbpb.DBGetFailureDomainRequest
	req.FailureDomainType = in.FailureDomainType
	rsp, err := client.DBClient.GetFailureDomain(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get failure domains error, %v", err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("get failure domains info from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("get failure domain type %d from db service succeed", req.FailureDomainType)
	for _, v := range rsp.FdList {
		out.FdList = append(out.FdList, &clusterpb.FailureDomainResource{
			FailureDomain: v.FailureDomain,
			Purpose:       v.Purpose,
			Spec:          v.Spec,
			Count:         v.Count,
		})
	}
	return nil
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

	dst.DiskReq = new(dbpb.DBDiskResource)
	if src.DiskReq != nil {
		dst.DiskReq.DiskId = src.DiskReq.DiskId
		dst.DiskReq.DiskName = src.DiskReq.DiskName
		dst.DiskReq.Path = src.DiskReq.Path
		dst.DiskReq.Type = src.DiskReq.Type
		dst.DiskReq.Capacity = src.DiskReq.Capacity
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

func (m *ResourceManager) UpdateHostStatus(ctx context.Context, in *clusterpb.UpdateHostStatusRequest, out *clusterpb.UpdateHostStatusResponse) error {
	var req dbpb.DBUpdateHostStatusRequest
	req.Status = in.Status
	req.HostIds = append(req.HostIds, in.HostIds...)

	rsp, err := client.DBClient.UpdateHostStatus(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update host status to %d for host[%v] error, %v", req.Status, req.HostIds, err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("update host status to %d in batch failed from db service: %d, %s", req.Status, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("update host status to  %d succeed for hosts %v from db service", req.Status, req.HostIds)
	return nil
}

func (m *ResourceManager) ReserveHost(ctx context.Context, in *clusterpb.ReserveHostRequest, out *clusterpb.ReserveHostResponse) error {
	var req dbpb.DBReserveHostRequest
	req.Reserved = in.Reserved
	req.HostIds = append(req.HostIds, in.HostIds...)

	rsp, err := client.DBClient.ReserveHost(ctx, &req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("set reserve to %v for host[%v] error, %v", req.Reserved, req.HostIds, err)
		return err
	}
	out.Rs = new(clusterpb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		framework.LogWithContext(ctx).Warnf("update host reserved to %v in batch failed from db service: %d, %s", req.Reserved, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	framework.LogWithContext(ctx).Infof("update host reserved to %v succeed for hosts %v from db service", req.Reserved, req.HostIds)
	return nil
}
