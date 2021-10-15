
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

const (
	ServiceLabel = "service"
	HandlerLabel = "handler"
	MethodLabel  = "method"
	CodeLabel    = "code"
)

// Metrics Define
var (
	BootTimeGaugeMetricDef = MetricDef{
		Name:       "boot_time",
		Help:       "A gauge of process boot time.",
		LabelNames: []string{ServiceLabel},
	}

	APIRequestsCounterMetricDef = MetricDef{
		Name:       "http_requests_total",
		Help:       "A counter for requests to the wrapped handler.",
		LabelNames: []string{ServiceLabel, HandlerLabel, MethodLabel, CodeLabel},
	}
	RequestDurationHistogramMetricDef = MetricDef{
		Name:       "http_request_duration_seconds",
		Help:       "A histogram of latencies for requests.",
		LabelNames: []string{ServiceLabel, HandlerLabel, MethodLabel, CodeLabel},
	}
	RequestSizeHistogramMetricDef = MetricDef{
		Name:       "http_request_size_bytes",
		Help:       "A histogram of request sizes for requests.",
		LabelNames: []string{ServiceLabel, HandlerLabel, MethodLabel, CodeLabel},
	}
	ResponseSizeHistogramMetricDef = MetricDef{
		Name:       "http_response_size_bytes",
		Help:       "A histogram of response sizes for requests.",
		LabelNames: []string{ServiceLabel, HandlerLabel, MethodLabel, CodeLabel},
	}

	MicroRequestsCounterMetricDef = MetricDef{
		Name:       "micro_requests_total",
		Help:       "A counter for requests to the micro service.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}
	MicroDurationHistogramMetricDef = MetricDef{
		Name:       "micro_request_duration_seconds",
		Help:       "A histogram for requests duration to the micro service.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}

	SqliteRequestsCounterMetricDef = MetricDef{
		Name:       "sqlite_requests_total",
		Help:       "A counter for requests to the sqlite.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}
	SqliteDurationHistogramMetricDef = MetricDef{
		Name:       "sqlite_request_duration_seconds",
		Help:       "A histogram for requests duration to the sqlite.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}

	TiUPRequestsCounterMetricDef = MetricDef{
		Name:       "tiup_requests_total",
		Help:       "A counter for requests to the tiup.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}
	TiUPDurationHistogramMetricDef = MetricDef{
		Name:       "tiup_request_duration_seconds",
		Help:       "A histogram for requests duration to the tiup.",
		LabelNames: []string{ServiceLabel, MethodLabel, CodeLabel},
	}

	ServerStartTimeGaugeMetricDef = MetricDef{
		Name:       "server_start_time",
		Help:       "A gauge of micro service start time.",
		LabelNames: []string{ServiceLabel},
	}
)
