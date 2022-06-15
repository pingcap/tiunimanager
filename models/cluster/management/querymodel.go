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

package management

import "github.com/pingcap/tiunimanager/common/constants"

type Filters struct {
	ClusterIDs    []string
	TenantId      string
	NameLike      string
	Type          string
	StatusFilters []constants.ClusterRunningStatus
	Tag           string
}

type Result struct {
	Cluster   *Cluster
	Instances []*ClusterInstance
	DBUsers   []*DBUser
}
