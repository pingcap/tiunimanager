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

package management

import "github.com/pingcap-inc/tiem/models/common"

type ClusterInstance struct {
	common.Entity

	Type           string
	Role           string
	Version        string
	ClusterID      string
	Status         string
	HostID         string
	SpecCode       string
	ZoneCode       string

	Addresses      []string
	Ports          []string
	DiskId         string
	DiskPath       string
	AllocRequestId string
	CpuCores       int8
	Memory         int8
}
