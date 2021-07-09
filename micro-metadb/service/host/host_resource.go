package host

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/models"
)

type FailureDomain uint32

func (fd FailureDomain) String() (str string) {
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
	extra             interface{}
	subItems          []*Item
}

func (item Item) GetId() string {
	return item.id
}

func (item Item) GetName() string {
	return item.name
}

func (item Item) GetStatus() int32 {
	return item.status
}

func (item Item) GetExtra() interface{} {
	return item.extra
}

func (item Item) isAvailable() bool {
	return item.status == 0
}

type HostExtraInfo struct {
	Id       string
	Ip       string
	CpuCores int
	Memory   int
}

type DiskExtraInfo struct {
	Id       string
	Path     string
	Capacity int32
}

func buildHierarchy() (rack2hosts map[string][]*Item, zone2racks map[string][]string, dc2zones map[string][]string, diskCount int, err error) {
	rack2hosts = make(map[string][]*Item)
	zone2racks = make(map[string][]string)
	dc2zones = make(map[string][]string)
	// For deduplicated
	tmp_rack_recorded := make(map[string]bool)
	tmp_zone_recorded := make(map[string]bool)
	// total disk count
	diskCount = 0

	hosts, _ := models.ListHosts()
	for _, host := range hosts {
		hostExtra := HostExtraInfo{
			Id:       host.ID,
			Ip:       host.IP,
			CpuCores: host.CpuCores,
			Memory:   host.Memory,
		}
		hostItem := Item{
			id:                host.ID,
			name:              host.Name,
			failureDomainType: HOST,
			status:            host.Status,
			extra:             &hostExtra,
		}
		for _, disk := range host.Disks {
			diskExtra := DiskExtraInfo{
				Id:       disk.ID,
				Path:     disk.Path,
				Capacity: disk.Capacity,
			}
			diskItem := Item{
				id:                disk.ID,
				name:              disk.Name,
				status:            disk.Status,
				failureDomainType: DISK,
				extra:             &diskExtra,
			}
			if diskItem.isAvailable() {
				diskItem.weight = calcDiskWeight(disk.Capacity)
			}
			if hostItem.isAvailable() {
				hostItem.weight += diskItem.weight
			}
			hostItem.subItems = append(hostItem.subItems, &diskItem)
			diskCount++
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

func GetHierarchyRoot() (root *Item, diskCnt int, err error) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "GetHierarchyRoot"})
	log := logger.WithContext(ctx)
	rack2hosts, zone2racks, dc2zones, diskCnt, err := buildHierarchy()
	if err != nil {
		log.Fatalln("BuildHierarchy failed, err:", err)
		return nil, 0, err
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
					if rackItem.isAvailable() && host.isAvailable() {
						rackItem.weight += host.weight
					}
				}
				zoneItem.subItems = append(zoneItem.subItems, &rackItem)
				if zoneItem.isAvailable() && rackItem.isAvailable() {
					zoneItem.weight += rackItem.weight
				}
			}
			dcItem.subItems = append(dcItem.subItems, &zoneItem)
			if dcItem.isAvailable() && zoneItem.isAvailable() {
				dcItem.weight += zoneItem.weight
			}
		}
		root.subItems = append(root.subItems, &dcItem)
		if root.isAvailable() && dcItem.isAvailable() {
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
	if high_item.weight == 0 {
		return nil
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

func adjustWeight(chosen_path []*Item, adjust_for_failed bool) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "adjustWeight"})
	log := logger.WithContext(ctx)
	var adjust_unit uint32
	var last_index = len(chosen_path) - 1
	if !adjust_for_failed {
		if chosen_path[last_index].failureDomainType != DISK {
			log.Fatalf("Expect last Item(%d) in path should be a DISK, but not %s\n", last_index, chosen_path[last_index].failureDomainType)
			return
		}
	}
	adjust_unit = chosen_path[last_index].weight
	for _, item := range chosen_path {
		log.Debugf("Adjust item (%s, %s) Weight, by (%d - %d)\n", item.name, item.failureDomainType, item.weight, adjust_unit)
		item.weight -= adjust_unit
	}
}

