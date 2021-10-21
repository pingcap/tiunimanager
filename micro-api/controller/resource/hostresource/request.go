
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

package hostresource

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type HostQuery struct {
	controller.PageRequest
	Purpose string `json:"purpose" form:"purpose"`
	Status  int    `json:"status" form:"status"`
}

type ExcelField int

const (
	HOSTNAME_FIELD ExcelField = iota
	IP_FILED
	USERNAME_FIELD
	PASSWD_FIELD
	REGION_FIELD
	ZONE_FIELD
	RACK_FIELD
	ARCH_FIELD
	OS_FIELD
	KERNEL_FIELD
	CPU_FIELD
	MEM_FIELD
	NIC_FIELD
	PURPOSE_FIELD
	DISKTYPE_FIELD
	DISKS_FIELD
)

type Allocation struct {
	FailureDomain string `json:"failureDomain"`
	CpuCores      int32  `json:"cpuCores"`
	Memory        int32  `json:"memory"`
	Count         int32  `json:"count"`
}

type AllocHostsReq struct {
	PdReq   []Allocation `json:"pdReq"`
	TidbReq []Allocation `json:"tidbReq"`
	TikvReq []Allocation `json:"tikvReq"`
}
