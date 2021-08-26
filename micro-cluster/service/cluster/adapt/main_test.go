package adapt

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"testing"
)

func TestMain(m *testing.M) {
	f := framework.InitBaseFrameworkForUt(framework.ClusterService)
	m.Run()
	f.Shutdown()
}