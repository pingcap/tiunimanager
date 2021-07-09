package host

import (
	"context"
	"fmt"

	"github.com/asim/go-micro/v3/logger"
	hostPb "github.com/pingcap/ticp/micro-manager/proto"
	dbClient "github.com/pingcap/ticp/micro-metadb/client"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func CopyHostToDBReq(src *hostPb.HostInfo, dst *dbPb.DBHostInfoDTO) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = dbPb.DBHostStatus(src.Status)
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbPb.DBDiskDTO{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   dbPb.DBDiskStatus(disk.Status),
		})
	}
}

func CopyHostFromDBRsp(src *dbPb.DBHostInfoDTO, dst *hostPb.HostInfo) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = hostPb.HostStatus(src.Status)
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &hostPb.Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   hostPb.DiskStatus(disk.Status),
		})
	}
}

func ImportHost(ctx context.Context, in *hostPb.ImportHostRequest, out *hostPb.ImportHostResponse) error {
	var req dbPb.DBAddHostRequest
	req.Host = new(dbPb.DBHostInfoDTO)
	CopyHostToDBReq(in.Host, req.Host)
	var err error
	rsp, err := dbClient.DBClient.AddHost(ctx, &req)
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if err != nil {
		logger.Fatalf("Add Host Failed, err: %v", err)
		return err
	}
	out.HostId = rsp.HostId
	return err
}

func RemoveHost(ctx context.Context, in *hostPb.RemoveHostRequest, out *hostPb.RemoveHostResponse) error {
	var req dbPb.DBRemoveHostRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.RemoveHost(ctx, &req)
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if err != nil {
		logger.Fatalf("Remove Host Failed, err: %v", err)
		return err
	}
	return err
}

func ListHost(ctx context.Context, in *hostPb.ListHostsRequest, out *hostPb.ListHostsResponse) error {
	var req dbPb.DBListHostsRequest
	req.Purpose = in.Purpose
	req.Status = dbPb.DBHostStatus(in.Status)
	rsp, err := dbClient.DBClient.ListHost(ctx, &req)
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if err != nil {
		logger.Fatalf("List Host Failed, err: %v", err)
		return err
	}
	for _, v := range rsp.HostList {
		var host hostPb.HostInfo
		CopyHostFromDBRsp(v, &host)
		out.HostList = append(out.HostList, &host)
	}
	return err
}

func CheckDetails(ctx context.Context, in *hostPb.CheckDetailsRequest, out *hostPb.CheckDetailsResponse) error {
	var req dbPb.DBCheckDetailsRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.CheckDetails(ctx, &req)
	out.Details = new(hostPb.HostInfo)
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if err != nil {
		logger.Fatal("Check Host", req.HostId, "Details Failed, err", err)
		return err
	}
	CopyHostFromDBRsp(rsp.Details, out.Details)
	return err
}

func getHostSpec(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dU%dG", cpuCores, mem)
}

func mergeReqs(zonesReqs map[string]map[string]*dbPb.DBPreAllocHostsRequest, reqs []*hostPb.AllocationReq) {
	for _, eachReq := range reqs {
		spec := getHostSpec(eachReq.CpuCores, eachReq.Memory)
		if r, ok := zonesReqs[eachReq.FailureDomain][spec]; ok {
			r.Req.Count += eachReq.Count
		} else {
			var req dbPb.DBPreAllocHostsRequest
			req.Req.FailureDomain = eachReq.FailureDomain
			req.Req.CpuCores = eachReq.CpuCores
			req.Req.Memory = eachReq.Memory
			req.Req.Count = eachReq.Count
			zonesReqs[eachReq.FailureDomain][spec] = &req
		}
	}
}

func fetchResults(zonesRsps map[string]map[string][]*dbPb.DBPreAllocation, reqs []*hostPb.AllocationReq) (res []*hostPb.AllocHost) {
	for _, eachReq := range reqs {
		spec := getHostSpec(eachReq.CpuCores, eachReq.Memory)
		zone := eachReq.FailureDomain
		for i := 0; i < int(eachReq.Count); i++ {
			rsp := zonesRsps[zone][spec][i]
			var host hostPb.AllocHost
			host.HostName = rsp.HostName
			host.Ip = rsp.Ip
			host.Disk = new(hostPb.Disk)
			host.Disk.Name = rsp.DiskName
			host.Disk.Capacity = rsp.DiskCap
			host.Disk.Path = rsp.DiskPath
			host.Disk.Status = 0
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
	i := 0
	for {
		for zone, specReqs := range zonesReqs {
			for spec, req := range specReqs {
				rsp, err := dbClient.DBClient.PreAllocHosts(ctx, req)
				// if PreAllocHosts failed, maybe no enough resources, no need to retry
				if err != nil {
					logger.Fatalf("PreAlloc %d Hosts with spec(%dU%dG) in %s failed, err: %v",
						req.Req.Count, req.Req.CpuCores, req.Req.Memory, req.Req.FailureDomain, err)
					out.Rs.Message = rsp.Rs.Message
					out.Rs.Code = rsp.Rs.Code
					return err
				}
				zonesRsps[zone][spec] = append(zonesRsps[zone][spec], rsp.Results...)
				lockReq.Req = append(lockReq.Req, rsp.Results...)
			}
		}

		rsp, err := dbClient.DBClient.LockHosts(ctx, &lockReq)
		// if lock resources failed, maybe some tx do allocation concurrently, retry...
		if err != nil {
			logger.Warnf("Lock PreAlloced Resources Failed in turn(%d), err: %v", i, err)
			i++
			if i < MaxRetries {
				zonesRsps = make(map[string]map[string][]*dbPb.DBPreAllocation)
				continue
			} else {
				logger.Fatal("Lock PreAlloced Resource Retry too many times...")
				out.Rs.Code = rsp.Rs.Code
				out.Rs.Message = rsp.Rs.Message
				return err
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
