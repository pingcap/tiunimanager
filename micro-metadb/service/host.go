package service

import (
	"context"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"

	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func (*DBServiceHandler) InitHostManager(ctx context.Context, req *dbPb.InitHostManagerRequest, rsp *dbPb.InitHostManagerResponse) error {
	var builtCnt int32
	var err error
	if builtCnt, err = models.CreateHostTable(); err != nil {
		rsp.Rs.Code = 1
		rsp.Rs.Msg = err.Error()
	}
	rsp.BuiltCnt = builtCnt
	return err
}

func (*DBServiceHandler) AddHost(ctx context.Context, req *dbPb.AddHostRequest, rsp *dbPb.AddHostResponse) error {
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
	if err != nil {
		log.Fatalf("Failed to Import host %v to DB, err: %v", host.IP, err)
		rsp.Rs.Code = 1
		rsp.Rs.Msg = err.Error()
		return err
	}
	rsp.HostId = hostId
	rsp.Rs.Code = 0
	return nil
}
func (*DBServiceHandler) RemoveHost(ctx context.Context, req *dbPb.RemoveHostRequest, rsp *dbPb.RemoveHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHost"})
	log := logger.WithContext(ctx)
	hostId := req.HostId
	err := models.DeleteHost(hostId)
	if err != nil {
		log.Fatalf("Failed to Delete host %v from DB, err: %v", hostId, err)
		rsp.Rs.Code = 1
		rsp.Rs.Msg = err.Error()
		return err
	}
	rsp.Rs.Code = 0
	return nil
}

func CopyHostInfo(src *models.Host, dst *dbPb.HostInfo) {
	dst.HostName = src.Name
	dst.Ip = src.IP
	dst.Os = src.OS
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Az = src.AZ
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

func (*DBServiceHandler) ListHost(ctx context.Context, req *dbPb.ListHostsRequest, rsp *dbPb.ListHostsResponse) error {
	// TODO: proto3 does not support `optional` by now
	hosts, _ := models.ListHosts()
	for _, v := range hosts {
		var host dbPb.HostInfo
		CopyHostInfo(&v, &host)
		rsp.HostList = append(rsp.HostList, &host)
	}
	return nil
}
func (*DBServiceHandler) CheckDetails(ctx context.Context, req *dbPb.CheckDetailsRequest, rsp *dbPb.CheckDetailsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckDetails"})
	log := logger.WithContext(ctx)
	host, err := models.FindHostById(req.HostId)
	if err != nil {
		log.Fatalf("Failed to Find host %v from DB, err: %v", req.HostId, err)
		rsp.Rs.Code = 1
		rsp.Rs.Msg = err.Error()
		return err
	}
	CopyHostInfo(host, rsp.Details)
	return err
}
func (*DBServiceHandler) AllocHosts(ctx context.Context, req *dbPb.AllocHostsRequest, rsp *dbPb.AllocHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AllocHosts"})
	log := logger.WithContext(ctx)
	log.Infof("DB Service Receive Alloc Host Request, pd %d, tidb %d, tikv %d", req.PdCount, req.TidbCount, req.TikvCount)
	hosts, _ := models.AllocHosts()
	for _, v := range hosts {
		var host dbPb.AllocHost
		host.HostName = v.Name
		host.Ip = v.IP
		host.Disk.Name = v.Disks[0].Name
		host.Disk.Capacity = v.Disks[0].Capacity
		host.Disk.Path = v.Disks[0].Path
		host.Disk.Status = dbPb.DiskStatus(v.Disks[0].Status)
		rsp.Hosts = append(rsp.Hosts, &host)
	}
	return nil
}
