package host

import (
	"testing"

	"github.com/pingcap-inc/tiem/library/framework"
)

var resourceManager *ResourceManager

func TestMain(m *testing.M) {
	f_ut := framework.InitBaseFrameworkForUt(framework.ClusterService)
	resourceManager = NewResourceManager(f_ut.GetRootLogger())
	m.Run()
}
