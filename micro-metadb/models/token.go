package models

import (
	"github.com/pingcap/errors"
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

func (t *Token) AddToken(db *gorm.DB,tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (token *Token, err error) {
	if nil == db || "" == tokenString || "" == accountName || "" == accountId || "" == tenantId {
		return nil, errors.Errorf("AddToken has invalid parameter, tokenString: %s, accountName: %s, tenantId: %s, accountId: %s", tokenString, accountName, tenantId, accountId)
	}
	token, err = t.FindToken(db,tokenString)
	if err == nil && token.TokenString == tokenString{
		token.ExpirationTime = expirationTime
		db.Save(&token)
	} else {
		token.TokenString = tokenString
		token.AccountId = accountId
		token.TenantId = tenantId
		token.ExpirationTime = expirationTime
		token.AccountName = accountName
		token.Status = 0
		db.Create(&token)
	}
	return token, err
}

func (t *Token) FindToken(db *gorm.DB, tokenString string) (token *Token, err error) {
	token = &Token{}
	return token, db.Where("token_string = ?", tokenString).First(token).Error
}