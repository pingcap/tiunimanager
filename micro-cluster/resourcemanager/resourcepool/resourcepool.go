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

package resourcepool

import (
	"context"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
)

type ResourcePool struct {
	hostProvider hostprovider.HostProvider
	// cloudHostProvider hostprovider.HostProvider
}

func (p *ResourcePool) InitResourcePool() {
	p.hostProvider = hostprovider.GetFileHostProvider()
}

func (p *ResourcePool) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (hostIds []string, err error) {
	return p.hostProvider.ImportHosts(ctx, hosts)
}

func (p *ResourcePool) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	return p.hostProvider.DeleteHosts(ctx, hostIds)
}

func (p *ResourcePool) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error) {
	return p.hostProvider.QueryHosts(ctx, filter, page)
}

func (p *ResourcePool) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	return p.hostProvider.UpdateHostStatus(ctx, hostIds, status)
}

func (p *ResourcePool) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	return p.hostProvider.UpdateHostReserved(ctx, hostIds, reserved)
}

func (p *ResourcePool) GetHierarchy(ctx context.Context, filter *structs.HostFilter, level int, depth int) (root *structs.HierarchyTreeNode, err error) {
	return p.hostProvider.GetHierarchy(ctx, filter, level, depth)
}

func (p *ResourcePool) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks *structs.Stocks, err error) {
	return p.hostProvider.GetStocks(ctx, location, hostFilter, diskFilter)
}
