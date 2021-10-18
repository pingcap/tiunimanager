package application

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/adapt"
	"testing"
)

var authManager *AuthManager
var tenantManager *TenantManager
var userManager *UserManager

func TestMain(m *testing.M) {
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			mock := adapt.NewMockRepo()
			userManager = NewUserManager(mock)
			tenantManager = NewTenantManager(mock)
			authManager = NewAuthManager(userManager, mock)

			return nil
		},
	)
	m.Run()
}
