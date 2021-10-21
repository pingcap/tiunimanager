
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
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	clusterService "github.com/pingcap-inc/tiem/micro-cluster/service"
	clusterAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/cluster/adapt"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.ClusterService,
		loadKnowledge,
		initLibForDev,
		initAdapter,
		initCronJob,
		defaultPortForLocal,
	)

	f.PrepareService(func(service micro.Service) error {
		return clusterpb.RegisterClusterServiceHandler(service.Server(), clusterService.NewClusterServiceHandler(f))
	})

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.MetaDBService: func(service micro.Service) error {
			client.DBClient = dbpb.NewTiEMDBService(string(framework.MetaDBService), service.Client())
			return nil
		},
	})

	f.StartService()
}

func initLibForDev(f *framework.BaseFramework) error {
	libtiup.MicroInit(f.GetDeployDir()+"/tiupcmd",
		"tiup",
		f.GetDataDir()+common.LogDirPrefix)
	libbr.MicroInit(f.GetDeployDir()+"/brcmd",
		f.GetDataDir()+common.LogDirPrefix)
	return nil
}

func loadKnowledge(f *framework.BaseFramework) error {
	knowledge.LoadKnowledge()
	return nil
}

func initAdapter(f *framework.BaseFramework) error {
	clusterAdapt.InjectionMetaDbRepo()
	return nil
}

func initCronJob(f *framework.BaseFramework) error {
	domain.InitAutoBackupCronJob()
	return nil
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroClusterPort
	}
	return nil
}
