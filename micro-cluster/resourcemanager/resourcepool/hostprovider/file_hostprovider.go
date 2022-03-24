/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 ******************************************************************************/

package hostprovider

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"

	cluster_rw "github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/resource"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
)

type FileHostProvider struct {
	rw resource.ReaderWriter
}

func NewFileHostProvider() *FileHostProvider {
	hostProvider := new(FileHostProvider)
	hostProvider.rw = models.GetResourceReaderWriter()
	return hostProvider
}

func (p *FileHostProvider) SetResourceReaderWriter(rw resource.ReaderWriter) {
	p.rw = rw
}

func (p *FileHostProvider) ValidateZoneInfo(ctx context.Context, host *structs.HostInfo) (err error) {
	productRW := models.GetProductReaderWriter()
	_, zones, _, err := productRW.GetVendor(ctx, host.Vendor)
	if err != nil {
		return err
	}

	// any matched
	for _, zone := range zones {
		if zone.RegionID == host.Region && zone.ZoneID == host.AZ {
			return nil
		}
	}
	errMsg := fmt.Sprintf("can not get zone info for host %s %s, with vendor %s, region %s, zone %s", host.HostName, host.IP, host.Vendor, host.Region, host.AZ)
	return errors.NewError(errors.TIEM_RESOURCE_INVALID_ZONE_INFO, errMsg)

}

func (p *FileHostProvider) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (hostIds []string, err error) {
	var dbModelHosts []resourcepool.Host
	for _, host := range hosts {
		var dbHost resourcepool.Host
		err = dbHost.ConstructFromHostInfo(&host)
		if err != nil {
			return nil, err
		}
		dbModelHosts = append(dbModelHosts, dbHost)
	}
	return p.rw.Create(ctx, dbModelHosts)
}

func (p *FileHostProvider) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	return p.rw.Delete(ctx, hostIds)
}

func (p *FileHostProvider) QueryHosts(ctx context.Context, location *structs.Location, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, total int64, err error) {
	dbhosts, total, err := p.rw.Query(ctx, location, filter, page.CalcOffset(), page.PageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, dbhost := range dbhosts {
		var host structs.HostInfo
		dbhost.ToHostInfo(&host)
		hosts = append(hosts, host)
	}

	err = p.getInstancesOnHosts(ctx, hosts)
	if err != nil {
		// maybe query cluster_instances table failed, return hosts info without instances relationship
		framework.LogWithContext(ctx).Warnf("get hosts instances failed, %v", err)
	}
	return
}

func (p *FileHostProvider) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	return p.rw.UpdateHostStatus(ctx, hostIds, status)
}

func (p *FileHostProvider) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	return p.rw.UpdateHostReserved(ctx, hostIds, reserved)
}

// Trim the whole depth tree by specified level and depth
func (p *FileHostProvider) trimTree(root *structs.HierarchyTreeNode, level constants.HierarchyTreeNodeLevel, depth int) *structs.HierarchyTreeNode {
	newRoot := structs.HierarchyTreeNode{
		Code: "root",
	}
	levelNodes := root.SubNodes

	// get to the specified level, and put all nodes in that level to newRoot.Subnodes[]
	for l := constants.REGION; l < level; l++ {
		var subnodes []*structs.HierarchyTreeNode
		for _, node := range levelNodes {
			subnodes = append(subnodes, node.SubNodes...)
		}
		levelNodes = subnodes
	}
	newRoot.SubNodes = levelNodes

	leafNodes := levelNodes
	// get to the specified depth from level, and set all nodes' subNode[] in that depth level to nil
	for d := 0; d < depth; d++ {
		var subnodes []*structs.HierarchyTreeNode
		for _, node := range leafNodes {
			subnodes = append(subnodes, node.SubNodes...)
		}
		leafNodes = subnodes
	}
	for _, node := range leafNodes {
		node.SubNodes = nil
	}

	return &newRoot
}

// Add child node to parent subNodes[], and RETURN:
//    - If parent node doest not existed, return new parent node
//    - If parent node already existed, return nil
func (p *FileHostProvider) addSubNode(current map[string]*structs.HierarchyTreeNode, code string, subNode *structs.HierarchyTreeNode) (parent *structs.HierarchyTreeNode) {
	if parent, ok := current[code]; ok {
		parent.SubNodes = append(parent.SubNodes, subNode)
		return nil
	} else {
		parent := structs.HierarchyTreeNode{
			Code:   code,
			Prefix: structs.GetDomainPrefixFromCode(code),
			Name:   structs.GetDomainNameFromCode(code),
		}
		parent.SubNodes = append(parent.SubNodes, subNode)
		current[code] = &parent
		return &parent
	}
}

// Build the hierarchy tree from 'host' node up to 'root'
func (p *FileHostProvider) buildHierarchy(Items []resource.HostItem) *structs.HierarchyTreeNode {
	root := structs.HierarchyTreeNode{
		Code: "root",
	}
	if len(Items) == 0 {
		return &root
	}
	var regions map[string]*structs.HierarchyTreeNode = make(map[string]*structs.HierarchyTreeNode)
	var zones map[string]*structs.HierarchyTreeNode = make(map[string]*structs.HierarchyTreeNode)
	var racks map[string]*structs.HierarchyTreeNode = make(map[string]*structs.HierarchyTreeNode)
	for _, item := range Items {
		hostItem := structs.HierarchyTreeNode{
			Code:     structs.GenDomainCodeByName(item.Ip, item.Name),
			Prefix:   item.Ip,
			Name:     item.Name,
			SubNodes: nil,
		}
		newRack := p.addSubNode(racks, item.Rack, &hostItem)
		if newRack == nil {
			continue
		}
		newZone := p.addSubNode(zones, item.Az, newRack)
		if newZone == nil {
			continue
		}
		newRegion := p.addSubNode(regions, item.Region, newZone)
		if newRegion != nil {
			root.SubNodes = append(root.SubNodes, newRegion)
		}
	}
	return &root
}

