package account

import "github.com/pingcap-inc/tiem/models/common"

type Account struct {
	common.Entity

	Name      string `gorm:"default:null;not null"`
	Salt      string `gorm:"default:null;not null;<-:create"`
	FinalHash string `gorm:"default:null;not null"`
}
