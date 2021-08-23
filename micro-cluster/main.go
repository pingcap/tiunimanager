package main

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/pingcap/tiem/library/firstparty/framework"
	"github.com/pingcap/tiem/library/secondparty/libbr"
	"github.com/pingcap/tiem/library/secondparty/libtiup"
	clusterAdapt "github.com/pingcap/tiem/micro-cluster/service/cluster/adapt"
	tenantAdapt "github.com/pingcap/tiem/micro-cluster/service/tenant/adapt"
	dbclient "github.com/pingcap/tiem/micro-metadb/client"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

func main() {
	f := framework.NewDefaultFramework(framework.ClusterService,
		initLib,
		initPort,
		initClient,
	)

	f.StartService()
}

func initClient(d *framework.DefaultServiceFramework) error {
	srv := framework.MetaDBService.BuildMicroService(d.GetRegistryAddress()...)
	dbclient.DBClient = db.NewTiEMDBService(string(framework.MetaDBService), srv.Client())
	return nil
}

func initLib(d *framework.DefaultServiceFramework) error {
	libtiup.MicroInit("./tiupmgr/tiupmgr", "tiup", "")
	libbr.MicroInit("./brmgr/brmgr", "br", "")
	return nil
}

func initPort(d *framework.DefaultServiceFramework) error {
	tenantAdapt.InjectionMetaDbRepo()
	clusterAdapt.InjectionMetaDbRepo()
	return nil
}
