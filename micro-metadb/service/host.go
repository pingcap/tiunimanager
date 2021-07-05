package service

import (
	"context"
	"fmt"
	"log"
	"strings"

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

type FailureDomain uint32

const (
	ROOT FailureDomain = iota
	DATACENTER
	ZONE
	RACK
	HOST
	DISK
)
const (
	MAX_TRIES     = 5
	GB_PER_WEIGHT = 256
)

type Item struct {
	id                string
	name              string
	failureDomainType FailureDomain
	status            int32
	weight            uint32
	subItems          []*Item
}

func isAvailable(item *Item) bool {
	return item.status == 0
}

func BuildHierarchy() (rack2hosts map[string][]*Item, zone2racks map[string][]string, dc2zones map[string][]string, err error) {
	rack2hosts = make(map[string][]*Item)
	zone2racks = make(map[string][]string)
	dc2zones = make(map[string][]string)
	// For deduplicated
	tmp_rack_recorded := make(map[string]bool)
	tmp_zone_recorded := make(map[string]bool)

	hosts, _ := models.ListHosts()
	for _, host := range hosts {
		hostItem := Item{
			id:                host.ID,
			name:              host.Name,
			failureDomainType: HOST,
			status:            host.Status,
		}
		for _, disk := range host.Disks {
			diskItem := Item{
				id:                disk.ID,
				name:              disk.Name,
				status:            disk.Status,
				failureDomainType: DISK,
				weight:            uint32(disk.Capacity) / GB_PER_WEIGHT,
			}
			if isAvailable(&hostItem) && isAvailable(&diskItem) {
				hostItem.weight += diskItem.weight
			}
			hostItem.subItems = append(hostItem.subItems, &diskItem)
		}
		rack2hosts[host.Rack] = append(rack2hosts[host.Rack], &hostItem)

		if _, ok := tmp_rack_recorded[host.Rack]; !ok {
			tmp_rack_recorded[host.Rack] = true
			zone2racks[host.AZ] = append(zone2racks[host.AZ], host.Rack)
			//fmt.Println(host.Rack, "has been put into zone2racks under key", host.AZ)
		}
		if _, ok := tmp_zone_recorded[host.AZ]; !ok {
			tmp_zone_recorded[host.AZ] = true
			dc2zones[host.DC] = append(dc2zones[host.DC], host.AZ)
			//fmt.Println(host.AZ, "has been put into dc2zones under key", host.DC)
		}
	}
	return
}

func GetHierarchyRoot() (root *Item, err error) {
	rack2hosts, zone2racks, dc2zones, err := BuildHierarchy()
	if err != nil {
		log.Fatal("BuildHierarchy failed, err:", err)
		return nil, err
	}
	root = new(Item)
	root.failureDomainType = ROOT
	root.name = "ROOT"
	root.status = 0
	for dc_name, zones := range dc2zones {
		dcItem := Item{
			name:              dc_name,
			failureDomainType: DATACENTER,
			status:            0,
		}
		for _, az := range zones {
			zoneItem := Item{
				name:              az,
				failureDomainType: ZONE,
				status:            0,
			}
			racks := zone2racks[az]
			for _, rack := range racks {
				rackItem := Item{
					name:              rack,
					failureDomainType: RACK,
					status:            0,
					subItems:          rack2hosts[rack],
				}
				for _, host := range rackItem.subItems {
					if isAvailable(&rackItem) && isAvailable(host) {
						rackItem.weight += host.weight
					}
				}
				zoneItem.subItems = append(zoneItem.subItems, &rackItem)
				if isAvailable(&zoneItem) && isAvailable(&rackItem) {
					zoneItem.weight += rackItem.weight
				}
			}
			dcItem.subItems = append(dcItem.subItems, &zoneItem)
			if isAvailable(&dcItem) && isAvailable(&zoneItem) {
				dcItem.weight += zoneItem.weight
			}
		}
		root.subItems = append(root.subItems, &dcItem)
		if isAvailable(root) && isAvailable(&dcItem) {
			root.weight += dcItem.weight
		}
	}
	return
}

// PrintHierarchy(root, "|----", false)
func PrintHierarchy(root *Item, pre string, has_brother bool) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "PrintHierarchy"})
	log := logger.WithContext(ctx)
	if root == nil {
		return
	}
	log.Debugln(pre, "Item name:", root.name, "Status:", root.status, "Id:", root.id, "Weight:", root.weight)
	if root.subItems != nil {
		for i, item := range root.subItems {
			var str string
			index := strings.LastIndex(pre, "|")
			if has_brother {
				str = fmt.Sprintf("%s|\t%s", pre[:index], pre[index:])
			} else {
				str = fmt.Sprintf("%s\t%s", pre[:index], pre[index:])
			}

			if i != len(root.subItems)-1 {
				PrintHierarchy(item, str, true)
			} else {
				PrintHierarchy(item, str, false)
			}
		}
	}
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
