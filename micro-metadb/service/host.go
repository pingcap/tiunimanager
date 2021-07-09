package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"

	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func (*DBServiceHandler) AddHost(ctx context.Context, req *dbPb.DBAddHostRequest, rsp *dbPb.DBAddHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AddHost"})
	log := logger.WithContext(ctx)
	var host models.Host
	host.HostName = req.Host.HostName
	host.IP = req.Host.Ip
	host.Status = int32(req.Host.Status)
	host.OS = req.Host.Os
	host.Kernel = req.Host.Kernel
	host.CpuCores = int(req.Host.CpuCores)
	host.Memory = int(req.Host.Memory)
	host.Nic = req.Host.Nic
	host.DC = req.Host.Dc
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
	dst.HostName = src.HostName
	dst.Ip = src.IP
	dst.Os = src.OS
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Dc = src.DC
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

func (*DBServiceHandler) PreAllocHosts(ctx context.Context, req *dbPb.DBPreAllocHostsRequest, rsp *dbPb.DBPreAllocHostsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckDetails"})
	log := logger.WithContext(ctx)
	log.Infof("DB Service Receive Alloc Host in %s for %d x (%dU%dG),", req.Req.FailureDomain, req.Req.Count, req.Req.CpuCores, req.Req.Memory)
	resources, err := models.PreAllocHosts(req.Req.FailureDomain, int(req.Req.Count), int(req.Req.CpuCores), int(req.Req.Memory))
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}
	if len(resources) < int(req.Req.Count) {
		rsp.Rs.Code = 1
		errMsg := fmt.Sprintf("No Enough host resources(%d/%d) in %s", len(resources), req.Req.Count, req.Req.FailureDomain)
		rsp.Rs.Message = errMsg
		return errors.New(errMsg)
	}
	for _, v := range resources {
		rsp.Results = append(rsp.Results, &dbPb.DBPreAllocation{
			FailureDomain: req.Req.FailureDomain,
			HostId:        v.HostId,
			DiskId:        v.Id,
			OriginCores:   int32(v.CpuCores),
			OriginMem:     int32(v.Memory),
			RequestCores:  req.Req.CpuCores,
			RequestMem:    req.Req.Memory,
			HostName:      v.HostName,
			Ip:            v.Ip,
			DiskName:      v.Name,
			DiskPath:      v.Path,
			DiskCap:       int32(v.Capacity),
		})
	}
	return nil
}

func (*DBServiceHandler) LockHosts(ctx context.Context, req *dbPb.DBLockHostsRequest, rsp *dbPb.DBLockHostsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckDetails"})
	log := logger.WithContext(ctx)
	var resources []models.ResourceLock
	for _, v := range req.Req {
		resources = append(resources, models.ResourceLock{
			HostId:       v.HostId,
			DiskId:       v.DiskId,
			OriginCores:  int(v.OriginCores),
			OriginMem:    int(v.OriginMem),
			RequestCores: int(v.RequestCores),
			RequestMem:   int(v.RequestMem),
		})
	}
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	err := models.LockHosts(resources)
	if err != nil {
		errMsg := fmt.Sprintf("LockHosts Failed, err: %v", err)
		log.Fatalln(errMsg)
		rsp.Rs.Code = 1
		rsp.Rs.Message = errMsg
		return err
	}
	return nil
}
