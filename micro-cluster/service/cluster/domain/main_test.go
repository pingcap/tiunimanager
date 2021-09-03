package domain

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"testing"
)

func TestMain(m *testing.M) {
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			setupMockAdapter()
			return nil
		},
		func(d *framework.BaseFramework) error {
			initFlow()
			return nil
		},
	)
	m.Run()
}
