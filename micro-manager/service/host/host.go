package host

import (
	"context"
	"log"

	hostPb "github.com/pingcap/ticp/micro-manager/proto"
	dbClient "github.com/pingcap/ticp/micro-metadb/client"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func CopyHostToDBReq(src *hostPb.HostInfo, dst *dbPb.HostInfo) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = dbPb.HostStatus(src.Status)
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbPb.Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   dbPb.DiskStatus(disk.Status),
		})
	}
}

func CopyHostFromDBRsp(src *dbPb.HostInfo, dst *hostPb.HostInfo) {
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
	var req dbPb.AddHostRequest
	CopyHostToDBReq(in.Host, req.Host)
	var err error
	rsp, err := dbClient.DBClient.AddHost(ctx, &req)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Msg = rsp.Rs.Msg
	if err != nil {
		log.Fatal("Add Host Failed, err", err)
		return err
	}
	out.HostId = rsp.HostId
	return err
}

func RemoveHost(ctx context.Context, in *hostPb.RemoveHostRequest, out *hostPb.RemoveHostResponse) error {
	var req dbPb.RemoveHostRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.RemoveHost(ctx, &req)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Msg = rsp.Rs.Msg
	if err != nil {
		log.Fatal("Remove Host Failed, err", err)
		return err
	}
	return err
}

func ListHost(ctx context.Context, in *hostPb.ListHostsRequest, out *hostPb.ListHostsResponse) error {
	var req dbPb.ListHostsRequest
	req.Purpose = in.Purpose
	req.Status = dbPb.HostStatus(in.Status)
	rsp, err := dbClient.DBClient.ListHost(ctx, &req)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Msg = rsp.Rs.Msg
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
	var req dbPb.CheckDetailsRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.CheckDetails(ctx, &req)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Msg = rsp.Rs.Msg
	if err != nil {
		log.Fatal("Check Host", req.HostId, "Details Failed, err", err)
		return err
	}
	return err
}

func AllocHosts(ctx context.Context, in *hostPb.AllocHostsRequest, out *hostPb.AllocHostResponse) error {
	var req dbPb.AllocHostsRequest
	req.PdCount = in.PdCount
	req.TidbCount = in.TidbCount
	req.TikvCount = in.TikvCount
	rsp, err := dbClient.DBClient.AllocHosts(ctx, &req)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Msg = rsp.Rs.Msg
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
