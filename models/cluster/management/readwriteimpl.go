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

	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/common/structs"
	"github.com/pingcap-inc/tiunimanager/library/framework"
	dbCommon "github.com/pingcap-inc/tiunimanager/models/common"
	"gorm.io/gorm"
)

type ClusterReadWrite struct {
	dbCommon.GormDB
}

func (g *ClusterReadWrite) Create(ctx context.Context, cluster *Cluster) (*Cluster, error) {
	err := g.DB(ctx).Create(cluster).Error

	if err != nil {
		// duplicated name
		existOrError := g.DB(ctx).Model(&Cluster{}).Where("name = ?", cluster.Name).First(&Cluster{}).Error
		if existOrError == nil {
			err = errors.NewErrorf(errors.TIUNIMANAGER_DUPLICATED_NAME, "%s:%s", errors.TIUNIMANAGER_DUPLICATED_NAME.Explain(), cluster.Name)
		} else {
			err = dbCommon.WrapDBError(err)
		}
	}
	return cluster, err
}

func (g *ClusterReadWrite) Delete(ctx context.Context, clusterID string) (err error) {
	got, err := g.Get(ctx, clusterID)
	if err != nil {
		return
	}
	err = g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return errors.OfNullable(nil).BreakIf(func() error {
			return tx.Delete(got).Error
		}).BreakIf(func() error {
			return tx.Where("cluster_id = ?", clusterID).Delete(&ClusterInstance{}).Error
		}).BreakIf(func() error {
			return tx.Where("subject_cluster_id = ?", clusterID).Delete(&ClusterRelation{}).Error
		}).BreakIf(func() error {
			return tx.Where("object_cluster_id = ?", clusterID).Delete(&ClusterRelation{}).Error
		}).BreakIf(func() error {
			return tx.Where("cluster_id = ?", clusterID).Delete(&DBUser{}).Error
		}).BreakIf(func() error {
			return tx.Where("cluster_id = ?", clusterID).Delete(&ClusterTopologySnapshot{}).Error
		}).If(func(e error) {
			framework.LogWithContext(ctx).Errorf("delete cluster %s failed, err = %s", clusterID, e.Error())
		}).Else(func() {
			framework.LogWithContext(ctx).Infof("delete cluster %s succeed", clusterID)
		}).Present()
	})

	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) DeleteInstance(ctx context.Context, ID string) error {
	instance := &ClusterInstance{}
	err := g.DB(ctx).First(instance, "id = ?", ID).Error
	if err != nil {
		return errors.WrapError(errors.TIUNIMANAGER_INSTANCE_NOT_FOUND, "", err)
	}
	err = g.DB(ctx).Delete(instance).Error
	if err != nil {
		return errors.WrapError(errors.TIUNIMANAGER_DELETE_INSTANCE_ERROR, "", err)
	}
	return nil
}

func (g *ClusterReadWrite) Get(ctx context.Context, clusterID string) (*Cluster, error) {
	if "" == clusterID {
		errInfo := "get cluster failed : empty clusterID"
		framework.LogWithContext(ctx).Error(errInfo)
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, errInfo)
	}

	cluster := &Cluster{}
	err := g.DB(ctx).First(cluster, "id = ?", clusterID).Error

	if err != nil {
		errInfo := fmt.Sprintf("get cluster failed : clusterID = %s, err = %s", clusterID, err.Error())
		framework.LogWithContext(ctx).Error(errInfo)

		return nil, errors.WrapError(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, errInfo, err)
	} else {
		return cluster, nil
	}

}

func (g *ClusterReadWrite) GetInstance(ctx context.Context, ID string) (*ClusterInstance, error) {
	instance := &ClusterInstance{}
	err := g.DB(ctx).First(instance, "id = ?", ID).Error
	if err != nil {
		return nil, errors.WrapError(errors.TIUNIMANAGER_INSTANCE_NOT_FOUND, "", err)
	}
	return instance, nil
}

