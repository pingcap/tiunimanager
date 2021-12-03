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
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
	"github.com/pingcap-inc/tiem/micro-api/controller/resource/warehouse"
	"math/rand"
	"time"
)

type ComponentGroup struct {
	ComponentType *knowledge.ClusterComponent
	Nodes         []*ComponentInstance
}

type ComponentInstance struct {
	ID string

	Code     string
	TenantId string

	Status        ClusterStatus
	ClusterId     string
	ComponentType *knowledge.ClusterComponent

	Role    string
	Version *knowledge.ClusterVersion

	HostId         string
	Host           string
	PortList       []int
	DiskId         string
	AllocRequestId string
	DiskPath       string

	Location        *resource.Location
	Compute         *resource.ComputeRequirement
	PortRequirement *resource.PortRequirement

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (c *ComponentInstance) SerializePortInfo() string {
	if c.PortList == nil || len(c.PortList) == 0 {
		return "[]"
	}
	portInfoBytes, err := json.Marshal(c.PortList)
	if err != nil {
		framework.Log().Errorf("serialize PortInfo error, %v", c.PortList)
		return "[]"
	}

	return string(portInfoBytes)
}

func (c *ComponentInstance) DeserializePortInfo(portInfo string) {
	c.PortList = make([]int, 0)
	if portInfo == "" {
		return
	}
	err := json.Unmarshal([]byte(portInfo), &c.PortList)
	if err != nil {
		framework.Log().Errorf("deserialize PortInfo error, %s, %s", portInfo, err.Error())
	}
}

func (p *ComponentInstance) SetLocation(location *resource.Location) {
	p.Location = location
}

func (p *ComponentInstance) SetDiskPath(location *resource.Location) {
	p.Location = location
}

// MockUsage TODO will be replaced with monitor implement
func MockUsage() controller.Usage {
	usage := controller.Usage{
		Total: 100,
		Used:  float32(rand.Intn(100)),
	}
	usage.UsageRate = usage.Used / usage.Total
	return usage
}

func mockRole() management.ComponentNodeRole {
	return management.ComponentNodeRole{
		RoleCode: "Leader",
		RoleName: "Flower",
	}
}

func mockSpec() warehouse.SpecBaseInfo {
	return warehouse.SpecBaseInfo{
		SpecCode: knowledge.GenSpecCode(4, 8),
		SpecName: knowledge.GenSpecCode(4, 8),
	}
}

func mockZone() warehouse.ZoneBaseInfo {
	return warehouse.ZoneBaseInfo{
		ZoneCode: "TEST_Zone1",
		ZoneName: "TEST_Zone1",
	}
}

func mockIops() []float32 {
	return []float32{10, 20}
}

func mockIoUtil() float32 {
	return 1
}
