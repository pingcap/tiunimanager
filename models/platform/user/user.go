package user

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models/common"
)

type DBUserRole struct {
	ClusterType constants.EMProductIDType
	RoleName    string
	RoleType    constants.DBRoleType
	Permission  []string
}

type DBUser struct {
	common.Entity
	ClusterID string     `gorm:"not null;type:varchar(22);default:null"`
	Name      string     `gorm:"default:null;not null;uniqueIndex;comment:'name of the user'"`
	Password  string     `gorm:"not null;size:64;comment:'password of the user'"`
	Role      DBUserRole
}
