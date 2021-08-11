package models

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
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

func TestMain(m *testing.M) {
	testFile := uuid.New().String() + ".db"
	MetaDB, _ = gorm.Open(sqlite.Open(testFile), &gorm.Config{})

	defer func() {
		os.Remove(testFile)
	}()

	err := MetaDB.Migrator().CreateTable(
		&TestEntity{},
		&TestEntity2{},
		&TestRecord{},
		&TestData{},
		&DemandRecordDO{},
		&ClusterDO{},
		&TiUPConfigDO{},
		&FlowDO{},
		&TaskDO{},
		&Token{},
		&BackupRecordDO{},
		&RecoverRecordDO{},
		&ParametersRecordDO{},
	)
	if err == nil {
		m.Run()
	} else {
		log.Error(err)
	}
}

func TestTime(t *testing.T)  {
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
	t.Run("test id", func(t *testing.T) {
		err := MetaDB.Create(&TestEntity{
			Entity{
				TenantId: "111",
				Code: "test_code",
			},
			"start",
		}).Error

		if err != nil {
			t.Errorf("TestEntity_BeforeCreate() error = %v", err)
		}

		err = MetaDB.Create(&TestEntity2{
			Entity{
				TenantId: "111",
				Code: "test_code",

			},
			"start",
		}).Error

		if err != nil {
			t.Errorf("TestEntity_BeforeCreate() error = %v", err)
		}

	})

}

func TestEntity_BeforeCreate(t *testing.T) {
	t.Run("test id", func(t *testing.T) {
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

	t.Run("test empty tenant", func(t *testing.T) {
		do := &TestEntity{
			Entity{},
			"start",
		}

		err := MetaDB.Create(do).Error

		if err == nil {
			t.Errorf("TestEntity_BeforeCreate() want error")
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
				BizId : "111",
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