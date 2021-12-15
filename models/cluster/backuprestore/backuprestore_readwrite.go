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
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"time"
)

type BRReadWrite struct {
	dbCommon.GormDB
}

func NewBRReadWrite(db *gorm.DB) *BRReadWrite {
	m := &BRReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *BRReadWrite) CreateBackupRecord(ctx context.Context, record *BackupRecord) (*BackupRecord, error) {
	return record, m.DB(ctx).Create(record).Error
}

func (m *BRReadWrite) UpdateBackupRecord(ctx context.Context, backupId string, status string, size uint64, backupTso uint64, endTime time.Time) (err error) {
	if "" == backupId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	record := &BackupRecord{}
	err = m.DB(ctx).First(record, "id = ?", backupId).Error
	if err != nil {
		return framework.SimpleError(common.TIEM_BACKUP_RECORD_NOT_FOUND)
	}

	db := m.DB(ctx).Model(record)
	if "" != status {
		db.Update("status", status)
	}
	if size > 0 {
		db.Update("size", size)
	}
	if backupTso > 0 {
		db.Update("backup_tso", backupTso)
	}
	if !endTime.IsZero() {
		db.Update("end_time", endTime)
	}

	return db.Error
}

func (m *BRReadWrite) GetBackupRecord(ctx context.Context, backupId string) (record *BackupRecord, err error) {
	if "" == backupId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	record = &BackupRecord{}
	err = m.DB(ctx).First(record, "id = ?", backupId).Error
	if err != nil {
		return nil, framework.SimpleError(common.TIEM_BACKUP_RECORD_NOT_FOUND)
	}
	return record, err
}

func (m *BRReadWrite) QueryBackupRecords(ctx context.Context, clusterId, backupId, backupMode string, startTime, endTime time.Time, page int, pageSize int) (records []*BackupRecord, total int64, err error) {
	records = make([]*BackupRecord, pageSize)
	query := m.DB(ctx).Model(BackupRecord{})
	if backupId != "" {
		query = query.Where("id = ?", backupId)
	}
	if backupMode != "" {
		query = query.Where("backup_mode = ?", backupMode)
	}
	if clusterId != "" {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}
	err = query.Order("id desc").Count(&total).Offset(pageSize * (page - 1)).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (m *BRReadWrite) DeleteBackupRecord(ctx context.Context, backupId string) (err error) {
	if "" == backupId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	record := &BackupRecord{}
	return m.DB(ctx).First(record, "id = ?", backupId).Delete(record).Error
}

func (m *BRReadWrite) CreateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (*BackupStrategy, error) {
	return strategy, m.DB(ctx).Create(strategy).Error
}

func (m *BRReadWrite) UpdateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (err error) {
	return m.DB(ctx).Model(strategy).Save(strategy).Error
}

func (m *BRReadWrite) GetBackupStrategy(ctx context.Context, clusterId string) (strategy *BackupStrategy, err error) {
	if "" == clusterId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	strategy = &BackupStrategy{}
	err = m.DB(ctx).First(strategy, "cluster_id = ?", clusterId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, framework.SimpleError(common.TIEM_BACKUP_STRATEGY_NOT_FOUND)
	}
	return strategy, err
}

func (m *BRReadWrite) QueryBackupStrategy(ctx context.Context, weekDay string, startHour uint32) (strategies []*BackupStrategy, err error) {
	query := m.DB(ctx).Model(BackupStrategy{})
	if weekDay != "" {
		query = query.Where("backup_date like '%" + weekDay + "%'")
	}
	if startHour != 0 {
		query = query.Where("start_hour = ?", startHour)
	}
	err = query.Find(&strategies).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else {
		return strategies, nil
	}
}

func (m *BRReadWrite) DeleteBackupStrategy(ctx context.Context, clusterId string) (err error) {
	if "" == clusterId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	strategy := &BackupStrategy{}
	return m.DB(ctx).First(strategy, "cluster_id = ?", clusterId).Delete(strategy).Error
}
