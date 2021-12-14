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
	"fmt"
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
	err = g.DB(ctx).Delete(got).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) Get(ctx context.Context, clusterID string) (*Cluster, error) {
	if "" == clusterID {
		errInfo := fmt.Sprint("get cluster failed : empty clusterID")
		framework.LogWithContext(ctx).Error(errInfo)
		return nil, framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, errInfo)
	}

	cluster := &Cluster{}
	err := g.DB(ctx).First(cluster, "id = ?", clusterID).Error

	if err != nil {
		errInfo := fmt.Sprintf("get cluster failed : clusterID = %s", clusterID)
		framework.LogWithContext(ctx).Error(errInfo)

		return nil, framework.WrapError(common.TIEM_CLUSTER_NOT_FOUND, errInfo, err)
	} else {
		return cluster, nil
	}

}

func (g *GormClusterReadWrite) GetMeta(ctx context.Context, clusterID string) (cluster *Cluster, instances []*ClusterInstance, err error) {
	cluster, err = g.Get(ctx, clusterID)

	if err != nil {
		err = dbCommon.WrapDBError(err)
		return
	}

	instances = make([]*ClusterInstance, 0)

	err = g.DB(ctx).Model(&ClusterInstance{}).Where("cluster_id = ?", clusterID).Find(&instances).Error
	err = dbCommon.WrapDBError(err)
	return

}

func (g *GormClusterReadWrite) UpdateInstance(ctx context.Context, instances ...*ClusterInstance) error {
	return g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		toCreate := make([]*ClusterInstance, 0)

		for _, instance := range instances {
			if instance.ID == "" {
				toCreate = append(toCreate, instance)
			} else {
				err := tx.Save(instance).Error
				if err != nil {
					err = dbCommon.WrapDBError(err)
					return err
				}
			}
		}
		if len(toCreate) > 0 {
			err := tx.CreateInBatches(toCreate, len(toCreate)).Error
			if err != nil {
				err = dbCommon.WrapDBError(err)
				return err
			}
		}
		return nil
	})
}

func (g *GormClusterReadWrite) UpdateBaseInfo(ctx context.Context, template *Cluster) error {
	if template == nil {
		errInfo := fmt.Sprint("update cluster base info failed : empty template")
		framework.LogWithContext(ctx).Error(errInfo)
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, errInfo)
	}

	_, err := g.Get(ctx, template.ID)

	if err != nil {
		err = dbCommon.WrapDBError(err)
		return err
	}

	err = g.DB(ctx).Save(template).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) UpdateStatus(ctx context.Context, clusterID string, status constants.ClusterRunningStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	cluster.Status = string(status)

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) SetMaintenanceStatus(ctx context.Context, clusterID string, targetStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != constants.ClusterMaintenanceNone {
		errInfo := fmt.Sprintf("set cluster maintenance status conflicted : current maintenance = %s, clusterID = %s", cluster.MaintenanceStatus, clusterID)
		framework.LogWithContext(ctx).Error(errInfo)
		return framework.NewTiEMError(common.TIEM_CLUSTER_MAINTENANCE_CONFLICT, errInfo)
	}

	cluster.MaintenanceStatus = targetStatus

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) ClearMaintenanceStatus(ctx context.Context, clusterID string, originalStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != originalStatus {
		errInfo := fmt.Sprintf("clear cluster maintenance status failed : unmatched original status, want %s, current %s", originalStatus, cluster.MaintenanceStatus)
		framework.LogWithContext(ctx).Error(errInfo)
		return framework.NewTiEMError(common.TIEM_CLUSTER_MAINTENANCE_CONFLICT, errInfo)
	}

	cluster.MaintenanceStatus = constants.ClusterMaintenanceNone

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) CreateRelation(ctx context.Context, relation *ClusterRelation) error {
	err := g.DB(ctx).Create(relation).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) DeleteRelation(ctx context.Context, relationID uint) error {
	relation := &ClusterRelation{}
	err := g.DB(ctx).First(relation, "id = ?", relationID).Delete(relation).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) CreateClusterTopologySnapshot(ctx context.Context, snapshot ClusterTopologySnapshot) error {
	if snapshot.ClusterID == "" || snapshot.TenantID == "" || snapshot.Config == "" {
		errInfo := fmt.Sprintf("CreateClusterTopologySnapshot failed : parameter invalid, ClusterID = %s, TenantID = %s, config = %s", snapshot.ClusterID, snapshot.TenantID, snapshot.Config)
		framework.LogWithContext(ctx).Error(errInfo)
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, errInfo)
	}
	err := g.DB(ctx).Create(&snapshot).Error
	return dbCommon.WrapDBError(err)
}

func (g *GormClusterReadWrite) GetLatestClusterTopologySnapshot(ctx context.Context, clusterID string) (snapshot ClusterTopologySnapshot, err error) {
	if "" == clusterID {
		errInfo := fmt.Sprint("get latest cluster topology snapshot failed : empty clusterID")
		framework.LogWithContext(ctx).Error(errInfo)
		err = framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, errInfo)
		return
	}

	err = g.DB(ctx).Model(snapshot).Where("cluster_id = ?", clusterID).Order("id desc").First(&snapshot).Error
	err = dbCommon.WrapDBError(err)
	return
}

func NewGormClusterReadWrite(db *gorm.DB) *GormClusterReadWrite {
	return &GormClusterReadWrite{
		dbCommon.WrapDB(db),
	}
}