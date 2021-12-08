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
)

type ReaderWriter interface {

	Create(ctx context.Context, cluster *Cluster) (*Cluster, error)
	Delete(ctx context.Context, clusterID string) (err error)
	Get(ctx context.Context, clusterID string) (*Cluster, error)
	GetMeta(ctx context.Context, clusterID string) (*Cluster, []*ClusterInstance, error)

	//
	// UpdateInstance update cluster instances
	//  @Description:
	//  @param ctx
	//  @param instances
	//  @return error
	//
	UpdateInstance(ctx context.Context, instances ...ClusterInstance) error

	//
	// UpdateBaseInfo update cluster base info
	//  @Description:
	//  @param ctx
	//  @param template
	//  @return error
	//
	UpdateBaseInfo(ctx context.Context, template *Cluster) error

	//
	// UpdateStatus
	//  @Description:
	//  @param ctx
	//  @param clusterID
	//  @param status
	//  @return error
	//
	UpdateStatus(ctx context.Context, clusterID string, status constants.ClusterRunningStatus) error

	//
	// SetMaintenanceStatus
	//  @Description:
	//  @param ctx
	//  @param clusterID
	//  @param targetStatus
	//  @return error
	//
	SetMaintenanceStatus(ctx context.Context, clusterID string, targetStatus constants.ClusterMaintenanceStatus) error

	//
	// ClearMaintenanceStatus
	//  @Description:
	//  @param ctx
	//  @param clusterID
	//  @param originalStatus
	//  @return error
	//
	ClearMaintenanceStatus(ctx context.Context, clusterID string, originalStatus constants.ClusterMaintenanceStatus) error
}
