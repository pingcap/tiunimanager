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

package adapt

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	resourceService "github.com/pingcap-inc/tiem/micro-cluster/service/resource"

	"text/template"
)

type DefaultTopologyPlanner struct {
}

func (d DefaultTopologyPlanner) BuildComponents(ctx context.Context, demands []*domain.ClusterComponentDemand, cluster *domain.Cluster) ([]*domain.ComponentGroup, error) {
	if len(demands) <= 0 {
		return nil, fmt.Errorf("demands: [%v] is empty", demands)
	}

	components := make([]*domain.ComponentGroup, 0)
	for _, demand := range demands {
		var componentGroup domain.ComponentGroup
		componentGroup.ComponentType = demand.ComponentType
		nodes := make([]*domain.ComponentInstance, 0, demand.TotalNodeCount)
		if len(demand.DistributionItems) > 0 {
			for _, items := range demand.DistributionItems {
				for i := 0; i < items.Count; i++ {
					portRange := knowledge.GetComponentPortRange(cluster.ClusterType.Code, cluster.ClusterVersion.Code, demand.ComponentType.ComponentType)
					node := &domain.ComponentInstance{
						TenantId:      cluster.TenantId,
						Status:        domain.ClusterStatusUnlined,
						ClusterId:     cluster.Id,
						ComponentType: demand.ComponentType,
						Version:       &cluster.ClusterVersion,
						Location: &resource.Location{
							Region: resource.GetDomainPrefixFromCode(items.ZoneCode),
							Zone:   resource.GetDomainNameFromCode(items.ZoneCode),
						},
						Compute: &resource.ComputeRequirement{
							CpuCores: int32(knowledge.ParseCpu(items.SpecCode)),
							Memory:   int32(knowledge.ParseMemory(items.SpecCode)),
						},
						PortRequirement: &resource.PortRequirement{
							Start:   int32(portRange.Start),
							End:     int32(portRange.End),
							PortCnt: int32(portRange.Count),
						},
					}
					nodes = append(nodes, node)
				}
			}
		} else {
			node := &domain.ComponentInstance{
				TenantId:        cluster.TenantId,
				Status:          domain.ClusterStatusUnlined,
				ClusterId:       cluster.Id,
				ComponentType:   demand.ComponentType,
				Version:         &cluster.ClusterVersion,
				Location:        &resource.Location{},
				Compute:         &resource.ComputeRequirement{},
				PortRequirement: &resource.PortRequirement{},
			}
			nodes = append(nodes, node)
		}

		componentGroup.Nodes = nodes
		components = append(components, &componentGroup)
	}

	return components, nil
}

func (d DefaultTopologyPlanner) AnalysisResourceRequest(ctx context.Context, cluster *domain.Cluster, components []*domain.ComponentGroup, takeover bool) (*clusterpb.BatchAllocRequest, error) {
	requirementList := make([]*clusterpb.AllocRequirement, 0)

	requestId := uuidutil.GenerateID()

	for _, component := range components {
		for _, instance := range component.Nodes {
			portRequirementList := make([]*clusterpb.PortRequirement, 0)
			if takeover { // takeover cluster
				for _, port := range instance.PortList {
					portRequirementList = append(portRequirementList, &clusterpb.PortRequirement{
						Start:   int32(port),
						End:     int32(port + 1),
						PortCnt: 1,
					})
				}
				requirementList = append(requirementList, &clusterpb.AllocRequirement{
					Location: &clusterpb.Location{Host: instance.Host},
					Require: &clusterpb.Requirement{
						Exclusive:  false,
						PortReq:    portRequirementList,
						DiskReq:    &clusterpb.DiskRequirement{NeedDisk: false},
						ComputeReq: &clusterpb.ComputeRequirement{CpuCores: 0, Memory: 0},
					},
					Count: 1,
					HostFilter: &clusterpb.Filter{
						Arch: cluster.CpuArchitecture,
					},
					Strategy: int32(resource.UserSpecifyHost),
				})
			} else {
				// no need to alloc for existed component or parasite component
				if instance.Status != domain.ClusterStatusUnlined ||
					knowledge.IsParasite(cluster.ClusterType.Code, cluster.ClusterVersion.Code, instance.ComponentType.ComponentType) {
					continue
				}

				instance.AllocRequestId = requestId

				portRequirementList = append(portRequirementList, &clusterpb.PortRequirement{
					Start:   instance.PortRequirement.Start,
					End:     instance.PortRequirement.End,
					PortCnt: instance.PortRequirement.PortCnt,
				})

				requirementList = append(requirementList, &clusterpb.AllocRequirement{
					Location: &clusterpb.Location{
						Region: instance.Location.Region,
						Zone:   instance.Location.Zone,
					},
					Require: &clusterpb.Requirement{
						Exclusive: false,
						PortReq:   portRequirementList,
						DiskReq:   &clusterpb.DiskRequirement{NeedDisk: true},
						ComputeReq: &clusterpb.ComputeRequirement{
							CpuCores: instance.Compute.CpuCores,
							Memory:   instance.Compute.Memory,
						},
					},
					Count: 1,
					HostFilter: &clusterpb.Filter{
						Arch: cluster.CpuArchitecture,
					},
					Strategy: int32(resource.RandomRack),
				})
			}
		}
	}

	allocReq := &clusterpb.BatchAllocRequest{
		BatchRequests: []*clusterpb.AllocRequest{
			{
				Applicant: &clusterpb.Applicant{
					HolderId:  cluster.Id,
					RequestId: requestId,
				},

				Requires: requirementList,
			},
		},
	}

	return allocReq, nil
}

