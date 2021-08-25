package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/framework"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	dbService "github.com/pingcap-inc/tiem/micro-metadb/service"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.MetaDBService,
		initSqliteDB,
		initTables,
		initTenantDataForDev,
		initResourceDataForDev,
	)

	f.PrepareService(func(service micro.Service) error {
		return dbPb.RegisterTiEMDBServiceHandler(service.Server(), new(dbService.DBServiceHandler))
	})

	f.StartService()
}
