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
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"math/rand"
	"strconv"
	"time"
)

type ComponentGroup struct {
	ComponentType *knowledge.ClusterComponent
	Nodes         []ComponentInstance
}

type ComponentInstance struct {
	ID string

	Code     string
	TenantId string

	Status        ClusterStatus
	ClusterId     string
	ComponentType *knowledge.ClusterComponent

	Role     string
	Version  *knowledge.ClusterVersion

	HostId         string
	Host           string
	PortList       []int
	DiskId         string
	PortInfo       string
	AllocRequestId string

	location *resource.Location
	diskPath string
	compute *resource.ComputeRequirement
	portRequirement *resource.PortRequirement

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (p *ComponentInstance) SetLocation (location *resource.Location) {
	p.location = location
}

func (p *ComponentInstance) SetDiskPath (location *resource.Location) {
	p.location = location
}


func (aggregation *ClusterAggregation) ExtractInstancesDTO() *clusterpb.ClusterInstanceDTO {
	dto := &clusterpb.ClusterInstanceDTO{
		Whitelist:       []string{},
		DiskUsage:       MockUsage(),
		CpuUsage:        MockUsage(),
		MemoryUsage:     MockUsage(),
		StorageUsage:    MockUsage(),
		BackupFileUsage: MockUsage(),
	}

	if record := aggregation.CurrentTopologyConfigRecord; aggregation.CurrentTopologyConfigRecord != nil && record.ConfigModel != nil {
		dto.IntranetConnectAddresses, dto.ExtranetConnectAddresses, dto.PortList = ConnectAddresses(record.ConfigModel)
	} else {
		dto.PortList = []int64{4000}
		dto.IntranetConnectAddresses = []string{"127.0.0.1:4000"}
		dto.ExtranetConnectAddresses = []string{"127.0.0.1:4000"}
	}

	return dto
}

func ConnectAddresses(spec *spec.Specification) ([]string, []string, []int64) {
	servers := spec.TiDBServers

	addressList := make([]string, 0)
	portList := make([]int64, 0)

	for _, v := range servers {
		addressList = append(addressList, v.Host+":"+strconv.Itoa(v.Port))
	}
	return addressList, addressList, portList
}

func (aggregation *ClusterAggregation) ExtractComponentDTOs() []*clusterpb.ComponentInstanceDTO {
	if record := aggregation.CurrentTopologyConfigRecord; aggregation.CurrentTopologyConfigRecord != nil && record.ConfigModel != nil {
		config := record.ConfigModel
		knowledge := knowledge.ClusterTypeSpecFromCode(aggregation.Cluster.ClusterType.Code)
		for _, v := range knowledge.VersionSpecs {
			if v.ClusterVersion.Code == aggregation.Cluster.ClusterVersion.Code {
				return appendAllComponentInstances(config, &v)
			}
		}
	}
	return make([]*clusterpb.ComponentInstanceDTO, 0)
}

func appendAllComponentInstances(config *spec.Specification, knowledge *knowledge.ClusterVersionSpec) []*clusterpb.ComponentInstanceDTO {
	components := make([]*clusterpb.ComponentInstanceDTO, 0, len(knowledge.ComponentSpecs))

	for _, v := range knowledge.ComponentSpecs {
		code := v.ClusterComponent.ComponentType
		componentDTO := &clusterpb.ComponentInstanceDTO{
			BaseInfo: &clusterpb.ComponentBaseInfoDTO{
				ComponentType: code,
				ComponentName: v.ClusterComponent.ComponentName,
			},
			Nodes: ComponentAppender[code](config, knowledge.ClusterVersion.Code),
		}

		components = append(components, componentDTO)
	}
	return components
}

var ComponentAppender = map[string]func(*spec.Specification, string) []*clusterpb.ComponentNodeDisplayInfoDTO{
	"TiDB":    tiDBComponent,
	"TiKV":    tiKVComponent,
	"PD":      pDComponent,
	"TiFlash": tiFlashComponent,
	//"TiCDC": tiCDCComponent,
}

func tiDBComponent(config *spec.Specification, version string) []*clusterpb.ComponentNodeDisplayInfoDTO {
	servers := config.TiDBServers
	dto := make([]*clusterpb.ComponentNodeDisplayInfoDTO, len(servers))
	for i, v := range servers {
		dto[i] = &clusterpb.ComponentNodeDisplayInfoDTO{
			NodeId:  v.Host,
			Version: version, // todo
			Status:  "运行中",   // todo
			Instance: &clusterpb.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port:   int32(v.Port),
				Role:   mockRole(),
				Spec:   mockSpec(),
				Zone:   mockZone(),
			},

			Usages: &clusterpb.ComponentNodeUsageDTO{
				IoUtil:       mockIoUtil(),
				Iops:         mockIops(),
				CpuUsage:     MockUsage(),
				MemoryUsage:  MockUsage(),
				StoregeUsage: MockUsage(),
			},
		}
	}
	return dto
}

