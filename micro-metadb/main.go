package main

import (
	"github.com/pingcap-inc/tiem/library/firstparty/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/service"
)

func main() {
	f := framework.NewDefaultFramework(framework.MetaDBService,
		initLogger,
		initSqliteDB,
		initTables,
		initTenantDataForDev,
		initResourceDataForDev,
	)

	f.StartService()
}

func initLogger(f *framework.DefaultServiceFramework) error {
	service.InitLogger(f.GetDefaultLogger())
	return nil
}
