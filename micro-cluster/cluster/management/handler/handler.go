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

package handler

import (
	"bytes"
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"text/template"
)

type ClusterMeta struct {
	Cluster              *management.Cluster
	Instances            map[string][]*management.ClusterInstance
	NodeExporterPort     int32
	BlackboxExporterPort int32
}

// AddInstances
// @Description add new instances into cluster topology, then alloc host ip, port and disk for these instances
// @Parameter	computes
// @Return		error
func (meta *ClusterMeta) AddInstances(ctx context.Context, computes []structs.ClusterResourceParameterCompute) error {
	framework.LogWithContext(ctx).Infof("Add new instances into cluster topology, include [%d] components", len(computes))
	if len(computes) <= 0 {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "Cluster resource parameter is empty!")
	}

	if meta.Instances == nil {
		meta.Instances = make(map[string][]*management.ClusterInstance)
	}

	for _, compute := range computes {
		for _, item := range compute.Resource {
			for i := 0; i < item.Count; i++ {
				instance := &management.ClusterInstance{
					Type:         compute.Type,
					Version:      meta.Cluster.Version,
					ClusterID:    meta.Cluster.ID,
					CpuCores:     int8(knowledge.ParseCpu(item.Spec)),
					Memory:       int8(knowledge.ParseMemory(item.Spec)),
					Zone:         resource.GetDomainNameFromCode(item.Zone),
					DiskType:     item.DiskType,
					DiskCapacity: int32(item.DiskCapacity),
				}
				instance.Status = string(constants.InstanceInitializing)
				meta.Instances[compute.Type] = append(meta.Instances[compute.Type], instance)
			}
		}
	}
	return nil
}

func (meta *ClusterMeta) generateAllocRequirements(instances []*management.ClusterInstance) []resource.AllocRequirement {
	requirements := make([]resource.AllocRequirement, 0)
	for _, instance := range instances {
		portRange := knowledge.GetComponentPortRange(meta.Cluster.Type, meta.Cluster.Version, instance.Type)
		requirements = append(requirements, resource.AllocRequirement{
			Location: resource.Location{
				Region: meta.Cluster.Region,
				Zone:   instance.Zone,
			},
			Require: resource.Requirement{
				Exclusive: meta.Cluster.Exclusive,
				PortReq: []resource.PortRequirement{
					{
						Start:   int32(portRange.Start),
						End:     int32(portRange.End),
						PortCnt: int32(portRange.Count),
					},
				},
				DiskReq: resource.DiskRequirement{
					NeedDisk: true,
					Capacity: instance.DiskCapacity,
					DiskType: instance.DiskType,
				},
				ComputeReq: resource.ComputeRequirement{
					CpuCores: int32(instance.CpuCores),
					Memory:   int32(instance.Memory),
				},
			},
			Count: 1,
			HostFilter: resource.Filter{
				Arch: string(meta.Cluster.CpuArchitecture),
			},
			Strategy: resource.RandomRack,
		})
	}
	return requirements
}

func (meta *ClusterMeta) generateAllocMonitoredPortRequirements() []resource.AllocRequirement {
	requirements := make([]resource.AllocRequirement, 0)

	portRange := knowledge.GetClusterPortRange(meta.Cluster.Type, meta.Cluster.Version)
	requirements = append(requirements, resource.AllocRequirement{
		Location: resource.Location{Region: meta.Cluster.Region},
		Require: resource.Requirement{
			PortReq: []resource.PortRequirement{
				{
					Start:   int32(portRange.Start),
					End:     int32(portRange.End),
					PortCnt: int32(portRange.Count),
				},
			},
			DiskReq:    resource.DiskRequirement{NeedDisk: false},
			ComputeReq: resource.ComputeRequirement{CpuCores: 0, Memory: 0},
		},
		Count: 1,
		HostFilter: resource.Filter{
			Arch: string(meta.Cluster.CpuArchitecture),
		},
		Strategy: resource.ClusterPorts,
	})

	return requirements
}