func tiKVComponent(config *spec.Specification, version string) []*clusterpb.ComponentNodeDisplayInfoDTO {
	servers := config.TiKVServers
	dto := make([]*clusterpb.ComponentNodeDisplayInfoDTO, len(servers))
	for i, v := range servers {
		dto[i] = &clusterpb.ComponentNodeDisplayInfoDTO{
			NodeId:  v.Host,
			Version: version, // todo
			Status:  "运行中",   // todo
			Instance: &clusterpb.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port:   20160,
				Role:   mockRole(),
				Spec:   mockSpec(),
				Zone:   mockZone(),
			},

			Usages: &clusterpb.ComponentNodeUsageDTO{
				IoUtil:       mockIoUtil(),
				Iops:         mockIops(),
				CpuUsage:     MockUsage(),
				MemoryUsage:  MockUsage(),
				StoregeUsage: MockUsage(),
			},
		}
	}
	return dto
}

func pDComponent(config *spec.Specification, version string) []*clusterpb.ComponentNodeDisplayInfoDTO {
	servers := config.PDServers
	dto := make([]*clusterpb.ComponentNodeDisplayInfoDTO, len(servers))
	for i, v := range servers {
		dto[i] = &clusterpb.ComponentNodeDisplayInfoDTO{
			NodeId:  v.Host,
			Version: version, // todo
			Status:  "运行中",   // todo
			Instance: &clusterpb.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port:   2379,
				Role:   mockRole(),
				Spec:   mockSpec(),
				Zone:   mockZone(),
			},

			Usages: &clusterpb.ComponentNodeUsageDTO{
				IoUtil:       mockIoUtil(),
				Iops:         mockIops(),
				CpuUsage:     MockUsage(),
				MemoryUsage:  MockUsage(),
				StoregeUsage: MockUsage(),
			},
		}
	}
	return dto
}

func tiFlashComponent(config *spec.Specification, version string) []*clusterpb.ComponentNodeDisplayInfoDTO {
	dto := make([]*clusterpb.ComponentNodeDisplayInfoDTO, 0)
	return dto
}

// MockUsage TODO will be replaced with monitor implement
func MockUsage() *clusterpb.UsageDTO {
	usage := &clusterpb.UsageDTO{
		Total: 100,
		Used:  float32(rand.Intn(100)),
	}
	usage.UsageRate = usage.Used / usage.Total
	return usage
}

func mockRole() *clusterpb.ComponentNodeRoleDTO {
	return &clusterpb.ComponentNodeRoleDTO{
		RoleCode: "Leader",
		RoleName: "Flower",
	}
}

func mockSpec() *clusterpb.SpecBaseInfoDTO {
	return &clusterpb.SpecBaseInfoDTO{
		SpecCode: knowledge.GenSpecCode(4, 8),
		SpecName: knowledge.GenSpecCode(4, 8),
	}
}

func mockZone() *clusterpb.ZoneBaseInfoDTO {
	return &clusterpb.ZoneBaseInfoDTO{
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
