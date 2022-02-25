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

	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/metrics"
	"github.com/pingcap-inc/tiem/proto/clusterservices"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/gin-contrib/cors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/pingcap-inc/tiem/docs"
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
		defaultPortForLocal,
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

	// enable cors access
	g.Use(cors.New(corsConfig()))

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
			d.GetRootLogger().ForkFile(constants.LogFileSystem).Fatal(err)
		}
	} else {
		if err := g.Run(addr); err != nil {
			d.GetRootLogger().ForkFile(constants.LogFileSystem).Fatal(err)
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
	etcdClient := framework.InitEtcdClientV2(f.GetServiceMeta().RegistryAddress)
	address := f.GetClientArgs().Host + f.GetServiceMeta().GetServiceAddress()
	key := "/micro/registry/" + f.GetServiceMeta().ServiceName.ServerName() + "/" + address
	// Register openapi-server every TTL-2 seconds, default TTL is 5s
	go func() {
		for {
			err := etcdClient.SetWithTtl(key, "{\"weight\":1, \"max_fails\":2, \"fail_timeout\":10}", 5)
			if err != nil {
				framework.LogForkFile(constants.LogFileSystem).Errorf("regitry openapi-server failed! error: %v", err)
			}
			time.Sleep(time.Second * 3)
		}
	}()
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = constants.DefaultMicroApiPort
	}
	return nil
}
