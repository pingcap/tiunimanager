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

package management

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormClusterReadWrite struct {
	dbCommon.GormDB
}

func (g *GormClusterReadWrite) Create(ctx context.Context, cluster *Cluster) (*Cluster, error) {
	return cluster, g.DB(ctx).Create(cluster).Error
}

func (g *GormClusterReadWrite) Delete(ctx context.Context, clusterID string) (err error) {
	got, err := g.Get(ctx, clusterID)
	if err != nil {
		return
	}
	return g.DB(ctx).Delete(got).Error
}

func (g *GormClusterReadWrite) Get(ctx context.Context, clusterID string) (*Cluster, error) {
	if "" == clusterID {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	cluster := &Cluster{}
	err := g.DB(ctx).First(cluster, "id = ?", clusterID).Error

	if err != nil {
		return nil, framework.WrapError(common.TIEM_CLUSTER_NOT_FOUND, common.TIEM_CLUSTER_NOT_FOUND.Explain(), err)
	} else {
		return cluster, nil
	}

}

func (g *GormClusterReadWrite) GetMeta(ctx context.Context, clusterID string) (cluster *Cluster, instances []*ClusterInstance, err error) {
	cluster, err = g.Get(ctx, clusterID)

	if err != nil {
		return
	}

	instances = make([]*ClusterInstance, 0)

	err = g.DB(ctx).Model(&ClusterInstance{}).Where("cluster_id = ?", clusterID).Find(&instances).Error
	return
}

func (g *GormClusterReadWrite) UpdateInstance(ctx context.Context, instances ...*ClusterInstance) error {
	return g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		toCreate := make([]*ClusterInstance, 0)

		for _,instance := range instances {
			if instance.ID == "" {
				toCreate = append(toCreate, instance)
			} else {
				err := tx.Save(instance).Error
				if err != nil {
					return err
				}
			}
		}
		if len(toCreate) > 0{
			err := tx.CreateInBatches(toCreate, len(toCreate)).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GormClusterReadWrite) UpdateBaseInfo(ctx context.Context, template *Cluster) error {
	if template == nil {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	_, err := g.Get(ctx, template.ID)

	if err != nil {
		return err
	}

	return g.DB(ctx).Save(template).Error
}

func (g *GormClusterReadWrite) UpdateStatus(ctx context.Context, clusterID string, status constants.ClusterRunningStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	cluster.Status = string(status)

	return g.DB(ctx).Save(cluster).Error
}

func (g *GormClusterReadWrite) SetMaintenanceStatus(ctx context.Context, clusterID string, targetStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != constants.ClusterMaintenanceNone {
		return framework.SimpleError(common.TIEM_CLUSTER_MAINTENANCE_CONFLICT)
	}

	cluster.MaintenanceStatus = targetStatus

	return g.DB(ctx).Save(cluster).Error
}

func (g *GormClusterReadWrite) ClearMaintenanceStatus(ctx context.Context, clusterID string, originalStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != originalStatus {
		return framework.SimpleError(common.TIEM_CLUSTER_MAINTENANCE_CONFLICT)
	}

	cluster.MaintenanceStatus = constants.ClusterMaintenanceNone

	return g.DB(ctx).Save(cluster).Error
}

func (g *GormClusterReadWrite) CreateRelation(ctx context.Context, relation *ClusterRelation) error {
	return g.DB(ctx).Create(relation).Error
}

func (g *GormClusterReadWrite) DeleteRelation(ctx context.Context, relationID uint) error {
	relation := &ClusterRelation{}
	return g.DB(ctx).First(relation, "id = ?", relationID).Delete(relation).Error
}

func NewGormClusterReadWrite(db *gorm.DB) *GormClusterReadWrite{
	return &GormClusterReadWrite{
		dbCommon.WrapDB(db),
	}
}
