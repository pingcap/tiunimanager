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
	if err != nil {
		t.Errorf("TestDAOClusterManager_CreateTransportRecord failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_CreateTransportRecord success, id: %d", id)
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
	if err != nil {
		t.Errorf("TestDAOClusterManager_UpdateTransportRecord create record failed, %s", err.Error())
		return
	}

	err = Dao.ClusterManager().UpdateTransportRecord(context.TODO(), id, record.ClusterId, time.Now())
	if err != nil {
		t.Errorf("TestDAOClusterManager_UpdateTransportRecord update record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_UpdateTransportRecord success")
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
	if err != nil {
		t.Errorf("TestDAOClusterManager_FindTransportRecordById create record failed, %s", err.Error())
		return
	}

	findRecord, err := Dao.ClusterManager().FindTransportRecordById(context.TODO(), id)
	if err != nil {
		t.Errorf("TestDAOClusterManager_FindTransportRecordById find record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_FindTransportRecordById success, record: %v", findRecord)
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
	_, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	if err != nil {
		t.Errorf("TestDAOClusterManager_ListTransportRecord create record failed, %s", err.Error())
		return
	}

	list, total, err := Dao.ClusterManager().ListTransportRecord(context.TODO(), record.ClusterId, 44, 0, 10)
	if err != nil {
		t.Errorf("TestDAOClusterManager_ListTransportRecord create record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_ListTransportRecord success, total: %d, list: %v", total, list)
}
