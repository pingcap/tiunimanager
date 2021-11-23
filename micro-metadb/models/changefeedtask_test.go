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

package models

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChangeFeedTask_Locked(t1 *testing.T) {
	type fields struct {
		Entity            Entity
		Name              string
		ClusterId         string
		DownstreamType    string
		StartTS           int64
		FilterRulesConfig string
		DownstreamConfig  string
		StatusLock        sql.NullTime
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"locked", fields{StatusLock: sql.NullTime{Time: time.Now(), Valid: true}}, true},
		{"invalid", fields{StatusLock: sql.NullTime{Time: time.Now(), Valid: false}}, false},
		{"expired", fields{StatusLock: sql.NullTime{Time: time.Now().Add(time.Minute * -2), Valid: true}}, false},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := ChangeFeedTask{
				Entity:            tt.fields.Entity,
				Name:              tt.fields.Name,
				ClusterId:         tt.fields.ClusterId,
				DownstreamType:    tt.fields.DownstreamType,
				StartTS:           tt.fields.StartTS,
				FilterRulesConfig: tt.fields.FilterRulesConfig,
				DownstreamConfig:  tt.fields.DownstreamConfig,
				StatusLock:        tt.fields.StatusLock,
			}
			if got := t.Locked(); got != tt.want {
				t1.Errorf("Locked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDAOChangeFeedManager_Create(t *testing.T) {
	existed, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}})
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), existed.ID)

	type args struct {
		ctx  context.Context
		task *ChangeFeedTask
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "dafadsfefesdf"}}, false},
		{"without tenant", args{context.TODO(), &ChangeFeedTask{Entity: Entity{ID: existed.ID}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Dao.ChangeFeedTaskManager()
			got, err := m.Create(tt.args.ctx, tt.args.task)
			defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), got.ID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestDAOChangeFeedManager_Delete(t *testing.T) {
	existed, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}})
	deleted, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "111"})
	Dao.Db().Delete(deleted)

	type args struct {
		ctx    context.Context
		taskId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), existed.ID}, false},
		{"deleted", args{context.TODO(), deleted.ID}, true},
		{"not existed", args{context.TODO(), "dfadsafsa"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Dao.ChangeFeedTaskManager()

			if err := m.Delete(tt.args.ctx, tt.args.taskId); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				_, e := m.Get(tt.args.ctx, tt.args.taskId)
				assert.Error(t, e)
			}

		})
	}
}

func TestDAOChangeFeedManager_LockStatus(t *testing.T) {
	locked, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{
		Entity: Entity{TenantId: "111"},
		StatusLock: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})

	unlocked, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{
		Entity: Entity{TenantId: "111"},
	})

	Dao.Db().Table(TABLE_NAME_CHANGE_FEED_TASKS).Where("id", unlocked.ID).Update("status_lock", sql.NullTime{
		Time:  time.Now(),
		Valid: false,
	})
	notExisted := "111"
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), locked.ID)
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), unlocked.ID)

	type args struct {
		ctx    context.Context
		taskId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"locked", args{context.TODO(), locked.ID}, true},
		{"unlocked", args{context.TODO(), unlocked.ID}, false},
		{"not existed", args{context.TODO(), notExisted}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Dao.ChangeFeedTaskManager()
			if err := m.LockStatus(tt.args.ctx, tt.args.taskId); (err != nil) != tt.wantErr {
				t.Errorf("LockStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				updated, e := m.Get(tt.args.ctx, tt.args.taskId)
				assert.NoError(t, e)
				assert.True(t, updated.Locked())
			}

		})
	}
}

func TestDAOChangeFeedManager_QueryByClusterId(t *testing.T) {
	t1, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "6666"})
	anotherCluster, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "3121"})
	deleted, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "6666"})
	t4, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "6666", StartTS: int64(9999)})
	t5, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}, ClusterId: "6666"})
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), t1.ID)
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), anotherCluster.ID)
	Dao.Db().Delete(deleted)
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), t4.ID)
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), t5.ID)

	tasks, total, err := Dao.ChangeFeedTaskManager().QueryByClusterId(context.TODO(), "6666", 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, int(total))
	assert.Equal(t, 9999, int(tasks[1].StartTS))
}

func TestDAOChangeFeedManager_UnlockStatus(t *testing.T) {
	newStatus := int8(2)
	locked, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{
		Entity: Entity{TenantId: "111"},
		StatusLock: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	unlocked, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{
		Entity: Entity{TenantId: "111"},
	})
	Dao.Db().Table(TABLE_NAME_CHANGE_FEED_TASKS).Where("id", unlocked.ID).Update("status_lock", sql.NullTime{
		Time:  time.Now().Add(time.Minute * -3),
		Valid: true,
	})

	notExisted := "111"
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), locked.ID)
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), unlocked.ID)

	type args struct {
		ctx          context.Context
		taskId       string
		targetStatus int8
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		finalStatus int8
	}{
		{"locked", args{context.TODO(), locked.ID, newStatus}, false, newStatus},
		{"unlocked", args{context.TODO(), unlocked.ID, newStatus},true, 0},
		{"notExisted", args{context.TODO(), notExisted, newStatus},true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Dao.ChangeFeedTaskManager()
			if err := m.UnlockStatus(tt.args.ctx, tt.args.taskId, tt.args.targetStatus); (err != nil) != tt.wantErr {
				t.Errorf("UnlockStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.finalStatus != 0 {
				final, _ := m.Get(tt.args.ctx, tt.args.taskId)
				assert.Equal(t, tt.finalStatus, final.Status)
				assert.False(t, final.Locked())
			}
		})
	}
}

func TestDAOChangeFeedManager_UpdateConfig(t *testing.T) {
	existed, _ := Dao.ChangeFeedTaskManager().Create(context.TODO(), &ChangeFeedTask{Entity: Entity{TenantId: "111"}})
	defer Dao.ChangeFeedTaskManager().Delete(context.TODO(), existed.ID)

	newString := "new"
	newInt := 99
	existed.DownstreamConfig = newString
	existed.DownstreamType = newString
	existed.FilterRulesConfig = newString
	existed.ClusterId = newString
	existed.StartTS = int64(newInt)
	existed.Entity.Status = int8(newInt)

	type args struct {
		ctx            context.Context
		updateTemplate ChangeFeedTask
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), *existed}, false},
		{"not existed", args{context.TODO(), ChangeFeedTask{
			Entity:Entity{ID: "111"},
		}}, true},
		{"without id", args{context.TODO(), ChangeFeedTask{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Dao.ChangeFeedTaskManager()
			if err := m.UpdateConfig(tt.args.ctx, tt.args.updateTemplate); (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				updated, _ := m.Get(tt.args.ctx, tt.args.updateTemplate.ID)
				assert.Equal(t, newString, updated.DownstreamConfig)
				assert.Equal(t, newString, updated.DownstreamType)
				assert.Equal(t, newString, updated.FilterRulesConfig)
				assert.NotEqual(t, newString, updated.ClusterId)
				assert.Equal(t, int64(newInt), updated.StartTS)
				assert.NotEqual(t, int8(newInt), updated.Status)
			}
		})
	}
}
