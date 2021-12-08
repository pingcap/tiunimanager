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
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
)

type ResourceManager struct {
	resourcePool *resourcepool.ResourcePool
}

func NewResourceManager() *ResourceManager {
	m := new(ResourceManager)
	m.resourcePool = new(resourcepool.ResourcePool)
	m.resourcePool.InitResourcePool()
	return m
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
