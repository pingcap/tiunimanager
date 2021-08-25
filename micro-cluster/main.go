package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/firstparty/client"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	clusterPb "github.com/pingcap-inc/tiem/micro-cluster/proto"
	clusterService "github.com/pingcap-inc/tiem/micro-cluster/service"
	clusterAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/cluster/adapt"
	tenantAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/adapt"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.MetaDBService,
		initLibForDev,
		initPort,
	)

	f.PrepareService(func(service micro.Service) error {
		return clusterPb.RegisterClusterServiceHandler(service.Server(), new(clusterService.ClusterServiceHandler))
	})

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.MetaDBService: func(service micro.Service) error {
			client.DBClient = dbPb.NewTiEMDBService(string(framework.MetaDBService), service.Client())
			return nil
		},
	})

	f.StartService()
}

func initLibForDev(f *framework.BaseFramework) error {
	libtiup.MicroInit(f.GetDeployDir() + "/tiupcmd", "tiup", "")
	libbr.MicroInit(f.GetDeployDir() + "/brcmd", "")
	return nil
}

func initPort(f *framework.BaseFramework) error {
	tenantAdapt.InjectionMetaDbRepo()
	clusterAdapt.InjectionMetaDbRepo()
	return nil
}
