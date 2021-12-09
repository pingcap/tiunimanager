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
	"errors"
	"github.com/pingcap-inc/tiem/library/framework"
	gormLog "gorm.io/gorm/logger"
	"testing"
	"time"
)

type TestEntity struct {
	Entity
	name string
}

type TestEntity2 struct {
	Entity
	name string
}

type TestRecord struct {
	Record
	name string
}

type TestData struct {
	Data
	name string
}

func TestTime(t *testing.T) {
	t.Run("test time", func(t *testing.T) {
		now := time.Now()
		do := &TestEntity{
			Entity{
				TenantId: "111",
			},
			"start",
		}

		MetaDB.Create(do)

		if now.After(do.CreatedAt) {
			t.Errorf("TestEntity_BeforeCreate() CreatedAt = %v", do.CreatedAt)
		}

		do.name = "updated"
		MetaDB.Updates(do)

		if !do.UpdatedAt.After(do.CreatedAt) {
			t.Errorf("TestEntity_BeforeCreate() CreatedAt = %v, UpdatedAt = %v", do.CreatedAt, do.UpdatedAt)
		}

		do.name = "deleted"
		MetaDB.Delete(do)
		if !do.DeletedAt.Time.After(do.CreatedAt) {
			t.Errorf("TestEntity_BeforeCreate() CreatedAt = %v, DeletedAt = %v", do.CreatedAt, do.DeletedAt)
		}
	})
}

func TestEntityCode(t *testing.T) {
	t.Run("repeated code ", func(t *testing.T) {
		err := MetaDB.Create(&TestEntity{
			Entity{
				TenantId: "111",
				Code:     "repeated_code",
			},
			"start",
		}).Error

		if err != nil {
			t.Errorf("TestEntityCode() error = %v", err)
		}

		err = MetaDB.Create(&TestEntity2{
			Entity{
				TenantId: "111",
				Code:     "repeated_code",
			},
			"start",
		}).Error

		if err != nil {
			t.Errorf("TestEntityCode() error = %v", err)
		}

		err = MetaDB.Create(&TestEntity2{
			Entity{
				TenantId: "111",
				Code:     "repeated_code",
			},
			"start",
		}).Error

		if err == nil {
			t.Errorf("TestEntityCode() want err, got nil")
		}

	})
	t.Run("too long", func(t *testing.T) {
		code := ""
		for i := 0; i < 128; i++ {
			code = code + "1"
		}
		err := MetaDB.Create(&TestEntity{
			Entity{
				TenantId: "111",
				Code:     code,
			},
			"start",
		}).Error

		if err != nil {
			t.Errorf("TestEntityCode() error = %v", err)
		}

		code = code + "1"
		te := &TestEntity{
			Entity{
				TenantId: "111",
				Code:     code,
			},
			"start",
		}
		err = MetaDB.Create(te).Error

		if err == nil {
			t.Errorf("TestEntityCode() want err, got nil")
		}
		te2 := &TestEntity{}
		err = MetaDB.Where("id = ?", te.ID).Find(te2).Error
		if err != nil {
			t.Errorf("TestEntityCode() want err, got nil")
		}
		if len(te2.Code) > 128 {
			t.Errorf("TestEntityCode() code too long, got = %v", te2.Code)
		}

	})
	t.Run("modify", func(t *testing.T) {
		te := &TestEntity{
			Entity{
				TenantId: "111",
				Code:     "test_code_modify",
			},
			"start",
		}
		MetaDB.Create(te)
		te.name = "modified"
		te.Code = "modified"
		MetaDB.Updates(te)
		MetaDB.Where("id = ?", te.ID).Find(te)
		if te.name != "modified" {
			t.Errorf("TestEntityCode() name want %v, got = %v", "modified", te.name)
		}
		if te.Code != "test_code_modify" {
			t.Errorf("TestEntityCode() code want %v, got = %v", "test_code_modify", te.Code)
		}
	})
}

