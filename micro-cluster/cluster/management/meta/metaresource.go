/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: metaresource.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/11
*******************************************************************************/

package meta

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	resource "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/platform/product"
)

func (p *ClusterMeta) GenerateInstanceResourceRequirements(ctx context.Context) ([]resource.AllocRequirement, []*management.ClusterInstance, error) {
	instances := p.GetInstanceByStatus(ctx, constants.ClusterInstanceInitializing)
	requirements := make([]resource.AllocRequirement, 0)

	_, _, components, err := models.GetProductReaderWriter().GetProduct(ctx, p.Cluster.Type)
	if err != nil {
		err = errors.WrapError(errors.TIEM_UNSUPPORT_PRODUCT, "get product failed", err)
		return nil, nil, err
	}

	componentsMap := make(map[string]*product.ProductComponentInfo)
	for _, c := range components {
		componentsMap[c.ComponentID] = c
	}

	allocInstances := make([]*management.ClusterInstance, 0)
	for _, instance := range instances {
		var componentInfo product.ProductComponentInfo
		if Contain(constants.ParasiteComponentIDs, constants.EMProductComponentIDType(instance.Type)) {
			continue
		}
		if c, ok := componentsMap[instance.Type]; !ok {
			return nil, nil, errors.NewError(errors.TIEM_UNSUPPORT_PRODUCT, "")
		} else {
			componentInfo = *c
		}
		requirements = append(requirements, resource.AllocRequirement{
			Location: structs.Location{
				Region: p.Cluster.Region,
				Zone:   instance.Zone,
			},
			Require: resource.Requirement{
				Exclusive: p.Cluster.Exclusive,
				PortReq: []resource.PortRequirement{
					{
						Start:   componentInfo.StartPort,
						End:     componentInfo.EndPort,
						PortCnt: componentInfo.MaxPort,
					},
				},
				DiskReq: resource.DiskRequirement{
					NeedDisk: true,
					Capacity: instance.DiskCapacity,
					DiskType: instance.DiskType,
				},
				ComputeReq: resource.ComputeRequirement{
					ComputeResource: resource.ComputeResource{
						CpuCores: int32(instance.CpuCores),
						Memory:   int32(instance.Memory),
					},
				},
			},
			Count: 1,
			HostFilter: resource.Filter{
				Arch: string(p.Cluster.CpuArchitecture),
			},
			Strategy: resource.RandomRack,
		})
		allocInstances = append(allocInstances, instance)
	}
	return requirements, allocInstances, nil
}

func (p *ClusterMeta) GenerateTakeoverResourceRequirements(ctx context.Context) ([]resource.AllocRequirement, []*management.ClusterInstance) {
	instances := p.GetInstanceByStatus(ctx, constants.ClusterInstanceRunning)
	requirements := make([]resource.AllocRequirement, 0)

	allocInstances := make([]*management.ClusterInstance, 0)
	for _, instance := range instances {
		portRequirements := make([]resource.PortRequirement, 0)
		for _, port := range instance.Ports {
			portRequirements = append(portRequirements, resource.PortRequirement{
				Start:   port,
				End:     port + 1,
				PortCnt: int32(1),
			})
		}
		requirements = append(requirements, resource.AllocRequirement{
			Location: structs.Location{
				HostIp: instance.HostIP[0],
			},
			Require: resource.Requirement{
				Exclusive:  p.Cluster.Exclusive,
				PortReq:    portRequirements,
				DiskReq:    resource.DiskRequirement{},
				ComputeReq: resource.ComputeRequirement{},
			},
			Count:    1,
			Strategy: resource.UserSpecifyHost,
		})
		allocInstances = append(allocInstances, instance)
	}
	return requirements, allocInstances
}

func (p *ClusterMeta) GenerateGlobalPortRequirements(ctx context.Context) ([]resource.AllocRequirement, error) {
	requirements := make([]resource.AllocRequirement, 0)

	if p.Cluster.Status != string(constants.ClusterInitializing) {
		framework.LogWithContext(ctx).Infof("cluster %s is not initializing, no need to alloc global port resource", p.Cluster.ID)
		return requirements, nil
	}

	portRange, err := getRetainedPortRange(ctx)
	if err != nil {
		return nil, err
	}
	requirements = append(requirements, resource.AllocRequirement{
		Location: structs.Location{Region: p.Cluster.Region},
		Require: resource.Requirement{
			PortReq: []resource.PortRequirement{
				{
					Start:   int32(portRange[0]),
					End:     int32(portRange[1]),
					PortCnt: int32(2),
				},
			},
			DiskReq: resource.DiskRequirement{NeedDisk: false},
			ComputeReq: resource.ComputeRequirement{
				ComputeResource: resource.ComputeResource{},
			},
		},
		Count: 1,
		HostFilter: resource.Filter{
			Arch: string(p.Cluster.CpuArchitecture),
		},
		Strategy: resource.ClusterPorts,
	})

	return requirements, nil
}

func (p *ClusterMeta) ApplyGlobalPortResource(nodeExporterPort, blackboxExporterPort int32) {
	p.NodeExporterPort = nodeExporterPort
	p.BlackboxExporterPort = blackboxExporterPort
}

func (p *ClusterMeta) ApplyInstanceResource(resource *resource.AllocRsp, instances []*management.ClusterInstance) {
	for i, instance := range instances {
		instance.HostID = resource.Results[i].HostId
		instance.HostIP = append(instance.HostIP, resource.Results[i].HostIp)
		instance.Ports = resource.Results[i].PortRes[0].Ports
		instance.DiskID = resource.Results[i].DiskRes.DiskId
		instance.DiskPath = resource.Results[i].DiskRes.Path
		instance.Rack = resource.Results[i].Location.Rack
	}
	if p.Cluster.Status == string(constants.ClusterInitializing) {
		pd := p.Instances[string(constants.ComponentIDPD)][0]
		for _, t := range constants.ParasiteComponentIDs {
			instance := p.Instances[string(t)][0]
			instance.HostID = pd.HostID
			instance.HostIP = pd.HostIP
			instance.DiskID = pd.DiskID
			instance.DiskPath = pd.DiskPath
			switch t {
			case constants.ComponentIDGrafana:
				instance.Ports = pd.Ports[3:4]
				continue
			case constants.ComponentIDPrometheus:
				instance.Ports = pd.Ports[2:3]
				continue
			case constants.ComponentIDAlertManger:
				instance.Ports = pd.Ports[4:6]
				continue
			default:
			}
		}
		pd.Ports = pd.Ports[:2]
	}
}
