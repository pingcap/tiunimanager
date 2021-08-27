package host

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"testing"
)

func TestMain(m *testing.M) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	m.Run()
}