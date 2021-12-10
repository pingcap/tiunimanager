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

import "github.com/pingcap-inc/tiem/common/structs"

type ComputeResource struct {
	CpuCores int32
	Memory   int32
}

type DiskResource struct {
	DiskId   string
	DiskName string
	Path     string
	Type     string
	Capacity int32
}

type PortResource struct {
	Start int32
	End   int32
	Ports []int32
}

type Compute struct {
	Reqseq     int32
	Location   structs.Location
	HostId     string
	HostName   string
	HostIp     string
	UserName   string
	Passwd     string
	ComputeRes ComputeRequirement
	DiskRes    DiskResource
	PortRes    []PortResource
}

type AllocRsp struct {
	Results []Compute
}

type BatchAllocResponse struct {
	BatchResults []*AllocRsp
}
