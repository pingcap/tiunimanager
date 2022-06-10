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

package importexport

import (
	"context"
	"github.com/pingcap-inc/tiunimanager/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var rw *ImportExportReadWrite

func TestImportExportReadWrite_CreateDataTransportRecord(t *testing.T) {
	record := &DataTransportRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "ImportInitStatus",
		},
		ClusterID:       "clusterId",
		TransportType:   "import",
		FilePath:        "/tmp/test",
		ZipName:         "data.zip",
		StorageType:     "s3",
		Comment:         "test",
		ReImportSupport: false,
		StartTime:       time.Now(),
	}
	_, err := rw.CreateDataTransportRecord(context.TODO(), record)
	assert.NoError(t, err)
}

func TestImportExportReadWrite_UpdateDataTransportRecord(t *testing.T) {
	record := &DataTransportRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "ImportInitStatus",
		},
		ClusterID:       "clusterId",
		TransportType:   "import",
		FilePath:        "/tmp/test",
		ZipName:         "data.zip",
		StorageType:     "s3",
		Comment:         "test",
		ReImportSupport: false,
		StartTime:       time.Now(),
	}
	recordCreate, errCreate := rw.CreateDataTransportRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	endTime := time.Now()
	errUpdate := rw.UpdateDataTransportRecord(context.TODO(), recordCreate.ID, "end", endTime)
	assert.NoError(t, errUpdate)

	recordGet, errGet := rw.GetDataTransportRecord(context.TODO(), recordCreate.ID)
	assert.NoError(t, errGet)
	assert.Equal(t, "end", recordGet.Status)
	assert.Equal(t, true, recordGet.EndTime.Equal(endTime))

	_, errGet = rw.GetDataTransportRecord(context.TODO(), "")
	assert.Error(t, errGet)

	errUpdate = rw.UpdateDataTransportRecord(context.TODO(), "", "end", endTime)
	assert.Error(t, errUpdate)

	errUpdate = rw.UpdateDataTransportRecord(context.TODO(), "aaaaa", "end", endTime)
	assert.Error(t, errUpdate)

}

func TestImportExportReadWrite_QueryDataTransportRecords(t *testing.T) {
	record := &DataTransportRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "ImportInitStatus",
		},
		ClusterID:       "clusterId",
		TransportType:   "import",
		FilePath:        "/tmp/import",
		ZipName:         "data.zip",
		StorageType:     "s3",
		Comment:         "test",
		ReImportSupport: false,
		StartTime:       time.Now(),
	}
	recordCreate, errCreate := rw.CreateDataTransportRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	t.Run("normal", func(t *testing.T) {

		recordQuery, total, errQuery := rw.QueryDataTransportRecords(context.TODO(), recordCreate.ID, "", false, 0, 0, 1, 10)
		assert.NoError(t, errQuery)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, recordCreate.ID, recordQuery[0].ID)
		assert.Equal(t, recordCreate.StorageType, recordQuery[0].StorageType)
	})

	t.Run("empty", func(t *testing.T) {
		_, count, errQuery := rw.QueryDataTransportRecords(context.TODO(), recordCreate.ID, "aaa", true, 11111111, 11111111111, 1, 10)
		assert.NoError(t, errQuery)
		assert.Equal(t, int64(0), count)

	})

}

func TestImportExportReadWrite_DeleteDataTransportRecord(t *testing.T) {
	record := &DataTransportRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "ImportInitStatus",
		},
		ClusterID:       "clusterId",
		TransportType:   "import",
		FilePath:        "/tmp/import",
		ZipName:         "data.zip",
		StorageType:     "s3",
		Comment:         "test",
		ReImportSupport: false,
		StartTime:       time.Now(),
	}
	recordCreate, errCreate := rw.CreateDataTransportRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	errDelete := rw.DeleteDataTransportRecord(context.TODO(), recordCreate.ID)
	assert.NoError(t, errDelete)

	recordGet, errGet := rw.GetDataTransportRecord(context.TODO(), recordCreate.ID)
	assert.Nil(t, recordGet)
	assert.NotNil(t, errGet)

	errDelete = rw.DeleteDataTransportRecord(context.TODO(), "")
	assert.Error(t, errDelete)

}
