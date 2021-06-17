package tenant

import (
	"time"
)

type TiCPToken struct {
	TokenString 	string
	AccountId		uint
	TenantId		uint
	ExpirationTime  time.Time
}
