package models

import (
	"gorm.io/gorm"
	"time"
)

type Token struct {
	gorm.Model

	TokenString    	string 		`gorm:"size:255"`
	AccountId      	string  	`gorm:"size:255"`
	AccountName		string  	`gorm:"size:255"`
	TenantId       	string  	`gorm:"size:255"`
	Status 			int8		`gorm:"size:255"`
	ExpirationTime 	time.Time  	`gorm:"size:255"`
}

func AddToken(tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (token Token, err error) {
	token, err = FindToken(tokenString)

	if err == nil && token.TokenString == tokenString{
		token.ExpirationTime = expirationTime
		MetaDB.Save(&token)
		return
	} else {
		token.TokenString = tokenString
		token.AccountId = accountId
		token.TenantId = tenantId
		token.ExpirationTime = expirationTime
		token.AccountName = accountName
		token.Status = 0

		MetaDB.Create(&token)
		return
	}
}

func FindToken(tokenString string) (token Token, err error) {
	MetaDB.Where("token_string = ?", tokenString).First(&token)
	return
}