package identification

import (
	"context"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

type TokenReadWrite struct {
	dbCommon.GormDB
}

func (g *TokenReadWrite) AddToken(ctx context.Context, tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (*Token, error) {
	if "" == tokenString /*|| "" == accountName || "" == accountId || "" == tenantId */ {
		return nil, errors.Errorf("AddToken has invalid parameter, tokenString: %s, accountName: %s, tenantId: %s, accountId: %s", tokenString, accountName, tenantId, accountId)
	}
	token, err := g.FindToken(ctx, tokenString)
	if err == nil && token.TokenString == tokenString {
		token.ExpirationTime = expirationTime
		err = g.DB(ctx).Save(&token).Error
	} else {
		token.TokenString = tokenString
		token.AccountId = accountId
		token.TenantId = tenantId
		token.ExpirationTime = expirationTime
		token.AccountName = accountName
		token.Status = 0
		err = g.DB(ctx).Create(&token).Error
	}
	return token, err
}

func (g *TokenReadWrite) FindToken(ctx context.Context, tokenString string) (*Token, error) {
	token := &Token{}
	return token, g.DB(ctx).Where("token_string = ?", tokenString).First(token).Error
}

func NewTokenReadWrite(db *gorm.DB) *TokenReadWrite{
	return &TokenReadWrite{
		dbCommon.WrapDB(db),
	}
}