func TestEntityTenantId(t *testing.T) {
	t.Run("test empty tenant", func(t *testing.T) {
		do := &TestEntity{
			Entity{},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err == nil {
			t.Errorf("TestEntityTenantId() want error")
		}
	})

	t.Run("modify", func(t *testing.T) {
		te := &TestEntity{
			Entity{
				TenantId: "111",
				Code:     "dfsafdaf",
			},
			"start",
		}
		MetaDB.Create(te)
		te.name = "modified"
		//TODO te.TenantId = "modified" ? tset update name ?
		MetaDB.Updates(te)
		MetaDB.Where("id = ?", te.ID).Find(te)
		if te.name != "modified" {
			t.Errorf("TestEntityTenantId() name want %v, got = %v", "modified", te.name)
		}
		if te.TenantId != "111" {
			t.Errorf("TestEntityTenantId() TenandId want %v, got = %v", "111", te.Code)
		}
	})
}

func TestEntity_BeforeCreate(t *testing.T) {
	t.Run("test generate id", func(t *testing.T) {
		do := &TestEntity{
			Entity{
				TenantId: "111",
			},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err != nil {
			t.Errorf("TestEntity_BeforeCreate() error = %v", err)
		}

		if do.ID == "" || len(do.ID) < 12 {
			t.Errorf("TestEntity_BeforeCreate() id = %v", do.ID)
		}

		if do.Code != do.ID {
			t.Errorf("TestEntity_BeforeCreate() id = %v, code = %v", do.ID, do.Code)
		}
	})

	t.Run("test generate code", func(t *testing.T) {
		do := &TestEntity{
			Entity{},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err == nil {
			t.Errorf("TestEntity_BeforeCreate() want error")
		}
		if do.Code == "" {
			t.Errorf("TestEntity_BeforeCreate() empty code")
		}
		if do.Code != do.ID {
			t.Errorf("TestEntity_BeforeCreate() entity code want = %v, got = %v", do.ID, do.Code)
		}

	})

	t.Run("test existing code", func(t *testing.T) {
		do := &TestEntity{
			Entity{Code: "existingCode"},
			"start",
		}
		err := MetaDB.Create(do).Error
		if err == nil {
			t.Errorf("TestEntity_BeforeCreate() want error")
		}
		if do.Code != "existingCode" {
			t.Errorf("TestEntity_BeforeCreate() entity code want = %v, got = %v", "existingCode", do.Code)
		}
	})

	t.Run("test status", func(t *testing.T) {
		do := &TestEntity{
			Entity{Status: 4},
			"start",
		}
		err := MetaDB.Create(do).Error
		if err == nil {
			t.Errorf("TestEntity_BeforeCreate() want error")
		}
		if do.Status != 4 {
			t.Errorf("TestEntity_BeforeCreate() entity status want = %v, got = %v", 0, do.Status)
		}
	})

}

func TestRecord_BeforeCreate(t *testing.T) {

	t.Run("test id", func(t *testing.T) {
		do := &TestRecord{
			Record{
				TenantId: "111",
			},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err != nil {
			t.Errorf("TestEntity_BeforeCreate() error = %v", err)
		}

		if do.ID < 1 {
			t.Errorf("TestEntity_BeforeCreate() id = %v", do.ID)
		}
	})

	t.Run("test empty tenant", func(t *testing.T) {
		do := &TestRecord{
			Record{},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err == nil {
			t.Errorf("TestEntity_BeforeCreate() want error")
		}
	})
}

func TestData_BeforeCreate(t *testing.T) {
	t.Run("test id", func(t *testing.T) {
		do := &TestData{
			Data{
				BizId: "111",
			},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err != nil {
			t.Errorf("TestEntity_BeforeCreate() error = %v", err)
		}

		if do.ID < 1 {
			t.Errorf("TestEntity_BeforeCreate() id = %v", do.ID)
		}
	})

}

func Test_generateEntityCode(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"中文"}, "_zhong_wen"},
		{"english", args{"english"}, "english"},
		{"mixed1", args{"english中mix"}, "english_zhong_mix"},
		{"mixed2", args{"中english结合"}, "_zhong_english_jie_he"},
		{"mixedAll", args{"中1文Eng2结4k_5v合''s"}, "_zhong_1_wen_Eng2_jie_4k_5v_he_''s"},
		{"empty", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateEntityCode(tt.args.name); got != tt.want {
				t.Errorf("generateEntityCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaoLogger_Log(t *testing.T) {
	DaoLogger := &DaoLogger{
		p:             framework.Current,
		SlowThreshold: 100 * time.Millisecond,
	}
	DaoLogger.LogMode(gormLog.Info)
	DaoLogger.Error(context.TODO(), "error")
	DaoLogger.Warn(context.TODO(), "warn")
	DaoLogger.Info(context.TODO(), "info")
}

func TestDaoLogger_Trace(t *testing.T) {
	DaoLogger := &DaoLogger{
		p:             framework.Current,
		SlowThreshold: 100 * time.Millisecond,
	}

	t.Run("error", func(t *testing.T) {
		DaoLogger.Trace(context.TODO(), time.Now(), func() (string, int64) { return "select", 100 }, errors.New("something wrong"))
	})

	t.Run("slow", func(t *testing.T) {
		DaoLogger.Trace(context.TODO(), time.Now().AddDate(0, 0, -1), func() (string, int64) { return "select", 100 }, nil)
	})

	t.Run("info", func(t *testing.T) {
		DaoLogger.Trace(context.TODO(), time.Now(), func() (string, int64) { return "select", 100 }, nil)
	})
}
