
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
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"strings"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func genDomainCodeByName(pre string, name string) string {
	return fmt.Sprintf("%s,%s", pre, name)
}

func GetDomainNameFromCode(failureDomain string) string {
	pos := strings.LastIndex(failureDomain, ",")
	return failureDomain[pos+1:]
}

func copyHostInfoFromReq(src *dbpb.DBHostInfoDTO, dst *resource.Host) {
	dst.HostName = src.HostName
	dst.IP = src.Ip
	dst.UserName = src.UserName
	dst.Passwd = src.Passwd
	dst.Arch = src.Arch
	dst.OS = src.Os
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.AZ = genDomainCodeByName(dst.Region, src.Az)
	dst.Rack = genDomainCodeByName(dst.AZ, src.Rack)
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.Reserved = src.Reserved
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, resource.Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Status:   int32(disk.Status),
			Capacity: disk.Capacity,
			Type:     disk.Type,
		})
	}
}

func (handler *DBServiceHandler) AddHost(ctx context.Context, req *dbpb.DBAddHostRequest, rsp *dbpb.DBAddHostResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	var host resource.Host
	copyHostInfoFromReq(req.Host, &host)

	hostId, err := resourceManager.CreateHost(&host)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to import host(%s) %s, %v", host.HostName, host.IP, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.HostId = hostId
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (handler *DBServiceHandler) AddHostsInBatch(ctx context.Context, req *dbpb.DBAddHostsInBatchRequest, rsp *dbpb.DBAddHostsInBatchResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	var hosts []*resource.Host
	for _, v := range req.Hosts {
		var host resource.Host
		copyHostInfoFromReq(v, &host)
		hosts = append(hosts, &host)
	}
	hostIds, err := resourceManager.CreateHostsInBatch(hosts)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to import hosts, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.HostIds = hostIds
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (handler *DBServiceHandler) RemoveHost(ctx context.Context, req *dbpb.DBRemoveHostRequest, rsp *dbpb.DBRemoveHostResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	hostId := req.HostId
	err := resourceManager.DeleteHost(hostId)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to delete host(%s), %v", hostId, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (handler *DBServiceHandler) RemoveHostsInBatch(ctx context.Context, req *dbpb.DBRemoveHostsInBatchRequest, rsp *dbpb.DBRemoveHostsInBatchResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	err := resourceManager.DeleteHostsInBatch(req.HostIds)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to delete host in batch, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func copyHostInfoToRsp(src *resource.Host, dst *dbpb.DBHostInfoDTO) {
	dst.HostId = src.ID
	dst.HostName = src.HostName
	dst.Ip = src.IP
	dst.Arch = src.Arch
	dst.Os = src.OS
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.Az = GetDomainNameFromCode(src.AZ)
	dst.Rack = GetDomainNameFromCode(src.Rack)
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.Reserved = src.Reserved
	dst.CreateAt = src.CreatedAt.Unix()
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbpb.DBDiskDTO{
			DiskId:   disk.ID,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
			Type:     disk.Type,
		})
	}
}

