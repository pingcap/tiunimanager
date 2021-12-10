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

/*******************************************************************************
 * @File: backuprestore.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package cluster

import (
	"github.com/pingcap-inc/tiem/common/structs"
	"time"
)

// BackupClusterDataReq Requests for manual data backup
type BackupClusterDataReq struct {
	ClusterID  string `json:"clusterId"`
	BackupType string `json:"backupType"` //full,incr
	BackupMode string `json:"backupMode"` //auto,manual
}

// BackupClusterDataResp Cluster backup reply message
type BackupClusterDataResp struct {
	structs.AsyncTaskWorkFlowInfo
	BackupID string `json:"backupId"`
}

// DeleteBackupDataReq Delete a backup file based on the cluster ID and the ID of the backup file
type DeleteBackupDataReq struct {
	ClusterID string `json:"clusterId"`
}

// DeleteBackupDataResp Delete a backup file reply message
type DeleteBackupDataResp struct {
	BackupID string `json:"backupId"`
}

// QueryBackupRecordsReq Query the list of backup files over time based on cluster ID
type QueryBackupRecordsReq struct {
	BackupID  string    `json:"backupId"`
	ClusterID string    `json:"clusterId" form:"clusterId"`
	StartTime time.Time `json:"startTime" form:"startTime"`
	EndTime   time.Time `json:"endTime" form:"endTime"`
	structs.PageRequest
}

// QueryBackupRecordsResp Query the return information of the backup file
type QueryBackupRecordsResp struct {
	BackupFileItems []*structs.BackupRecord
}

// UpdateBackupStrategyReq Request to update backup data strategy
type UpdateBackupStrategyReq struct {
	ClusterID string                 `json:"clusterId"`
	Strategy  structs.BackupStrategy `json:"strategy"`
}

// UpdateBackupStrategyResp update backup strategy reply message
type UpdateBackupStrategyResp struct {
}

// QueryBackupStrategyReq Query messages for cluster backup strategy
type QueryBackupStrategyReq struct {
	ClusterID string `json:"clusterId" form:"clusterId"`
}

// QueryBackupStrategyResp Query backup strategy reply message
type QueryBackupStrategyResp struct {
	Strategy structs.BackupStrategy `json:"strategy"`
}

// DeleteBackupStrategyReq Request to delete backup data strategy
type DeleteBackupStrategyReq struct {
	ClusterID string                 `json:"clusterId"`
	Strategy  structs.BackupStrategy `json:"strategy"`
}

// DeleteBackupStrategyResp delete backup strategy reply message
type DeleteBackupStrategyResp struct {
}
