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

package domain

import (
	"encoding/json"
	"time"

	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

type Cluster struct {
	Id       string
	Code     string
	TenantId string

	ClusterName    string
	DbPassword     string
	ClusterType    knowledge.ClusterType
	ClusterVersion knowledge.ClusterVersion
	Tags           []string
	Tls            bool
	RecoverInfo    RecoverInfo
	Status         ClusterStatus

	Demands []*ClusterComponentDemand

	WorkFlowId uint

	OwnerId string

	CreateTime time.Time
	UpdateTime time.Time
	DeleteTime time.Time
}

func (c *Cluster) Online() {
	c.Status = ClusterStatusOnline
}
func (c *Cluster) Delete() {
	c.Status = ClusterStatusDeleted
}
func (c *Cluster) Restart() {
	c.Status = ClusterStatusRestart
}

type ClusterComponentDemand struct {
	ComponentType     *knowledge.ClusterComponent
	TotalNodeCount    int
	DistributionItems []*ClusterNodeDistributionItem
}

type ClusterDemandRecord struct {
	Id         uint
	TenantId   string
	ClusterId  string
	Content    []*ClusterComponentDemand
	CreateTime time.Time
}

type ClusterNodeDistributionItem struct {
	ZoneCode string
	SpecCode string
	Count    int
}

type TopologyConfigRecord struct {
	Id          uint
	TenantId    string
	ClusterId   string
	ConfigModel *spec.Specification
	CreateTime  time.Time
}

type RecoverInfo struct {
	SourceClusterId string
	BackupRecordId  int64
}

func (r TopologyConfigRecord) Content() string {
	bytes, _ := json.Marshal(r.ConfigModel)
	return string(bytes)
}

