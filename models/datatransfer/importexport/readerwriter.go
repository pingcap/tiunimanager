/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package importexport

import (
	"context"
	"time"
)

type ReaderWriter interface {
	// CreateDataTransportRecord
	// @Description: create new data transport record
	// @Receiver m
	// @Parameter ctx
	// @Parameter record
	// @Return *DataTransportRecord
	// @Return error
	CreateDataTransportRecord(ctx context.Context, record *DataTransportRecord) (*DataTransportRecord, error)

	// UpdateDataTransportRecord
	// @Description: update data transport record
	// @Receiver m
	// @Parameter ctx
	// @Parameter recordId
	// @Parameter status
	// @Parameter endTime
	// @Return error
	UpdateDataTransportRecord(ctx context.Context, recordId string, status string, endTime time.Time) (err error)

	// GetDataTransportRecord
	// @Description: get data transport record by id
	// @Receiver m
	// @Parameter ctx
	// @Parameter recordId
	// @Return *DataTransportRecord
	// @Return error
	GetDataTransportRecord(ctx context.Context, recordId string) (record *DataTransportRecord, err error)

	// QueryDataTransportRecords
	// @Description: query data transport records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter recordId
	// @Parameter clusterId
	// @Parameter reImport
	// @Parameter startTime
	// @Parameter endTime
	// @Parameter page
	// @Parameter pageSize
	// @return []*DataTransportRecord
	// @Return total
	// @Return error
	QueryDataTransportRecords(ctx context.Context, recordId string, clusterId string, reImport bool, startTime, endTime int64, page int, pageSize int) (records []*DataTransportRecord, total int64, err error)

	// DeleteDataTransportRecord
	// @Description: delete data transport record by id
	// @Receiver m
	// @Parameter ctx
	// @Parameter recordId
	// @Return error
	DeleteDataTransportRecord(ctx context.Context, recordId string) (err error)
}
