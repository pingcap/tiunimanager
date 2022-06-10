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
 * @File: metaoperation.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/11
*******************************************************************************/

package meta

import (
	"context"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/common/structs"
	"github.com/pingcap-inc/tiunimanager/library/framework"
	"github.com/pingcap-inc/tiunimanager/models"
	"github.com/pingcap-inc/tiunimanager/models/cluster/management"
	dbCommon "github.com/pingcap-inc/tiunimanager/models/common"
)

// AddInstances
// @Description add new instances into cluster topology, then alloc host ip, port and disk for these instances
// @Parameter	computes
// @Return		error
func (p *ClusterMeta) AddInstances(ctx context.Context, computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter is invalid!")
	}

	if p.Cluster == nil {
		return errors.NewError(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR, "cluster is nil!")
	}

	if len(p.Instances) == 0 {
		p.Instances = make(map[string][]*management.ClusterInstance)
	}

	for _, compute := range computes {
		for _, item := range compute.Resource {
			for i := 0; i < item.Count; i++ {
				ips := make([]string, 0)
				if len(item.HostIP) > 0 {
					ips = append(ips, item.HostIP)
				}
				instance := &management.ClusterInstance{
					Entity: dbCommon.Entity{
						TenantId: p.Cluster.TenantId,
						Status:   string(constants.ClusterInstanceInitializing),
					},
					Type:         compute.Type,
					Version:      p.Cluster.Version,
					ClusterID:    p.Cluster.ID,
					CpuCores:     int8(structs.ParseCpu(item.Spec)),
					Memory:       int8(structs.ParseMemory(item.Spec)),
					Zone:         structs.GetDomainNameFromCode(item.Zone),
					DiskType:     item.DiskType,
					DiskCapacity: int32(item.DiskCapacity),
					DiskID:       item.DiskID,
					HostIP:       ips,
				}
				p.Instances[compute.Type] = append(p.Instances[compute.Type], instance)
			}
		}
	}
	framework.LogWithContext(ctx).Infof("add new instances into cluster%s topology", p.Cluster.ID)
	return nil
}

func (p *ClusterMeta) AddDefaultInstances(ctx context.Context) error {
	if string(constants.EMProductIDTiDB) == p.Cluster.Type {
		for _, t := range constants.ParasiteComponentIDs {
			instance := &management.ClusterInstance{
				Entity: dbCommon.Entity{
					TenantId: p.Cluster.TenantId,
					Status:   string(constants.ClusterInstanceInitializing),
				},
				Type:      string(t),
				Version:   p.Cluster.Version,
				ClusterID: p.Cluster.ID,
			}
			p.Instances[string(t)] = append(p.Instances[string(t)], instance)
		}

	}

	return nil
}

// UpdateClusterStatus
// @Description update cluster status
// @Return		error
func (p *ClusterMeta) UpdateClusterStatus(ctx context.Context, status constants.ClusterRunningStatus) error {
	err := models.GetClusterReaderWriter().UpdateStatus(ctx, p.Cluster.ID, status)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster%s status into %s failed", p.Cluster.ID, status)
	} else {
		p.Cluster.Status = string(status)
		framework.LogWithContext(ctx).Infof("update cluster%s status into %s succeed", p.Cluster.ID, status)
	}
	return err
}

// DeleteInstance
// @Description delete instance from cluster topology based on instance id
// @Parameter	instance id
// @Return		error
func (p *ClusterMeta) DeleteInstance(ctx context.Context, instanceID string) (*management.ClusterInstance, error) {
	instance, err := p.GetInstance(ctx, instanceID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"get instance %s error: %s", instanceID, err.Error())
		return nil, err
	}

	if err = models.GetClusterReaderWriter().DeleteInstance(ctx, instance.ID); err != nil {
		return nil, err
	}

	// delete instance from cluster topology
	for componentType, components := range p.Instances {
		for index, item := range components {
			if item.ID == instance.ID {
				components = append(components[:index], components[index+1:]...)
				p.Instances[componentType] = components
			}
		}
	}

	return instance, nil
}

// StartMaintenance
// @Description: try to start a maintenance
// @Receiver p
// @Parameter ctx
// @Parameter maintenanceStatus
// @return error
func (p *ClusterMeta) StartMaintenance(ctx context.Context, maintenanceStatus constants.ClusterMaintenanceStatus) error {
	// deleting will end all maintenance
	err := models.GetClusterReaderWriter().SetMaintenanceStatus(ctx, p.Cluster.ID, maintenanceStatus)

	if err == nil {
		p.Cluster.MaintenanceStatus = maintenanceStatus
	}
	return err
}

// EndMaintenance
// @Description: clear maintenance status after maintenance finished or failed
// @Receiver p
// @Parameter ctx
// @Parameter maintenanceStatus
// @return error
func (p *ClusterMeta) EndMaintenance(ctx context.Context, originStatus constants.ClusterMaintenanceStatus) error {
	err := models.GetClusterReaderWriter().ClearMaintenanceStatus(ctx, p.Cluster.ID, originStatus)

	if err == nil {
		p.Cluster.MaintenanceStatus = constants.ClusterMaintenanceNone
	}
	return err
}
