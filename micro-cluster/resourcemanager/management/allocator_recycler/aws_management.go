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

	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/resource"
)

type AwsManagement struct {
	rw resource.ReaderWriter
}

func NewAwsManagement() structs.AllocatorRecycler {
	awsManagement := new(AwsManagement)
	awsManagement.rw = models.GetResourceReaderWriter()
	return awsManagement
}

func (m *AwsManagement) SetResourceReaderWriter(rw resource.ReaderWriter) {
	m.rw = rw
}

func (m *AwsManagement) AllocResources(ctx context.Context, batchReq *structs.BatchAllocRequest) (results *structs.BatchAllocResponse, err error) {
	return m.rw.AllocResources(ctx, batchReq)
}
func (m *AwsManagement) RecycleResources(ctx context.Context, request *structs.RecycleRequest) (err error) {
	return m.rw.RecycleResources(ctx, request)
}
