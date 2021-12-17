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

package backuprestore

import (
	"context"
	"time"
)

type ReaderWriter interface {
	// CreateBackupRecord
	// @Description: create new backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter record
	// @Return *BackupRecord
	// @Return error
	CreateBackupRecord(ctx context.Context, record *BackupRecord) (*BackupRecord, error)

	// UpdateBackupRecord
	// @Description: update backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter backupId
	// @Parameter status
	// @Parameter size
	// @Parameter backupTso
	// @Parameter endTime
	// @Return error
	UpdateBackupRecord(ctx context.Context, backupId string, status string, size uint64, backupTso uint64, endTime time.Time) (err error)

	// GetBackupRecord
	// @Description: get backup record by Id
	// @Receiver m
	// @Parameter ctx
	// @Parameter backupId
	// @Return *BackupRecord
	// @Return error
	GetBackupRecord(ctx context.Context, backupId string) (record *BackupRecord, err error)

	// QueryBackupRecords
	// @Description: query backup records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter backupId
	// @Parameter backupMode
	// @Parameter clusterId
	// @Parameter startTime
	// @Parameter endTime
	// @Parameter page
	// @Parameter pageSize
	// @Return []*BackupRecord
	// @Return total
	// @Return error
	QueryBackupRecords(ctx context.Context, clusterId, backupId, backupMode string, startTime, endTime int64, page int, pageSize int) (records []*BackupRecord, total int64, err error)

	// DeleteBackupRecord
	// @Description: delete backup record by Id
	// @Receiver m
	// @Parameter ctx
	// @Parameter backupId
	// @Return error
	DeleteBackupRecord(ctx context.Context, backupId string) (err error)

	// CreateBackupStrategy
	// @Description: create new backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter strategy
	// @Return *BackupStrategy
	// @Return error
	CreateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (*BackupStrategy, error)

	// UpdateBackupStrategy
	// @Description: update backup strategy
	// @Receiver m
	// @Parameter ctx
	// @Parameter strategy
	// @Return error
	UpdateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (err error)

	// GetBackupStrategy
	// @Description: get backup strategy by clusterId
	// @Receiver m
	// @Parameter ctx
	// @Parameter clusterId
	// @Return *BackupStrategy
	// @Return error
	GetBackupStrategy(ctx context.Context, clusterId string) (strategy *BackupStrategy, err error)

	// QueryBackupStrategy
	// @Description: query backup strategies by time
	// @Receiver m
	// @Parameter ctx
	// @Parameter weekDay
	// @Parameter startHour
	// @Return []*BackupStrategy
	// @Return error
	QueryBackupStrategy(ctx context.Context, weekDay string, startHour uint32) (strategies []*BackupStrategy, err error)

	// DeleteBackupStrategy
	// @Description: delete backup strategy by clusterId
	// @Receiver m
	// @Parameter ctx
	// @Parameter clusterId
	// @Return error
	DeleteBackupStrategy(ctx context.Context, clusterId string) (err error)
}
