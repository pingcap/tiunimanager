package port

import (
	"github.com/pingcap/ticp/micro-manager/service/tenant"
	"time"
)

var TokenMNG TokenManager

type TokenManager interface {

	// Provide 提供一个有效的token
	Provide (accountId uint, tenantId uint, expirationTime time.Time) (string, error)

	// TakeBack 收回token
	TakeBack (tokenString string) error

	// Renew 续期token
	Renew (tiCPToken *tenant.TiCPToken, expirationTime time.Time) error

	// Verify 校验token并返回token对象
	Verify (tokenString string) (tenant.TiCPToken, error)
}