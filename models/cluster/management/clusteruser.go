package management

import (
	"gorm.io/gorm"
	"time"
)

type DBUser struct {
	gorm.Model
	ClusterID                string     `gorm:"not null;type:varchar(22);default:null"`
	Name                     string     `gorm:"default:null;not null;uniqueIndex;comment:'name of the user'"`
	Password                 string     `gorm:"not null;size:64;comment:'password of the user'"`
	RoleType                 string
	LastPasswordGenerateTime time.Time
}
