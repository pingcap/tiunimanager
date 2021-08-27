package models

import (
	"errors"
	"github.com/mozillazg/go-pinyin"
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/gorm"
	"time"
)

const (
	TALBE_NAME_CLUSTER = "cluster"
	TABLE_NAME_DEMAND_RECORD  = "demand_record"
	TABLE_NAME_ACCOUNT = "account"
	TABLE_NAME_TENANT  = "tenant"
	TABLE_NAME_ROLE    = "role"
	TABLE_NAME_ROLE_BINDING = "role_binding"
	TABLE_NAME_PERMISSION = "permission"
	TABLE_NAME_PERMISSION_BINDING = "permission_binding"
	TABLE_NAME_TOKEN   = "token"
	TABLE_NAME_TASK    = "task"
	TABLE_NAME_HOST	   = "host"
	TABLE_NAME_DISK    = "disk"
	TABLE_NAME_TIUP_CONFIG    = "tiup_config"
	TABLE_NAME_TIUP_TASK = "tiup_task"
	TABLE_NAME_FLOW    = "flow"
	TABLE_NAME_PARAMETERS_RECORD = "parameters_record"
	TABLE_NAME_BACKUP_RECORD = "backup_record"
	TABLE_NAME_RECOVER_RECORD = "recover_record"
)

func getLogger() *framework.LogRecord {
	return framework.GetLogger()
}

type Entity struct {
	ID        	string 				`gorm:"primaryKey;"`
	CreatedAt 	time.Time			`gorm:"<-:create"`
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	Code		string				`gorm:"uniqueIndex;default:null;not null;<-:create"`
	TenantId    string				`gorm:"default:null;not null;<-:create"`
	Status 		int8				`gorm:"type:SMALLINT;default:0"`
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = GenerateID()
	if e.Code == "" {
		e.Code = e.ID
	}
	e.Status = 0

	if len(e.Code) > 128 {
		return errors.New("entity code is too long, code = " + e.Code)
	}

	return nil
}

type Record struct {
	ID        	uint 				`gorm:"primaryKey"`
	CreatedAt 	time.Time			`gorm:"<-:create"`
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	TenantId    string				`gorm:"default:null;not null;<-:create"`
}

type Data struct {
	ID        	uint 				`gorm:"primaryKey"`
	CreatedAt 	time.Time			`gorm:"<-:create"`
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	BizId       string				`gorm:"default:null;<-:create"`
	Status 		int8				`gorm:"type:SMALLINT;default:0"`
}

var split = []byte("_")

func generateEntityCode(name string) string {
	a := pinyin.NewArgs()

	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	code := pinyin.Pinyin(name, a)

	bytes := make([]byte, 0, len(name) * 4)

	// split code with "_" before and after chinese word
	previousSplitFlag := false
	for _, v := range code {
		currentWord := v[0]

		if previousSplitFlag || len(currentWord) > 1 {
			bytes = append(bytes, split...)
		}

		bytes = append(bytes, []byte(currentWord)...)
		previousSplitFlag = len(currentWord) > 1
	}
	return string(bytes)
}

