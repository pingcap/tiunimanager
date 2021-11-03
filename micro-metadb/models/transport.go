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
	"fmt"
	"time"
)

type TransportRecord struct {
	Record
	ClusterId     string
	TransportType string
	FilePath      string
	StorageType   string
	FlowId        int64
	StartTime     time.Time
	EndTime       time.Time
}

type TransportRecordFetchResult struct {
	TransportRecord *TransportRecord
	Flow            *FlowDO
}

func (m *DAOClusterManager) CreateTransportRecord(ctx context.Context, record *TransportRecord) (recordId int, err error) {
	err = m.Db(ctx).Create(record).Error
	if err != nil {
		return 0, err
	}
	return int(record.ID), nil
}

func (m *DAOClusterManager) UpdateTransportRecord(ctx context.Context, recordId int, clusterId string, endTime time.Time) (err error) {
	record := TransportRecord{
		ClusterId: clusterId,
	}
	record.ID = uint(recordId)
	err = m.Db(ctx).Model(&record).Updates(map[string]interface{}{"EndTime": endTime}).Error
	return err
}

func (m *DAOClusterManager) FindTransportRecordById(ctx context.Context, recordId int) (record *TransportRecord, err error) {
	record = &TransportRecord{}
	err = m.Db(ctx).Where("id = ?", recordId).First(record).Error
	if err != nil {
		return record, err
	}
	return record, nil
}

func (m *DAOClusterManager) ListTransportRecord(ctx context.Context, clusterId string, recordId int, offset int32, length int32) (dos []*TransportRecordFetchResult, total int64, err error) {
	records := make([]*TransportRecord, length)

	db := m.Db(ctx).Table(TABLE_NAME_TRANSPORT_RECORD)
	if clusterId != "" {
		db.Where("cluster_id = ?", clusterId)
	}
	if recordId > 0 {
		db.Where("id = ?", recordId)
	}
	err = db.Count(&total).Order("id desc").Offset(int(offset)).Limit(int(length)).Find(&records).Error
	if err == nil {
		// query flows
		flowIds := make([]int64, len(records))
		dos = make([]*TransportRecordFetchResult, len(records))
		for i, r := range records {
			flowIds[i] = r.FlowId
			dos[i] = &TransportRecordFetchResult{
				TransportRecord: r,
			}
		}

		flows := make([]*FlowDO, len(records))
		err = m.Db(ctx).Find(&flows, flowIds).Error
		m.HandleMetrics(TABLE_NAME_FLOW, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("ListTransportRecord, query record failed, clusterId: %s, error: %v", clusterId, err)
		}

		flowMap := make(map[int64]*FlowDO)
		for _, v := range flows {
			flowMap[int64(v.ID)] = v
		}
		for i, v := range records {
			dos[i].TransportRecord = v
			dos[i].Flow = flowMap[v.FlowId]
		}
	}

	return
}
