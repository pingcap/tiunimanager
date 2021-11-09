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
	"strings"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

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

func getDomainPrefixFromCode(failureDomain string) string {
	pos := strings.LastIndex(failureDomain, ",")
	if pos == -1 {
		// No found ","
		return failureDomain
	}
	return failureDomain[:pos]
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
	log := framework.LogWithContext(ctx)
	var host resource.Host
	copyHostInfoFromReq(req.Host, &host)

	hostId, err := resourceManager.CreateHost(ctx, &host)
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
	log := framework.LogWithContext(ctx)
	var hosts []*resource.Host
	for _, v := range req.Hosts {
		var host resource.Host
		copyHostInfoFromReq(v, &host)
		hosts = append(hosts, &host)
	}
	hostIds, err := resourceManager.CreateHostsInBatch(ctx, hosts)
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
	log := framework.LogWithContext(ctx)
	hostId := req.HostId
	err := resourceManager.DeleteHost(ctx, hostId)
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
	log := framework.LogWithContext(ctx)
	err := resourceManager.DeleteHostsInBatch(ctx, req.HostIds)
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
	dst.UpdateAt = src.UpdatedAt.Unix()
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
	log := framework.LogWithContext(ctx)
	var hostReq models.ListHostReq
	hostReq.Purpose = req.Purpose
	hostReq.Stat = resource.HostStat(req.Stat)
	hostReq.Status = resource.HostStatus(req.Status)
	hostReq.Limit = int(req.Page.PageSize)
	if req.Page.Page >= 1 {
		hostReq.Offset = (int(req.Page.Page) - 1) * int(req.Page.PageSize)
	} else {
		hostReq.Offset = 0
	}
	hosts, err := resourceManager.ListHosts(ctx, hostReq)
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
	log := framework.LogWithContext(ctx)
	host, err := resourceManager.FindHostById(ctx, req.HostId)
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
	log := framework.LogWithContext(ctx)
	// Build up allocHosts request for model
	req := make(models.AllocReqs)
	copyAllocReq("PD", req, in.PdReq)
	copyAllocReq("TiDB", req, in.TidbReq)
	copyAllocReq("TiKV", req, in.TikvReq)

	resources, err := resourceManager.AllocHosts(ctx, req)
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

	log := framework.LogWithContext(ctx)
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
	resources, err := resourceManager.GetFailureDomain(ctx, domain)
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
	log := framework.LogWithContext(ctx)
	log.Infof("Receive %d allocation requirement from %s in requestID %s\n", len(in.Requires), in.Applicant.HolderId, in.Applicant.RequestId)
	resourceManager := handler.Dao().ResourceManager()
	resources, err := resourceManager.AllocResources(ctx, in)
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
	out.Rs.Code = int32(common.TIEM_SUCCESS)
	return nil
}

func (handler *DBServiceHandler) AllocResourcesInBatch(ctx context.Context, in *dbpb.DBBatchAllocRequest, out *dbpb.DBBatchAllocResponse) error {
	log := framework.LogWithContext(ctx)
	log.Infof("Receive batch allocation with %d requests", len(in.BatchRequests))
	out.Rs = new(dbpb.DBAllocResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	resources, err := resourceManager.AllocResourcesInBatch(ctx, in)
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
		rsp.Rs.Code = int32(common.TIEM_SUCCESS)
		for _, r := range result.Results {
			var hostResource dbpb.DBHostResource
			copyResultToRsp(&r, &hostResource)
			rsp.Results = append(rsp.Results, &hostResource)
		}
		out.BatchResults = append(out.BatchResults, &rsp)
	}
	out.Rs.Code = int32(common.TIEM_SUCCESS)
	return nil
}

