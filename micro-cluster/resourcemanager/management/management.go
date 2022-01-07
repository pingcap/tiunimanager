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
 *                                                                            *
 ******************************************************************************/

package management

import (
	"context"
	"sync"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	allocrecycle "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/allocator_recycler"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
)

type Management struct {
	localHostManage structs.AllocatorRecycler
	AwsManage       structs.AllocatorRecycler
}

var management *Management
var once sync.Once

func GetManagement() *Management {
	once.Do(func() {
		if management == nil {
			management = new(Management)
			management.InitManagement()
		}
	})
	return management
}

func (m *Management) InitManagement() {
	m.localHostManage = allocrecycle.NewLocalHostManagement()
	m.AwsManage = allocrecycle.NewAwsManagement()
}

func (m *Management) SetAllocatorRecycler(localHostManage structs.AllocatorRecycler) {
	m.localHostManage = localHostManage
}

func (m *Management) GetAllocatorRecycler() structs.AllocatorRecycler {
	return m.localHostManage
}

func (m *Management) AllocResources(ctx context.Context, batchReq *structs.BatchAllocRequest) (results *structs.BatchAllocResponse, err error) {
	framework.LogWithContext(ctx).Infof("AllocResources in vendor %s", batchReq.Vendor)
	if batchReq.IsLocalVendor() {
		return m.localHostManage.AllocResources(ctx, batchReq)
	} else if batchReq.IsAwsVendor() {
		return m.AwsManage.AllocResources(ctx, batchReq)
	} else {
		return nil, errors.NewEMErrorf(errors.TIEM_RESOURCE_UNKNOWN_VENDOR, "alloc with unknown vendor %s", batchReq.Vendor)
	}
}
func (m *Management) RecycleResources(ctx context.Context, request *structs.RecycleRequest) (err error) {
	if request.IsLocalVendor() {
		return m.localHostManage.RecycleResources(ctx, request)
	} else if request.IsAwsVendor() {
		return m.AwsManage.RecycleResources(ctx, request)
	} else {
		return errors.NewEMErrorf(errors.TIEM_RESOURCE_UNKNOWN_VENDOR, "recycle with unknown vendor %s", request.Vendor)
	}

}