// AllocInstanceResource
// @Description alloc host ip, port and disk for all new instances
// @Return		alloc request id
// @Return		error
func (meta *ClusterMeta) AllocInstanceResource(ctx context.Context) (string, error) {
	// Build alloc resource request
	requestID := uuidutil.GenerateID()
	instances := make([]*management.ClusterInstance, 0)
	for _, components := range meta.Instances {
		for _, instance := range components {
			if instance.Status != string(constants.InstanceInitializing) {
				continue
			}
			instances = append(instances, instance)
		}
	}
	request := &resource.BatchAllocRequest{
		BatchRequests: []resource.AllocReq{
			{
				Applicant: resource.Applicant{
					HolderId:  meta.Cluster.ID,
					RequestId: requestID,
				},
				Requires: meta.generateAllocRequirements(instances),
			},
			{
				Applicant: resource.Applicant{
					HolderId:  meta.Cluster.ID,
					RequestId: requestID,
				},
				Requires: meta.generateAllocMonitoredPortRequirements(),
			},
		},
	}

	// Alloc resource
	resourceManager := resourceManagement.NewResourceManager()
	response, err := resourceManager.AllocResources(ctx, request)
	if err != nil {
		framework.LogWithContext(ctx).Infof(
			"Cluster[%s] alloc resource error: %s", meta.Cluster.ID, err.Error())
		return "", err
	}

	// Set host ip, port and disk resource into new instances
	if len(response.BatchResults) != 2 || len(response.BatchResults[1].Results[0].PortRes[0].Ports) != 2 {
		return "", framework.NewTiEMError(common.TIEM_UNRECOGNIZED_ERROR, "Alloc resource is wrong!")
	}
	for i, instance := range instances {
		instance.HostID = response.BatchResults[0].Results[i].HostId
		instance.HostIP = append(instance.HostIP, response.BatchResults[0].Results[i].HostIp)
		instance.Ports = response.BatchResults[0].Results[i].PortRes[0].Ports
		instance.DiskID = response.BatchResults[0].Results[i].DiskRes.DiskId
		instance.DiskPath = response.BatchResults[0].Results[i].DiskRes.Path
	}
	meta.NodeExporterPort = response.BatchResults[1].Results[0].PortRes[0].Ports[0]
	meta.BlackboxExporterPort = response.BatchResults[1].Results[0].PortRes[0].Ports[1]

	return requestID, nil
}

// GenerateTopologyConfig
// @Description generate yaml config based on cluster topology
// @Return		yaml config
// @Return		error
func (meta *ClusterMeta) GenerateTopologyConfig(ctx context.Context) (string, error) {
	if meta.Cluster == nil || len(meta.Instances) == 0 {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "Cluster topology is empty, please check it!")
	}

	t, err := template.New("cluster_topology.yaml").ParseFiles("template/cluster_topology.yaml")
	if err != nil {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, meta); err != nil {
		return "", framework.NewTiEMError(common.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("Generate topology config: %s", topology.String())

	return topology.String(), nil
}

// SetClusterOnline
// @Description set cluster and all instances status into running
// @Return		error
func (meta *ClusterMeta) SetClusterOnline(ctx context.Context) error {
	meta.Cluster.Status = string(constants.ClusterRunning)
	for _, components := range meta.Instances {
		for _, instance := range components {
			if instance.Status == string(constants.InstanceInitializing) {
				instance.Status = string(constants.InstanceRunning)
			}
		}
	}
	// TODO: write db
	return nil
}

// SetClusterMaintenanceStatus
// @Description set cluster maintenance status
// @Return		error
func (meta *ClusterMeta) SetClusterMaintenanceStatus(ctx context.Context, status constants.ClusterMaintenanceStatus) error {
	meta.Cluster.MaintenanceStatus = status
	// TODO: write db
	framework.LogWithContext(ctx).Infof("Set cluster[%s] maintenance status into %s", meta.Cluster.ID, status)
	return nil
}

/*
// Persist
// @Description persist new instances and cluster info into db
// @Return		error
func (meta *ClusterMeta) Persist() error {
	return nil
}

func (p *ClusterMeta) CloneMeta(ctx context.Context) *ClusterMeta {
	return nil
}

func (p *ClusterMeta) TryMaintenance(ctx context.Context, maintenanceStatus constants.ClusterMaintenanceStatus) error {
	// check Maintenance status
	if true {
		p.cluster.MaintenanceStatus = maintenanceStatus
		return nil
	} else {
		return framework.NewTiEMError(common.TIEM_TASK_CONFLICT, "// todo")
	}
}

func (p *ClusterMeta) GetCluster() *management.Cluster {
	return p.cluster
}

func (p *ClusterMeta) GetID() string {
	return p.cluster.ID
}

func (p *ClusterMeta) GetInstances() []*management.ClusterInstance {
	return p.instances
}

func (p *ClusterMeta) GetDeployConfig() string {
	return ""
}

func (p *ClusterMeta) GetScaleInConfig() (string, error) {
	return "", nil
}

func (p *ClusterMeta) GetScaleOutConfig() (string, error) {
	return "", nil
}

func (p *ClusterMeta) GetConnectAddresses() []string {
	// got all tidb instances, then get connect addresses
	return nil
}

func (p *ClusterMeta) GetMonitorAddresses() []string {
	// got all tidb instances, then get connect addresses
	return nil
}

func Get(ctx context.Context, clusterID string) (*ClusterMeta, error) {
	cluster, instances, err := models.GetClusterReaderWriter().GetMeta(ctx, clusterID)
	// todo
	if err != nil {
		return &ClusterMeta{
			cluster: cluster,
			instances: instances,
		}, err
	}

	return nil, framework.WrapError(common.TIEM_CLUSTER_NOT_FOUND, "", err)
}

func Create(ctx context.Context, template management.Cluster) (*ClusterMeta, error) {
	return nil, nil
}
*/
