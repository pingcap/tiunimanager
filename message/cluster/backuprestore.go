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

// CancelBackupReq Requests for cancel backup
type CancelBackupReq struct {
	ClusterID string `json:"clusterId"`
	BackupID  string `json:"backupId"`
}

// CancelBackupResp Response for cancel backup
type CancelBackupResp struct {
}

// DeleteBackupDataReq Delete a backup file based on the cluster ID and the ID of the backup file
type DeleteBackupDataReq struct {
	ClusterID        string   `json:"clusterId"`
	BackupMode       string   `json:"backupMode"`
	BackupID         string   `json:"backupId" swaggerignore:"true"`
	ExcludeBackupIDs []string `json:"excludeBackupId" swaggerignore:"true"`
}

// DeleteBackupDataResp Delete a backup file reply message
type DeleteBackupDataResp struct {
}

// QueryBackupRecordsReq Query the list of backup files over time based on cluster ID
type QueryBackupRecordsReq struct {
	BackupID  string `json:"backupId" form:"backupId"`
	ClusterID string `json:"clusterId" form:"clusterId"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
	structs.PageRequest
}

// QueryBackupRecordsResp Query the return information of the backup file
type QueryBackupRecordsResp struct {
	BackupRecords []*structs.BackupRecord `json:"backupRecords"`
}

// SaveBackupStrategyReq Request to update backup data strategy
type SaveBackupStrategyReq struct {
	ClusterID string                 `json:"clusterId" swaggerignore:"true"`
	Strategy  structs.BackupStrategy `json:"strategy"`
}

// SaveBackupStrategyResp update backup strategy reply message
type SaveBackupStrategyResp struct {
}

// GetBackupStrategyReq Query messages for cluster backup strategy
type GetBackupStrategyReq struct {
	ClusterID string `json:"clusterId" form:"clusterId" swaggerignore:"true"`
}

// GetBackupStrategyResp Query backup strategy reply message
type GetBackupStrategyResp struct {
	Strategy structs.BackupStrategy `json:"strategy"`
}

// DeleteBackupStrategyReq Request to delete backup data strategy
type DeleteBackupStrategyReq struct {
	ClusterID string `json:"clusterId"`
}

// DeleteBackupStrategyResp delete backup strategy reply message
type DeleteBackupStrategyResp struct {
}
