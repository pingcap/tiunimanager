package host

import (
	"context"
	"log"

	"github.com/pingcap/ticp/addon/logger"
	hostPb "github.com/pingcap/ticp/micro-manager/proto"
	dbClient "github.com/pingcap/ticp/micro-metadb/client"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
	"google.golang.org/grpc/codes"
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
	dst.HostId = src.HostId
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
			DiskId:   disk.DiskId,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   hostPb.DiskStatus(disk.Status),
		})
	}
}

func ImportHost(ctx context.Context, in *hostPb.ImportHostRequest, out *hostPb.ImportHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "ImportHost"})
	log := logger.WithContext(ctx)
	var req dbPb.DBAddHostRequest
	req.Host = new(dbPb.DBHostInfoDTO)
	CopyHostToDBReq(in.Host, req.Host)
	var err error
	rsp, err := dbClient.DBClient.AddHost(ctx, &req)
	if err != nil {
		log.Errorf("Add Host %s error, %v", req.Host.Ip, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Import Host Failed from DB Service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	log.Infof("Import Host(%s) Succeed from DB Service: %s", in.Host.Ip, rsp.HostId)
	out.HostId = rsp.HostId

	return nil
}

func ImportHostsInBatch(ctx context.Context, in *hostPb.ImportHostsInBatchRequest, out *hostPb.ImportHostsInBatchResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "ImportHostInBatch"})
	log := logger.WithContext(ctx)
	var req dbPb.DBAddHostsInBatchRequest
	for _, v := range in.Hosts {
		var host dbPb.DBHostInfoDTO
		CopyHostToDBReq(v, &host)
		req.Hosts = append(req.Hosts, &host)
	}
	var err error
	rsp, err := dbClient.DBClient.AddHostsInBatch(ctx, &req)
	if err != nil {
		log.Errorf("Add Host In Batch error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Import Host Failed from DB Service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}
	log.Infof("Import %d Hosts In Batch Succeed from DB Service.", len(rsp.HostIds))
	out.HostIds = rsp.HostIds

	return nil
}

func RemoveHost(ctx context.Context, in *hostPb.RemoveHostRequest, out *hostPb.RemoveHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHost"})
	log := logger.WithContext(ctx)
	var req dbPb.DBRemoveHostRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.RemoveHost(ctx, &req)
	if err != nil {
		log.Errorf("Remove Host %s error, %v", req.HostId, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Remove Host %s Failed from DB Service: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	log.Infof("Remove Host %s succeed from DB Service", req.HostId)
	return nil
}

func RemoveHostsInBatch(ctx context.Context, in *hostPb.RemoveHostsInBatchRequest, out *hostPb.RemoveHostsInBatchResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHostInBatch"})
	log := logger.WithContext(ctx)
	var req dbPb.DBRemoveHostsInBatchRequest
	req.HostIds = in.HostIds
	rsp, err := dbClient.DBClient.RemoveHostsInBatch(ctx, &req)
	if err != nil {
		log.Errorf("Remove Hosts In Batch error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message
	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Remove Hosts In Batch Failed from DB Service: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	log.Infof("Remove %d Hosts succeed from DB Service", len(req.HostIds))
	return nil
}

func ListHost(ctx context.Context, in *hostPb.ListHostsRequest, out *hostPb.ListHostsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "ListHost"})
	log := logger.WithContext(ctx)
	var req dbPb.DBListHostsRequest
	req.Purpose = in.Purpose
	req.Status = dbPb.DBHostStatus(in.Status)
	rsp, err := dbClient.DBClient.ListHost(ctx, &req)
	if err != nil {
		log.Errorf("List Host error, %v", err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Get Hosts Info from DB Service Failed: %d, %s", rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	log.Infof("Get %d Hosts Info from DB Service Succeed", len(rsp.HostList))
	for _, v := range rsp.HostList {
		var host hostPb.HostInfo
		CopyHostFromDBRsp(v, &host)
		out.HostList = append(out.HostList, &host)
	}
	return nil
}

func CheckDetails(ctx context.Context, in *hostPb.CheckDetailsRequest, out *hostPb.CheckDetailsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "ListHost"})
	log := logger.WithContext(ctx)
	var req dbPb.DBCheckDetailsRequest
	req.HostId = in.HostId
	rsp, err := dbClient.DBClient.CheckDetails(ctx, &req)
	if err != nil {
		log.Errorf("Check Host %s Details Failed, %v", req.HostId, err)
		return err
	}
	out.Rs = new(hostPb.ResponseStatus)
	out.Rs.Code = rsp.Rs.Code
	out.Rs.Message = rsp.Rs.Message

	if rsp.Rs.Code != int32(codes.OK) {
		log.Warnf("Get Host %s Info from DB Service Failed: %d, %s", req.HostId, rsp.Rs.Code, rsp.Rs.Message)
		return nil
	}

	log.Infof("Get Host %s Info from DB Service Succeed", req.HostId)
	out.Details = new(hostPb.HostInfo)
	CopyHostFromDBRsp(rsp.Details, out.Details)

	return nil
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
	for _, v := range rsp.Hosts {
		var host hostPb.AllocHost
		host.HostName = v.HostName
		host.Ip = v.Ip
		host.Disk = new(hostPb.Disk)
		host.Disk.Name = v.Disk.Name
		host.Disk.Capacity = v.Disk.Capacity
		host.Disk.Path = v.Disk.Path
		host.Disk.Status = hostPb.DiskStatus(v.Disk.Status)
		out.Hosts = append(out.Hosts, &host)
	}
	return err
}
