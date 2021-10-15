
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

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