func (d DefaultTopologyPlanner) ApplyResourceToComponents(ctx context.Context, cluster *domain.Cluster, response *clusterpb.BatchAllocResponse, components []*domain.ComponentGroup) error {
	// handle response error
	if response.Rs.Code != 0 {
		return fmt.Errorf(response.Rs.Message)
	}

	// get host resource
	var count int
	for _, component := range components {
		for _, instance := range component.Nodes {
			if instance.Status != domain.ClusterStatusUnlined ||
				knowledge.IsParasite(cluster.ClusterType.Code, cluster.ClusterVersion.Code, instance.ComponentType.ComponentType) {
				continue
			}

			if len(response.BatchResults) <= 0 {
				return fmt.Errorf("alloc resources is empty")
			}
			portList := make([]int, 0)
			for _, port := range response.BatchResults[0].Results[count].PortRes[0].Ports {
				portList = append(portList, int(port))
			}
			instance.HostId = response.BatchResults[0].Results[count].HostId
			instance.Host = response.BatchResults[0].Results[count].HostIp
			instance.PortList = portList
			instance.DiskId = response.BatchResults[0].Results[count].DiskRes.DiskId
			instance.DiskPath = response.BatchResults[0].Results[count].DiskRes.Path
			count += 1
		}
	}

	return nil
}

func (d DefaultTopologyPlanner) getClusterPorts(ctx context.Context, cluster *domain.Cluster, requestId string) (ports []int, err error) {
	clusterPortRange := knowledge.GetClusterPortRange(cluster.ClusterType.Code, cluster.ClusterVersion.Code)
	portRequirementList := make([]*clusterpb.PortRequirement, 0)
	portRequirementList = append(portRequirementList, &clusterpb.PortRequirement{
		Start:   int32(clusterPortRange.Start),
		End:     int32(clusterPortRange.End),
		PortCnt: int32(clusterPortRange.Count),
	})
	var requirementList []*clusterpb.AllocRequirement
	requirementList = append(requirementList, &clusterpb.AllocRequirement{
		Location: &clusterpb.Location{Region: cluster.Region},
		Require: &clusterpb.Requirement{
			PortReq:    portRequirementList,
			DiskReq:    &clusterpb.DiskRequirement{NeedDisk: false},
			ComputeReq: &clusterpb.ComputeRequirement{CpuCores: 0, Memory: 0},
		},
		Count: 1,
		HostFilter: &clusterpb.Filter{
			Arch: cluster.CpuArchitecture,
		},
		Strategy: int32(resource.ClusterPorts),
	})

	allocReq := &clusterpb.BatchAllocRequest{
		BatchRequests: []*clusterpb.AllocRequest{
			{
				Applicant: &clusterpb.Applicant{
					HolderId:  cluster.Id,
					RequestId: requestId,
				},

				Requires: requirementList,
			},
		},
	}

	var rsp clusterpb.BatchAllocResponse
	resourceManager := resourceService.NewResourceManager()
	err = resourceManager.AllocResourcesInBatch(ctx, allocReq, &rsp)
	if err != nil {
		return nil, err
	} else if rsp.Rs.Code != 0 {
		err = framework.NewTiEMErrorf(common.TIEM_CLUSTER_GET_CLUSTER_PORT_ERROR, rsp.Rs.Message)
		return nil, err
	}

	portRes := rsp.BatchResults[0].Results[0].PortRes[0]
	for _, port := range portRes.Ports {
		ports = append(ports, int(port))
	}
	return ports, nil
}

func (d DefaultTopologyPlanner) GenerateTopologyConfig(ctx context.Context, components []*domain.ComponentGroup, cluster *domain.Cluster) (string, error) {
	if len(components) <= 0 {
		return "", fmt.Errorf("components is empty")
	}

	t, err := template.New("cluster_topology.yaml").ParseFiles("template/cluster_topology.yaml")
	if err != nil {
		return "", err
	}
	requestId := components[0].Nodes[0].AllocRequestId
	ports, err := d.getClusterPorts(ctx, cluster, requestId)
	if err != nil {
		return "", err
	}
	topology := new(bytes.Buffer)
	data := struct {
		Cluster              *domain.Cluster
		Components           []*domain.ComponentGroup
		NodeExporterPort     int
		BlackboxExporterPort int
	}{
		cluster,
		components,
		ports[0],
		ports[1],
	}
	if err = t.Execute(topology, data); err != nil {
		return "", err
	}

	return topology.String(), nil
}
