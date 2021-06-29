package service

import (
	"context"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"

	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func (*DBServiceHandler) AddHost(ctx context.Context, req *dbPb.DBAddHostRequest, rsp *dbPb.DBAddHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AddHost"})
	log := logger.WithContext(ctx)
	var host models.Host
	host.Name = req.Host.HostName
	host.IP = req.Host.Ip
	host.Status = int32(req.Host.Status)
	host.OS = req.Host.Os
	host.Kernel = req.Host.Kernel
	host.CpuCores = int(req.Host.CpuCores)
	host.Memory = int(req.Host.Memory)
	host.Nic = req.Host.Nic
	host.AZ = req.Host.Az
	host.Rack = req.Host.Rack
	host.Purpose = req.Host.Purpose
	for _, v := range req.Host.Disks {
		host.Disks = append(host.Disks, models.Disk{
			Name:     v.Name,
			Path:     v.Path,
			Status:   int32(v.Status),
			Capacity: v.Capacity,
		})
	}
	hostId, err := models.CreateHost(&host)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		log.Fatalf("Failed to Import host %v to DB, err: %v", host.IP, err)
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}
	rsp.HostId = hostId
	rsp.Rs.Code = 0
	return nil
}
func (*DBServiceHandler) RemoveHost(ctx context.Context, req *dbPb.DBRemoveHostRequest, rsp *dbPb.DBRemoveHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHost"})
	log := logger.WithContext(ctx)
	hostId := req.HostId
	err := models.DeleteHost(hostId)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		log.Fatalf("Failed to Delete host %v from DB, err: %v", hostId, err)
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}
	rsp.Rs.Code = 0
	return nil
}

func CopyHostInfo(src *models.Host, dst *dbPb.DBHostInfoDTO) {
	dst.HostName = src.Name
	dst.Ip = src.IP
	dst.Os = src.OS
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Az = src.AZ
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

func (*DBServiceHandler) ListHost(ctx context.Context, req *dbPb.DBListHostsRequest, rsp *dbPb.DBListHostsResponse) error {
	// TODO: proto3 does not support `optional` by now
	hosts, _ := models.ListHosts()
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	for _, v := range hosts {
		var host dbPb.DBHostInfoDTO
		CopyHostInfo(&v, &host)
		rsp.HostList = append(rsp.HostList, &host)
	}
	return nil
}
func (*DBServiceHandler) CheckDetails(ctx context.Context, req *dbPb.DBCheckDetailsRequest, rsp *dbPb.DBCheckDetailsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckDetails"})
	log := logger.WithContext(ctx)
	host, err := models.FindHostById(req.HostId)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		log.Fatalf("Failed to Find host %v from DB, err: %v", req.HostId, err)
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}
	rsp.Details = new(dbPb.DBHostInfoDTO)
	CopyHostInfo(host, rsp.Details)
	return err
}
func (*DBServiceHandler) AllocHosts(ctx context.Context, req *dbPb.DBAllocHostsRequest, rsp *dbPb.DBAllocHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AllocHosts"})
	log := logger.WithContext(ctx)
	log.Infof("DB Service Receive Alloc Host Request, pd %d, tidb %d, tikv %d", req.PdCount, req.TidbCount, req.TikvCount)
	hosts, _ := models.AllocHosts()
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	for _, v := range hosts {
		var host dbPb.DBAllocHostDTO
		host.HostName = v.Name
		host.Ip = v.IP
		host.Disk = new(dbPb.DBDiskDTO)
		host.Disk.Name = v.Disks[0].Name
		host.Disk.Capacity = v.Disks[0].Capacity
		host.Disk.Path = v.Disks[0].Path
		host.Disk.Status = dbPb.DBDiskStatus(v.Disks[0].Status)
		rsp.Hosts = append(rsp.Hosts, &host)
	}
	return nil
}
