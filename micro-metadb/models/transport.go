
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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"context"
	"strconv"
	"time"
)

type TransportRecord struct {
	Record
	ClusterId     string
	TransportType string
	FilePath      string
	TenantId      string
	Status        string
	StartTime     time.Time
	EndTime       time.Time
}

func (m *DAOClusterManager) CreateTransportRecord(ctx context.Context, record *TransportRecord) (id string, err error) {
	err = m.Db(ctx).Create(record).Error
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(record.ID)), nil
}

func (m *DAOClusterManager) UpdateTransportRecord(ctx context.Context, id, clusterId, status string, endTime time.Time) (err error) {
	record := TransportRecord{
		ClusterId: clusterId,
	}
	uintId, _ := strconv.ParseInt(id, 10, 64)
	record.ID = uint(uintId)
	err = m.Db(ctx).Model(&record).Updates(map[string]interface{}{"Status": status, "EndTime": endTime}).Error
	return err
}

func (m *DAOClusterManager) FindTransportRecordById(ctx context.Context, id string) (record *TransportRecord, err error) {
	record = &TransportRecord{}
	uintId, _ := strconv.ParseInt(id, 10, 64)

	err = m.Db(ctx).Where("id = ?", uintId).First(record).Error
	if err != nil {
		return record, err
	}
	return record, nil
}

func (m *DAOClusterManager) ListTransportRecord(ctx context.Context, clusterId string, recordId string, offset int32, length int32) (records []*TransportRecord, total int64, err error) {
	records = make([]*TransportRecord, length)

	db := m.Db(ctx).Table(TABLE_NAME_TRANSPORT_RECORD).Where("cluster_id = ?", clusterId)
	if recordId != "" {
		db.Where("id = ?", recordId)
	}
	err = db.Count(&total).Order("id desc").Offset(int(offset)).Limit(int(length)).Find(&records).Error

	return
}
