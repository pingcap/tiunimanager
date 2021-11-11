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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDAOClusterManager_CreateTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       11,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	assert.Equal(t, 11, id)
	assert.NoError(t, err)
}

func TestDAOClusterManager_UpdateTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       22,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	assert.Equal(t, 22, id)
	assert.NoError(t, err)

	err = Dao.ClusterManager().UpdateTransportRecord(context.TODO(), id, record.ClusterId, time.Now())
	assert.NoError(t, err)
}

func TestDAOClusterManager_FindTransportRecordById(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       33,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	assert.Equal(t, 33, id)
	assert.NoError(t, err)

	_, err = Dao.ClusterManager().FindTransportRecordById(context.TODO(), id)
	assert.NoError(t, err)
}

func TestDAOClusterManager_ListTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       44,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	assert.Equal(t, 44, id)
	assert.NoError(t, err)

	_, _, err = Dao.ClusterManager().ListTransportRecord(context.TODO(), record.ClusterId, 44, 0, 10)
	assert.NoError(t, err)
}

func TestDAOClusterManager_DeleteTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       55,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	assert.Equal(t, 55, id)
	assert.NoError(t, err)

	_, err = Dao.ClusterManager().DeleteTransportRecord(context.TODO(), id)
	assert.NoError(t, err)
}