func (g *ClusterReadWrite) GetMeta(ctx context.Context, clusterID string) (cluster *Cluster, instances []*ClusterInstance, users []*DBUser, err error) {
	cluster, err = g.Get(ctx, clusterID)

	if err != nil {
		err = dbCommon.WrapDBError(err)
		return
	}

	instances = make([]*ClusterInstance, 0)

	err = g.DB(ctx).Model(&ClusterInstance{}).Where("cluster_id = ?", clusterID).Find(&instances).Error
	if err != nil {
		err = dbCommon.WrapDBError(err)
		return
	}

	users, err = g.GetDBUser(ctx, clusterID)
	if err != nil {
		err = dbCommon.WrapDBError(err)
		return
	}

	return
}

func (g *ClusterReadWrite) RelationsResetSyncChangeFeedTaskIDs(ctx context.Context, masterClusterID string, slavesClusterIDMapToSyncTaskID map[string]string) error {
	return g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		slaves, err := g.GetSlaves(ctx, masterClusterID)
		if err != nil {
			return err
		}
		if len(slaves) != len(slavesClusterIDMapToSyncTaskID) {
			return fmt.Errorf("len(slaves) != len(slavesClusterIDMapToSyncTaskID): %d != %d", len(slaves), len(slavesClusterIDMapToSyncTaskID))
		}
		for _, rel := range slaves {
			newTaskID, ok := slavesClusterIDMapToSyncTaskID[rel.ObjectClusterID]
			if !ok {
				return fmt.Errorf("slaveClusterID %s not match slavesClusterIDMapToSyncTaskID", rel.ObjectClusterID)
			}
			if len(newTaskID) <= 0 {
				return fmt.Errorf("unexpected zero-length newTaskID, slaveID:%s", rel.ObjectClusterID)
			}
			if newTaskID != rel.SyncChangeFeedTaskID {
				err := g.DB(ctx).Model(rel).Update("sync_change_feed_task_id", newTaskID).Error
				if err != nil {
					err = dbCommon.WrapDBError(err)
					return err
				}
			}
		}
		return nil
	})
}

func (g *ClusterReadWrite) GetRelations(ctx context.Context, clusterID string) ([]*ClusterRelation, error) {
	relations := make([]*ClusterRelation, 0)
	err := g.DB(ctx).Model(&ClusterRelation{}).Where("object_cluster_id  = ? ", clusterID).Find(&relations).Error
	if err != nil {
		err = dbCommon.WrapDBError(err)
	}

	return relations, err
}

func (g *ClusterReadWrite) GetMasters(ctx context.Context, clusterID string) ([]*ClusterRelation, error) {
	if "" == clusterID {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "cluster id is invalid")
	}
	relations := make([]*ClusterRelation, 0)
	err := g.DB(ctx).Model(&ClusterRelation{}).Where("object_cluster_id  = ? ", clusterID).Find(&relations).Error
	return relations, err
}

func (g *ClusterReadWrite) GetSlaves(ctx context.Context, clusterID string) ([]*ClusterRelation, error) {
	if "" == clusterID {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "cluster id is invalid")
	}
	relations := make([]*ClusterRelation, 0)
	err := g.DB(ctx).Model(&ClusterRelation{}).Where("subject_cluster_id  = ? ", clusterID).Find(&relations).Error
	return relations, err
}

