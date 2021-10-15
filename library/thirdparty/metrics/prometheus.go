package metrics

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricDef struct {
	Name       string
	Help       string
	LabelNames []string
}

func RegisterNewGaugeVec(metricDef MetricDef) *prometheus.GaugeVec {
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: common.TiEM,
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
			Namespace: common.TiEM,
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
			Namespace: common.TiEM,
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
			Namespace: common.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}
