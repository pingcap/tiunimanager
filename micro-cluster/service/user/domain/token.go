package domain

import (
	"time"
)

type TiEMToken struct {
	TokenString 	string
	AccountName		string
	AccountId		string
	TenantId		string
	TenantName		string
	ExpirationTime  time.Time
}

func (token *TiEMToken) IsValid() bool {
	now := time.Now()
	return now.Before(token.ExpirationTime)
}

func (token *TiEMToken) Destroy() {
	token.ExpirationTime = time.Now()
}

type UnauthorizedError struct{}

func (*UnauthorizedError) Error() string {
	return "Unauthorized"
}

type ForbiddenError struct{}

func (*ForbiddenError) Error() string {
	return "Access Forbidden"
}
