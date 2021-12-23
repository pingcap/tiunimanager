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
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockbr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBRService(t *testing.T) {
	service := GetBRService()
	assert.NotNil(t, service)
}

func TestBRManager_DeleteBackupRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	records := make([]*backuprestore.BackupRecord, 1)
	records[0] = &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxx",
		},
		FilePath: "./testdata",
	}
	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(records, int64(1), nil)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make([]*backuprestore.BackupRecord, 0), int64(0), nil)
	brRW.EXPECT().DeleteBackupRecord(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.DeleteBackupRecords(context.TODO(), cluster.DeleteBackupDataReq{})
	assert.Nil(t, err)
}

func TestBRManager_GetBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().GetBackupStrategy(gomock.Any(), gomock.Any()).Return(&backuprestore.BackupStrategy{
		ClusterID:  "cls-xxxx",
		BackupDate: "Monday,Friday",
		StartHour:  0,
		EndHour:    1,
	}, nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	resp, err := service.GetBackupStrategy(context.TODO(), cluster.GetBackupStrategyReq{})
	assert.Nil(t, err)
	assert.Equal(t, "0:00-1:00", resp.Strategy.Period)
}

func TestBRManager_DeleteBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().DeleteBackupStrategy(gomock.Any(), gomock.Any()).Return(nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.DeleteBackupStrategy(context.TODO(), cluster.DeleteBackupStrategyReq{})
	assert.Nil(t, err)
}

func TestBRManager_SaveBackupStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().CreateBackupStrategy(gomock.Any(), gomock.Any()).Return(nil, nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	_, err := service.SaveBackupStrategy(context.TODO(), cluster.SaveBackupStrategyReq{
		ClusterID: "cls-xxxx",
		Strategy: structs.BackupStrategy{
			ClusterID:  "cls-xxxx",
			BackupDate: "Monday",
			Period:     "0:00-1:00",
		},
	})
	assert.Nil(t, err)
}

func TestBRManager_QueryClusterBackupRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	records := make([]*backuprestore.BackupRecord, 1)
	records[0] = &backuprestore.BackupRecord{
		Entity: common.Entity{
			ID: "record-xxx",
		},
		FilePath: "./testdata",
	}
	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().QueryBackupRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(records, int64(1), nil)
	models.SetBRReaderWriter(brRW)

	service := GetBRService()
	resp, _, err := service.QueryClusterBackupRecords(context.TODO(), cluster.QueryBackupRecordsReq{})
	assert.Nil(t, err)
	assert.Equal(t, records[0].FilePath, resp.BackupRecords[0].FilePath)
}
