package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/client"
	common "github.com/pingcap-inc/tiem/library/common"
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
	f := framework.InitBaseFrameworkFromArgs(framework.ClusterService,
		initLibForDev,
		initAdapter,
		defaultPortForLocal,
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
	libtiup.MicroInit(f.GetDeployDir() + "/tiupcmd",
		"tiup",
		f.GetDataDir() + common.LogDirPrefix)
	libbr.MicroInit(f.GetDeployDir() + "/brcmd",
		f.GetDataDir() + common.LogDirPrefix)
	return nil
}

func initAdapter(f *framework.BaseFramework) error {
	tenantAdapt.InjectionMetaDbRepo()
	clusterAdapt.InjectionMetaDbRepo()
	return nil
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroClusterPort
	}
	return nil
}