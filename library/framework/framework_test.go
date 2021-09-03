package framework

import (
	"testing"
)

func TestInitBaseFrameworkForUt(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		f := InitBaseFrameworkForUt(MetaDBService,
			func(d *BaseFramework) error {
				d.GetServiceMeta().ServicePort = 99999
				return nil
			},
		)
		if f.GetServiceMeta().ServicePort != 99999 {
			t.Errorf("InitBaseFrameworkFromArgs() service port wrong, want = %v, got %v", 99999, f.GetServiceMeta().ServicePort)
		}
	})
}
