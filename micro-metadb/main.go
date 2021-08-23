package main

import (
	"github.com/pingcap/tiem/library/firstparty/framework"
)

func main() {
	f := framework.NewDefaultFramework(framework.MetaDBService,
		initSqliteDB,
		initTables,
		initTenantDataForDev,
		initResourceDataForDev,
	)

	f.StartService()
}
