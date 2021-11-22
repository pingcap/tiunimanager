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

package models

import (
	"context"
	"fmt"
)

type ComponentInstance struct {
	Entity
	ClusterId      string `gorm:"not null;type:varchar(22);default:null"`
	ComponentType  string `gorm:"not null;"`
	Role           string
	Version        string `gorm:"not null;"`
	HostId         string
	Host           string
	PortInfo       string
	DiskId         string
	DiskPath       string
	AllocRequestId string
	CpuCores       int8
	Memory         int8
}

func (m *DAOClusterManager) ListComponentInstances(ctx context.Context, clusterId string) (componentInstances []*ComponentInstance, err error) {
	if clusterId == "" {
		return nil, fmt.Errorf("ListComponentInstances has invalid parameter, clusterId: %s", clusterId)
	}
	componentInstances = make([]*ComponentInstance, 0, 10)

	err = m.Db(ctx).Table(TABLE_NAME_COMPONENT_INSTANCE).Where("cluster_id = ?", clusterId).Find(&componentInstances).Error

	return componentInstances, err
}

func (m *DAOClusterManager) ListComponentInstancesByHost(ctx context.Context, hostId string) (componentInstances []*ComponentInstance, err error) {
	if hostId == "" {
		return nil, fmt.Errorf("ListComponentInstancesByHost has invalid parameter, hostId: %s", hostId)
	}
	componentInstances = make([]*ComponentInstance, 0)

	err = m.Db(ctx).Table(TABLE_NAME_COMPONENT_INSTANCE).Where("host_id = ?", hostId).Find(&componentInstances).Error

	return componentInstances, err
}

func (m *DAOClusterManager) AddClusterComponentInstance(ctx context.Context, clusterId string, componentInstances []*ComponentInstance) ([]*ComponentInstance, error) {
	if clusterId == "" {
		return nil, fmt.Errorf("AddClusterComponentInstance has invalid parameter, clusterId: %s", clusterId)
	}
	if len(componentInstances) == 0 {
		return nil, fmt.Errorf("AddClusterComponentInstance has invalid parameter, componentInstances: %v", componentInstances)
	}
	err := m.db.CreateInBatches(componentInstances, len(componentInstances)).Error
	return componentInstances, err
}

func (m *DAOClusterManager) DeleteClusterComponentInstance(ctx context.Context, instanceId string) error {
	if "" == instanceId {
		return fmt.Errorf("DeleteInstance has invalid parameter, instanceId: %s", instanceId)
	}
	instance := &ComponentInstance{}
	return m.Db(ctx).First(instance, "id = ?", instanceId).Delete(instance).Error
}

