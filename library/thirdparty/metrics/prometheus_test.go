package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterNewGaugeVec(t *testing.T) {
	md := MetricDef{
		Name:       "test_gauge",
		Help:       "A gauge test metrics",
		LabelNames: []string{ServiceLabel},
	}
	t.Run("normal", func(t *testing.T) {
		m := RegisterNewGaugeVec(md)
		assert.NotNil(t, m)
	})
}

func TestRegisterNewCounterVec(t *testing.T) {
	md := MetricDef{
		Name:       "test_counter",
		Help:       "A counter test metrics",
		LabelNames: []string{ServiceLabel},
	}
	t.Run("normal", func(t *testing.T) {
		m := RegisterNewCounterVec(md)
		assert.NotNil(t, m)
	})
}

func TestRegisterNewHistogramVec(t *testing.T) {
	md := MetricDef{
		Name:       "test_histogram",
		Help:       "A histogram test metrics",
		LabelNames: []string{ServiceLabel},
	}
	t.Run("normal", func(t *testing.T) {
		m := RegisterNewHistogramVec(md)
		assert.NotNil(t, m)
	})
}

func TestRegisterNewSummaryVec(t *testing.T) {
	md := MetricDef{
		Name:       "test_summary",
		Help:       "A summary test metrics",
		LabelNames: []string{ServiceLabel},
	}
	t.Run("normal", func(t *testing.T) {
		m := RegisterNewSummaryVec(md)
		assert.NotNil(t, m)
	})
}
