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
	"github.com/pingcap-inc/tiem/common/errors"
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
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "backup id cannot be empty")
	}

	record := &BackupRecord{}
	err = m.DB(ctx).First(record, "id = ?", backupId).Error
	if err != nil {
		return err
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
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "backup id cannot be empty")
	}
	record = &BackupRecord{}
	err = m.DB(ctx).First(record, "id = ?", backupId).Error
	if err != nil {
		return nil, err
	}
	return record, err
}

func (m *BRReadWrite) QueryBackupRecords(ctx context.Context, clusterId, backupId, backupMode string, startTime, endTime int64, page int, pageSize int) (records []*BackupRecord, total int64, err error) {
	records = make([]*BackupRecord, pageSize)
	query := m.DB(ctx).Model(&BackupRecord{}).Where("deleted_at is null")
	if backupId != "" {
		query = query.Where("id = ?", backupId)
	}
	if backupMode != "" {
		query = query.Where("backup_mode = ?", backupMode)
	}
	if clusterId != "" {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if startTime > 0 {
		query = query.Where("start_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("end_time <= ?", endTime)
	}
	err = query.Order("created_at desc").Count(&total).Offset(pageSize * (page - 1)).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (m *BRReadWrite) DeleteBackupRecord(ctx context.Context, backupId string) (err error) {
	if "" == backupId {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "backup id cannot be empty")
	}
	record := &BackupRecord{}
	return m.DB(ctx).First(record, "id = ?", backupId).Unscoped().Delete(record).Error
}

func (m *BRReadWrite) DeleteBackupRecords(ctx context.Context, backupIds []string) (err error) {
	if len(backupIds) <= 0 {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "backup ids cannot be empty")
	}
	records := &[]BackupRecord{}
	return m.DB(ctx).Find(records, "id in ?", backupIds).Unscoped().Delete(records).Error
}

func (m *BRReadWrite) CreateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (*BackupStrategy, error) {
	return strategy, m.DB(ctx).Create(strategy).Error
}

func (m *BRReadWrite) UpdateBackupStrategy(ctx context.Context, strategy *BackupStrategy) (err error) {
	columnMap := make(map[string]interface{})
	columnMap["backup_date"] = strategy.BackupDate
	columnMap["start_hour"] = strategy.StartHour
	columnMap["end_hour"] = strategy.EndHour
	return m.DB(ctx).Model(strategy).Where("cluster_id = ?", strategy.ClusterID).Updates(columnMap).Error
}

func (m *BRReadWrite) SaveBackupStrategy(ctx context.Context, strategy *BackupStrategy) (*BackupStrategy, error) {
	existStrategy, err := m.GetBackupStrategy(ctx, strategy.ClusterID)
	if err != nil || existStrategy.ID == "" {
		return m.CreateBackupStrategy(ctx, strategy)
	} else {
		return strategy, m.UpdateBackupStrategy(ctx, strategy)
	}
}

func (m *BRReadWrite) GetBackupStrategy(ctx context.Context, clusterId string) (strategy *BackupStrategy, err error) {
	if "" == clusterId {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "cluster id cannot be empty")
	}

	strategy = &BackupStrategy{}
	err = m.DB(ctx).First(strategy, "cluster_id = ?", clusterId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return strategy, nil
}

func (m *BRReadWrite) QueryBackupStrategy(ctx context.Context, weekDay string, startHour uint32) (strategies []*BackupStrategy, err error) {
	query := m.DB(ctx).Model(&BackupStrategy{})
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
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "cluster id cannot be empty")
	}
	strategy, err := m.GetBackupStrategy(ctx, clusterId)
	if err != nil {
		return err
	}
	if strategy.ID == "" {
		return nil
	}

	return m.DB(ctx).First(strategy, "cluster_id = ?", clusterId).Unscoped().Delete(strategy).Error
}
