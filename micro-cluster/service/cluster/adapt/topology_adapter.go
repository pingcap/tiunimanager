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
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"path/filepath"
)

type DefaultTopologyPlanner struct {
}

func (d DefaultTopologyPlanner) BuildComponents(ctx context.Context, demands []*domain.ClusterComponentDemand, components []*domain.ComponentGroup, cluster *domain.Cluster) error {
	if len(demands) <= 0 {
		return fmt.Errorf("demands: [%v] is empty", demands)
	}

	for _, demand := range demands {
		var componentGroup domain.ComponentGroup
		componentGroup.ComponentType = demand.ComponentType
		nodes := make([]*domain.ComponentInstance, 0, demand.TotalNodeCount)
		for _, items := range demand.DistributionItems {
			for i := 0; i < items.Count; i++ {
				portRange := knowledge.GetComponentPortRange(cluster.ClusterType.Code, cluster.ClusterVersion.Code, demand.ComponentType.ComponentType)
				node := &domain.ComponentInstance{
					TenantId: cluster.TenantId,
					Status: domain.ClusterStatusUnlined,
					ClusterId: cluster.Id,
					ComponentType: demand.ComponentType,
					Version: &cluster.ClusterVersion,
					Location: &resource.Location{
						Region: cluster.ClusterDemand.Region,
						Zone: items.ZoneCode,
					},
					Compute: &resource.ComputeRequirement{
						CpuCores: int32(knowledge.ParseCpu(items.SpecCode)),
						Memory: int32(knowledge.ParseMemory(items.SpecCode)),
					},
					PortRequirement: &resource.PortRequirement{
						Start:	int32(portRange.Start),
						End: int32(portRange.End),
						PortCnt: int32(portRange.Count),
					},
				}
				nodes = append(nodes, node)
			}
		}

		componentGroup.Nodes = nodes
		components = append(components, &componentGroup)
	}

	return nil
}

func (d DefaultTopologyPlanner) AnalysisResourceRequest(ctx context.Context, cluster *domain.Cluster, components []*domain.ComponentGroup, takeover bool) (*clusterpb.BatchAllocRequest, error) {
	requirementList := make([]*clusterpb.AllocRequirement, 0)

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
					Count:      1,
					HostFilter: &clusterpb.Filter{},
					Strategy:   int32(resource.UserSpecifyHost),
				})
			} else {
				if instance.Status != domain.ClusterStatusUnlined {
					continue
				}
				portRequirementList = append(portRequirementList, &clusterpb.PortRequirement{
					Start: instance.PortRequirement.Start,
					End: instance.PortRequirement.End,
					PortCnt: instance.PortRequirement.PortCnt,
				})

				requirementList = append(requirementList, &clusterpb.AllocRequirement{
					Location: &clusterpb.Location{
						Region: instance.Location.Region,
						Zone: instance.Location.Zone,
					},
					Require: &clusterpb.Requirement{
						Exclusive: 	false,
						PortReq: portRequirementList,
						DiskReq: &clusterpb.DiskRequirement{NeedDisk: true},
						ComputeReq: &clusterpb.ComputeRequirement{
							CpuCores: instance.Compute.CpuCores,
							Memory: instance.Compute.Memory,
						},
					},
					Count: 1,
					HostFilter: &clusterpb.Filter{},
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
					RequestId: uuidutil.GenerateID(),
				},

				Requires: requirementList,
			},
		},
	}

	return allocReq, nil
}

func (d DefaultTopologyPlanner) ApplyResourceToComponents(ctx context.Context, response *clusterpb.BatchAllocResponse, components []*domain.ComponentGroup) error {
	// handle response error
	if response.Rs.Code != 0 {
		return fmt.Errorf(response.Rs.Message)
	}

	// get host resource
	var count int
	for _, component := range components {
		for _, instance := range component.Nodes {
			if instance.Status != domain.ClusterStatusUnlined {
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

func (d DefaultTopologyPlanner) GenerateTopologyConfig(ctx context.Context, components []*domain.ComponentGroup, cluster *domain.Cluster) (*spec.Specification, error) {
	if len(components) <= 0 {
		return nil, fmt.Errorf("components is empty")
	}
	tiupConfig := new(spec.Specification)

	if cluster.Status == domain.ClusterStatusUnlined { // create cluster
		// Deal with Global Settings
		tiupConfig.GlobalOptions.DataDir = filepath.Join(cluster.Id, "tidb-data")
		tiupConfig.GlobalOptions.DeployDir = filepath.Join(cluster.Id, "tidb-deploy")
		tiupConfig.GlobalOptions.User = "tidb"
		tiupConfig.GlobalOptions.SSHPort = 22
		tiupConfig.GlobalOptions.Arch = "amd64"
		tiupConfig.GlobalOptions.LogDir = filepath.Join(cluster.Id, "tidb-log")
	}

	for _, component := range components {
		for _, instance := range component.Nodes {
			if instance.Status != domain.ClusterStatusUnlined {
				continue
			}
			if component.ComponentType.ComponentType == "TiDB" {
				tiupConfig.TiDBServers = append(tiupConfig.TiDBServers, &spec.TiDBSpec{
					Host: instance.Host,
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "tidb-deploy"),
					Port: domain.DefaultTidbPort,
				})

			} else if component.ComponentType.ComponentType == "TiKV" {
				tiupConfig.TiKVServers = append(tiupConfig.TiKVServers, &spec.TiKVSpec{
					Host: instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "tikv-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "tikv-deploy"),
				})
			} else if component.ComponentType.ComponentType == "PD" {
				tiupConfig.PDServers = append(tiupConfig.PDServers, &spec.PDSpec{
					Host: instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "pd-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "pd-deploy"),
				})
			} else if component.ComponentType.ComponentType == "TiFlash" {
				tiupConfig.TiFlashServers = append(tiupConfig.TiFlashServers, &spec.TiFlashSpec{
					Host: instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "tiflash-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "tiflash-deploy"),
				})
			} else if component.ComponentType.ComponentType == "TiCDC" {
				tiupConfig.CDCServers = append(tiupConfig.CDCServers, &spec.CDCSpec{
					Host: instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "cdc-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "cdc-deploy"),
					LogDir: filepath.Join(instance.DiskPath, cluster.Id, "cdc-log"),
				})
			} else if component.ComponentType.ComponentType == "Grafana" {
				tiupConfig.Grafanas = append(tiupConfig.Grafanas, &spec.GrafanaSpec{
					Host:            instance.Host,
					DeployDir:       filepath.Join(instance.DiskPath, cluster.Id, "grafanas-deploy"),
					AnonymousEnable: true,
					DefaultTheme:    "light",
					OrgName:         "Main Org.",
					OrgRole:         "Viewer",
				})
			} else if component.ComponentType.ComponentType == "Prometheus" {
				tiupConfig.Monitors = append(tiupConfig.Monitors, &spec.PrometheusSpec{
					Host:      instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "prometheus-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "prometheus-deploy"),
				})
			} else if component.ComponentType.ComponentType == "AlertManger" {
				tiupConfig.Alertmanagers = append(tiupConfig.Alertmanagers, &spec.AlertmanagerSpec{
					Host:      instance.Host,
					DataDir:   filepath.Join(instance.DiskPath, cluster.Id, "alertmanagers-data"),
					DeployDir: filepath.Join(instance.DiskPath, cluster.Id, "alertmanagers-deploy"),
				})
			}
		}
	}

	return tiupConfig, nil
}
