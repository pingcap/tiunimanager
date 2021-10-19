package resource

import (
	"testing"

	"github.com/pingcap-inc/tiem/library/framework"
)

var resourceManager *ResourceManager

func TestMain(m *testing.M) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	resourceManager = NewResourceManager()
	m.Run()
}
