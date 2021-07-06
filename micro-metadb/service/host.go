package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
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

func (fd FailureDomain) GetStr() (str string) {
	switch fd {
	case ROOT:
		str = "Root"
	case DATACENTER:
		str = "DataCenter"
	case ZONE:
		str = "Zone"
	case RACK:
		str = "Rack"
	case HOST:
		str = "Host"
	case DISK:
		str = "Disk"
	default:
		str = "Unknown"
	}
	return str
}

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

func calcDiskWeight(c int32) uint32 {
	return uint32(c) / GB_PER_WEIGHT
}

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
			}
			if isAvailable(&diskItem) {
				diskItem.weight = calcDiskWeight(disk.Capacity)
			}
			if isAvailable(&hostItem) {
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

func doHash(item *Item, n int) uint32 {
	a := fnv.New32()
	str := fmt.Sprintf("%d,%s,%s", n, item.id, item.name)
	a.Write([]byte(str))
	return a.Sum32() & 0x00ffffff
}

func bucketChoose(bucket *Item, trial int) (item *Item) {
	var high_item *Item
	var high_draw uint32
	const Magic = 47
	for index, item := range bucket.subItems {
		draw := doHash(item, trial*Magic)
		draw *= item.weight
		if index == 0 || draw > high_draw {
			high_item = item
			high_draw = draw
		}
	}
	return high_item

}

func hasChosen(items []*Item, item *Item) bool {
	for _, eachItem := range items {
		if item.id == eachItem.id && item.name == eachItem.name {
			return true
		}
	}
	return false
}

func adjustWeight(chosen_path []*Item) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "adjustWeight"})
	log := logger.WithContext(ctx)
	var last_index = len(chosen_path) - 1
	if chosen_path[last_index].failureDomainType != DISK {
		log.Fatalf("Expect last Item(%d) in path should be a DISK, but not %s\n", last_index, chosen_path[last_index].failureDomainType.GetStr())
		return
	}
	diskWeight := chosen_path[last_index].weight
	for _, item := range chosen_path {
		log.Debugf("Adjust item (%s, %s) Weight, by (%d - %d)\n", item.name, item.failureDomainType.GetStr(), item.weight, diskWeight)
		item.weight -= diskWeight
	}
}

func ChooseFirstn(take *Item, numReps int32, failureDomain FailureDomain, chooseLeaf bool, chosen_path []*Item) (result []*Item, diskItem []*Item, err error) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "ChooseFirstn"})
	log := logger.WithContext(ctx)
	var chooseOne bool
	reserved_path_len := len(chosen_path)
	for i := 0; i < int(numReps); i++ {
		log.Debugf("----- Start to choose (%d/%d) for %s -----\n", i+1, numReps, failureDomain.GetStr())
		bucket := take
		trial := i
		chooseOne = false
		chosen_path = chosen_path[:reserved_path_len]
		for trial < MAX_TRIES+i {
			chosen_path = append(chosen_path, bucket)
			item := bucketChoose(bucket, trial)
			if item.failureDomainType != failureDomain {
				log.Debugf("%s failure domain %s contains target failure domain %s, will go into next\n",
					item.name, item.failureDomainType.GetStr(), failureDomain.GetStr())
				bucket = item
				continue
			}
			if hasChosen(result, item) {
				log.Warnf("%s (%s) has been chosen already, will re-choose another (retries: %d), chosen set by now:[%v]\n",
					item.name, item.failureDomainType.GetStr(), trial, result)
				goto RETRY
			}
			if !isAvailable(item) {
				log.Warnf("%s (%s) is not available (Status: %v), will re-choose another (retries: %d), chosen set by now:[%v]\n",
					item.name, item.failureDomainType.GetStr(), item.status, trial, result)
				goto RETRY
			}
			if failureDomain == DISK {
				log.Debugf("\tChoose Disk %v Succeed\n", *item)
				result = append(result, item)
				diskItem = append(diskItem, item)
				chosen_path = append(chosen_path, item)
				chooseOne = true
				adjustWeight(chosen_path)
				break
			} else if chooseLeaf {
				log.Debugf("=== Try to Choose Disk in %s (%s) ===\n", item.name, item.failureDomainType.GetStr())
				_, leafDisk, err := ChooseFirstn(item, 1, DISK, false, chosen_path)
				if err != nil {
					log.Warnf("=== Try to Choose Disk in %s (%s) Failed, will retry(%d) err: %v===\n",
						item.name, item.failureDomainType.GetStr(), trial, err)
					goto RETRY
				} else {
					log.Debugf("=== Choose Disk %s in %s (%s) Succeed ===\n",
						leafDisk[0].name, item.name, item.failureDomainType.GetStr())
					result = append(result, item)
					diskItem = append(diskItem, leafDisk[0])
					chooseOne = true
					break
				}
			} else {
				log.Debugf("Choose %s (%s) Succeed, Try to Get another\n", item.name, item.failureDomainType.GetStr())
				result = append(result, item)
				chosen_path = append(chosen_path, item)
				chooseOne = true
				adjustWeight(chosen_path)
				break
			}
		RETRY:
			trial++
			bucket = take
			chosen_path = chosen_path[:reserved_path_len]
			continue
		}
		if !chooseOne {
			errMsg := fmt.Sprintf("Could not find a %s in Round (%d/%d)", failureDomain.GetStr(), i+1, numReps)
			log.Errorln("-----", errMsg, "-----")
			err = errors.New(errMsg)
			return
		}
		log.Debugf("----- End to choose (%d/%d) for %s -----\n", i+1, numReps, failureDomain.GetStr())
	}
	return
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