func (handler *DBServiceHandler) ListHost(ctx context.Context, req *dbpb.DBListHostsRequest, rsp *dbpb.DBListHostsResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	var hostReq models.ListHostReq
	hostReq.Purpose = req.Purpose
	hostReq.Status = resource.HostStatus(req.Status)
	hostReq.Limit = int(req.Page.PageSize)
	if req.Page.Page >= 1 {
		hostReq.Offset = (int(req.Page.Page) - 1) * int(req.Page.PageSize)
	} else {
		hostReq.Offset = 0
	}
	hosts, err := resourceManager.ListHosts(hostReq)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to list hosts, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	for _, v := range hosts {
		var host dbpb.DBHostInfoDTO
		copyHostInfoToRsp(&v, &host)
		rsp.HostList = append(rsp.HostList, &host)
	}
	rsp.Page = new(dbpb.DBHostPageDTO)
	rsp.Page.Page = req.Page.Page
	rsp.Page.PageSize = req.Page.PageSize
	rsp.Page.Total = int32(len(hosts))
	rsp.Rs.Code = int32(codes.OK)
	return nil
}
func (handler *DBServiceHandler) CheckDetails(ctx context.Context, req *dbpb.DBCheckDetailsRequest, rsp *dbpb.DBCheckDetailsResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	host, err := resourceManager.FindHostById(req.HostId)
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("failed to list hosts %s, %v", req.HostId, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}

	rsp.Details = new(dbpb.DBHostInfoDTO)
	copyHostInfoToRsp(host, rsp.Details)
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func copyAllocReq(component string, req models.AllocReqs, in []*dbpb.DBAllocationReq) {
	for _, eachReq := range in {
		if eachReq.Count == 0 {
			continue
		}
		req[component] = append(req[component], &models.HostAllocReq{
			FailureDomain: eachReq.FailureDomain,
			CpuCores:      eachReq.CpuCores,
			Memory:        eachReq.Memory,
			Count:         int(eachReq.Count),
		})
	}
}

func buildAllocRsp(componet string, req models.AllocRsps, out *[]*dbpb.DBAllocHostDTO) {
	for _, result := range req[componet] {
		*out = append(*out, &dbpb.DBAllocHostDTO{
			HostName: result.HostName,
			Ip:       result.Ip,
			UserName: result.UserName,
			Passwd:   result.Passwd,
			CpuCores: int32(result.CpuCores),
			Memory:   int32(result.Memory),
			Disk: &dbpb.DBDiskDTO{
				DiskId:   result.DiskId,
				Name:     result.DiskName,
				Path:     result.Path,
				Capacity: int32(result.Capacity),
				Status:   int32(resource.DISK_AVAILABLE),
			},
		})
	}
}

func (handler *DBServiceHandler) AllocHosts(ctx context.Context, in *dbpb.DBAllocHostsRequest, out *dbpb.DBAllocHostsResponse) error {
	resourceManager := handler.Dao().ResourceManager()
	log := framework.Log()
	// Build up allocHosts request for model
	req := make(models.AllocReqs)
	copyAllocReq("PD", req, in.PdReq)
	copyAllocReq("TiDB", req, in.TidbReq)
	copyAllocReq("TiKV", req, in.TikvReq)

	resources, err := resourceManager.AllocHosts(req)
	out.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("db service receive alloc hosts error, %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}

	buildAllocRsp("PD", resources, &out.PdHosts)
	buildAllocRsp("TiDB", resources, &out.TidbHosts)
	buildAllocRsp("TiKV", resources, &out.TikvHosts)

	out.Rs.Code = int32(codes.OK)
	return nil
}

func getFailureDomainByType(fd resource.FailureDomain) (domain string, err error) {
	switch fd {
	case resource.REGION:
		domain = "region"
	case resource.ZONE:
		domain = "az"
	case resource.RACK:
		domain = "rack"
	default:
		err = status.Errorf(codes.InvalidArgument, "%d is invalid domain type(1-Region, 2-Zone, 3-Rack)", fd)
	}
	return
}

func (handler *DBServiceHandler) GetFailureDomain(ctx context.Context, req *dbpb.DBGetFailureDomainRequest, rsp *dbpb.DBGetFailureDomainResponse) error {

	log := framework.Log()
	domainType := req.FailureDomainType
	domain, err := getFailureDomainByType(resource.FailureDomain(domainType))
	rsp.Rs = new(dbpb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("get failure domain resources failed, err: %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}

	resourceManager := handler.Dao().ResourceManager()
	resources, err := resourceManager.GetFailureDomain(domain)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("get failure domain resources failed, err: %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	for _, v := range resources {
		rsp.FdList = append(rsp.FdList, &dbpb.DBFailureDomainResource{
			FailureDomain: v.FailureDomain,
			Purpose:       v.Purpose,
			Spec:          knowledge.GenSpecCode(int32(v.CpuCores), int32(v.Memory)),
			Count:         int32(v.Count),
		})
	}
	return nil
}

func copyResultToRsp(src *resource.HostResource, dst *dbpb.DBHostResource) {
	dst.Reqseq = src.Reqseq
	dst.Location = new(dbpb.DBLocation)
	dst.Location.Region = src.Location.Region
	dst.Location.Zone = src.Location.Zone
	dst.Location.Rack = src.Location.Rack
	dst.Location.Host = src.Location.Host
	dst.HostId = src.HostId
	dst.HostName = src.HostName
	dst.HostIp = src.HostIp
	dst.UserName = src.UserName
	dst.Passwd = src.Passwd
	dst.ComputeRes = new(dbpb.DBComputeRequirement)
	dst.ComputeRes.CpuCores = src.ComputeRes.CpuCores
	dst.ComputeRes.Memory = src.ComputeRes.Memory
	dst.DiskRes = new(dbpb.DBDiskResource)
	dst.DiskRes.DiskId = src.DiskRes.DiskId
	dst.DiskRes.DiskName = src.DiskRes.DiskName
	dst.DiskRes.Path = src.DiskRes.Path
	dst.DiskRes.Capacity = src.DiskRes.Capacity
	dst.DiskRes.Type = src.DiskRes.Type
	for _, portRes := range src.PortRes {
		dst.PortRes = append(dst.PortRes, &dbpb.DBPortResource{
			Start: portRes.Start,
			End:   portRes.End,
			Ports: portRes.Ports,
		})
	}
}

func (handler *DBServiceHandler) AllocResources(ctx context.Context, in *dbpb.DBAllocRequest, out *dbpb.DBAllocResponse) error {
	log := framework.Log()
	log.Infof("Receive %d allocation requirement from %s in requestID %s\n", len(in.Requires), in.Applicant.HolderId, in.Applicant.RequestId)
	resourceManager := handler.Dao().ResourceManager()
	resources, err := resourceManager.AllocResources(in)
	out.Rs = new(dbpb.DBAllocResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("alloc resources failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	for _, r := range resources.Results {
		var hostResource dbpb.DBHostResource
		copyResultToRsp(&r, &hostResource)
		out.Results = append(out.Results, &hostResource)
	}
	out.Rs.Code = common.TIEM_SUCCESS
	return nil
}

func (handler *DBServiceHandler) AllocResourcesInBatch(ctx context.Context, in *dbpb.DBBatchAllocRequest, out *dbpb.DBBatchAllocResponse) error {
	log := framework.Log()
	log.Infof("Receive batch allocation with %d requests", len(in.BatchRequests))
	out.Rs = new(dbpb.DBAllocResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	resources, err := resourceManager.AllocResourcesInBatch(in)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("alloc resources in batch failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	for _, result := range resources.BatchResults {
		var rsp dbpb.DBAllocResponse
		rsp.Rs = new(dbpb.DBAllocResponseStatus)
		rsp.Rs.Code = common.TIEM_SUCCESS
		for _, r := range result.Results {
			var hostResource dbpb.DBHostResource
			copyResultToRsp(&r, &hostResource)
			rsp.Results = append(rsp.Results, &hostResource)
		}
		out.BatchResults = append(out.BatchResults, &rsp)
	}
	out.Rs.Code = common.TIEM_SUCCESS
	return nil
}

func (handler *DBServiceHandler) RecycleResources(ctx context.Context, in *dbpb.DBRecycleRequest, out *dbpb.DBRecycleResponse) error {
	log := framework.Log()
	log.Infof("Receive recycle with %d requires", len(in.RecycleReqs))
	out.Rs = new(dbpb.DBAllocResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	err := resourceManager.RecycleAllocResources(in)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("recycle resources failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	out.Rs.Code = common.TIEM_SUCCESS
	return nil
}
