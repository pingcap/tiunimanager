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

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models/common"
)

type Cluster struct {
	common.Entity
	Name              string   `gorm:"not null;size:64;comment:'user name of the cluster''"`
	DBUser            string   `gorm:"not null;size:64;comment:'user name of the database''"`
	DBPassword        string   `gorm:"not null;size:64;comment:'user password of the database''"`
	Type              string   `gorm:"not null;size:16;comment:'type of the cluster, eg. TiDB、TiDB Migration';"`
	Version           string   `gorm:"not null;size:64;comment:'version of the cluster'"`
	TLS               bool     `gorm:"default:false;comment:'whether to enable TLS, value: true or false'"`
	Tags              []string `gorm:"comment:'cluster tag information'"`
	OwnerId           string   `gorm:"not null:size:32;<-:create;->"`
	ParameterGroupID  string   `gorm:"comment: parameter group id"`
	Exclusive         bool
	Region            string
	CpuArchitecture   constants.ArchType
	MaintenanceStatus constants.ClusterMaintenanceStatus
}
