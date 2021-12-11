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

	allocrecycle "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/allocator_recycler"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/structs"
)

type Management struct {
	localHostManage structs.AllocatorRecycler
	// cloudManage allocrecycle.AllocatorRecycler
}

func (m *Management) InitManagement() {
	m.localHostManage = allocrecycle.GetLocalHostManagement()
}

func (m *Management) AllocResources(ctx context.Context, batchReq *structs.BatchAllocRequest) (results *structs.BatchAllocResponse, err error) {
	return nil, nil
}
func (m *Management) RecycleResources(ctx context.Context, request *structs.RecycleRequest) (err error) {
	return nil
}
