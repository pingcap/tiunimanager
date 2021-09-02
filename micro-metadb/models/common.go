package models

import (
	"errors"
	"github.com/mozillazg/go-pinyin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

const (
	TABLE_NAME_CLUSTER            = "clusters"
	TABLE_NAME_DEMAND_RECORD      = "demand_records"
	TABLE_NAME_ACCOUNT            = "accounts"
	TABLE_NAME_TENANT             = "tenants"
	TABLE_NAME_ROLE               = "roles"
	TABLE_NAME_ROLE_BINDING       = "role_bindings"
	TABLE_NAME_PERMISSION         = "permissions"
	TABLE_NAME_PERMISSION_BINDING = "permission_bindings"
	TABLE_NAME_TOKEN              = "tokens"
	TABLE_NAME_TASK               = "tasks"
	TABLE_NAME_HOST               = "hosts"
	TABLE_NAME_DISK               = "disks"
	TABLE_NAME_TIUP_CONFIG        = "tiup_configs"
	TABLE_NAME_TIUP_TASK          = "tiup_tasks"
	TABLE_NAME_FLOW               = "flows"
	TABLE_NAME_PARAMETERS_RECORD  = "parameters_records"
	TABLE_NAME_BACKUP_RECORD      = "backup_records"
	TABLE_NAME_BACKUP_STRATEGY    = "backup_strategys"
	TABLE_NAME_TRANSPORT_RECORD    = "transport_records"
	TABLE_NAME_RECOVER_RECORD     = "recover_records"
)

func getLogger() *log.Entry {
	return framework.Log()
}

type Entity struct {
	ID        string    `gorm:"primaryKey;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Code     string `gorm:"uniqueIndex;default:null;not null;<-:create"`
	TenantId string `gorm:"default:null;not null;<-:create"`
	Status   int8   `gorm:"type:SMALLINT;default:0"`
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
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
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TenantId string `gorm:"default:null;not null;<-:create"`
}

type Data struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	BizId  string `gorm:"default:null;<-:create"`
	Status int8   `gorm:"type:SMALLINT;default:0"`
}

var split = []byte("_")

func generateEntityCode(name string) string {
	a := pinyin.NewArgs()

	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	code := pinyin.Pinyin(name, a)

	bytes := make([]byte, 0, len(name)*4)

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
