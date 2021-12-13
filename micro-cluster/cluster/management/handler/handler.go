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
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"strconv"
	"strings"
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
func (p *ClusterMeta) AddInstances(ctx context.Context, computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "cluster resource parameter is empty!")
	}

	if p.Instances == nil {
		p.Instances = make(map[string][]*management.ClusterInstance)
	}

	for _, compute := range computes {
		for _, item := range compute.Resource {
			for i := 0; i < item.Count; i++ {
				instance := &management.ClusterInstance{
					Type:         compute.Type,
					Version:      p.Cluster.Version,
					ClusterID:    p.Cluster.ID,
					CpuCores:     int8(knowledge.ParseCpu(item.Spec)),
					Memory:       int8(knowledge.ParseMemory(item.Spec)),
					Zone:         resource.GetDomainNameFromCode(item.Zone),
					DiskType:     item.DiskType,
					DiskCapacity: int32(item.DiskCapacity),
				}
				instance.Status = string(constants.ClusterInitializing)
				p.Instances[compute.Type] = append(p.Instances[compute.Type], instance)
			}
		}
	}
	framework.LogWithContext(ctx).Infof("add new instances into cluster[%s] topology", p.Cluster.Name)
	return nil
}

func (p *ClusterMeta) generateAllocRequirements(instances []*management.ClusterInstance) []resource.AllocRequirement {
	requirements := make([]resource.AllocRequirement, 0)
	for _, instance := range instances {
		portRange := knowledge.GetComponentPortRange(p.Cluster.Type, p.Cluster.Version, instance.Type)
		requirements = append(requirements, resource.AllocRequirement{
			Location: resource.Location{
				Region: p.Cluster.Region,
				Zone:   instance.Zone,
			},
			Require: resource.Requirement{
				Exclusive: p.Cluster.Exclusive,
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
				Arch: string(p.Cluster.CpuArchitecture),
			},
			Strategy: resource.RandomRack,
		})
	}
	return requirements
}

func (p *ClusterMeta) generateAllocMonitoredPortRequirements() []resource.AllocRequirement {
	requirements := make([]resource.AllocRequirement, 0)

	portRange := knowledge.GetClusterPortRange(p.Cluster.Type, p.Cluster.Version)
	requirements = append(requirements, resource.AllocRequirement{
		Location: resource.Location{Region: p.Cluster.Region},
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
			Arch: string(p.Cluster.CpuArchitecture),
		},
		Strategy: resource.ClusterPorts,
	})

	return requirements
}

// AllocInstanceResource
// @Description alloc host ip, port and disk for all new instances
// @Return		alloc request id
// @Return		error
func (p *ClusterMeta) AllocInstanceResource(ctx context.Context) (string, error) {
	requestID := uuidutil.GenerateID()
	instances := make([]*management.ClusterInstance, 0)
	for _, components := range p.Instances {
		for _, instance := range components {
			if instance.Status != string(constants.ClusterInitializing) {
				continue
			}
			instances = append(instances, instance)
		}
	}

	// Alloc instances resource
	resourceManager := resourceManagement.NewResourceManager()
	request := &resource.BatchAllocRequest{
		BatchRequests: []resource.AllocReq{
			{
				Applicant: resource.Applicant{
					HolderId:  p.Cluster.ID,
					RequestId: requestID,
				},
				Requires: p.generateAllocRequirements(instances),
			},
		},
	}
	response, err := resourceManager.AllocResources(ctx, request)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] alloc instances resource error: %s", p.Cluster.Name, err.Error())
		return "", err
	}
	for i, instance := range instances {
		instance.HostID = response.BatchResults[0].Results[i].HostId
		instance.HostIP = append(instance.HostIP, response.BatchResults[0].Results[i].HostIp)
		instance.Ports = response.BatchResults[0].Results[i].PortRes[0].Ports
		instance.DiskID = response.BatchResults[0].Results[i].DiskRes.DiskId
		instance.DiskPath = response.BatchResults[0].Results[i].DiskRes.Path
	}

	// Alloc monitored ports resource
	request = &resource.BatchAllocRequest{
		BatchRequests: []resource.AllocReq{
			{
				Applicant: resource.Applicant{
					HolderId:  p.Cluster.ID,
					RequestId: requestID,
				},
				Requires: p.generateAllocMonitoredPortRequirements(),
			},
		},
	}
	response, err = resourceManager.AllocResources(ctx, request)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"cluster[%s] alloc monitored ports resource error: %s", p.Cluster.Name, err.Error())
		return "", err
	}

	p.NodeExporterPort = response.BatchResults[0].Results[0].PortRes[0].Ports[0]
	p.BlackboxExporterPort = response.BatchResults[0].Results[0].PortRes[0].Ports[1]

	return requestID, nil
}

