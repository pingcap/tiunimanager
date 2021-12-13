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

package resourcemanager

import (
	"context"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/structs"
	"github.com/pingcap-inc/tiem/models"
)

type ResourceManager struct {
	resourcePool *resourcepool.ResourcePool
	management   *management.Management
}

func NewResourceManager() *ResourceManager {
	m := new(ResourceManager)
	m.resourcePool = new(resourcepool.ResourcePool)
	m.management = new(management.Management)
	m.resourcePool.InitResourcePool(models.GetResourceReaderWriter())
	m.management.InitManagement(models.GetResourceReaderWriter())
	return m
}

func (m *ResourceManager) GetResourcePool() *resourcepool.ResourcePool {
	return m.resourcePool
}

func (m *ResourceManager) GetManagement() *management.Management {
	return m.management
}

func (m *ResourceManager) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (hostIds []string, err error) {
	hostIds, err = m.resourcePool.ImportHosts(ctx, hosts)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("import hosts in batch failed from db service: %v", err)
	} else {
		framework.LogWithContext(ctx).Infof("import %d hosts in batch succeed from db service.", len(hosts))
	}

	return
}

func (m *ResourceManager) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error) {
	hosts, err = m.resourcePool.QueryHosts(ctx, filter, page)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("query hosts in filter %v failed from db service: %v", *filter, err)
	} else {
		framework.LogWithContext(ctx).Infof("query %d hosts in filter %v succeed from db service.", len(hosts), *filter)
	}

	return
}

func (m *ResourceManager) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	err = m.resourcePool.DeleteHosts(ctx, hostIds)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("delete %d hosts in failed from db service: %v", len(hostIds), err)
	} else {
		framework.LogWithContext(ctx).Infof("delete %d hosts %v succeed from db service.", len(hostIds), hostIds)
	}

	return
}

func (m *ResourceManager) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	err = m.resourcePool.UpdateHostReserved(ctx, hostIds, reserved)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update %d hosts to reserved (%v) failed from db service: %v", len(hostIds), reserved, err)
	} else {
		framework.LogWithContext(ctx).Infof("update %d hosts[%v] to reserved (%v) succeed from db service.", len(hostIds), hostIds, reserved)
	}

	return
}

func (m *ResourceManager) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	err = m.resourcePool.UpdateHostStatus(ctx, hostIds, status)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update %d hosts to status %s failed from db service: %v", len(hostIds), status, err)
	} else {
		framework.LogWithContext(ctx).Infof("update %d hosts[%v] to status %s succeed from db service.", len(hostIds), hostIds, status)
	}

	return
}

func (m *ResourceManager) GetHierarchy(ctx context.Context, filter *structs.HostFilter, level int, depth int) (root *structs.HierarchyTreeNode, err error) {
	root, err = m.resourcePool.GetHierarchy(ctx, filter, level, depth)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get hierarchy filter %v, level %d, depth %d failed from db service: %v", *filter, level, depth, err)
	} else {
		framework.LogWithContext(ctx).Infof("get hierarchy filter %v, level %d, depth %d succeed from db service.", *filter, level, depth)
	}

	return
}

func (m *ResourceManager) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks *structs.Stocks, err error) {
	stocks, err = m.resourcePool.GetStocks(ctx, location, hostFilter, diskFilter)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get stocks on location %v, host filter %v, disk filter %v failed from db service: %v", *location, *hostFilter, *diskFilter, err)
	} else {
		framework.LogWithContext(ctx).Infof("get stocks on location %v, host filter %v, disk filter %v succeed from db service.", *location, *hostFilter, *diskFilter)
	}

	return
}

func (m *ResourceManager) AllocResources(ctx context.Context, batchReq *resource_structs.BatchAllocRequest) (results *resource_structs.BatchAllocResponse, err error) {
	results, err = m.management.AllocResources(ctx, batchReq)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("alloc resource failed on request %v from db service: %v", *batchReq, err)
	} else {
		framework.LogWithContext(ctx).Infof("alloc resources %v succeed from db service for request %v.", *results, *batchReq)
	}

	return
}

func (m *ResourceManager) RecycleResources(ctx context.Context, request *resource_structs.RecycleRequest) (err error) {
	err = m.management.RecycleResources(ctx, request)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("recycle resources failed on request %v from db service: %v", *request, err)
	} else {
		framework.LogWithContext(ctx).Infof("recycle resources %v succeed from db service.", *request)
	}
	return
}
