/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package backuprestore

import (
	"context"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
)

type BRService interface {
	// BackupCluster
	// @Description: backup cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.BackupClusterDataResp
	// @Return error
	BackupCluster(ctx context.Context, request *cluster.BackupClusterDataReq) (*cluster.BackupClusterDataResp, error)

	// RestoreNewCluster
	// @Description: restore a new cluster by backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.RestoreNewClusterResp
	// @Return error
	//RestoreNewCluster(ctx context.Context, request *cluster.RestoreNewClusterReq) (*cluster.RestoreNewClusterResp, error) // todo: move to cluster manager

	// RestoreExistCluster
	// @Description: restore exist cluster by backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.RestoreExistClusterResp
	// @Return error
	RestoreExistCluster(ctx context.Context, request *cluster.RestoreExistClusterReq) (*cluster.RestoreExistClusterResp, error)

	// QueryClusterBackupRecords
	// @Description: query backup records of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.QueryBackupRecordsResp
	// @Return *structs.Page
	// @Return error
	QueryClusterBackupRecords(ctx context.Context, request *cluster.QueryBackupRecordsReq) (*cluster.QueryBackupRecordsResp, *structs.Page, error)

	// DeleteBackupRecords
	// @Description: delete backup records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.QueryBackupRecordsResp
	// @Return error
	DeleteBackupRecords(ctx context.Context, request *cluster.DeleteBackupDataReq) (*cluster.DeleteBackupDataResp, error)

	// GetBackupStrategy
	// @Description: get backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.GetBackupStrategyResp
	// @Return error
	GetBackupStrategy(ctx context.Context, request *cluster.GetBackupStrategyReq) (*cluster.GetBackupStrategyResp, error)

	// SaveBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.UpdateBackupStrategyResp
	// @Return error
	SaveBackupStrategy(ctx context.Context, request *cluster.UpdateBackupStrategyReq) (*cluster.UpdateBackupStrategyResp, error)

	// DeleteBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *cluster.UpdateBackupStrategyResp
	// @Return error
	DeleteBackupStrategy(ctx context.Context, request *cluster.DeleteBackupStrategyReq) (*cluster.DeleteBackupStrategyResp, error)
}
