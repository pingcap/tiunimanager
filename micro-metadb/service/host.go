package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"

	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
	resource "github.com/pingcap/ticp/micro-metadb/service/host"
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
	dst.HostName = src.Name
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

func (*DBServiceHandler) AllocHosts(ctx context.Context, req *dbPb.DBAllocHostsRequest, rsp *dbPb.DBAllocHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AllocHosts"})
	log := logger.WithContext(ctx)
	log.Infof("DB Service Receive Alloc Host Request, pd %d, tidb %d, tikv %d", req.PdCount, req.TidbCount, req.TikvCount)
	rsp.Rs = new(dbPb.DBHostResponseStatus)

	// Build the resources hierarchy Tree according to the hosts table in database
	root, err := resource.GetHierarchyRoot()
	if err != nil {
		log.Fatalln("GetHierarchyRoot Failed, err:", err)
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}
	resource.PrintHierarchy(root, "|----", false)

	// chosen_path is used to record the path from root to chosen disk in hierarchy tree, then adjust the resource weight
	chosen_path := make([]*resource.Item, 0, resource.DISK)

	// Choose nums of resources in HOST failure domain
	hosts, disks, err := resource.ChooseFirstn(root, req.PdCount, resource.HOST, true, chosen_path)
	if err != nil {
		log.Fatalf("ChooseFirstn Failed for choose %d pd, err: %v\n", req.PdCount, err)
		rsp.Rs.Code = 1
		rsp.Rs.Message = err.Error()
		return err
	}

	// TODO: Change the status of chosen disks and hosts' status to 'inused'
	// models.ChangeStatus()

	// Buildup AllocHostsResponse and Return
	for k, v := range hosts {
		var host dbPb.DBAllocHostDTO
		hostExtra, ok := v.GetExtra().(*resource.HostExtraInfo)
		if !ok {
			errMsg := fmt.Sprintf("Get Host Item(%s) Extra Info Failed", v.GetName())
			log.Fatalln(errMsg)
			rsp.Rs.Code = 1
			rsp.Rs.Message = errMsg
			return errors.New(errMsg)
		}
		host.HostName = v.GetName()
		host.Ip = hostExtra.Ip
		host.Disk = new(dbPb.DBDiskDTO)
		host.Disk.Name = disks[k].GetName()
		diskExtra, ok := disks[k].GetExtra().(*resource.DiskExtraInfo)
		if !ok {
			errMsg := fmt.Sprintf("Get Disk Item(%s) Extra Info Failed", disks[k].GetName())
			log.Fatalln(errMsg)
			rsp.Rs.Code = 1
			rsp.Rs.Message = errMsg
			return errors.New(errMsg)
		}
		host.Disk.Capacity = diskExtra.Capacity
		host.Disk.Path = diskExtra.Path
		host.Disk.Status = dbPb.DBDiskStatus(disks[k].GetStatus())
		rsp.Hosts = append(rsp.Hosts, &host)
	}
	return nil
}
