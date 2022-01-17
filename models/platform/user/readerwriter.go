package user

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
)

type ReaderWriter interface {
	AddUser(ctx context.Context, name, password string) (*DBUser, error)
	DeleteUser(ctx context.Context, name string) error
	FindUserByClusterID(ctx context.Context, clusterID string) ([]DBUser, error)
	FindUserByBDRoleType(ctx context.Context, clusterID string, roleType constants.DBUserRoleType) (*DBUser, error)
}