func (p *FileHostProvider) GetHierarchy(ctx context.Context, filter *structs.HostFilter, level int, depth int) (*structs.HierarchyTreeNode, error) {
	log := framework.LogWithContext(ctx)
	items, err := p.rw.GetHostItems(ctx, filter, int32(level), int32(depth))
	if err != nil {
		log.Errorf("get host items on hostFilter %v failed, %v", *filter, err)
		return nil, err
	}
	// build the whole depth tree (root-region-zone-rack-host) using the items
	wholeTree := p.buildHierarchy(items)
	var root *structs.HierarchyTreeNode
	if wholeTree.SubNodes != nil {
		// trim the whole depth tree by level and depth, eg. trimed tree: root-zone-rack (level = 2, depth = 1)
		root = p.trimTree(wholeTree, constants.HierarchyTreeNodeLevel(level), depth)
	} else {
		root = wholeTree
		log.Warnf("no stocks with filter:%v", filter)
	}

	return root, nil
}

func (p *FileHostProvider) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks map[string]*structs.Stocks, err error) {
	log := framework.LogWithContext(ctx)
	hostStocks, err := p.rw.GetHostStocks(ctx, location, hostFilter, diskFilter)
	if err != nil {
		log.Errorf("get host stocks on location %v, hostFilter %v, diskFilter %v failed, %v", *location, *hostFilter, *diskFilter, err)
		return nil, err
	}
	stocks = make(map[string]*structs.Stocks)
	for i := range hostStocks {
		if zoneStock, ok := stocks[hostStocks[i].Zone]; ok {
			zoneStock.FreeHostCount++
			zoneStock.FreeCpuCores += hostStocks[i].FreeCpuCores
			zoneStock.FreeMemory += hostStocks[i].FreeMemory
			zoneStock.FreeDiskCount += hostStocks[i].FreeDiskCount
			zoneStock.FreeDiskCapacity += hostStocks[i].FreeDiskCapacity
		} else {
			// First host stock in the zone
			hostStocks[i].FreeHostCount = 1
			stocks[hostStocks[i].Zone] = &hostStocks[i]
		}
	}
	return stocks, nil
}

func (p *FileHostProvider) buildInstancesOnHost(ctx context.Context, items []cluster_rw.HostInstanceItem) map[string]map[string][]string {
	result := make(map[string]map[string][]string)
	for i := range items {
		if instances, ok := result[items[i].HostID]; !ok {
			result[items[i].HostID] = make(map[string][]string)
			result[items[i].HostID][items[i].ClusterID] = []string{items[i].Component}
		} else {
			instances[items[i].ClusterID] = append(instances[items[i].ClusterID], items[i].Component)
		}
	}
	return result
}

func (p *FileHostProvider) getInstancesOnHosts(ctx context.Context, hosts []structs.HostInfo) (err error) {
	clusterRW := models.GetClusterReaderWriter()
	var hostIds []string
	for i := range hosts {
		hostIds = append(hostIds, hosts[i].ID)
	}

	items, err := clusterRW.QueryHostInstances(ctx, hostIds)
	if err != nil {
		return err
	}

	instances := p.buildInstancesOnHost(ctx, items)

	for i := range hosts {
		if hostInstances, ok := instances[hosts[i].ID]; ok {
			hosts[i].Instances = hostInstances
		} else {
			hosts[i].Instances = make(map[string][]string)
		}
	}

	return nil
}

func (p *FileHostProvider) UpdateHostInfo(ctx context.Context, host structs.HostInfo) (err error) {
	if host.ID == "" {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "update host failed without host id")
	}
	if host.UsedCpuCores != 0 || host.UsedMemory != 0 {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "used cpu cores or used memory field should not be set while update host %s", host.ID)
	}
	var dbHost resourcepool.Host
	err = dbHost.ConstructFromHostInfo(&host)
	if err != nil {
		return err
	}
	dbHost.ID = host.ID
	return p.rw.UpdateHostInfo(ctx, dbHost)
}

func (p *FileHostProvider) CreateDisks(ctx context.Context, hostId string, disks []structs.DiskInfo) (diskIds []string, err error) {
	var dbDisks []resourcepool.Disk
	for i := range disks {
		if err = disks[i].ValidateDisk(hostId, ""); err != nil {
			return nil, err
		}
		var dbDisk resourcepool.Disk
		dbDisk.ConstructFromDiskInfo(&disks[i])
		dbDisks = append(dbDisks, dbDisk)
	}
	return p.rw.CreateDisks(ctx, hostId, dbDisks)
}

func (p *FileHostProvider) DeleteDisks(ctx context.Context, diskIds []string) (err error) {
	return p.rw.DeleteDisks(ctx, diskIds)
}

func (p *FileHostProvider) UpdateDisk(ctx context.Context, disk structs.DiskInfo) (err error) {
	if disk.ID == "" {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "update disk failed without disk id")
	}
	var dbDisk resourcepool.Disk
	dbDisk.ConstructFromDiskInfo(&disk)
	dbDisk.ID = disk.ID
	return p.rw.UpdateDisk(ctx, dbDisk)
}
