
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

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	// process boot metrics
	BootTimeGaugeMetric *prometheus.GaugeVec

	// api metrics
	APIRequestsCounterMetric       *prometheus.CounterVec
	RequestDurationHistogramMetric *prometheus.HistogramVec
	RequestSizeHistogramMetric     *prometheus.HistogramVec
	ResponseSizeHistogramMetric    *prometheus.HistogramVec

	// micro service metrics
	MicroRequestsCounterMetric   *prometheus.CounterVec
	MicroDurationHistogramMetric *prometheus.HistogramVec

	// sqlite metrics
	SqliteRequestsCounterMetric   *prometheus.CounterVec
	SqliteDurationHistogramMetric *prometheus.HistogramVec

	// tiup metrics
	TiUPRequestsCounterMetric   *prometheus.CounterVec
	TiUPDurationHistogramMetric *prometheus.HistogramVec

	// micro server start time metrics
	ServerStartTimeGaugeMetric *prometheus.GaugeVec
}

func InitMetrics() *Metrics {
	m := Metrics{
		BootTimeGaugeMetric:            RegisterNewGaugeVec(BootTimeGaugeMetricDef),
		APIRequestsCounterMetric:       RegisterNewCounterVec(APIRequestsCounterMetricDef),
		RequestDurationHistogramMetric: RegisterNewHistogramVec(RequestDurationHistogramMetricDef),
		RequestSizeHistogramMetric:     RegisterNewHistogramVec(RequestSizeHistogramMetricDef),
		ResponseSizeHistogramMetric:    RegisterNewHistogramVec(ResponseSizeHistogramMetricDef),
		MicroRequestsCounterMetric:     RegisterNewCounterVec(MicroRequestsCounterMetricDef),
		MicroDurationHistogramMetric:   RegisterNewHistogramVec(MicroDurationHistogramMetricDef),
		SqliteRequestsCounterMetric:    RegisterNewCounterVec(SqliteRequestsCounterMetricDef),
		SqliteDurationHistogramMetric:  RegisterNewHistogramVec(SqliteDurationHistogramMetricDef),
		TiUPRequestsCounterMetric:      RegisterNewCounterVec(TiUPRequestsCounterMetricDef),
		TiUPDurationHistogramMetric:    RegisterNewHistogramVec(TiUPDurationHistogramMetricDef),
		ServerStartTimeGaugeMetric:     RegisterNewGaugeVec(ServerStartTimeGaugeMetricDef),
	}
	return &m
}
