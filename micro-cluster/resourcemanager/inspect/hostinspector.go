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
)

type HostInspector interface {
	CheckCpuCores(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error)
	CheckMemorySize(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error)
	CheckCpuAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]structs.CheckInt32, err error)
	CheckMemAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]structs.CheckInt32, err error)
	CheckDiskAllocated(ctx context.Context, hosts []structs.HostInfo) (inconsistDisks map[string]map[string]structs.CheckString, err error)
	CheckDiskSize(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]structs.CheckInt32, err error)
	CheckDiskRatio(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]structs.CheckInt32, err error)
}
