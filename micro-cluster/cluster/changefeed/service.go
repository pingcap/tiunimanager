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

package changefeed

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
)

type Service interface {
	//
	// CreateBetweenClusters
	// @Description: create a change feed task for replicating the incremental data of source cluster to target cluster
	// @param ctx
	// @param sourceClusterID
	// @param targetClusterID
	// @param relationType
	// @return ID
	// @return err
	//
	CreateBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error)

	//
	// ReverseBetweenClusters reverse change feed task
	// @Description: it will delete change feed task of source cluster, then create a new one for target
	// @param ctx
	// @param sourceClusterID
	// @param targetClusterID
	// @param relationType
	// @return ID
	// @return err
	//
	ReverseBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error)
}

func GetChangeFeedService() Service {
	return nil
}