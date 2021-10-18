package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitMetrics(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		m := InitMetrics()
		assert.NotNil(t, m.APIRequestsCounterMetric)
		assert.NotNil(t, m.RequestDurationHistogramMetric)
		assert.NotNil(t, m.RequestSizeHistogramMetric)
		assert.NotNil(t, m.ResponseSizeHistogramMetric)
		assert.NotNil(t, m.MicroRequestsCounterMetric)
		assert.NotNil(t, m.MicroDurationHistogramMetric)
		assert.NotNil(t, m.SqliteRequestsCounterMetric)
		assert.NotNil(t, m.SqliteDurationHistogramMetric)
		assert.NotNil(t, m.TiUPRequestsCounterMetric)
		assert.NotNil(t, m.TiUPDurationHistogramMetric)
		assert.NotNil(t, m.ServerStartTimeGaugeMetric)
	})
}
