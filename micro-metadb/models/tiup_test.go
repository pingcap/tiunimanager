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
