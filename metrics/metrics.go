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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// process boot metrics
	BootTimeGaugeMetric *prometheus.GaugeVec

	// api metrics
	APIRequestsCounterMetric       *prometheus.CounterVec
	RequestDurationHistogramMetric *prometheus.HistogramVec
	RequestSizeHistogramMetric     *prometheus.HistogramVec
	ResponseSizeHistogramMetric    *prometheus.HistogramVec

	// cluster service metrics
	MicroRequestsCounterMetric   *prometheus.CounterVec
	MicroDurationHistogramMetric *prometheus.HistogramVec

	// db metrics
	SqliteRequestsCounterMetric   *prometheus.CounterVec
	SqliteDurationHistogramMetric *prometheus.HistogramVec

	// micro server start time metrics
	ServerStartTimeGaugeMetric *prometheus.GaugeVec

	// work flow metrics
	WorkFlowCounterMetric     *prometheus.CounterVec
	WorkFlowNodeCounterMetric *prometheus.CounterVec
}

func RegisterNewGaugeVec(metricDef MetricDef) *prometheus.GaugeVec {
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: constants.EM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}

func RegisterNewCounterVec(metricDef MetricDef) *prometheus.CounterVec {
	metric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: constants.EM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}

func RegisterNewHistogramVec(metricDef MetricDef) *prometheus.HistogramVec {
	metric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: constants.EM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}

func RegisterNewSummaryVec(metricDef MetricDef) *prometheus.SummaryVec {
	metric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: constants.EM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}

func GetMetrics() *Metrics {
	once.Do(func() {
		if metrics == nil {
			metrics = &Metrics{
				BootTimeGaugeMetric:            RegisterNewGaugeVec(BootTimeGaugeMetricDef),
				APIRequestsCounterMetric:       RegisterNewCounterVec(APIRequestsCounterMetricDef),
				RequestDurationHistogramMetric: RegisterNewHistogramVec(RequestDurationHistogramMetricDef),
				RequestSizeHistogramMetric:     RegisterNewHistogramVec(RequestSizeHistogramMetricDef),
				ResponseSizeHistogramMetric:    RegisterNewHistogramVec(ResponseSizeHistogramMetricDef),
				MicroRequestsCounterMetric:     RegisterNewCounterVec(MicroRequestsCounterMetricDef),
				MicroDurationHistogramMetric:   RegisterNewHistogramVec(MicroDurationHistogramMetricDef),
				SqliteRequestsCounterMetric:    RegisterNewCounterVec(SqliteRequestsCounterMetricDef),
				SqliteDurationHistogramMetric:  RegisterNewHistogramVec(SqliteDurationHistogramMetricDef),
				ServerStartTimeGaugeMetric:     RegisterNewGaugeVec(ServerStartTimeGaugeMetricDef),
				WorkFlowCounterMetric:          RegisterNewCounterVec(WorkFlowCounterMetricDef),
				WorkFlowNodeCounterMetric:      RegisterNewCounterVec(WorkFlowNodeCounterMetricDef),
			}
		}
	})
	return metrics
}

func HandleMetrics(metricsType constants.MetricsType) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		code := fmt.Sprintf("%d", c.Writer.Status())
		serviceName := OpenApiServer

		GetMetrics().APIRequestsCounterMetric.
			With(prometheus.Labels{ServiceLabel: serviceName, HandlerLabel: string(metricsType), MethodLabel: c.Request.Method, CodeLabel: code}).
			Inc()
		GetMetrics().RequestDurationHistogramMetric.
			With(prometheus.Labels{ServiceLabel: serviceName, HandlerLabel: string(metricsType), MethodLabel: c.Request.Method, CodeLabel: code}).
			Observe(duration.Seconds())
		GetMetrics().RequestSizeHistogramMetric.
			With(prometheus.Labels{ServiceLabel: serviceName, HandlerLabel: string(metricsType), MethodLabel: c.Request.Method, CodeLabel: code}).
			Observe(float64(computeApproximateRequestSize(c.Request)))
		GetMetrics().ResponseSizeHistogramMetric.
			With(prometheus.Labels{ServiceLabel: serviceName, HandlerLabel: string(metricsType), MethodLabel: c.Request.Method, CodeLabel: code}).
			Observe(float64(c.Writer.Size()))
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

func HandleClusterMetrics(start time.Time, funcName string, code int) {
	duration := time.Since(start)
	GetMetrics().MicroDurationHistogramMetric.With(prometheus.Labels{
		ServiceLabel: ClusterServer,
		MethodLabel:  funcName,
		CodeLabel:    strconv.Itoa(code)}).
		Observe(duration.Seconds())
	GetMetrics().MicroRequestsCounterMetric.With(prometheus.Labels{
		ServiceLabel: ClusterServer,
		MethodLabel:  funcName,
		CodeLabel:    strconv.Itoa(code)}).
		Inc()
}

type WorkFlowLabel struct {
	BizType string
	Name    string
	Status  string
}

func HandleWorkFlowMetrics(label WorkFlowLabel) {
	GetMetrics().WorkFlowCounterMetric.With(prometheus.Labels{
		ServiceLabel:    ClusterServer,
		BizTypeLabel:    label.BizType,
		FlowNameLabel:   label.Name,
		FlowStatusLabel: label.Status,
	}).
		Inc()
}

type WorkFlowNodeLabel struct {
	BizType  string
	FlowName string
	Node     string
	Status   string
}

func HandleWorkFlowNodeMetrics(label WorkFlowNodeLabel) {
	GetMetrics().WorkFlowNodeCounterMetric.With(prometheus.Labels{
		ServiceLabel:        ClusterServer,
		BizTypeLabel:        label.BizType,
		FlowNameLabel:       label.FlowName,
		FlowNodeLabel:       label.Node,
		FlowNodeStatusLabel: label.Status,
	}).
		Inc()
}
