package models

import (
	"gorm.io/gorm"
)

type Cluster struct {
	gorm.Model
	TenantId 		uint
	Name 			string
	Status  		int
	Version 		int

}

type TiUPConfig struct {
	gorm.Model
	TenantId 		uint	`gorm:"size:32"`
	ClusterId		uint	`gorm:"size:32"`
	Latest			bool	`gorm:"size:32"`
	Config			string	`gorm:"type:text"`
}