func (handler *DBServiceHandler) RecycleResources(ctx context.Context, in *dbpb.DBRecycleRequest, out *dbpb.DBRecycleResponse) error {
	log := framework.LogWithContext(ctx)
	log.Infof("Receive recycle with %d requires", len(in.RecycleReqs))
	out.Rs = new(dbpb.DBAllocResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	err := resourceManager.RecycleAllocResources(ctx, in)
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
	out.Rs.Code = int32(common.TIEM_SUCCESS)
	return nil
}

func (handler *DBServiceHandler) UpdateHostStatus(ctx context.Context, in *dbpb.DBUpdateHostStatusRequest, out *dbpb.DBUpdateHostStatusResponse) error {
	log := framework.LogWithContext(ctx)
	log.Infof("update host %v status to %d", in.HostIds, in.Status)
	out.Rs = new(dbpb.DBHostResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	err := resourceManager.UpdateHostStatus(ctx, in)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("update host status failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	out.Rs.Code = common.TIEM_SUCCESS
	return nil
}

func (handler *DBServiceHandler) ReserveHost(ctx context.Context, in *dbpb.DBReserveHostRequest, out *dbpb.DBReserveHostResponse) error {
	log := framework.LogWithContext(ctx)
	log.Infof("set host %v reserved status to %v", in.HostIds, in.Reserved)
	out.Rs = new(dbpb.DBHostResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	err := resourceManager.ReserveHost(ctx, in)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("update host status failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	out.Rs.Code = common.TIEM_SUCCESS
	return nil
}

type node struct {
	Code     string
	Prefix   string
	Name     string
	subNodes []*node
}

func addSubNode(current map[string]*node, code string, subNode *node) (parent *node) {
	if parent, ok := current[code]; ok {
		parent.subNodes = append(parent.subNodes, subNode)
		return nil
	} else {
		parent := node{
			Code:   code,
			Prefix: getDomainPrefixFromCode(code),
			Name:   GetDomainNameFromCode(code),
		}
		parent.subNodes = append(parent.subNodes, subNode)
		current[code] = &parent
		return &parent
	}
}

func (handler *DBServiceHandler) buildHierarchy(Items []models.Item) *node {
	if len(Items) == 0 {
		return nil
	}
	root := node{
		Code: "root",
	}
	var regions map[string]*node = make(map[string]*node)
	var zones map[string]*node = make(map[string]*node)
	var racks map[string]*node = make(map[string]*node)
	for _, item := range Items {
		hostItem := node{
			Code:     genDomainCodeByName(item.Ip, item.Name),
			Prefix:   item.Ip,
			Name:     item.Name,
			subNodes: nil,
		}
		newRack := addSubNode(racks, item.Rack, &hostItem)
		if newRack == nil {
			continue
		}
		newZone := addSubNode(zones, item.Az, newRack)
		if newZone == nil {
			continue
		}
		newRegion := addSubNode(regions, item.Region, newZone)
		if newRegion != nil {
			root.subNodes = append(root.subNodes, newRegion)
		}
	}
	return &root
}

func (handler *DBServiceHandler) trimTree(root *node, level resource.FailureDomain, depth int) *node {
	newRoot := node{
		Code: "root",
	}
	levelNodes := root.subNodes

	for l := resource.REGION; l < level; l++ {
		var subnodes []*node
		for _, node := range levelNodes {
			subnodes = append(subnodes, node.subNodes...)
		}
		levelNodes = subnodes
	}
	newRoot.subNodes = levelNodes

	leafNodes := levelNodes
	for d := 0; d < depth; d++ {
		var subnodes []*node
		for _, node := range leafNodes {
			subnodes = append(subnodes, node.subNodes...)
		}
		leafNodes = subnodes
	}
	for _, node := range leafNodes {
		node.subNodes = nil
	}

	return &newRoot
}

func copyHierarchyToRsp(root *node, dst **dbpb.DBNode) {
	*dst = &dbpb.DBNode{
		Code:   root.Code,
		Name:   root.Name,
		Prefix: root.Prefix,
	}
	(*dst).SubNodes = make([]*dbpb.DBNode, len(root.subNodes))
	for i, subNode := range root.subNodes {
		copyHierarchyToRsp(subNode, &((*dst).SubNodes[i]))
	}
}

func (handler *DBServiceHandler) GetHierarchy(ctx context.Context, in *dbpb.DBGetHierarchyRequest, out *dbpb.DBGetHierarchyResponse) error {
	log := framework.LogWithContext(ctx)
	filter := resource.Filter{
		Arch:     in.Filter.Arch,
		Purpose:  in.Filter.Purpose,
		DiskType: in.Filter.DiskType,
	}
	log.Infof("Receive GetHierarchy Request for arch = %s, purpose = %s, disktype = %s\n", filter.Arch, filter.Purpose, filter.DiskType)
	out.Rs = new(dbpb.DBHostResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	Items, err := resourceManager.GetHostItems(ctx, filter, in.Level, in.Depth)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("get host items failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	wholeTree := handler.buildHierarchy(Items)
	if wholeTree == nil {
		out.Rs.Code = common.TIEM_RESOURCE_NO_STOCK
		out.Rs.Message = fmt.Sprintf("no stocks with filter:%v", filter)
		return nil
	}
	root := handler.trimTree(wholeTree, resource.FailureDomain(in.Level), int(in.Depth))
	copyHierarchyToRsp(root, &out.Root)

	out.Rs.Code = common.TIEM_SUCCESS

	return nil
}

func (handler *DBServiceHandler) buildStockCondition(cond *models.StockCondition, in *dbpb.DBGetStocksRequest) {
	if in.Location != nil {
		cond.Location.Region = in.Location.Region
		cond.Location.Zone = in.Location.Zone
		cond.Location.Rack = in.Location.Rack
		cond.Location.Host = in.Location.Host
	}
	if in.HostFilter != nil {
		if in.HostFilter.Status != int32(resource.HOST_WHATEVER) {
			cond.HostCondition.Status = &in.HostFilter.Status
		}
		if in.HostFilter.Stat != int32(resource.HOST_STAT_WHATEVER) {
			cond.HostCondition.Stat = &in.HostFilter.Stat
		}
		if in.HostFilter.Arch != "" {
			cond.HostCondition.Arch = &in.HostFilter.Arch
		}
	}
	if in.DiskFilter != nil {
		if in.DiskFilter.Status != int32(resource.DISK_STATUS_WHATEVER) {
			cond.DiskCondition.Status = &in.DiskFilter.Status
		}
		if in.DiskFilter.Capacity != 0 {
			cond.DiskCondition.Capacity = &in.DiskFilter.Capacity
		}
		if in.DiskFilter.Type != "" {
			cond.DiskCondition.Type = &in.DiskFilter.Type
		}
	}
}

func (handler *DBServiceHandler) GetStocks(ctx context.Context, in *dbpb.DBGetStocksRequest, out *dbpb.DBGetStocksResponse) error {
	log := framework.LogWithContext(ctx)
	var cond models.StockCondition
	handler.buildStockCondition(&cond, in)
	log.Infof("Receive GetStocks Request for cond = %v\n", cond)
	out.Rs = new(dbpb.DBHostResponseStatus)

	resourceManager := handler.Dao().ResourceManager()
	stocks, err := resourceManager.GetStocks(ctx, cond)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			out.Rs.Code = int32(st.Code())
			out.Rs.Message = st.Message()
		} else {
			out.Rs.Code = int32(codes.Internal)
			out.Rs.Message = fmt.Sprintf("get host items failed, err: %v", err)
		}
		log.Warnln(out.Rs.Message)

		// return nil to use rsp
		return nil
	}
	out.Stocks = new(dbpb.DBStocks)
	out.Stocks.FreeHostCount = int32(len(stocks))
	for _, stock := range stocks {
		out.Stocks.FreeCpuCores += int32(stock.FreeCpuCores)
		out.Stocks.FreeMemory += int32(stock.FreeMemory)
		out.Stocks.FreeDiskCount += int32(stock.FreeDiskCount)
		out.Stocks.FreeDiskCapacity += int32(stock.FreeDiskCapacity)
	}
	out.Rs.Code = common.TIEM_SUCCESS

	return nil
}
