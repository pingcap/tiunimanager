package main

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	dbclient "github.com/pingcap-inc/tiem/library/firstparty/client"
	"github.com/pingcap-inc/tiem/library/firstparty/framework"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	clusterAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/cluster/adapt"
	tenantAdapt "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/adapt"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

func main() {
	f := framework.NewDefaultFramework(framework.ClusterService,
		initLibForDev,
		initPort,
		initClient,
	)

	f.StartService()
}

func initClient(d *framework.DefaultServiceFramework) error {
	srv := framework.MetaDBService.BuildMicroService(d.GetRegistryAddress()...)
	dbclient.DBClient = db.NewTiEMDBService(framework.MetaDBService.ToString(), srv.Client())
	return nil
}

func initLibForDev(d *framework.DefaultServiceFramework) error {
	libtiup.MicroInit("./../bin/tiupcmd", "tiup", "")
	libbr.MicroInit("./../bin/brcmd", "")
	return nil
}

func initPort(d *framework.DefaultServiceFramework) error {
	tenantAdapt.InjectionMetaDbRepo()
	clusterAdapt.InjectionMetaDbRepo()
	return nil
}
