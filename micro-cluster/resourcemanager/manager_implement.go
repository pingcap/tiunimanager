/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
	"sync"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/management"
	resource_structs "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool"
)

type ResourceManager struct {
	resourcePool *resourcepool.ResourcePool
	management   *management.Management
}

var manager *ResourceManager
var once sync.Once

func NewResourceManager() *ResourceManager {
	once.Do(func() {
		if manager == nil {
			manager = new(ResourceManager)
			manager.InitResourceManager()
		}
	})

	return manager
}

func (m *ResourceManager) InitResourceManager() {
	m.resourcePool = resourcepool.GetResourcePool()
	m.management = management.GetManagement()
}

func (m *ResourceManager) GetResourcePool() *resourcepool.ResourcePool {
	return m.resourcePool
}

func (m *ResourceManager) GetManagement() *management.Management {
	return m.management
}

func (m *ResourceManager) ImportHosts(ctx context.Context, hosts []structs.HostInfo, condition *structs.ImportCondition) (flowIds []string, hostIds []string, err error) {
	ok := framework.CheckAndSetInEMTiupProcess()
	if !ok {
		errMsg := "a import/delete hosts workflow is running, please retry later"
		framework.LogWithContext(ctx).Errorln(errMsg)
		return nil, nil, errors.NewError(errors.TIUNIMANAGER_TASK_CONFLICT, errMsg)
	}
	flowIds, hostIds, err = m.resourcePool.ImportHosts(ctx, hosts, condition)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("import hosts %v in batch failed from db service: %v", hosts, err)
		framework.UnsetInEmTiupProcess()
	} else {
		framework.LogWithContext(ctx).Infof("import %d hosts in batch %v succeed from db service.", len(hosts), flowIds)
		// if err is nil, call framework.UnsetInEmTiupProcess() after async background import flows done
	}

	return
}

func (m *ResourceManager) QueryHosts(ctx context.Context, location *structs.Location, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, total int64, err error) {
	hosts, total, err = m.resourcePool.QueryHosts(ctx, location, filter, page)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("query hosts in location %v with filter %v failed from db service: %v", *location, *filter, err)
	} else {
		framework.LogWithContext(ctx).Infof("query %d hosts in location %v with filter %v succeed from db service.", len(hosts), *location, *filter)
	}

	return
}

func (m *ResourceManager) DeleteHosts(ctx context.Context, hostIds []string, force bool) (flowIds []string, err error) {
	ok := framework.CheckAndSetInEMTiupProcess()
	if !ok {
		errMsg := "a import/delete hosts workflow is running, please retry later"
		framework.LogWithContext(ctx).Errorln(errMsg)
		return nil, errors.NewError(errors.TIUNIMANAGER_TASK_CONFLICT, errMsg)
	}
	flowIds, err = m.resourcePool.DeleteHosts(ctx, hostIds, force)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("delete %d hosts %v in batch failed from db service: %v", len(hostIds), hostIds, err)
		framework.UnsetInEmTiupProcess()
	} else {
		framework.LogWithContext(ctx).Infof("delete %d hosts %v in batch %v succeed from db service.", len(hostIds), hostIds, flowIds)
		// if err is nil, call framework.UnsetInEmTiupProcess() after async background delete flows done
	}

	return
}

func (m *ResourceManager) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	err = m.resourcePool.UpdateHostReserved(ctx, hostIds, reserved)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update %d hosts %v to reserved (%v) failed from db service: %v", len(hostIds), hostIds, reserved, err)
	} else {
		framework.LogWithContext(ctx).Infof("update %d hosts %v to reserved (%v) succeed from db service.", len(hostIds), hostIds, reserved)
	}

	return
}

func (m *ResourceManager) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	err = m.resourcePool.UpdateHostStatus(ctx, hostIds, status)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update %d hosts %v to status %s failed from db service: %v", len(hostIds), hostIds, status, err)
	} else {
		framework.LogWithContext(ctx).Infof("update %d hosts %v to status %s succeed from db service.", len(hostIds), hostIds, status)
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

func (m *ResourceManager) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks map[string]*structs.Stocks, err error) {
	stocks, err = m.resourcePool.GetStocks(ctx, location, hostFilter, diskFilter)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get stocks on location %v, host filter %v, disk filter %v failed from db service: %v", *location, *hostFilter, *diskFilter, err)
	} else {
		framework.LogWithContext(ctx).Infof("get stocks on location %v, host filter %v, disk filter %v succeed from db service.", *location, *hostFilter, *diskFilter)
	}

	return
}

func (m *ResourceManager) UpdateHostInfo(ctx context.Context, host structs.HostInfo) (err error) {
	err = m.resourcePool.UpdateHostInfo(ctx, host)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update host %s failed from db service: %v", host.ID, err)
	} else {
		framework.LogWithContext(ctx).Infof("update host %s succeed from db service.", host.ID)
	}

	return
}

func (m *ResourceManager) CreateDisks(ctx context.Context, hostId string, disks []structs.DiskInfo) (diskIds []string, err error) {
	diskIds, err = m.resourcePool.CreateDisks(ctx, hostId, disks)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("create %d disks for host %s failed from db service: %v", len(disks), hostId, err)
	} else {
		framework.LogWithContext(ctx).Infof("create %d disks for host %s succeed from db service.", len(disks), hostId)
	}

	return
}

func (m *ResourceManager) DeleteDisks(ctx context.Context, diskIds []string) (err error) {
	err = m.resourcePool.DeleteDisks(ctx, diskIds)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("delete %d disks failed from db service: %v", len(diskIds), err)
	} else {
		framework.LogWithContext(ctx).Infof("delete %d disks succeed from db service.", len(diskIds))
	}

	return
}

func (m *ResourceManager) UpdateDisk(ctx context.Context, disk structs.DiskInfo) (err error) {
	err = m.resourcePool.UpdateDisk(ctx, disk)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update disk %s failed from db service: %v", disk.ID, err)
	} else {
		framework.LogWithContext(ctx).Infof("update disk %s succeed from db service.", disk.ID)
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
