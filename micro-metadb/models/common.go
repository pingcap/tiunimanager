package models

import (
	"github.com/mozillazg/go-pinyin"
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
	e.ID = GenerateID()
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

