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
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/metrics"
	"github.com/pingcap/tiunimanager/micro-cluster/platform/system"
	"github.com/pingcap/tiunimanager/micro-cluster/registry"
	clusterService "github.com/pingcap/tiunimanager/micro-cluster/service"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/proto/clusterservices"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.ClusterService,
		initLibForDev,
		openDatabase,
		initEmbedEtcd,
		notifySystemEvent,
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

func initEmbedEtcd(b *framework.BaseFramework) error {
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
}