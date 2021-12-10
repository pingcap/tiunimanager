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

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/models/resource"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
)

type FileHostProvider struct {
	rw resource.ResourceReaderWriter
}

func GetFileHostProvider() HostProvider {
	hostProvider := new(FileHostProvider)
	hostProvider.rw = resource.NewGormChangeFeedReadWrite()
	return hostProvider
}

func (p *FileHostProvider) SetResourceReaderWriter(rw resource.ResourceReaderWriter) {
	p.rw = rw
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

func (p *FileHostProvider) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error) {
	dbhosts, err := p.rw.Query(ctx, filter, page.CalcOffset(), page.PageSize)
	if err != nil {
		return nil, err
	}
	for _, dbhost := range dbhosts {
		var host structs.HostInfo
		dbhost.ToHostInfo(&host)
		hosts = append(hosts, host)
	}
	return
}

func (p *FileHostProvider) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	return p.rw.UpdateHostStatus(ctx, hostIds, status)
}

func (p *FileHostProvider) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	return p.rw.UpdateHostReserved(ctx, hostIds, reserved)
}

func (p *FileHostProvider) trimTree(root *structs.HierarchyTreeNode, level constants.HierarchyTreeNodeLevel, depth int) *structs.HierarchyTreeNode {
	newRoot := structs.HierarchyTreeNode{
		Code: "root",
	}
	levelNodes := root.SubNodes

	for l := constants.REGION; l < level; l++ {
		var subnodes []*structs.HierarchyTreeNode
		for _, node := range levelNodes {
			subnodes = append(subnodes, node.SubNodes...)
		}
		levelNodes = subnodes
	}
	newRoot.SubNodes = levelNodes

	leafNodes := levelNodes
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
	wholeTree := p.buildHierarchy(items)
	var root *structs.HierarchyTreeNode
	if wholeTree.SubNodes != nil {
		root = p.trimTree(wholeTree, constants.HierarchyTreeNodeLevel(level), depth)
	} else {
		root = wholeTree
		log.Warnf("no stocks with filter:%v", filter)
	}

	return root, nil
}

func (p *FileHostProvider) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks *structs.Stocks, err error) {
	log := framework.LogWithContext(ctx)
	hostStocks, err := p.rw.GetHostStocks(ctx, location, hostFilter, diskFilter)
	if err != nil {
		log.Errorf("get host stocks on location %v, hostFilter %v, diskFilter %v failed, %v", *location, *hostFilter, *diskFilter, err)
		return nil, err
	}
	stocks = new(structs.Stocks)
	stocks.FreeHostCount = int32(len(hostStocks))
	for _, stock := range hostStocks {
		stocks.FreeCpuCores += int32(stock.FreeCpuCores)
		stocks.FreeMemory += int32(stock.FreeMemory)
		stocks.FreeDiskCount += int32(stock.FreeDiskCount)
		stocks.FreeDiskCapacity += int32(stock.FreeDiskCapacity)
	}
	return stocks, nil
}
