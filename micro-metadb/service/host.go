package service

import (
	"context"
	"fmt"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

func copyHostInfoFromReq(src *dbPb.DBHostInfoDTO, dst *models.Host) {
	dst.HostName = src.HostName
	dst.IP = src.Ip
	dst.OS = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int(src.CpuCores)
	dst.Memory = int(src.Memory)
	dst.Nic = src.Nic
	dst.DC = src.Dc
	dst.AZ = src.Az
	dst.Rack = src.Rack
	dst.Status = int32(src.Status)
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, models.Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Status:   int32(disk.Status),
			Capacity: disk.Capacity,
		})
	}
}

func (*DBServiceHandler) AddHost(ctx context.Context, req *dbPb.DBAddHostRequest, rsp *dbPb.DBAddHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AddHost"})
	log := logger.WithContext(ctx)
	var host models.Host
	copyHostInfoFromReq(req.Host, &host)

	hostId, err := models.CreateHost(&host)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to Import Host(%s) %s, %v", host.HostName, host.IP, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.HostId = hostId
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (*DBServiceHandler) AddHostsInBatch(ctx context.Context, req *dbPb.DBAddHostsInBatchRequest, rsp *dbPb.DBAddHostsInBatchResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "AddHostInBatch"})
	log := logger.WithContext(ctx)
	var hosts []*models.Host
	for _, v := range req.Hosts {
		var host models.Host
		copyHostInfoFromReq(v, &host)
		hosts = append(hosts, &host)
	}
	hostIds, err := models.CreateHostsInBatch(hosts)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to Import Hosts, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.HostIds = hostIds
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (*DBServiceHandler) RemoveHost(ctx context.Context, req *dbPb.DBRemoveHostRequest, rsp *dbPb.DBRemoveHostResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHost"})
	log := logger.WithContext(ctx)
	hostId := req.HostId
	err := models.DeleteHost(hostId)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to Delete HostId(%s), %v", hostId, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (*DBServiceHandler) RemoveHostsInBatch(ctx context.Context, req *dbPb.DBRemoveHostsInBatchRequest, rsp *dbPb.DBRemoveHostsInBatchResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "RemoveHostInBatch"})
	log := logger.WithContext(ctx)

	err := models.DeleteHostsInBatch(req.HostIds)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to Delete HostId In Batch, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func copyHostInfoToRsp(src *models.Host, dst *dbPb.DBHostInfoDTO) {
	dst.HostId = src.ID
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
	dst.Status = src.Status
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, &dbPb.DBDiskDTO{
			DiskId:   disk.ID,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
		})
	}
}

func (*DBServiceHandler) ListHost(ctx context.Context, req *dbPb.DBListHostsRequest, rsp *dbPb.DBListHostsResponse) error {
	// TODO: proto3 does not support `optional` by now
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "ListHost"})
	log := logger.WithContext(ctx)
	hosts, err := models.ListHosts()
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to List Hosts, %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	for _, v := range hosts {
		var host dbPb.DBHostInfoDTO
		copyHostInfoToRsp(&v, &host)
		rsp.HostList = append(rsp.HostList, &host)
	}
	rsp.Rs.Code = int32(codes.OK)
	return nil
}
func (*DBServiceHandler) CheckDetails(ctx context.Context, req *dbPb.DBCheckDetailsRequest, rsp *dbPb.DBCheckDetailsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckDetails"})
	log := logger.WithContext(ctx)
	host, err := models.FindHostById(req.HostId)
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("Failed to List Hosts %s, %v", req.HostId, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}

	rsp.Details = new(dbPb.DBHostInfoDTO)
	copyHostInfoToRsp(host, rsp.Details)
	rsp.Rs.Code = int32(codes.OK)
	return nil
}

func (*DBServiceHandler) PreAllocHosts(ctx context.Context, req *dbPb.DBPreAllocHostsRequest, rsp *dbPb.DBPreAllocHostsResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "PreAllocHost"})
	log := logger.WithContext(ctx)
	log.Infof("DB Service Receive Alloc Host in %s for %d x (%dU%dG)", req.Req.FailureDomain, req.Req.Count, req.Req.CpuCores, req.Req.Memory)
	resources, err := models.PreAllocHosts(req.Req.FailureDomain, int(req.Req.Count), int(req.Req.CpuCores), int(req.Req.Memory))
	rsp.Rs = new(dbPb.DBHostResponseStatus)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("DB Service Receive Alloc Host in %s for %d x (%dU%dG) Error, err: %v",
				req.Req.FailureDomain, req.Req.Count, req.Req.CpuCores, req.Req.Memory, err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}

	if len(resources) < int(req.Req.Count) {
		errMsg := fmt.Sprintf("No Enough host resources(%d/%d) in %s", len(resources), req.Req.Count, req.Req.FailureDomain)
		log.Errorln(errMsg)
		rsp.Rs.Code = int32(codes.ResourceExhausted)
		rsp.Rs.Message = errMsg
		return nil
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
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "LockHost"})
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
		st, ok := status.FromError(err)
		if ok {
			rsp.Rs.Code = int32(st.Code())
			rsp.Rs.Message = st.Message()
		} else {
			rsp.Rs.Code = int32(codes.Internal)
			rsp.Rs.Message = fmt.Sprintf("LockHosts Failed, err: %v", err)
		}
		log.Warnln(rsp.Rs.Message)

		// return nil to use rsp
		return nil
	}
	return nil
}
