
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

package models

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"gorm.io/gorm"
	"testing"
)

func TestCreateTiupTask(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		gotId, err := CreateTiupTask(Dao.Db(), context.TODO(), dbpb.TiupTaskType_Deploy, 12321)
		if err != nil {
			t.Errorf("CreateTiupTask() error = %v, wantErr nil", err)
			return
		}
		if gotId <= 0 {
			t.Errorf("CreateTiupTask() gotId = %v, want > 0", gotId)
		}
	})
}

func TestFindTiupTaskByID(t *testing.T) {
	id, err := CreateTiupTask(Dao.Db(), context.TODO(), dbpb.TiupTaskType_Deploy, 12321)
	if err != nil {
		t.Errorf("mock data CreateTiupTask error, err = %v", err)
	}
	t.Run("normal", func(t *testing.T) {
		tiupTask, err := FindTiupTaskByID(Dao.Db(), context.TODO(), id)
		if err != nil {
			t.Errorf("FindTiupTaskByID() error = %v, wantErr nil", err)
			return
		}
		if tiupTask.ID != id {
			t.Errorf("FindTiupTaskByID() gotId = %v, want = %v", tiupTask.ID, id)
		}
	})
}

func TestFindTiupTasksByBizID(t *testing.T) {
	bizId := uint64(12321)
	_, err := CreateTiupTask(Dao.Db(), context.TODO(), dbpb.TiupTaskType_Deploy, bizId)
	if err != nil {
		t.Errorf("mock data CreateTiupTask error, err = %v", err)
	}
	t.Run("normal", func(t *testing.T) {
		tiupTask, err := FindTiupTasksByBizID(Dao.Db(), context.TODO(), bizId)
		if err != nil {
			t.Errorf("FindTiupTaskByID() error = %v, wantErr nil", err)
			return
		}
		if tiupTask[0].BizID != bizId {
			t.Errorf("FindTiupTaskByID() gotId = %v, want = %v", tiupTask[0].BizID, bizId)
		}
	})
}

func TestUpdateTiupTaskStatus(t *testing.T) {
	type args struct {
		db         *gorm.DB
		ctx        context.Context
		id         uint64
		taskStatus dbpb.TiupTaskStatus
		errStr     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateTiupTaskStatus(tt.args.db, tt.args.ctx, tt.args.id, tt.args.taskStatus, tt.args.errStr); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTiupTaskStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
