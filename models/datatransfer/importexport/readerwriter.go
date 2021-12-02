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

package importexport

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type DataTransportReaderWriter interface {
	SetDb(db *gorm.DB)
	Db(ctx context.Context) *gorm.DB

	CreateDataTransportRecord(ctx context.Context, record *DataTransportRecord) (err error)
	UpdateDataTransportRecord(ctx context.Context, recordId string, status string, endTime time.Time) (err error)
	GetDataTransportRecord(ctx context.Context, recorId string) (record *DataTransportRecord, err error)
	QueryDataTransportRecords(ctx context.Context, recordId, clusterId string, reImport bool, startTime, endTime time.Time, page int, pageSize int) (records []*DataTransportRecord, total int64, err error)
	DeleteDataTransportRecord(ctx context.Context, recordId string) (err error)
}
