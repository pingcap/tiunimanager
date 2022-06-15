/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"time"

	"github.com/pingcap/tiunimanager/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/proto/clusterservices"

	"github.com/pingcap/tiunimanager/common/constants"

	"github.com/pingcap/tiunimanager/file-server/service"

	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/pingcap/tiunimanager/docs"
	"github.com/pingcap/tiunimanager/file-server/interceptor"
	"github.com/pingcap/tiunimanager/file-server/route"
	"github.com/pingcap/tiunimanager/library/framework"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.FileService,
		defaultPortForLocal,
		initManager,
	)

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.ClusterService: func(service micro.Service) error {
			client.ClusterClient = clusterservices.NewClusterService(string(framework.ClusterService), service.Client())
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

	//g.Use(promMiddleware(d)) //todo: no need metrics

	route.Route(g)

	port := d.GetServiceMeta().ServicePort

	addr := fmt.Sprintf(":%d", port)
	// file-server service registry
	serviceRegistry(d)

	d.GetMetrics().ServerStartTimeGaugeMetric.
		With(prometheus.Labels{metrics.ServiceLabel: d.GetServiceMeta().ServiceName.ServerName()}).
		SetToCurrentTime()

	if d.GetClientArgs().EnableHttps {
		g.Use(interceptor.TlsHandler(addr))
		if err := g.RunTLS(addr, d.GetCertificateInfo().CertificateCrtFilePath, d.GetCertificateInfo().CertificateKeyFilePath); err != nil {
			d.GetRootLogger().ForkFile(constants.LogFileSystem).Fatal(err)
		}
	} else {
		if err := g.Run(addr); err != nil {
			d.GetRootLogger().ForkFile(constants.LogFileSystem).Fatal(err)
		}
	}

	return nil
}

// serviceRegistry registry file-server service
func serviceRegistry(f *framework.BaseFramework) {
	etcdClient := framework.InitEtcdClientV2(f.GetServiceMeta().RegistryAddress)
	address := f.GetClientArgs().Host + f.GetServiceMeta().GetServiceAddress()
	key := "/micro/registry/" + f.GetServiceMeta().ServiceName.ServerName() + "/" + address
	// Register file-server every TTL-2 seconds, default TTL is 5s
	go func() {
		for {
			err := etcdClient.SetWithTtl(key, "{\"weight\":1, \"max_fails\":2, \"fail_timeout\":10}", 5)
			if err != nil {
				framework.LogForkFile(constants.LogFileSystem).Errorf("regitry file-server failed! error: %v", err)
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = constants.DefaultMicroFilePort
	}
	return nil
}

/*
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
*/
func initManager(f *framework.BaseFramework) error {
	service.InitFileManager()
	service.InitDirManager()
	return nil
}
