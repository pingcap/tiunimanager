package host

import (
	"context"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"

	hostPb "github.com/pingcap-inc/tiem/micro-cluster/proto"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"google.golang.org/grpc/codes"
)

func getLogger() *framework.LogRecord {
	return framework.GetLogger()
}

func CopyHostToDBReq(src *hostPb.HostInfo, dst *dbPb.DBHostInfoDTO) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.UserName = src.UserName
	dst.Passwd = src.Passwd
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Spec = src.Spec
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbPb.DBDiskDTO{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
		})
	}
}

func CopyHostFromDBRsp(src *dbPb.DBHostInfoDTO, dst *hostPb.HostInfo) {
	dst.HostId = src.HostId
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Spec = src.Spec
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Purpose = src.Purpose
	dst.CreateAt = src.CreateAt
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &hostPb.Disk{
			DiskId:   disk.DiskId,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
		})
	}
}

func ImportHost(ctx context.Context, in *hostPb.ImportHostRequest, out *hostPb.ImportHostResponse) error {
	var req dbPb.DBAddHostRequest
	req.Host = new(dbPb.DBHostInfoDTO)
	CopyHostToDBReq(in.Host, req.Host)
	var err error
	rsp, err := client.DBClient.AddHost(ctx, &req)
	if err != nil {
		getLogger().Errorf("import host %s error, %v", req.Host.Ip, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("import host failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	getLogger().Infof("import host(%s) succeed from db service: %s", in.Host.Ip, rsp.HostId)
	out.HostId = rsp.HostId

	return nil
}

func ImportHostsInBatch(ctx context.Context, in *hostPb.ImportHostsInBatchRequest, out *hostPb.ImportHostsInBatchResponse) error {
	var req dbPb.DBAddHostsInBatchRequest
	for _, v := range in.Hosts {
		var host dbPb.DBHostInfoDTO
		CopyHostToDBReq(v, &host)
		req.Hosts = append(req.Hosts, &host)
	}
	var err error
	rsp, err := client.DBClient.AddHostsInBatch(ctx, &req)
	if err != nil {
		getLogger().Errorf("import hosts in batch error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("import hosts in batch failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	getLogger().Infof("import %d hosts in batch succeed from db service.", len(rsp.HostIds))
	out.HostIds = rsp.HostIds

	return nil
}

func RemoveHost(ctx context.Context, in *hostPb.RemoveHostRequest, out *hostPb.RemoveHostResponse) error {
	var req dbPb.DBRemoveHostRequest
	req.HostId = in.HostId
	rsp, err := client.DBClient.RemoveHost(ctx, &req)
	if err != nil {
		getLogger().Errorf("remove host %s error, %v", req.HostId, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("remove host %s failed from db service: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	getLogger().Infof("remove host %s succeed from db service", req.HostId)
	return nil
}

func RemoveHostsInBatch(ctx context.Context, in *hostPb.RemoveHostsInBatchRequest, out *hostPb.RemoveHostsInBatchResponse) error {
	var req dbPb.DBRemoveHostsInBatchRequest
	req.HostIds = in.HostIds
	rsp, err := client.DBClient.RemoveHostsInBatch(ctx, &req)
	if err != nil {
		getLogger().Errorf("remove hosts in batch error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("remove hosts in batch failed from db service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	getLogger().Infof("remove %d hosts succeed from db service", len(req.HostIds))
	return nil
}

func ListHost(ctx context.Context, in *hostPb.ListHostsRequest, out *hostPb.ListHostsResponse) error {
	var req dbPb.DBListHostsRequest
	req.Purpose = in.Purpose
	req.Status = in.Status
	req.Page = new(dbPb.DBHostPageDTO)
	req.Page.Page = in.PageReq.Page
	req.Page.PageSize = in.PageReq.PageSize
	rsp, err := client.DBClient.ListHost(ctx, &req)
	if err != nil {
		getLogger().Errorf("list hosts error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.PageReq = new(hostPb.PageDTO)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("list hosts info from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	getLogger().Infof("list %d hosts info from db service succeed", len(rsp.HostList))
	for _, v := range rsp.HostList {
		var host hostPb.HostInfo
		CopyHostFromDBRsp(v, &host)
		out.HostList = append(out.HostList, &host)
	}
	out.PageReq.Page = rsp.Page.Page
	out.PageReq.PageSize = rsp.Page.PageSize
	out.PageReq.Total = rsp.Page.Total
	return nil
}

func CheckDetails(ctx context.Context, in *hostPb.CheckDetailsRequest, out *hostPb.CheckDetailsResponse) error {
	var req dbPb.DBCheckDetailsRequest
	req.HostId = in.HostId
	rsp, err := client.DBClient.CheckDetails(ctx, &req)
	if err != nil {
		getLogger().Errorf("check host %s details failed, %v", req.HostId, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("check host %s details from db service failed: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	getLogger().Infof("check host %s details from db service succeed", req.HostId)
	out.Details = new(hostPb.HostInfo)
	CopyHostFromDBRsp(rsp.Details, out.Details)

	return nil
}

/*
func getHostSpec(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}

func mergeReqs(zonesReqs map[string]map[string]*dbPb.DBPreAllocHostsRequest, reqs []*hostPb.AllocationReq) {
	for _, eachReq := range reqs {
		if eachReq.Count == 0 {
			continue
		}
		spec := getHostSpec(eachReq.CpuCores, eachReq.Memory)
		if r, ok := zonesReqs[eachReq.FailureDomain][spec]; ok {
			r.Req.Count += eachReq.Count
		} else {
			var req dbPb.DBPreAllocHostsRequest
			req.Req = new(dbPb.DBAllocationReq)
			req.Req.FailureDomain = eachReq.FailureDomain
			req.Req.CpuCores = eachReq.CpuCores
			req.Req.Memory = eachReq.Memory
			req.Req.Count = eachReq.Count
			if zonesReqs[eachReq.FailureDomain] == nil {
				zonesReqs[eachReq.FailureDomain] = make(map[string]*dbPb.DBPreAllocHostsRequest)
			}
			zonesReqs[eachReq.FailureDomain][spec] = &req
		}
	}
}

func fetchResults(zonesRsps map[string]map[string][]*dbPb.DBPreAllocation, reqs []*hostPb.AllocationReq) (res []*hostPb.AllocHost) {
	for _, eachReq := range reqs {
		if eachReq.Count == 0 {
			continue
		}
		spec := getHostSpec(eachReq.CpuCores, eachReq.Memory)
		zone := eachReq.FailureDomain
		for i := 0; i < int(eachReq.Count); i++ {
			rsp := zonesRsps[zone][spec][i]
			var host hostPb.AllocHost
			host.HostName = rsp.HostName
			host.Ip = rsp.Ip
			host.UserName = rsp.UserName
			//plain, err := crypto.AesDecryptCFB(rsp.Passwd)
			host.Passwd = rsp.Passwd
			host.CpuCores = rsp.RequestCores
			host.Memory = rsp.RequestMem
			host.Disk = new(hostPb.Disk)
			host.Disk.Name = rsp.DiskName
			host.Disk.Capacity = rsp.DiskCap
			host.Disk.Path = rsp.DiskPath
			host.Disk.Status = 0
			host.Disk.DiskId = rsp.DiskId
			res = append(res, &host)
		}
		zonesRsps[zone][spec] = zonesRsps[zone][spec][eachReq.Count:]
	}
	return
}

func AllocHosts(ctx context.Context, in *hostPb.AllocHostsRequest, out *hostPb.AllocHostResponse) error {
	var zonesReqs map[string]map[string]*dbPb.DBPreAllocHostsRequest = make(map[string]map[string]*dbPb.DBPreAllocHostsRequest)
	var zonesRsps map[string]map[string][]*dbPb.DBPreAllocation = make(map[string]map[string][]*dbPb.DBPreAllocation)

	mergeReqs(zonesReqs, in.PdReq)
	mergeReqs(zonesReqs, in.TidbReq)
	mergeReqs(zonesReqs, in.TikvReq)

	out.Rs = new(hostPb.ResponseStatus)
	var lockReq dbPb.DBLockHostsRequest

	const MaxRetries = 5
	retry := 0
	for {
		for zone, specReqs := range zonesReqs {
			for spec, req := range specReqs {
				rsp, err := client.DBClient.PreAllocHosts(ctx, req)
				// if PreAllocHosts failed, maybe no enough resources, no need to retry
				if err != nil {
					getLogger().Errorf("pre-alloc %d hosts with spec(%du%dg) in %s failed, err: %v",
						req.Req.Count, req.Req.CpuCores, req.Req.Memory, req.Req.FailureDomain, err)
					return err
				}
				out.Rs.Message = rsp.Rs.Message
				out.Rs.Code = rsp.Rs.Code
				if rsp.Rs.Code != int32(codes.OK) {
					getLogger().Errorf("pre-alloc %d hosts with spec(%du%dg) in %s failed: %d, %s",
						req.Req.Count, req.Req.CpuCores, req.Req.Memory, req.Req.FailureDomain, rsp.Rs.Code, rsp.Rs.Message)
					return nil
				}
				if zonesRsps[zone] == nil {
					zonesRsps[zone] = make(map[string][]*dbPb.DBPreAllocation)
				}
				zonesRsps[zone][spec] = append(zonesRsps[zone][spec], rsp.Results...)
				lockReq.Req = append(lockReq.Req, rsp.Results...)
			}
		}

		rsp, err := client.DBClient.LockHosts(ctx, &lockReq)
		if err != nil {
			getLogger().Warnf("lock pre-alloced resources failed in turn(%d), err: %v", retry, err)
			return err
		}
		out.Rs.Code = rsp.Rs.Code
		out.Rs.Message = rsp.Rs.Message
		// if lock resources failed, maybe some tx do allocation concurrently, retry...
		if rsp.Rs.Code != int32(codes.OK) {
			getLogger().Errorf("lock hosts failed for %d times: %d, %s", retry, rsp.Rs.Code, rsp.Rs.Message)
			retry++
			if retry < MaxRetries {
				zonesRsps = make(map[string]map[string][]*dbPb.DBPreAllocation)
				continue
			} else {
				getLogger().Errorln("lock pre-alloced resource retry too many times...")
				return nil
			}

		} else {
			break
		}
	}

	out.PdHosts = fetchResults(zonesRsps, in.PdReq)
	out.TidbHosts = fetchResults(zonesRsps, in.TidbReq)
	out.TikvHosts = fetchResults(zonesRsps, in.TikvReq)
	return nil
}
*/

func makeAllocReq(src []*hostPb.AllocationReq) (dst []*dbPb.DBAllocationReq) {
	for _, req := range src {
		dst = append(dst, &dbPb.DBAllocationReq{
			FailureDomain: req.FailureDomain,
			CpuCores:      req.CpuCores,
			Memory:        req.Memory,
			Count:         req.Count,
			Purpose:       req.Purpose,
		})
	}
	return dst
}

func buildDiskFromDB(disk *dbPb.DBDiskDTO) *hostPb.Disk {
	var hd hostPb.Disk
	hd.DiskId = disk.DiskId
	hd.Name = disk.Name
	hd.Path = disk.Path
	hd.Capacity = disk.Capacity
	hd.Status = disk.Status

	return &hd
}

func getAllocRsp(src []*dbPb.DBAllocHostDTO) (dst []*hostPb.AllocHost) {
	for _, rsp := range src {
		dst = append(dst, &hostPb.AllocHost{
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

func AllocHosts(ctx context.Context, in *hostPb.AllocHostsRequest, out *hostPb.AllocHostResponse) error {
	log := framework.GetLogger()
	req := new(dbPb.DBAllocHostsRequest)
	req.PdReq = makeAllocReq(in.PdReq)
	req.TidbReq = makeAllocReq(in.TidbReq)
	req.TikvReq = makeAllocReq(in.TikvReq)

	rsp, err := client.DBClient.AllocHosts(ctx, req)
	if err != nil {
		log.Errorf("alloc hosts error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("alloc hosts from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	out.PdHosts = getAllocRsp(rsp.PdHosts)
	out.TidbHosts = getAllocRsp(rsp.TidbHosts)
	out.TikvHosts = getAllocRsp(rsp.TikvHosts)
	return nil
}

func GetFailureDomain(ctx context.Context, in *hostPb.GetFailureDomainRequest, out *hostPb.GetFailureDomainResponse) error {
	var req dbPb.DBGetFailureDomainRequest
	req.FailureDomainType = in.FailureDomainType
	rsp, err := client.DBClient.GetFailureDomain(ctx, &req)
	if err != nil {
		getLogger().Errorf("get failure domains error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		getLogger().Warnf("get failure domains info from db service failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	getLogger().Infof("get failure domain type %d from db service succeed", req.FailureDomainType)
	for _, v := range rsp.FdList {
		out.FdList = append(out.FdList, &hostPb.FailureDomainResource{
			FailureDomain: v.FailureDomain,
			Purpose:       v.Purpose,
			Spec:          v.Spec,
			Count:         v.Count,
		})
	}
	return nil
}
