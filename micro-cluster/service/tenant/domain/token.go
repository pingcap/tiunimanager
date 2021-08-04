package domain

import (
	commons2 "github.com/pingcap/ticp/micro-cluster/service/tenant/commons"
	"time"
)

type TiCPToken struct {
	TokenString 	string
	AccountName		string
	AccountId		string
	TenantId		string
	TenantName		string
	ExpirationTime  time.Time
}

func (token *TiCPToken) destroy() error {
	token.ExpirationTime = time.Now()
	return TokenMNG.Modify(token)
}

func (token *TiCPToken) renew() error {
	token.ExpirationTime = time.Now().Add(commons2.DefaultTokenValidPeriod)
	return TokenMNG.Modify(token)
}

func (token *TiCPToken) isValid() bool {
	now := time.Now()

	return now.Before(token.ExpirationTime)
}

func createToken(accountId string, accountName string, tenantId string) (TiCPToken, error) {
	token := TiCPToken{
		AccountName: accountName,
		AccountId: accountId,
		TenantId: tenantId,
		ExpirationTime: time.Now().Add(commons2.DefaultTokenValidPeriod),
	}

	tokenString, err := TokenMNG.Provide(&token)
	token.TokenString = tokenString
	return token, err
}