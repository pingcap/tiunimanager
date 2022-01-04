
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
			Namespace: constants.TiEM,
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
			Namespace: constants.TiEM,
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
			Namespace: constants.TiEM,
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
			Namespace: constants.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	prometheus.MustRegister(metric)
	return metric
}

func RegisterNewGaugeVecForUT(metricDef MetricDef) *prometheus.GaugeVec {
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: constants.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	return metric
}

func RegisterNewCounterVecForUT(metricDef MetricDef) *prometheus.CounterVec {
	metric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: constants.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	return metric
}

func RegisterNewHistogramVecForUT(metricDef MetricDef) *prometheus.HistogramVec {
	metric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: constants.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	return metric
}

func RegisterNewSummaryVecForUT(metricDef MetricDef) *prometheus.SummaryVec {
	metric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: constants.TiEM,
			Name:      metricDef.Name,
			Help:      metricDef.Help,
		},
		metricDef.LabelNames,
	)
	return metric
}