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
 *                                                                            *
 ******************************************************************************/

package hostresource

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
	CLUSTER_TYPE_FIELD
	PURPOSE_FIELD
	DISKTYPE_FIELD
	DISKS_FIELD
)

//Constants for importing host information
const (
	ImportHostTemplateFileName string = "hostInfo_template.xlsx"
	ImportHostTemplateFilePath string = "./resource"
	ImportHostTemplateSheet    string = "Host Information"
)