func (g *ClusterReadWrite) QueryClusters(ctx context.Context, tenantID string) ([]*Result, error) {
	if "" == tenantID {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "tenant id is invalid")
	}

	clusters := make([]*Cluster, 0)
	err := g.DB(ctx).Table("clusters").Where("tenant_id = ?",
		tenantID).Where("deleted_at is null").Find(&clusters).Error
	if err != nil {
		return nil, errors.WrapError(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, "", err)
	}

	results := make([]*Result, 0)
	for _, c := range clusters {
		instances := make([]*ClusterInstance, 0)

		err = g.DB(ctx).Model(&ClusterInstance{}).Where("cluster_id = ?", c.ID).Find(&instances).Error

		if err != nil {
			return nil, errors.WrapError(errors.TIUNIMANAGER_INSTANCE_NOT_FOUND, "", err)
		}

		users, err := g.GetDBUser(ctx, c.ID)
		if err != nil {
			return nil, dbCommon.WrapDBError(err)
		}
		results = append(results, &Result{
			Cluster:   c,
			Instances: instances,
			DBUsers:   users,
		})
	}

	return results, nil
}

// todo
func (g *ClusterReadWrite) QueryMetas(ctx context.Context, filters Filters, pageReq structs.PageRequest) ([]*Result, structs.Page, error) {
	page := structs.Page{
		Page:     pageReq.Page,
		PageSize: pageReq.PageSize,
	}

	if len(filters.TenantId) == 0 {
		err := errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "tenant id required in QueryMetas")
		return nil, page, err
	}

	clusters := make([]*Cluster, 0)

	total := int64(0)
	query := g.DB(ctx).Table("clusters").Where("tenant_id = ?", filters.TenantId).Where("deleted_at is null")
	if len(filters.ClusterIDs) > 0 {
		query = query.Where("id in ?", filters.ClusterIDs)
	}
	if filters.NameLike != "" {
		query = query.Where("name like '%" + filters.NameLike + "%'")
	}

	if len(filters.Type) > 0 {
		query = query.Where("type = ?", filters.Type)
	}

	if len(filters.StatusFilters) > 0 {
		query = query.Where("status in ?", filters.StatusFilters)
	}

	if len(filters.Tag) > 0 {
		query = query.Where("tag_info like '%\"" + filters.Tag + "\"%'")
	}

	err := query.Count(&total).Order("updated_at desc").Offset(pageReq.GetOffset()).Limit(pageReq.PageSize).Find(&clusters).Error
	if err != nil {
		err = errors.WrapError(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, "", err)
		return nil, page, err
	} else {
		page.Total = int(total)
	}

	results := make([]*Result, 0)

	for _, c := range clusters {
		instances := make([]*ClusterInstance, 0)

		err = g.DB(ctx).Model(&ClusterInstance{}).Where("cluster_id = ?", c.ID).Find(&instances).Error

		if err != nil {
			err = errors.WrapError(errors.TIUNIMANAGER_INSTANCE_NOT_FOUND, "", err)
			return nil, page, err
		}

		users, err := g.GetDBUser(ctx, c.ID)
		if err != nil {
			err = dbCommon.WrapDBError(err)
			return nil, page, err
		}
		results = append(results, &Result{
			Cluster:   c,
			Instances: instances,
			DBUsers:   users,
		})
	}
	return results, page, nil
}

func (g *ClusterReadWrite) UpdateMeta(ctx context.Context, cluster *Cluster, instances []*ClusterInstance) error {
	return g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		err := g.UpdateClusterInfo(dbCommon.CtxWithTransaction(ctx, tx), cluster)
		if err == nil {
			err = g.UpdateInstance(dbCommon.CtxWithTransaction(ctx, tx), instances...)
		}

		if err != nil {
			msg := fmt.Sprintf("cluster update meta failed, clusterId = %s", cluster.ID)
			framework.LogWithContext(ctx).Error(msg)
			return dbCommon.WrapDBError(err)
		}
		return nil
	})

}

