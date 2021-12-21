package identification

import (
	"gorm.io/gorm"
	"time"
)

type Token struct {
	gorm.Model

	TokenString    string    `gorm:"size:255"`
	AccountId      string    `gorm:"size:255"`
	AccountName    string    `gorm:"size:255"`
	TenantId       string    `gorm:"size:255"`
	Status         int8      `gorm:"size:255"`
	ExpirationTime time.Time `gorm:"size:255"`
}