// GenerateTopologyConfig
// @Description generate yaml config based on cluster topology
// @Return		yaml config
// @Return		error
func (p *ClusterMeta) GenerateTopologyConfig(ctx context.Context) (string, error) {
	if p.Cluster == nil || len(p.Instances) == 0 {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "cluster topology is empty, please check it!")
	}

	t, err := template.New("cluster_topology.yaml").ParseFiles("template/cluster_topology.yaml")
	if err != nil {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", framework.NewTiEMError(common.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

// UpdateClusterStatus
// @Description update cluster status
// @Return		error
func (p *ClusterMeta) UpdateClusterStatus(ctx context.Context, status constants.ClusterRunningStatus) error {
	p.Cluster.Status = string(status)
	// TODO: write db
	framework.LogWithContext(ctx).Infof("update cluster[%s] status into %s", p.Cluster.Name, status)
	return nil
}

// UpdateClusterMaintenanceStatus
// @Description update cluster maintenance status
// @Return		error
func (p *ClusterMeta) UpdateClusterMaintenanceStatus(ctx context.Context, status constants.ClusterMaintenanceStatus) error {
	p.Cluster.MaintenanceStatus = status
	// TODO: write db
	framework.LogWithContext(ctx).Infof("update cluster[%s] maintenance status into %s", p.Cluster.Name, status)
	return nil
}

// UpdateInstancesStatus
// @Description update cluster Instances status
// @Return		error
func (p *ClusterMeta) UpdateInstancesStatus(ctx context.Context,
	originStatus constants.ClusterRunningStatus, status constants.ClusterRunningStatus) error {
	for _, components := range p.Instances {
		for _, instance := range components {
			if instance.Status == string(originStatus) {
				instance.Status = string(status)
			}
		}
	}
	// TODO: write db
	framework.LogWithContext(ctx).Infof("update cluster[%s] instances status into %s", p.Cluster.ID, status)
	return nil
}

// GetInstance
// @Description get instance based on instanceID
// @Return		instance
// @Return		error
func (p *ClusterMeta) GetInstance(instanceID string) (*management.ClusterInstance, error) {
	host := strings.Split(instanceID, ":")
	if len(host) != 2 {
		return nil, framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "parameter format is wrong")
	}
	port, err := strconv.Atoi(host[1])
	if err != nil {
		return nil, framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "parameter format is wrong")
	}

	for _, components := range p.Instances {
		for _, instance := range components {
			if Contain(instance.HostIP, host[0]) && Contain(instance.Ports, int32(port)) {
				return instance, nil
			}
		}
	}
	return nil, framework.NewTiEMError(common.TIEM_INSTANCE_NOT_FOUND, "instance not found")
}

func (p *ClusterMeta) CloneMeta(ctx context.Context) *ClusterMeta {
	return nil
}

func (p *ClusterMeta) TryMaintenance(ctx context.Context, maintenanceStatus constants.ClusterMaintenanceStatus) error {
	// check Maintenance status
	if true {
		p.Cluster.MaintenanceStatus = maintenanceStatus
		return nil
	} else {
		return framework.NewTiEMError(common.TIEM_TASK_CONFLICT, "// todo")
	}
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
	cluster, _, err := models.GetClusterReaderWriter().GetMeta(ctx, clusterID)
	// todo
	if err != nil {
		return &ClusterMeta{
			Cluster: cluster,
			//Instances: instances,
		}, err
	}

	return nil, framework.WrapError(common.TIEM_CLUSTER_NOT_FOUND, "", err)
}

func Create(ctx context.Context, template management.Cluster) (*ClusterMeta, error) {
	return nil, nil
}


