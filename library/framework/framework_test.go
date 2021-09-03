package framework

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

func TestGetRootLogger(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		got := GetRootLogger()
		assert.Equal(t, "service", got.LogFileName, )
	})
	t.Run("service", func(t *testing.T) {
		Current = nil
		InitBaseFrameworkForUt(MetaDBService)
		got := GetRootLogger()
		assert.Equal(t, MetaDBService.ServerName(), got.LogFileName)
	})
}

func TestLog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		entry := Log()
		_, ok := entry.Data[RecordFunField]
		assert.True(t, ok, "RecordFunField not found")
	})
	t.Run("service", func(t *testing.T) {
		Current = nil
		InitBaseFrameworkForUt(MetaDBService)
		entry := Log()
		_, ok := entry.Data[RecordFunField]
		assert.True(t, ok, "RecordFunField not found")

	})
}

func TestBaseFramework_GetLoggerWithContext(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_NAME, "111")

	got := GetLoggerWithContext(ctx)
	assert.Equal(t, "111", got.Data[TiEM_X_TRACE_ID_NAME])
}

func TestBaseFramework_Get(t *testing.T) {
	f := InitBaseFrameworkForUt(MetaDBService)
	assert.NotNil(t, f.GetClientArgs())
	assert.NotNil(t, f.GetConfiguration())
	assert.NotNil(t, f.GetTracer())
	assert.NotNil(t, f.GetDataDir())
	assert.NotNil(t, f.GetDeployDir())

	assert.NotNil(t, f.GetServiceMeta())
	assert.NoError(t, f.Shutdown())
	assert.NoError(t, f.StopService())

	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_NAME, "111")
	assert.Equal(t, "111", f.GetLoggerWithContext(ctx).Data[TiEM_X_TRACE_ID_NAME])
}

func TestBaseFramework_parseArgs(t *testing.T) {

}