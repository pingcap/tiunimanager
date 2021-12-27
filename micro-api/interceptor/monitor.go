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

package interceptor

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		framework.Current.GetMetrics().MicroDurationHistogramMetric.With(prometheus.Labels{
			metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
			metrics.MethodLabel:  c.Request.RequestURI + ":" + c.Request.Method,
			metrics.CodeLabel:    fmt.Sprint(c.Writer.Status())}).
			Observe(float64(duration.Microseconds()))
		framework.Current.GetMetrics().MicroRequestsCounterMetric.With(prometheus.Labels{
			metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
			metrics.MethodLabel:  c.Request.RequestURI + ":" + c.Request.Method,
			metrics.CodeLabel:    fmt.Sprint(c.Writer.Status())}).
			Inc()
	}
}

func HandleMetrics(start time.Time, funcName string, code int) {
	duration := time.Since(start)
	framework.Current.GetMetrics().MicroDurationHistogramMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Observe(duration.Seconds())
	framework.Current.GetMetrics().MicroRequestsCounterMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Inc()
}
