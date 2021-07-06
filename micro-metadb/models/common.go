package models

import (
	"gorm.io/gorm"
	"time"
)

type Entity struct {
	ID        	string 				`gorm:"PrimaryKey"`
	Code		string				`gorm:"uniqueIndex:code_index"`
	TenantId    string
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"uniqueIndex:code_index"`
}

type Record struct {
	ID        	uint 				`gorm:"PrimaryKey"`
	TenantId    string
	CreatedAt 	time.Time
}