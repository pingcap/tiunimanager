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

package allocrecycle

import (
	"context"

	"github.com/pingcap-inc/tiunimanager/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiunimanager/models"
	"github.com/pingcap-inc/tiunimanager/models/resource"
)

type LocalHostManagement struct {
	rw resource.ReaderWriter
}

func NewLocalHostManagement() structs.AllocatorRecycler {
	localManagement := new(LocalHostManagement)
	localManagement.rw = models.GetResourceReaderWriter()
	return localManagement
}

func (m *LocalHostManagement) SetResourceReaderWriter(rw resource.ReaderWriter) {
	m.rw = rw
}

func (m *LocalHostManagement) AllocResources(ctx context.Context, batchReq *structs.BatchAllocRequest) (results *structs.BatchAllocResponse, err error) {
	return m.rw.AllocResources(ctx, batchReq)
}
func (m *LocalHostManagement) RecycleResources(ctx context.Context, request *structs.RecycleRequest) (err error) {
	return m.rw.RecycleResources(ctx, request)
}
