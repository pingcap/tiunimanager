/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package importexport

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
	"time"
)

type ImportExportReadWrite struct {
	dbCommon.GormDB
}

func NewImportExportReadWrite(db *gorm.DB) *ImportExportReadWrite {
	m := &ImportExportReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *ImportExportReadWrite) CreateDataTransportRecord(ctx context.Context, record *DataTransportRecord) (*DataTransportRecord, error) {
	return record, m.DB(ctx).Create(record).Error
}

func (m *ImportExportReadWrite) UpdateDataTransportRecord(ctx context.Context, recordId string, status string, endTime time.Time) (err error) {
	if "" == recordId {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "record id required")
	}

	record := &DataTransportRecord{}
	err = m.DB(ctx).First(record, "id = ?", recordId).Error
	if err != nil {
		return err
	}

	return m.DB(ctx).Model(record).
		Update("status", status).
		Update("end_time", endTime).Error
}

func (m *ImportExportReadWrite) GetDataTransportRecord(ctx context.Context, recordId string) (record *DataTransportRecord, err error) {
	if "" == recordId {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "record id required")
	}
	record = &DataTransportRecord{}
	err = m.DB(ctx).First(record, "id = ?", recordId).Error
	if err != nil {
		return nil, err
	}
	return record, err
}

func (m *ImportExportReadWrite) QueryDataTransportRecords(ctx context.Context, recordId string, clusterId string, reImport bool, startTime, endTime int64, page int, pageSize int) (records []*DataTransportRecord, total int64, err error) {
	records = make([]*DataTransportRecord, 0)
	query := m.DB(ctx).Model(&DataTransportRecord{}).Where("deleted_at is null")
	if recordId != "" {
		query = query.Where("id = ?", recordId)
	}
	if clusterId != "" {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if reImport {
		query = query.Where("re_import_support = ?", reImport).Where("status = ?", constants.DataImportExportFinished)
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

func (m *ImportExportReadWrite) DeleteDataTransportRecord(ctx context.Context, recordId string) (err error) {
	if "" == recordId {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "record id required")
	}
	record := &DataTransportRecord{}
	return m.DB(ctx).First(record, "id = ?", recordId).Delete(record).Error
}
