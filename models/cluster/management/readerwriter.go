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
	"github.com/pingcap-inc/tiem/common/structs"
)

type ReaderWriter interface {
	Create(ctx context.Context, cluster *Cluster) (*Cluster, error)
	Delete(ctx context.Context, clusterID string) (err error)
	Get(ctx context.Context, clusterID string) (*Cluster, error)
	GetMeta(ctx context.Context, clusterID string) (*Cluster, []*ClusterInstance, []*DBUser, error)
	GetRelations(ctx context.Context, clusterID string) ([]*ClusterRelation, error)
	//
	// GetMasters
	// @Description: get masters for specified cluster
	// @param ctx
	// @param cluster id
	// @return []*ClusterRelation
	// @return error
	//
	GetMasters(ctx context.Context, clusterID string) ([]*ClusterRelation, error)
	//
	// GetSlaves
	// @Description: get slaves for specified cluster
	// @param ctx
	// @param cluster id
	// @return []*ClusterRelation
	// @return error
	//
	GetSlaves(ctx context.Context, clusterID string) ([]*ClusterRelation, error)

	QueryMetas(ctx context.Context, filters Filters, pageReq structs.PageRequest) ([]*Result, structs.Page, error)

	QueryClusters(ctx context.Context, tenantID string) ([]*Result, error)

	GetInstance(ctx context.Context, ID string) (*ClusterInstance, error)

	DeleteInstance(ctx context.Context, ID string) error

	//
	// UpdateMeta
	// @Description: update cluster and instances, use Update and UpdateInstance
	// @param ctx
	// @param cluster
	// @param instances[]*ClusterInstance
	// @return error
	//
	UpdateMeta(ctx context.Context, cluster *Cluster, instances []*ClusterInstance) error
	//
	// UpdateInstance update cluster instances
	//  @Description:
	//  @param ctx
	//  @param instances
	//  @return error
	//
	UpdateInstance(ctx context.Context, instances ...*ClusterInstance) error

	//
	// QueryInstancesByHost
	// @Description: query all instances located at host
	// @param ctx
	// @param hostId
	// @param typeFilter
	// @param statusFilter
	// @return []*ClusterInstance
	// @return error
	//
	QueryInstancesByHost(ctx context.Context, hostId string, typeFilter []string, statusFilter []string) ([]*ClusterInstance, error)

	//
	// UpdateClusterInfo
	// @Description:  update cluster base info, excluding Status, MaintenanceStatus
	// please update status with specific method, UpdateStatus, SetMaintenanceStatus, ClearMaintenanceStatus
	// @param ctx
	// @param template
	// @return error
	//
	UpdateClusterInfo(ctx context.Context, template *Cluster) error

	//
	// UpdateStatus
	//  @Description: update cluster status
	//  @param ctx
	//  @param clusterID
	//  @param status
	//  @return error
	//
	UpdateStatus(ctx context.Context, clusterID string, status constants.ClusterRunningStatus) error

	//
	// SetMaintenanceStatus
	//  @Description: set maintenance status to targetStatus,
	//  (current MaintenanceStatus == constants.ClusterMaintenanceNone) is a precondition
	//  @param ctx
	//  @param clusterID
	//  @param targetStatus
	//  @return error
	//
	SetMaintenanceStatus(ctx context.Context, clusterID string, targetStatus constants.ClusterMaintenanceStatus) error

	//
	// ClearMaintenanceStatus
	// @Description: set maintenance status to constant.ClusterMaintenanceNone
	// (current MaintenanceStatus == originalStatus) is a precondition
	// @param ctx
	// @param clusterID
	// @param originalStatus
	// @return error
	//
	ClearMaintenanceStatus(ctx context.Context, clusterID string, originalStatus constants.ClusterMaintenanceStatus) error

	CreateRelation(ctx context.Context, relation *ClusterRelation) error
	DeleteRelation(ctx context.Context, relationID uint) error
	SwapMasterSlaveRelation(ctx context.Context, oldMasterClusterId, oldSlaveClusterId, newSyncChangeFeedTaskId string) error

	CreateClusterTopologySnapshot(ctx context.Context, snapshot ClusterTopologySnapshot) error
	GetCurrentClusterTopologySnapshot(ctx context.Context, clusterID string) (ClusterTopologySnapshot, error)
	UpdateTopologySnapshotConfig(ctx context.Context, clusterID string, config string) error

	//
	// ClearClusterPhysically
	// @Description: If you don't know why you should use it, then don't use it
	// @param ctx
	// @param clusterID
	// @return err
	//
	ClearClusterPhysically(ctx context.Context, clusterID string) (err error)
	//
	// CreateDBUser
	// @Description: create cluster users
	// @param ctx
	// @param user
	// @return error
	//
	CreateDBUser(ctx context.Context, user *DBUser) error
	//
	// GetDBUser
	// @Description: get cluster users by clusterID
	// @param ctx
	// @param clusterID
	// @return []*DBUser
	// @return error
	//
	GetDBUser(ctx context.Context, clusterID string) ([]*DBUser, error)
	//
	// UpdateDBUser
	// @Description: update cluster users
	// @param ctx
	// @param user
	// @return error
	//
	UpdateDBUser(ctx context.Context, user *DBUser) error
	//
	// DeleteDBUser
	// @Description: delete cluster users
	// @param ctx
	// @param ID
	// @return error
	//
	DeleteDBUser(ctx context.Context, ID uint) error
}
