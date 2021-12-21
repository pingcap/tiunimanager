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
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var rw *BRReadWrite

func TestBRReadWrite_CreateBackupRecord(t *testing.T) {
	record := &BackupRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:    "clusterId",
		FilePath:     "/tmp/test",
		StorageType:  "s3",
		BackupType:   "full",
		BackupMethod: "logic",
		BackupMode:   "auto",
		Size:         23346546456,
		BackupTso:    42353454343234,
		StartTime:    time.Now(),
	}
	_, err := rw.CreateBackupRecord(context.TODO(), record)
	assert.NoError(t, err)
}

func TestBRReadWrite_GetBackupRecord(t *testing.T) {
	record := &BackupRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:    "clusterId",
		FilePath:     "/tmp/test",
		StorageType:  "s3",
		BackupType:   "full",
		BackupMethod: "logic",
		BackupMode:   "auto",
		Size:         23346546456,
		BackupTso:    42353454343234,
		StartTime:    time.Now(),
	}
	recordCreate, errCreate := rw.CreateBackupRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	recordGet, errGet := rw.GetBackupRecord(context.TODO(), recordCreate.ID)
	assert.NoError(t, errGet)
	assert.Equal(t, recordCreate.ID, recordGet.ID)
	assert.Equal(t, recordCreate.Status, recordGet.Status)
	assert.Equal(t, recordCreate.BackupMode, recordGet.BackupMode)
}

func TestBRReadWrite_UpdateBackupRecord(t *testing.T) {
	record := &BackupRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:    "clusterId",
		FilePath:     "/tmp/test",
		StorageType:  "s3",
		BackupType:   "full",
		BackupMethod: "logic",
		BackupMode:   "auto",
		Size:         23346546456,
		BackupTso:    42353454343234,
		StartTime:    time.Now(),
	}
	recordCreate, errCreate := rw.CreateBackupRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	recordCreate.Size = 123
	recordCreate.BackupTso = 354365767866
	recordCreate.Status = "BackupEndStatus"
	recordCreate.EndTime = time.Now()
	errUpdate := rw.UpdateBackupRecord(context.TODO(), recordCreate.ID, recordCreate.Status, recordCreate.Size, recordCreate.BackupTso, recordCreate.EndTime)
	assert.NoError(t, errUpdate)

	recordGet, errGet := rw.GetBackupRecord(context.TODO(), recordCreate.ID)
	assert.NoError(t, errGet)
	assert.Equal(t, recordCreate.Size, recordGet.Size)
	assert.Equal(t, recordCreate.Status, recordGet.Status)
	assert.Equal(t, recordCreate.BackupTso, recordGet.BackupTso)
	assert.Equal(t, true, recordGet.EndTime.Equal(recordCreate.EndTime))
}

func TestBRReadWrite_QueryBackupRecords(t *testing.T) {
	record := &BackupRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:    "clusterId",
		FilePath:     "/tmp/test",
		StorageType:  "s3",
		BackupType:   "full",
		BackupMethod: "logic",
		BackupMode:   "auto",
		Size:         23346546,
		BackupTso:    42353454343234,
		StartTime:    time.Now(),
	}
	recordCreate, errCreate := rw.CreateBackupRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	recordQuery, total, errQuery := rw.QueryBackupRecords(context.TODO(), "", recordCreate.ID, "", 0, 0, 1, 10)
	assert.NoError(t, errQuery)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, recordCreate.ID, recordQuery[0].ID)
	assert.Equal(t, recordCreate.Status, recordQuery[0].Status)
	assert.Equal(t, recordCreate.BackupMode, recordQuery[0].BackupMode)
}

func TestBRReadWrite_DeleteBackupRecord(t *testing.T) {
	record := &BackupRecord{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:    "clusterId",
		FilePath:     "/tmp/test",
		StorageType:  "s3",
		BackupType:   "full",
		BackupMethod: "logic",
		BackupMode:   "auto",
		Size:         23346546,
		BackupTso:    42353454343234,
		StartTime:    time.Now(),
	}
	recordCreate, errCreate := rw.CreateBackupRecord(context.TODO(), record)
	assert.NoError(t, errCreate)

	errDelete := rw.DeleteBackupRecord(context.TODO(), recordCreate.ID)
	assert.NoError(t, errDelete)

	recordGet, errGet := rw.GetBackupRecord(context.TODO(), recordCreate.ID)
	assert.NotNil(t, errGet)
	assert.Nil(t, recordGet)
}

func TestBRReadWrite_CreateBackupStrategy(t *testing.T) {
	strategy := &BackupStrategy{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:  "clusterId",
		BackupDate: "Monday,Friday",
		StartHour:  11,
		EndHour:    12,
	}
	_, err := rw.CreateBackupStrategy(context.TODO(), strategy)
	assert.NoError(t, err)
}

func TestBRReadWrite_GetBackupStrategy(t *testing.T) {
	strategy := &BackupStrategy{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:  "clusterIdGet",
		BackupDate: "Monday,Friday",
		StartHour:  11,
		EndHour:    12,
	}
	strategyCreate, errCreate := rw.CreateBackupStrategy(context.TODO(), strategy)
	assert.NoError(t, errCreate)

	strategyCreate.BackupDate = "Friday"
	errUpdate := rw.UpdateBackupStrategy(context.TODO(), strategyCreate)
	assert.NoError(t, errUpdate)

	strategyGet, errGet := rw.GetBackupStrategy(context.TODO(), strategyCreate.ClusterID)
	assert.NoError(t, errGet)
	assert.Equal(t, strategyCreate.BackupDate, strategyGet.BackupDate)
}

func TestBRReadWrite_SaveBackupStrategy(t *testing.T) {
	strategy := &BackupStrategy{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:  "clusterIdGet",
		BackupDate: "Monday,Friday",
		StartHour:  11,
		EndHour:    12,
	}
	strategySave, errCreate := rw.SaveBackupStrategy(context.TODO(), strategy)
	assert.NoError(t, errCreate)

	strategyGet, errGet := rw.GetBackupStrategy(context.TODO(), strategySave.ClusterID)
	assert.NoError(t, errGet)
	assert.Equal(t, strategySave.BackupDate, strategyGet.BackupDate)
}

func TestBRReadWrite_QueryBackupStrategy(t *testing.T) {
	strategy := &BackupStrategy{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:  "clusterId",
		BackupDate: "Monday,Friday",
		StartHour:  11,
		EndHour:    12,
	}
	strategyCreate, errCreate := rw.CreateBackupStrategy(context.TODO(), strategy)
	assert.NoError(t, errCreate)

	strategyQuery, errQuery := rw.QueryBackupStrategy(context.TODO(), "Friday", 11)
	assert.NoError(t, errQuery)
	assert.Equal(t, strategyCreate.BackupDate, strategyQuery[0].BackupDate)
}

func TestBRReadWrite_DeleteBackupStrategy(t *testing.T) {
	strategy := &BackupStrategy{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "BackupInitStatus",
		},
		ClusterID:  "clusterIdDelete",
		BackupDate: "Monday,Friday",
		StartHour:  11,
		EndHour:    12,
	}
	strategyCreate, errCreate := rw.CreateBackupStrategy(context.TODO(), strategy)
	assert.NoError(t, errCreate)

	errDelete := rw.DeleteBackupStrategy(context.TODO(), strategyCreate.ClusterID)
	assert.NoError(t, errDelete)

	_, errGet := rw.GetBackupStrategy(context.TODO(), strategyCreate.ClusterID)
	assert.NotNil(t, errGet)
}
