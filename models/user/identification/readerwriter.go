package identification

import (
	"context"
	"time"
)

type ReaderWriter interface {
	AddToken(ctx context.Context, tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (*Token, error)
	FindToken(ctx context.Context, tokenString string) (*Token, error)
}