func (g *ClusterReadWrite) QueryInstancesByHost(ctx context.Context, hostId string, typeFilter []string, statusFilter []string) ([]*ClusterInstance, error) {
	if len(hostId) == 0 {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "empty hostId")
	}
	query := g.DB(ctx).Table("cluster_instances").Where("host_id = ?", hostId)
	if len(typeFilter) > 0 {
		query = query.Where("type in ?", typeFilter)
	}
	if len(statusFilter) > 0 {
		query = query.Where("status in ?", statusFilter)
	}

	instances := make([]*ClusterInstance, 0)
	err := query.Find(&instances).Error

	return instances, dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) QueryHostInstances(ctx context.Context, hostIds []string) ([]HostInstanceItem, error) {
	db := g.DB(ctx).Model(&ClusterInstance{}).Select("host_id, cluster_id, type as component")
	if hostIds != nil {
		db.Where("host_id in ?", hostIds)
	}
	db.Group("host_id").Group("cluster_id").Group("type")

	items := make([]HostInstanceItem, 0)
	err := db.Scan(&items).Error

	return items, dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) UpdateInstance(ctx context.Context, instances ...*ClusterInstance) error {
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

func (g *ClusterReadWrite) UpdateClusterInfo(ctx context.Context, template *Cluster) error {
	if template == nil {
		errInfo := "update cluster base info failed : empty template"
		framework.LogWithContext(ctx).Error(errInfo)
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, errInfo)
	}

	cluster, err := g.Get(ctx, template.ID)

	if err != nil {
		err = dbCommon.WrapDBError(err)
		return err
	}
	err = g.DB(ctx).Model(cluster).Omit("status", "maintenance_status").Updates(template).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) UpdateStatus(ctx context.Context, clusterID string, status constants.ClusterRunningStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	cluster.Status = string(status)

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) SetMaintenanceStatus(ctx context.Context, clusterID string, targetStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != constants.ClusterMaintenanceNone &&
		targetStatus != constants.ClusterMaintenanceDeleting ||
		cluster.MaintenanceStatus == constants.ClusterMaintenanceDeleting {

		errInfo := fmt.Sprintf("set cluster maintenance status conflicted : current maintenance = %s, target maintenance = %s, clusterID = %s", cluster.MaintenanceStatus, targetStatus, clusterID)
		framework.LogWithContext(ctx).Error(errInfo)
		return errors.NewError(errors.TIUNIMANAGER_CLUSTER_MAINTENANCE_CONFLICT, errInfo)
	}

	cluster.MaintenanceStatus = targetStatus

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) ClearMaintenanceStatus(ctx context.Context, clusterID string, originalStatus constants.ClusterMaintenanceStatus) error {
	cluster, err := g.Get(ctx, clusterID)

	if err != nil {
		return err
	}

	if cluster.MaintenanceStatus != originalStatus {
		errInfo := fmt.Sprintf("clear cluster maintenance status failed : unmatched original status, want %s, current %s", originalStatus, cluster.MaintenanceStatus)
		framework.LogWithContext(ctx).Error(errInfo)
		return errors.NewError(errors.TIUNIMANAGER_CLUSTER_MAINTENANCE_CONFLICT, errInfo)
	}

	cluster.MaintenanceStatus = constants.ClusterMaintenanceNone

	err = g.DB(ctx).Save(cluster).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) CreateRelation(ctx context.Context, relation *ClusterRelation) error {
	err := g.DB(ctx).Create(relation).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) DeleteRelation(ctx context.Context, relationID uint) error {
	relation := &ClusterRelation{}
	err := g.DB(ctx).First(relation, "id = ?", relationID).Delete(relation).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) SwapMasterSlaveRelations(ctx context.Context, oldMasterClusterId, slaveToBeMasterClusterId string, newSlaveClusterIdMapToSyncCDCTaskId map[string]string) error {
	tx := g.DB(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}
	relations := make([]*ClusterRelation, 0)
	err := tx.Model(&ClusterRelation{}).
		Where("subject_cluster_id  = ? ", oldMasterClusterId).
		Where("relation_type  = ? ", string(constants.ClusterRelationStandBy)).
		Find(&relations).Error
	if err != nil {
		err = dbCommon.WrapDBError(err)
		tx.Rollback()
		return err
	}
	if len(relations) <= 0 {
		err = dbCommon.WrapDBError(fmt.Errorf("gorm SwapMasterSlaveRelations no relation found"))
		tx.Rollback()
		return err
	}
	if len(relations) != len(newSlaveClusterIdMapToSyncCDCTaskId) {
		err = dbCommon.WrapDBError(fmt.Errorf("gorm SwapMasterSlaveRelations relations amount does not match: %d != %d",
			len(relations), len(newSlaveClusterIdMapToSyncCDCTaskId)))
		tx.Rollback()
		return err
	}
	dupNewSlaveClusterIdMapToSyncCDCTaskId := make(map[string]string)
	for k, v := range newSlaveClusterIdMapToSyncCDCTaskId {
		dupNewSlaveClusterIdMapToSyncCDCTaskId[k] = v
	}
	_, ok := dupNewSlaveClusterIdMapToSyncCDCTaskId[oldMasterClusterId]
	if ok {
		delete(dupNewSlaveClusterIdMapToSyncCDCTaskId, oldMasterClusterId)
	} else {
		err = dbCommon.WrapDBError(fmt.Errorf("gorm SwapMasterSlaveRelations dupNewSlaveClusterIdMapToSyncCDCTaskId has no oldMasterClusterId"))
		tx.Rollback()
	}
	equalSlaveToBeMasterClusterIdCt := 0
	for _, relation := range relations {
		framework.Assert(relation.SubjectClusterID == oldMasterClusterId)
		framework.LogWithContext(ctx).Debugf("gorm SwapMasterSlaveRelations get relation %v %s %s %s",
			relation.ID, relation.SubjectClusterID, relation.ObjectClusterID, relation.SyncChangeFeedTaskID)
		if relation.ObjectClusterID == slaveToBeMasterClusterId {
			equalSlaveToBeMasterClusterIdCt++
		} else {
			delete(dupNewSlaveClusterIdMapToSyncCDCTaskId, relation.ObjectClusterID)
		}
		err = tx.Debug().Delete(relation).Error
		if err != nil {
			framework.LogWithContext(ctx).Errorf("gorm SwapMasterSlaveRelations %s", err)
			tx.Rollback()
			return err
		}
	}
	if equalSlaveToBeMasterClusterIdCt != 1 || len(dupNewSlaveClusterIdMapToSyncCDCTaskId) != 0 {
		framework.LogWithContext(ctx).Errorf("gorm SwapMasterSlaveRelations check failed, %v,%v",
			equalSlaveToBeMasterClusterIdCt == 1, len(dupNewSlaveClusterIdMapToSyncCDCTaskId) == 0)
		tx.Rollback()
		return err
	}
	for newSlaveId, newSyncCDCId := range newSlaveClusterIdMapToSyncCDCTaskId {
		err = tx.Debug().Create(&ClusterRelation{
			RelationType:         constants.ClusterRelationStandBy,
			ObjectClusterID:      newSlaveId,
			SubjectClusterID:     slaveToBeMasterClusterId,
			SyncChangeFeedTaskID: newSyncCDCId,
		}).Error
		if err != nil {
			framework.LogWithContext(ctx).Errorf("gorm SwapMasterSlaveRelations create relation failed, %s", err)
			tx.Rollback()
			return err
		}
	}

	return dbCommon.WrapDBError(tx.Commit().Error)
}

