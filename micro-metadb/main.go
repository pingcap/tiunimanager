package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/pingcap-inc/tiem/micro-metadb/registry"
	dbService "github.com/pingcap-inc/tiem/micro-metadb/service"
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
	log := framework.LogWithCaller()
	log.Info("etcd client connect success")

	f.PrepareService(func(service micro.Service) error {
		return dbPb.RegisterTiEMDBServiceHandler(service.Server(), dbService.NewDBServiceHandler(f.GetDataDir(), f))
	})

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