func ChooseFirstn(take *Item, numReps int32, maxRetries int, failureDomain FailureDomain, chooseLeaf bool, chosen_path []*Item) (result []*Item, diskItem []*Item, err error) {
	ctx := logger.NewContext(context.Background(), logger.Fields{"micro-service": "ChooseFirstn"})
	log := logger.WithContext(ctx)
	var chooseOne bool
	reserved_path_len := len(chosen_path)
	for i := 0; i < int(numReps); i++ {
		log.Debugf("----- Start to choose (%d/%d) for %s -----\n", i+1, numReps, failureDomain)
		bucket := take
		trial := i
		chooseOne = false
		chosen_path = chosen_path[:reserved_path_len]
		for trial < maxRetries+i {
			chosen_path = append(chosen_path, bucket)
			item := bucketChoose(bucket, trial)
			if item == nil {
				log.Warnf("Resources under %s are all not available", bucket.failureDomainType)
				break
			}
			if item.failureDomainType != failureDomain {
				log.Debugf("%s failure domain %s contains target failure domain %s, will go into next\n",
					item.name, item.failureDomainType, failureDomain)
				bucket = item
				continue
			}
			if hasChosen(result, item) {
				log.Warnf("%s (%s) has been chosen already, will re-choose another (retries: %d), chosen set by now:[%v]\n",
					item.name, item.failureDomainType, trial, result)
				// append item to chosen_path to adjust weight, after which will reset chosen_path in RETRY for a new loop
				chosen_path = append(chosen_path, item)
				adjustWeight(chosen_path, true)
				goto RETRY
			}
			if !item.isAvailable() {
				log.Warnf("%s (%s) is not available (Status: %v), will re-choose another (retries: %d), chosen set by now:[%v]\n",
					item.name, item.failureDomainType, item.status, trial, result)
				// append item to chosen_path to adjust weight, after which will reset chosen_path in RETRY for a new loop
				chosen_path = append(chosen_path, item)
				adjustWeight(chosen_path, true)
				goto RETRY
			}
			if failureDomain == DISK {
				log.Debugf("\tChoose Disk %v Succeed\n", *item)
				result = append(result, item)
				diskItem = append(diskItem, item)
				chosen_path = append(chosen_path, item)
				chooseOne = true
				adjustWeight(chosen_path, false)
				break
			} else if chooseLeaf {
				log.Debugf("=== Try to Choose Disk in %s (%s) ===\n", item.name, item.failureDomainType)
				_, leafDisk, err := ChooseFirstn(item, 1, maxRetries, DISK, false, chosen_path)
				if err != nil {
					log.Warnf("=== Try to Choose Disk in %s (%s) Failed, will retry(%d) err: %v===\n",
						item.name, item.failureDomainType, trial, err)
					goto RETRY
				} else {
					log.Debugf("=== Choose Disk %s in %s (%s) Succeed ===\n",
						leafDisk[0].name, item.name, item.failureDomainType)
					result = append(result, item)
					diskItem = append(diskItem, leafDisk[0])
					chooseOne = true
					break
				}
			} else {
				log.Debugf("Choose %s (%s) Succeed, Try to Get another\n", item.name, item.failureDomainType)
				result = append(result, item)
				chosen_path = append(chosen_path, item)
				chooseOne = true
				adjustWeight(chosen_path, false)
				break
			}
		RETRY:
			trial++
			bucket = take
			chosen_path = chosen_path[:reserved_path_len]
			continue
		}
		if !chooseOne {
			errMsg := fmt.Sprintf("Could not find a %s in Round (%d/%d)", failureDomain, i+1, numReps)
			log.Errorln("-----", errMsg, "-----")
			err = errors.New(errMsg)
			return
		}
		log.Debugf("----- End to choose (%d/%d) for %s -----\n", i+1, numReps, failureDomain)
	}
	return
}
