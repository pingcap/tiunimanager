package host

import (
	"context"
	"log"

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
		log.Fatal("Add Host Failed, err", err)
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
		log.Fatal("Remove Host Failed, err", err)
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
		log.Fatal("List Host Failed, err", err)
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
		log.Fatal("Check Host", req.HostId, "Details Failed, err", err)
		return err
	}
	CopyHostFromDBRsp(rsp.Details, out.Details)
	return err
}

func AllocHosts(ctx context.Context, in *hostPb.AllocHostsRequest, out *hostPb.AllocHostResponse) error {
	var req dbPb.DBAllocHostsRequest
	req.PdCount = in.PdCount
	req.TidbCount = in.TidbCount
	req.TikvCount = in.TikvCount
	rsp, err := dbClient.DBClient.AllocHosts(ctx, &req)
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if err != nil {
		log.Fatal("Alloc Hosts", req.PdCount, req.TidbCount, req.TikvCount, "Failed, err", err)
		return err
	}
	for _, v := range out.Hosts {
		var host hostPb.AllocHost
		host.HostName = v.HostName
		host.Ip = v.Ip
		host.Disk.Name = v.Disk.Name
		host.Disk.Capacity = v.Disk.Capacity
		host.Disk.Path = v.Disk.Path
		host.Disk.Status = hostPb.DiskStatus(v.Disk.Status)
		out.Hosts = append(out.Hosts, &host)
	}
	return err
}
