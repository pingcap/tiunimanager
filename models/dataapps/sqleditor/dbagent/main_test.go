package dbagent

import (
	"os"
	"testing"

	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
)

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			models.MockDB()
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)
			models.MockDB()
			return models.Open(d)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}