func (g *ClusterReadWrite) CreateClusterTopologySnapshot(ctx context.Context, snapshot ClusterTopologySnapshot) error {
	if len(snapshot.ClusterID) == 0 || len(snapshot.TenantID) == 0 {
		errInfo := fmt.Sprintf("CreateClusterTopologySnapshot failed : parameter invalid, ClusterID = %s, TenantID = %s", snapshot.ClusterID, snapshot.TenantID)
		framework.LogWithContext(ctx).Error(errInfo)
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, errInfo)
	}
	if len(snapshot.PrivateKey) == 0 || len(snapshot.PublicKey) == 0 {
		errInfo := "CreateClusterTopologySnapshot failed : connection key required"
		framework.LogWithContext(ctx).Error(errInfo)
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, errInfo)
	}

	err := g.DB(ctx).Create(&snapshot).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) GetCurrentClusterTopologySnapshot(ctx context.Context, clusterID string) (snapshot ClusterTopologySnapshot, err error) {
	if len(clusterID) == 0 {
		errInfo := "get cluster topology snapshot failed : cluster id required"
		framework.LogWithContext(ctx).Error(errInfo)
		err = errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, errInfo)
		return
	}

	err = g.DB(ctx).Model(snapshot).Where("cluster_id = ?", clusterID).First(&snapshot).Error
	err = dbCommon.WrapDBError(err)
	return
}

