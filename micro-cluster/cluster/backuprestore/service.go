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
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/message/cluster"
)

type BRService interface {
	// BackupCluster
	// @Description: backup cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter maintenanceStatusChange
	// @Return cluster.BackupClusterDataResp
	// @Return error
	BackupCluster(ctx context.Context, request cluster.BackupClusterDataReq, maintenanceStatusChange bool) (resp cluster.BackupClusterDataResp, backupErr error)

	// RestoreExistCluster
	// @Description: restore exist cluster by backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter maintenanceStatusChange
	// @Return cluster.RestoreExistClusterResp
	// @Return error
	RestoreExistCluster(ctx context.Context, request cluster.RestoreExistClusterReq, maintenanceStatusChange bool) (resp cluster.RestoreExistClusterResp, restoreErr error)

	// CancelBackup
	// @Description: cancel backup cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.CancelBackupResp
	// @Return error
	CancelBackup(ctx context.Context, request cluster.CancelBackupReq) (resp cluster.CancelBackupResp, cancelErr error)

	// QueryClusterBackupRecords
	// @Description: query backup records of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.QueryBackupRecordsResp
	// @Return structs.Page
	// @Return error
	QueryClusterBackupRecords(ctx context.Context, request cluster.QueryBackupRecordsReq) (resp cluster.QueryBackupRecordsResp, page structs.Page, err error)

	// DeleteBackupRecords
	// @Description: delete backup records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.QueryBackupRecordsResp
	// @Return error
	DeleteBackupRecords(ctx context.Context, request cluster.DeleteBackupDataReq) (resp cluster.DeleteBackupDataResp, err error)

	// SaveBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.SaveBackupStrategyResp
	// @Return error
	SaveBackupStrategy(ctx context.Context, request cluster.SaveBackupStrategyReq) (resp cluster.SaveBackupStrategyResp, err error)

	// GetBackupStrategy
	// @Description: get backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.GetBackupStrategyResp
	// @Return error
	GetBackupStrategy(ctx context.Context, request cluster.GetBackupStrategyReq) (resp cluster.GetBackupStrategyResp, err error)

	// DeleteBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.UpdateBackupStrategyResp
	// @Return error
	DeleteBackupStrategy(ctx context.Context, request cluster.DeleteBackupStrategyReq) (resp cluster.DeleteBackupStrategyResp, err error)
}
