/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package metrics

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestMetrics(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		m := GetMetrics()
		assert.NotNil(t, m.APIRequestsCounterMetric)
		assert.NotNil(t, m.RequestDurationHistogramMetric)
		assert.NotNil(t, m.RequestSizeHistogramMetric)
		assert.NotNil(t, m.ResponseSizeHistogramMetric)
		assert.NotNil(t, m.MicroRequestsCounterMetric)
		assert.NotNil(t, m.MicroDurationHistogramMetric)
		assert.NotNil(t, m.SqliteRequestsCounterMetric)
		assert.NotNil(t, m.SqliteDurationHistogramMetric)
		assert.NotNil(t, m.ServerStartTimeGaugeMetric)
	})

	t.Run("handle metrics", func(t *testing.T) {
		got := HandleMetrics(constants.MetricsClusterCreate)
		assert.NotNil(t, got)
		HandleClusterMetrics(time.Now(), "test", 200)
	})
}
