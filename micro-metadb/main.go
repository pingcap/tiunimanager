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
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/pingcap-inc/tiem/micro-metadb/registry"
	dbService "github.com/pingcap-inc/tiem/micro-metadb/service"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.MetaDBService,
		defaultPortForLocal,
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
	log := f.GetRootLogger().ForkFile(common.LogFileSystem)
	log.Info("etcd client connect success")

	f.PrepareService(func(service micro.Service) error {
		return dbpb.RegisterTiEMDBServiceHandler(service.Server(), dbService.NewDBServiceHandler(f.GetDataDir(), f))
	})

	f.GetMetrics().ServerStartTimeGaugeMetric.
		With(prometheus.Labels{metrics.ServiceLabel: f.GetServiceMeta().ServiceName.ServerName()}).
		SetToCurrentTime()

	err := f.StartService()
	if nil != err {
		log.Errorf("start meta-server failed, error: %v", err)
	}
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroMetaDBPort
	}
	return nil
}
