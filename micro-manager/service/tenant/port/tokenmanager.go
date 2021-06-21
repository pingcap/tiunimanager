package port

import (
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
)

var TokenMNG TokenManager

type TokenManager interface {

	// Provide 提供一个有效的token
	Provide  (tiCPToken *domain.TiCPToken) (string, error)

	// Modify 修改token
	Modify (tiCPToken *domain.TiCPToken) error

	// GetToken 获取一个token
	GetToken(tokenString string) (domain.TiCPToken, error)
}