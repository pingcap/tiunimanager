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
	"context"
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/common/client"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/metrics"
	"github.com/pingcap-inc/tiem/micro-cluster/platform/system"
	"github.com/pingcap-inc/tiem/micro-cluster/registry"
	clusterService "github.com/pingcap-inc/tiem/micro-cluster/service"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/proto/clusterservices"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.ClusterService,
		initLibForDev,
		openDatabase,
		notifySystemEvent,
		func(b *framework.BaseFramework) error {
			go func() {
				// init embed etcd.
				err := registry.InitEmbedEtcd(b)
				if err != nil {
					b.GetRootLogger().ForkFile(b.GetServiceMeta().ServiceName.ServerName()).
						Errorf("init embed etcd failed, error: %v", err)
					return
				}
			}()
			return nil
		},
	)

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.ClusterService: func(service micro.Service) error {
			client.ClusterClient = clusterservices.NewClusterService(string(framework.ClusterService), service.Client())
			return nil
		},
	})

	f.PrepareService(func(service micro.Service) error {
		return clusterservices.RegisterClusterServiceHandler(service.Server(), clusterService.NewClusterServiceHandler(f))
	})

	f.GetMetrics().ServerStartTimeGaugeMetric.
		With(prometheus.Labels{metrics.ServiceLabel: f.GetServiceMeta().ServiceName.ServerName()}).
		SetToCurrentTime()

	f.StartService()
}

func initLibForDev(f *framework.BaseFramework) error {
	secondparty.Manager = &secondparty.SecondPartyManager{
		TiUPBinPath: constants.TiUPBinPath,
	}
	secondparty.Manager.Init()
	deployment.M = &deployment.Manager{
		TiUPBinPath: constants.TiUPBinPath,
	}
	return nil
}

func openDatabase(f *framework.BaseFramework) error {
	return models.Open(f)
}

func notifySystemEvent(f *framework.BaseFramework) error {
	return system.GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessStarted)
}