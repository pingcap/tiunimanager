package models

import (
	"context"
	"github.com/pingcap/errors"
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

func (m *DAOAccountManager) AddToken(ctx context.Context, tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (token *Token, err error) {
	if "" == tokenString /*|| "" == accountName || "" == accountId || "" == tenantId */ {
		return nil, errors.Errorf("AddToken has invalid parameter, tokenString: %s, accountName: %s, tenantId: %s, accountId: %s", tokenString, accountName, tenantId, accountId)
	}
	token, err = m.FindToken(ctx, tokenString)
	if err == nil && token.TokenString == tokenString {
		token.ExpirationTime = expirationTime
		err = m.Db(ctx).Save(&token).Error
	} else {
		token.TokenString = tokenString
		token.AccountId = accountId
		token.TenantId = tenantId
		token.ExpirationTime = expirationTime
		token.AccountName = accountName
		token.Status = 0
		err = m.Db(ctx).Create(&token).Error
	}
	return token, err
}

func (m *DAOAccountManager) FindToken(ctx context.Context, tokenString string) (token *Token, err error) {
	token = &Token{}
	return token, m.Db(ctx).Where("token_string = ?", tokenString).First(token).Error
}
