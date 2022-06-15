/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package hostinitiator

import (
	"context"

	"github.com/pingcap/tiunimanager/common/structs"
)

type HostInitiator interface {
	AuthHost(ctx context.Context, deployUser, userGroup string, h *structs.HostInfo) (err error)
	Prepare(ctx context.Context, h *structs.HostInfo) (err error)
	Verify(ctx context.Context, h *structs.HostInfo) (err error)
	InstallSoftware(ctx context.Context, hosts []structs.HostInfo) (err error)
	PreCheckHostInstallFilebeat(ctx context.Context, hosts []structs.HostInfo) (installed bool, err error)
	JoinEMCluster(ctx context.Context, hosts []structs.HostInfo) (operationID string, err error)
	LeaveEMCluster(ctx context.Context, nodeId string) (operationID string, err error)
}
