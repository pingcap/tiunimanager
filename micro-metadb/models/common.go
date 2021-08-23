package models

import (
	"github.com/pingcap/tiem/library/firstparty/util/uuidutil"
	"gorm.io/gorm"
	"time"
)

type Entity struct {
	ID        	string 				`gorm:"PrimaryKey"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"uniqueIndex"`

	Code		string				`gorm:"uniqueIndex"`
	TenantId    string				`gorm:"not null;type:varchar(36);default:null"`
	Status 		int8				`gorm:"default:0"`
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	if e.Code == "" {
		e.Code = e.ID
	}
	e.Status = 0
	return nil
}

type Record struct {
	ID        	uint 				`gorm:"primarykey"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	TenantId    string				`gorm:"not null;type:varchar(36);default:null"`
}

type Data struct {
	ID        	uint 				`gorm:"primarykey"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	BizId       string				`gorm:"type:varchar(64);default:null"`
	Status      int8				`gorm:"default:0"`
}

