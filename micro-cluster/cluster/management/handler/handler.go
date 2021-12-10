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
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
)

type ClusterMeta struct {
	cluster *management.Cluster
	instances []*management.ClusterInstance
}

func (p *ClusterMeta) Save(ctx context.Context) error {
	return nil
}

func (p *ClusterMeta) ScaleIn(ctx context.Context, instancesIDs []string) error {
	return nil
}

func (p *ClusterMeta) ScaleOut(ctx context.Context, instances []*management.ClusterInstance) error {
	if p.instances == nil {
		p.instances = instances
	} else {
		p.instances = append(p.instances, instances...)
	}
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
