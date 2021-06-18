package models

import (
	"github.com/pingcap/ticp/micro-db/service"
	"gorm.io/gorm"
	"time"
)

type Token struct {
	gorm.Model

	TokenString    	string 		`gorm:"size:255"`
	AccountId      	int32  		`gorm:"size:255"`
	TenantId       	int32  		`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
	ExpirationTime 	time.Time  	`gorm:"size:255"`
}

func AddToken(tokenString string, accountId,tenantId int32, expirationTime time.Time) (token Token, err error) {
	token.TokenString = tokenString
	token.AccountId = accountId
	token.TenantId = tenantId
	token.ExpirationTime = expirationTime
	token.Status = 1

	service.DB.Create(&token)
	return
}

func FindToken(tokenString string) (token Token, err error) {
	service.DB.Where("token_string = ?", tokenString).First(&token)
	return
}