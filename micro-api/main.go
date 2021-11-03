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

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap-inc/tiem/library/thirdparty/etcd_clientv2"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"

	"github.com/pingcap-inc/tiem/library/knowledge"

	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/pingcap-inc/tiem/docs"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	"github.com/pingcap-inc/tiem/micro-api/route"
)

// @title TiEM UI API
// @version 1.0
// @description TiEM UI API

// @contact.name zhangpeijin
// @contact.email zhangpeijin@pingcap.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:4116
// @BasePath /api/v1/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.ApiService,
		loadKnowledge,
		defaultPortForLocal,
	)

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.ClusterService: func(service micro.Service) error {
			client.ClusterClient = clusterpb.NewClusterService(string(framework.ClusterService), service.Client())
			return nil
		},
	})

	f.PrepareService(func(service micro.Service) error {
		return initGinEngine(f)
	})

	//f.StartService()
}

func initGinEngine(d *framework.BaseFramework) error {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	// enable cors access
	g.Use(cors.New(corsConfig()))

	g.Use(promMiddleware(d))

	route.Route(g)

	port := d.GetServiceMeta().ServicePort

	addr := fmt.Sprintf(":%d", port)
	// openapi-server service registry
	serviceRegistry(d)

	d.GetMetrics().ServerStartTimeGaugeMetric.
		With(prometheus.Labels{metrics.ServiceLabel: d.GetServiceMeta().ServiceName.ServerName()}).
		SetToCurrentTime()

	if d.GetClientArgs().EnableHttps {
		g.Use(interceptor.TlsHandler(addr))
		if err := g.RunTLS(addr, d.GetCertificateInfo().CertificateCrtFilePath, d.GetCertificateInfo().CertificateKeyFilePath); err != nil {
			d.GetRootLogger().ForkFile(common.LogFileSystem).Fatal(err)
		}
	} else {
		if err := g.Run(addr); err != nil {
			d.GetRootLogger().ForkFile(common.LogFileSystem).Fatal(err)
		}
	}

	return nil
}

// corsConfig
// @Description: build cors config
// @return cors.Config
func corsConfig() cors.Config {
	config := cors.DefaultConfig()
	config.AllowHeaders = []string{"*"}
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	return config
}

// serviceRegistry registry openapi-server service
func serviceRegistry(f *framework.BaseFramework) {
	etcdClient := etcd_clientv2.InitEtcdClient(f.GetServiceMeta().RegistryAddress)
	address := f.GetClientArgs().Host + f.GetServiceMeta().GetServiceAddress()
	key := common.RegistryMicroServicePrefix + f.GetServiceMeta().ServiceName.ServerName() + "/" + address
	// Register openapi-server every TTL-2 seconds, default TTL is 5s
	go func() {
		for {
			err := etcdClient.SetWithTtl(key, "{\"weight\":1, \"max_fails\":2, \"fail_timeout\":10}", 5)
			if err != nil {
				f.Log().Errorf("regitry openapi-server failed! error: %v", err)
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
func loadKnowledge(f *framework.BaseFramework) error {
	knowledge.LoadKnowledge()
	return nil
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroApiPort
	}
	return nil
}

// prometheus http metrics
func promMiddleware(d *framework.BaseFramework) gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := c.Request.URL.Path
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		code := fmt.Sprintf("%d", c.Writer.Status())
		serviceName := d.GetServiceMeta().ServiceName.ServerName()

		d.GetMetrics().APIRequestsCounterMetric.
			With(prometheus.Labels{metrics.ServiceLabel: serviceName, metrics.HandlerLabel: relativePath, metrics.MethodLabel: c.Request.Method, metrics.CodeLabel: code}).
			Inc()
		d.GetMetrics().RequestDurationHistogramMetric.
			With(prometheus.Labels{metrics.ServiceLabel: serviceName, metrics.HandlerLabel: relativePath, metrics.MethodLabel: c.Request.Method, metrics.CodeLabel: code}).
			Observe(duration.Seconds())
		d.GetMetrics().RequestSizeHistogramMetric.
			With(prometheus.Labels{metrics.ServiceLabel: serviceName, metrics.HandlerLabel: relativePath, metrics.MethodLabel: c.Request.Method, metrics.CodeLabel: code}).
			Observe(float64(computeApproximateRequestSize(c.Request)))
		d.GetMetrics().ResponseSizeHistogramMetric.
			With(prometheus.Labels{metrics.ServiceLabel: serviceName, metrics.HandlerLabel: relativePath, metrics.MethodLabel: c.Request.Method, metrics.CodeLabel: code}).
			Observe(float64(c.Writer.Size()))
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
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
