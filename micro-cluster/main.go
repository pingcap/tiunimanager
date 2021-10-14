package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty"
	clusterService "github.com/pingcap-inc/tiem/micro-cluster/service"
	clusterAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/cluster/adapt"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	tenantAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/adapt"
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
	var secondMicro secondparty.MicroSrv
	secondMicro = &secondparty.SecondMicro{
		TiupBinPath: "tiup",
	}
	secondMicro.MicroInit(f.GetDataDir()+common.LogDirPrefix)
	return nil
}

func loadKnowledge(f *framework.BaseFramework) error {
	knowledge.LoadKnowledge()
	return nil
}

func initAdapter(f *framework.BaseFramework) error {
	tenantAdapt.InjectionMetaDbRepo()
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
