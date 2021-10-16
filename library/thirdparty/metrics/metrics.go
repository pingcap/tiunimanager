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
