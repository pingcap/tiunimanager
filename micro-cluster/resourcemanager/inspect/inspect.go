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

package hostInspector

import (
	"context"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models"
	cluster_mgm "github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/resource"
)

type HostInspect struct {
	resouceRW  resource.ReaderWriter
	instanceRW cluster_mgm.ReaderWriter
}

func NewHostInspector() HostInspector {
	hostInspector := new(HostInspect)
	hostInspector.resouceRW = models.GetResourceReaderWriter()
	hostInspector.instanceRW = models.GetClusterReaderWriter()
	return hostInspector
}

func (p *HostInspect) CheckCpuCores(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckMemorySize(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckDiskSize(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckDiskRatio(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]structs.CheckInt32, err error) {
	return
}

func (p *HostInspect) CheckCpuAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]structs.CheckInt32, err error) {
	//result = new(structs.CheckInt32)
	//result.ExpectedValue = p.resouceRW.
	return
}
func (p *HostInspect) CheckMemAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckDiskAllocated(ctx context.Context, hosts []structs.HostInfo) (inconsistDisks map[string]map[string]structs.CheckString, err error) {
	return
}