func (g *ClusterReadWrite) UpdateTopologySnapshotConfig(ctx context.Context, clusterID string, config string) error {
	snapshot := &ClusterTopologySnapshot{}
	err := g.DB(ctx).Model(&ClusterTopologySnapshot{}).Where("cluster_id = ?", clusterID).First(snapshot).Error
	if err != nil {
		errInfo := "update cluster topology snapshot failed : record not found"
		framework.LogWithContext(ctx).Error(errInfo)
		err = errors.NewError(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, errInfo)
		return err
	}
	snapshot.Config = config

	return dbCommon.WrapDBError(g.DB(ctx).Save(snapshot).Error)
}

func (g *ClusterReadWrite) ClearClusterPhysically(ctx context.Context, clusterID string, reason string) error {
	if len(clusterID) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "clusterId is empty")
	}
	if len(reason) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "reason is empty")
	}
	err := g.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return errors.OfNullable(nil).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Delete(&Cluster{Entity: dbCommon.Entity{ID: clusterID}}).Error
		}).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Where("cluster_id = ?", clusterID).Delete(&ClusterInstance{}).Error
		}).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Where("cluster_id = ?", clusterID).Delete(&ClusterTopologySnapshot{}).Error
		}).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Where("cluster_id = ?", clusterID).Delete(&DBUser{}).Error
		}).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Where("subject_cluster_id = ?", clusterID).Delete(&ClusterRelation{}).Error
		}).BreakIf(func() error {
			return g.DB(ctx).Unscoped().Where("object_cluster_id = ?", clusterID).Delete(&ClusterRelation{}).Error
		}).If(func(err error) {
			framework.LogWithContext(ctx).Errorf("clear cluster data physically failed, clusterId = %s, err = %s", clusterID, err.Error())
		}).Else(func() {
			framework.LogWithContext(ctx).Warnf("clear data of cluster %s physically, operatorId = %s, reason = %s", clusterID, framework.GetUserIDFromContext(ctx), reason)
		}).Present()
	})
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) CreateDBUser(ctx context.Context, user *DBUser) error {
	err := g.DB(ctx).Create(user).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) GetDBUser(ctx context.Context, clusterID string) ([]*DBUser, error) {
	users := make([]*DBUser, 0)
	err := g.DB(ctx).Model(&DBUser{}).Where("cluster_id = ? ", clusterID).Find(&users).Error
	if err != nil {
		err = dbCommon.WrapDBError(err)
	}
	return users, err
}

func (g *ClusterReadWrite) DeleteDBUser(ctx context.Context, ID uint) error {
	user := &DBUser{}
	err := g.DB(ctx).First(user, "id = ?", ID).Delete(user).Error
	return dbCommon.WrapDBError(err)
}

func (g *ClusterReadWrite) UpdateDBUser(ctx context.Context, user *DBUser) error {
	g.DB(ctx).Save(user)
	return nil
}

func NewClusterReadWrite(db *gorm.DB) *ClusterReadWrite {
	return &ClusterReadWrite{
		dbCommon.WrapDB(db),
	}
}
